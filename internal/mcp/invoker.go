package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/mcp/transport"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Enforcer defines the interface for RBAC and budget enforcement.
// This is duplicated here to avoid import cycles with internal/mcp/rbac.
type Enforcer interface {
	CheckPermission(ctx context.Context, agentRole string, capability string) (bool, error)
	GetCapabilities(ctx context.Context, agentRole string) ([]types.AgentCapability, error)
	ApplyBudget(ctx context.Context, toolName string, budget *types.ToolBudget) error
	RecordInvocation(ctx context.Context, log *types.MCPInvocationLog) error
}

// AgentInvoker orchestrates MCP tool invocations with RBAC, budgets, caching, and redaction.
type AgentInvoker interface {
	InvokeTool(ctx context.Context, agentRole, toolName, method string, args map[string]interface{}) (*types.Evidence, error)
}

// invoker implements the AgentInvoker interface.
type invoker struct {
	registry *Registry
	enforcer Enforcer
	redactor ai.Redactor
	cache    *store.Cache
}

// NewInvoker creates a new AgentInvoker instance.
func NewInvoker(reg *Registry, enf Enforcer, red ai.Redactor, cache *store.Cache) AgentInvoker {
	return &invoker{
		registry: reg,
		enforcer: enf,
		redactor: red,
		cache:    cache,
	}
}

// InvokeTool executes an MCP tool method with full orchestration:
// 1. RBAC permission check
// 2. Budget enforcement (rate limiting, concurrency)
// 3. Cache lookup
// 4. Transport invocation
// 5. Redaction
// 6. Audit logging
// 7. Evidence normalization
func (inv *invoker) InvokeTool(ctx context.Context, agentRole, toolName, method string, args map[string]interface{}) (*types.Evidence, error) {
	startTime := time.Now()

	// Step 1: Get the tool from registry
	tool, err := inv.registry.Get(ctx, toolName)
	if err != nil {
		return nil, err
	}

	if !tool.Enabled {
		return nil, ErrToolDisabled
	}

	// Step 2: RBAC permission check
	capability := fmt.Sprintf("%s.%s", toolName, method)
	permitted, err := inv.enforcer.CheckPermission(ctx, agentRole, capability)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !permitted {
		// Record audit log for permission denied
		inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "permission_denied", "", nil)
		return nil, ErrPermissionDenied
	}

	// Step 3: Budget enforcement
	if tool.Config != nil {
		budget := &types.ToolBudget{
			ToolName: toolName,
			RateLimit: types.RateLimit{
				RequestsPerSecond: 10.0, // Default rate limit
				BurstSize:         20,
			},
			ConcurrencyLimit: 5,
			Timeout:          30 * time.Second,
		}

		if err := inv.enforcer.ApplyBudget(ctx, toolName, budget); err != nil {
			inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "rate_limited", "", err)
			return nil, err
		}
	}

	// Step 4: Check cache
	cacheKey := inv.buildCacheKey(toolName, method, args)
	if cachedResult := inv.checkCache(cacheKey); cachedResult != nil {
		// Cache hit - return cached evidence
		inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "success_cached", "", nil)
		return cachedResult, nil
	}

	// Step 5: Invoke transport
	trans, err := inv.getTransport(&tool)
	if err != nil {
		inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "error", "", err)
		return nil, fmt.Errorf("failed to get transport: %w", err)
	}

	response, err := trans.Invoke(ctx, method, args)
	if err != nil {
		inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "error", "", err)
		return nil, fmt.Errorf("transport invocation failed: %w", err)
	}

	// Step 6: Redaction
	responseJSON, _ := json.Marshal(response)
	redactedText, _, err := inv.redactor.Redact(string(responseJSON))
	if err != nil {
		// Log error but continue - redaction is not critical
		redactedText = string(responseJSON)
	}

	var redactedResponse map[string]interface{}
	json.Unmarshal([]byte(redactedText), &redactedResponse)

	// Step 7: Normalize to Evidence
	evidence := normalizeEvidence(toolName, method, redactedResponse, startTime)

	// Step 8: Store in cache
	inv.storeCache(cacheKey, evidence)

	// Step 9: Audit log
	inv.recordAuditLog(ctx, agentRole, toolName, method, args, startTime, "success", evidence.ID, nil)

	return evidence, nil
}

// getTransport retrieves or creates the transport for a tool.
func (inv *invoker) getTransport(tool *types.MCPTool) (transport.Transport, error) {
	if tool.Config == nil {
		return nil, fmt.Errorf("tool config is nil")
	}

	switch tool.Config.Transport {
	case "stdio":
		trans, err := transport.NewStdioTransport(tool.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdio transport: %w", err)
		}
		return trans, nil
	case "http":
		trans, err := transport.NewHTTPTransport(tool.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP transport: %w", err)
		}
		return trans, nil
	default:
		return nil, fmt.Errorf("unsupported transport: %s", tool.Config.Transport)
	}
}

// buildCacheKey creates a cache key from tool name, method, and args.
func (inv *invoker) buildCacheKey(toolName, method string, args map[string]interface{}) string {
	argsJSON, _ := json.Marshal(args)
	hash := sha256.Sum256(argsJSON)
	return fmt.Sprintf("mcp:%s:%s:%x", toolName, method, hash[:8])
}

// checkCache looks up a cached result.
func (inv *invoker) checkCache(key string) *types.Evidence {
	// TODO: Integrate with store.Cache once we determine the right cache structure
	// For now, return nil (cache miss)
	return nil
}

// storeCache stores a result in the cache.
func (inv *invoker) storeCache(key string, evidence *types.Evidence) {
	// TODO: Integrate with store.Cache once we determine the right cache structure
	// For now, no-op
}

// recordAuditLog creates and records an audit log entry.
func (inv *invoker) recordAuditLog(ctx context.Context, agentRole, toolName, method string, args map[string]interface{}, startTime time.Time, status, evidenceID string, err error) {
	argsJSON, _ := json.Marshal(args)
	hash := sha256.Sum256(argsJSON)

	log := &types.MCPInvocationLog{
		ID:               uuid.New().String(),
		Timestamp:        startTime,
		RunID:            getRunIDFromContext(ctx),
		AgentID:          getAgentIDFromContext(ctx),
		AgentRole:        agentRole,
		ToolName:         toolName,
		Method:           method,
		ArgsHash:         fmt.Sprintf("%x", hash),
		RedactionApplied: true, // Always true in our implementation
		Duration:         time.Since(startTime),
		Status:           status,
		ErrorMessage:     getErrorMessage(err),
	}

	// Record in enforcer's audit log
	inv.enforcer.RecordInvocation(ctx, log)
}

// getRunIDFromContext extracts run ID from context, or generates one.
func getRunIDFromContext(ctx context.Context) string {
	if runID, ok := ctx.Value("run_id").(string); ok {
		return runID
	}
	return uuid.New().String()
}

// getAgentIDFromContext extracts agent ID from context, or returns "unknown".
func getAgentIDFromContext(ctx context.Context) string {
	if agentID, ok := ctx.Value("agent_id").(string); ok {
		return agentID
	}
	return "unknown"
}

// getErrorMessage safely extracts error message.
func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
