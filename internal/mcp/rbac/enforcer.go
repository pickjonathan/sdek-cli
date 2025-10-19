package rbac

import (
	"context"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Enforcer defines the interface for RBAC and budget enforcement.
type Enforcer interface {
	CheckPermission(ctx context.Context, agentRole string, capability string) (bool, error)
	GetCapabilities(ctx context.Context, agentRole string) ([]types.AgentCapability, error)
	ApplyBudget(ctx context.Context, toolName string, budget *types.ToolBudget) error
	RecordInvocation(ctx context.Context, log *types.MCPInvocationLog) error
}

// DefaultEnforcer implements the Enforcer interface with in-memory role management.
type DefaultEnforcer struct {
	roles   map[string][]types.AgentCapability
	audit   *AuditLogger
	budgets *BudgetManager
}

// NewEnforcer creates a new RBAC enforcer with default settings.
func NewEnforcer() *DefaultEnforcer {
	return &DefaultEnforcer{
		roles:   make(map[string][]types.AgentCapability),
		audit:   NewAuditLogger(),
		budgets: NewBudgetManager(),
	}
}

// CheckPermission verifies if the agent role has permission for the capability.
func (e *DefaultEnforcer) CheckPermission(ctx context.Context, agentRole string, capability string) (bool, error) {
	agentCaps, exists := e.roles[agentRole]
	if !exists {
		return false, nil
	}

	for _, agentCap := range agentCaps {
		for _, cap := range agentCap.Capabilities {
			if cap == capability {
				return true, nil
			}

			if matchWildcard(cap, capability) {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetCapabilities returns all capabilities available to the agent role.
func (e *DefaultEnforcer) GetCapabilities(ctx context.Context, agentRole string) ([]types.AgentCapability, error) {
	capabilities, exists := e.roles[agentRole]
	if !exists {
		return []types.AgentCapability{}, nil
	}

	return capabilities, nil
}

// ApplyBudget checks and enforces rate limits and concurrency limits for the tool.
func (e *DefaultEnforcer) ApplyBudget(ctx context.Context, toolName string, budget *types.ToolBudget) error {
	return e.budgets.CheckBudget(ctx, toolName, budget)
}

// RecordInvocation creates an audit log entry for the tool invocation.
func (e *DefaultEnforcer) RecordInvocation(ctx context.Context, log *types.MCPInvocationLog) error {
	return e.audit.Write(ctx, log)
}

// AddRole adds a role with its capabilities to the enforcer.
func (e *DefaultEnforcer) AddRole(role string, capabilities []types.AgentCapability) {
	e.roles[role] = capabilities
}

// matchWildcard checks if a capability pattern matches a specific capability.
func matchWildcard(pattern string, capability string) bool {
	if len(pattern) == 0 {
		return false
	}

	if pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(capability) >= len(prefix) && capability[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
