package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp/transport"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Registry manages the lifecycle of MCP tool connections.
type Registry struct {
	tools      map[string]*types.MCPTool
	transports map[string]transport.Transport // Keep transports alive
	mu         sync.RWMutex
	loader     *Loader
	validator  *Validator
	watcher    *Watcher
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// NewRegistry creates a new MCP registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:      make(map[string]*types.MCPTool),
		transports: make(map[string]transport.Transport),
		loader:     NewLoader("", ""),
		validator:  NewValidator(),
		stopCh:     make(chan struct{}),
	}
}

// Init discovers and initializes all MCP tools from configured paths.
func (r *Registry) Init(ctx context.Context) (int, error) {
	configs, err := r.loader.LoadConfigs()
	if err != nil {
		return 0, fmt.Errorf("failed to load configs: %w", err)
	}

	if len(configs) == 0 {
		return 0, nil
	}

	successCount := 0
	var initWg sync.WaitGroup

	for _, config := range configs {
		initWg.Add(1)
		go func(cfg *types.MCPConfig) {
			defer initWg.Done()

			if err := r.initTool(ctx, cfg); err != nil {
				fmt.Printf("warning: failed to init tool %s: %v\n", cfg.Name, err)
				r.mu.Lock()
				r.tools[cfg.Name] = &types.MCPTool{
					Name:            cfg.Name,
					Config:          cfg,
					Status:          types.ToolStatusDegraded,
					LastError:       err,
					LastHealthCheck: time.Now(),
					Enabled:         true,
					CircuitBreaker: &types.CircuitBreaker{
						State:    types.CircuitBreakerOpen,
						Failures: 1,
					},
				}
				r.mu.Unlock()
			} else {
				r.mu.Lock()
				successCount++
				r.mu.Unlock()
			}
		}(config)
	}

	initWg.Wait()

	r.wg.Add(1)
	go r.healthMonitor(ctx)

	if r.watcher != nil {
		r.wg.Add(1)
		go r.watcher.Watch(ctx, r, r.stopCh, nil)
	}

	return successCount, nil
}

// initTool initializes a single tool with handshake.
func (r *Registry) initTool(ctx context.Context, config *types.MCPConfig) error {
	var trans transport.Transport
	var err error

	switch config.Transport {
	case "stdio":
		trans, err = transport.NewStdioTransport(config)
	case "http":
		trans, err = transport.NewHTTPTransport(config)
	default:
		return fmt.Errorf("unsupported transport: %s", config.Transport)
	}

	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	handshakeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	start := time.Now()
	err = trans.HealthCheck(handshakeCtx)
	handshakeLatency := time.Since(start)

	tool := &types.MCPTool{
		Name:            config.Name,
		Config:          config,
		LastHealthCheck: time.Now(),
		Enabled:         true,
		Metrics: types.ToolMetrics{
			HandshakeLatency: handshakeLatency,
		},
		CircuitBreaker: &types.CircuitBreaker{
			State: types.CircuitBreakerClosed,
		},
	}

	if err != nil {
		tool.Status = types.ToolStatusDegraded
		tool.LastError = err
		tool.CircuitBreaker.Failures = 1
		trans.Close() // Close only on failure
		return fmt.Errorf("handshake failed: %w", err)
	}

	tool.Status = types.ToolStatusReady

	r.mu.Lock()
	r.tools[config.Name] = tool
	r.transports[config.Name] = trans // Store transport for reuse
	r.mu.Unlock()

	return nil
}

// Close gracefully shuts down all tool connections.
func (r *Registry) Close(ctx context.Context) error {
	close(r.stopCh)
	r.wg.Wait()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Close all transports
	for _, trans := range r.transports {
		trans.Close()
	}

	return nil
}

// GetTransport returns the transport for a tool (for use by invoker).
func (r *Registry) GetTransport(toolName string) (transport.Transport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	trans, ok := r.transports[toolName]
	if !ok {
		return nil, fmt.Errorf("transport not found for tool: %s", toolName)
	}

	return trans, nil
}

// Reload re-scans config directories and hot-reloads changed tools.
func (r *Registry) Reload(ctx context.Context) (int, error) {
	configs, err := r.loader.LoadConfigs()
	if err != nil {
		return 0, fmt.Errorf("failed to load configs: %w", err)
	}

	configMap := make(map[string]*types.MCPConfig)
	for _, cfg := range configs {
		configMap[cfg.Name] = cfg
	}

	reloadCount := 0

	r.mu.Lock()
	defer r.mu.Unlock()

	for name, newConfig := range configMap {
		if _, exists := r.tools[name]; exists {
			reloadCount++
		} else {
			go func(cfg *types.MCPConfig) {
				r.initTool(ctx, cfg)
			}(newConfig)
			reloadCount++
		}
	}

	for name := range r.tools {
		if _, exists := configMap[name]; !exists {
			r.tools[name].Status = types.ToolStatusOffline
			r.tools[name].Enabled = false
			delete(r.tools, name)
			reloadCount++
		}
	}

	return reloadCount, nil
}

// List returns all discovered tools with their current status.
func (r *Registry) List(ctx context.Context) ([]types.MCPTool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]types.MCPTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, *tool)
	}

	return tools, nil
}

// Get retrieves a specific tool by name.
func (r *Registry) Get(ctx context.Context, name string) (types.MCPTool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return types.MCPTool{}, ErrToolNotFound
	}

	return *tool, nil
}

// Enable marks a tool as administratively enabled.
func (r *Registry) Enable(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tool, exists := r.tools[name]
	if !exists {
		return ErrToolNotFound
	}

	tool.Enabled = true

	if tool.CircuitBreaker.State != types.CircuitBreakerOpen {
		tool.Status = types.ToolStatusReady
	}

	return nil
}

// Disable marks a tool as administratively disabled.
func (r *Registry) Disable(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tool, exists := r.tools[name]
	if !exists {
		return ErrToolNotFound
	}

	tool.Enabled = false
	tool.Status = types.ToolStatusOffline

	return nil
}

// Validate validates one or more config files against the schema.
func (r *Registry) Validate(ctx context.Context, paths ...string) ([]types.SchemaError, error) {
	var allErrors []types.SchemaError

	for _, path := range paths {
		errors := r.validator.Validate(path)
		allErrors = append(allErrors, errors...)
	}

	return allErrors, nil
}

// Test performs a health check and handshake on a tool.
func (r *Registry) Test(ctx context.Context, name string) (types.MCPHealthReport, error) {
	r.mu.RLock()
	tool, exists := r.tools[name]
	r.mu.RUnlock()

	if !exists {
		return types.MCPHealthReport{}, ErrToolNotFound
	}

	report := types.MCPHealthReport{
		ToolName:         tool.Name,
		Status:           tool.Status,
		HandshakeLatency: tool.Metrics.HandshakeLatency,
		Capabilities:     tool.Config.Capabilities,
		LastError:        tool.LastError,
		Timestamp:        time.Now(),
	}

	return report, nil
}

// healthMonitor runs periodic health checks in the background.
func (r *Registry) healthMonitor(ctx context.Context) {
	defer r.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.performHealthChecks(ctx)
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// performHealthChecks checks health of all tools.
func (r *Registry) performHealthChecks(ctx context.Context) {
	r.mu.RLock()
	tools := make([]*types.MCPTool, 0, len(r.tools))
	for _, tool := range r.tools {
		if tool.Enabled {
			tools = append(tools, tool)
		}
	}
	r.mu.RUnlock()

	for _, tool := range tools {
		go func(t *types.MCPTool) {
			r.mu.Lock()
			t.LastHealthCheck = time.Now()
			r.mu.Unlock()
		}(tool)
	}
}
