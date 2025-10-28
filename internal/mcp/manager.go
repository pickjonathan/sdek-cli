package mcp

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ServerStatus represents the health status of an MCP server
type ServerStatus string

const (
	StatusUnknown  ServerStatus = "unknown"
	StatusHealthy  ServerStatus = "healthy"
	StatusDegraded ServerStatus = "degraded"
	StatusDown     ServerStatus = "down"
)

// MCPServer represents runtime state for an MCP server connection
type MCPServer struct {
	Name            string
	Config          types.MCPServerConfig
	Client          *MCPClient
	HealthStatus    ServerStatus
	LastHealthCheck time.Time
	Tools           []types.Tool
	Stats           ServerStats
	mu              sync.RWMutex
}

// ServerStats tracks runtime statistics for an MCP server
type ServerStats struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	ErrorRate          float64
	AvgLatencyMs       int
	LastError          string
	LastErrorTime      time.Time
	ConsecutiveFailures int
}

// MCPManager orchestrates multiple MCP server connections
type MCPManager struct {
	config  types.MCPConfig
	servers map[string]*MCPServer
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewMCPManager creates a new MCP manager
func NewMCPManager(config types.MCPConfig) *MCPManager {
	return &MCPManager{
		config:  config,
		servers: make(map[string]*MCPServer),
		stopCh:  make(chan struct{}),
	}
}

// Initialize initializes all configured MCP servers
func (m *MCPManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, serverConfig := range m.config.Servers {
		// Create MCP client
		client, err := NewMCPClient(serverConfig)
		if err != nil {
			return fmt.Errorf("failed to create client for server %s: %w", name, err)
		}

		// Initialize client (perform handshake)
		if err := client.Initialize(ctx); err != nil {
			// Log error but continue with other servers (graceful degradation)
			fmt.Printf("Warning: failed to initialize server %s: %v\n", name, err)

			// Create server entry with down status
			m.servers[name] = &MCPServer{
				Name:         name,
				Config:       serverConfig,
				Client:       nil,
				HealthStatus: StatusDown,
				Tools:        []types.Tool{},
				Stats: ServerStats{
					LastError:     err.Error(),
					LastErrorTime: time.Now(),
				},
			}
			continue
		}

		// Create server entry
		server := &MCPServer{
			Name:            name,
			Config:          serverConfig,
			Client:          client,
			HealthStatus:    StatusHealthy,
			LastHealthCheck: time.Now(),
			Tools:           client.ListTools(),
		}

		m.servers[name] = server
	}

	// Start health check goroutine if enabled
	if m.config.HealthCheckInterval > 0 {
		m.wg.Add(1)
		go m.healthCheckLoop()
	}

	return nil
}

// DiscoverTools aggregates tools from all healthy servers
func (m *MCPManager) DiscoverTools() []types.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allTools []types.Tool
	for _, server := range m.servers {
		if server.HealthStatus == StatusHealthy || server.HealthStatus == StatusDegraded {
			server.mu.RLock()
			allTools = append(allTools, server.Tools...)
			server.mu.RUnlock()
		}
	}

	return allTools
}

// ExecuteTool routes a tool execution to the appropriate MCP server
func (m *MCPManager) ExecuteTool(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	server, exists := m.servers[serverName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	// Check server health
	if server.HealthStatus == StatusDown {
		return nil, fmt.Errorf("server %s is down", serverName)
	}

	// Execute with retry logic
	maxRetries := m.config.Retry.MaxAttempts
	if maxRetries == 0 {
		maxRetries = 3
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		startTime := time.Now()

		// Call the tool
		result, err := server.Client.CallTool(ctx, toolName, arguments)
		latency := time.Since(startTime).Milliseconds()

		// Update stats
		server.mu.Lock()
		server.Stats.TotalRequests++
		if err == nil {
			server.Stats.SuccessfulRequests++
			server.Stats.ConsecutiveFailures = 0

			// Update average latency
			if server.Stats.AvgLatencyMs == 0 {
				server.Stats.AvgLatencyMs = int(latency)
			} else {
				server.Stats.AvgLatencyMs = (server.Stats.AvgLatencyMs + int(latency)) / 2
			}
		} else {
			server.Stats.FailedRequests++
			server.Stats.LastError = err.Error()
			server.Stats.LastErrorTime = time.Now()
			server.Stats.ConsecutiveFailures++
		}
		server.Stats.ErrorRate = float64(server.Stats.FailedRequests) / float64(server.Stats.TotalRequests)
		server.mu.Unlock()

		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !m.isRetryable(err) {
			m.markServerStatus(serverName, StatusDown)
			return nil, fmt.Errorf("permanent error from server %s: %w", serverName, err)
		}

		// Calculate backoff delay
		if attempt < maxRetries-1 {
			delay := m.calculateBackoff(attempt)
			time.Sleep(delay)
		}
	}

	// Mark server as degraded after max retries
	m.markServerStatus(serverName, StatusDegraded)
	return nil, fmt.Errorf("server %s failed after %d retries: %w", serverName, maxRetries, lastErr)
}

// Health checks the health of a specific server or all servers
func (m *MCPManager) Health(serverName string) (ServerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if serverName == "" {
		// Check all servers
		healthyCount := 0
		totalCount := len(m.servers)
		for _, server := range m.servers {
			if server.HealthStatus == StatusHealthy {
				healthyCount++
			}
		}

		if healthyCount == 0 {
			return StatusDown, fmt.Errorf("all servers are down")
		} else if healthyCount < totalCount {
			return StatusDegraded, fmt.Errorf("%d/%d servers are healthy", healthyCount, totalCount)
		}
		return StatusHealthy, nil
	}

	// Check specific server
	server, exists := m.servers[serverName]
	if !exists {
		return StatusUnknown, fmt.Errorf("server not found: %s", serverName)
	}

	return server.HealthStatus, nil
}

// GetServer returns a specific MCP server
func (m *MCPManager) GetServer(name string) (*MCPServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[name]
	return server, exists
}

// ListServers returns all MCP servers
func (m *MCPManager) ListServers() []*MCPServer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*MCPServer, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	return servers
}

// Close closes all MCP server connections
func (m *MCPManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop health check loop
	close(m.stopCh)
	m.wg.Wait()

	// Close all server connections
	var errs []error
	for name, server := range m.servers {
		if server.Client != nil {
			if err := server.Client.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close server %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// healthCheckLoop periodically checks server health
func (m *MCPManager) healthCheckLoop() {
	defer m.wg.Done()

	interval := time.Duration(m.config.HealthCheckInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performHealthChecks()
		case <-m.stopCh:
			return
		}
	}
}

// performHealthChecks checks health of all servers
func (m *MCPManager) performHealthChecks() {
	m.mu.RLock()
	servers := make([]*MCPServer, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	m.mu.RUnlock()

	for _, server := range servers {
		// Skip down servers (they won't respond anyway)
		if server.HealthStatus == StatusDown && server.Stats.ConsecutiveFailures > 5 {
			continue
		}

		// Ping the server by listing tools
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := server.Client.CallTool(ctx, "ping", nil)
		cancel()

		server.mu.Lock()
		server.LastHealthCheck = time.Now()

		if err == nil {
			// Server is healthy
			if server.HealthStatus == StatusDown || server.HealthStatus == StatusDegraded {
				server.HealthStatus = StatusHealthy
				fmt.Printf("Server %s recovered\n", server.Name)
			}
			server.Stats.ConsecutiveFailures = 0
		} else {
			// Server is unhealthy
			server.Stats.ConsecutiveFailures++

			// Determine new status based on error rate and consecutive failures
			if server.Stats.ErrorRate >= 0.50 || server.Stats.ConsecutiveFailures > 5 {
				server.HealthStatus = StatusDown
			} else if server.Stats.ErrorRate > 0.10 {
				server.HealthStatus = StatusDegraded
			}
		}
		server.mu.Unlock()
	}
}

// markServerStatus updates a server's health status
func (m *MCPManager) markServerStatus(serverName string, status ServerStatus) {
	m.mu.RLock()
	server, exists := m.servers[serverName]
	m.mu.RUnlock()

	if !exists {
		return
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	oldStatus := server.HealthStatus
	server.HealthStatus = status

	if oldStatus != status {
		fmt.Printf("Server %s status changed: %s -> %s\n", serverName, oldStatus, status)
	}
}

// isRetryable determines if an error is retryable
func (m *MCPManager) isRetryable(err error) bool {
	// Check for specific error types
	if err == ErrTimeout || err == ErrTransportFailed {
		return true
	}

	// TODO: Add more sophisticated retry logic based on error messages
	return false
}

// calculateBackoff calculates exponential backoff delay
func (m *MCPManager) calculateBackoff(attempt int) time.Duration {
	baseDelay := time.Duration(m.config.Retry.InitialDelayMS) * time.Millisecond
	maxDelay := time.Duration(m.config.Retry.MaxDelayMS) * time.Millisecond

	switch m.config.Retry.Backoff {
	case "exponential":
		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		if delay > maxDelay {
			return maxDelay
		}
		return delay

	case "linear":
		delay := baseDelay + time.Duration(attempt)*baseDelay
		if delay > maxDelay {
			return maxDelay
		}
		return delay

	default: // "constant"
		return baseDelay
	}
}
