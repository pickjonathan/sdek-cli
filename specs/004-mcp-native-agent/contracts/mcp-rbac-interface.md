# MCP RBAC Interface Contract

**Feature**: 004-mcp-native-agent  
**Date**: 2025-10-19  
**Status**: Contract Definition (Implementation Pending)

---

## Overview

The `RBACEnforcer` interface provides capability-based access control for MCP tool invocations. It enforces which agents can invoke which tool methods and applies execution budgets (rate limits, concurrency, timeouts).

---

## Interface Definition

```go
package rbac

import (
	"context"
	"time"
	
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// RBACEnforcer enforces capability-based access control for MCP tools.
type RBACEnforcer interface {
	// CheckPermission verifies if an agent has permission to invoke a tool method.
	// Returns nil if allowed, error if denied.
	CheckPermission(ctx context.Context, agent types.Agent, tool, method string) error
	
	// GetCapabilities returns all capabilities granted to an agent role.
	GetCapabilities(ctx context.Context, role string) ([]string, error)
	
	// ApplyBudget checks if invocation is within tool's execution budget.
	// Returns nil if allowed, error if rate limit / concurrency exceeded.
	ApplyBudget(ctx context.Context, tool string) error
	
	// RecordInvocation records a successful invocation for audit and budget tracking.
	RecordInvocation(ctx context.Context, log types.MCPInvocationLog) error
}

// Agent represents an entity invoking MCP tools.
type Agent struct {
	ID   string // e.g., "evidence-collector-1"
	Role string // e.g., "evidence-collector"
}
```

---

## Method Contracts

### CheckPermission(ctx, agent, tool, method) error

**Purpose**: Verify agent has required capability.

**Behavior**:
1. Construct capability string: `<tool>.<method>` (e.g., "github.commits.list")
2. Lookup agent's role in RBAC policy
3. Check if capability is granted (exact match or wildcard)
4. Return nil if allowed, `ErrPermissionDenied` if denied

**Capability Matching**:
- Exact: `github.commits.list` matches `github.commits.list`
- Wildcard tool: `github.*` matches `github.commits.list` or any github method
- Wildcard all: `*.*` matches any tool and method

**Returns**:
- nil if permission granted
- `ErrPermissionDenied` if denied

**Acceptance Criteria** (AC-04):
- Given agent without capability, when invoked, returns ErrPermissionDenied

---

### GetCapabilities(ctx, role) ([]string, error)

**Purpose**: Retrieve all capabilities for a role.

**Behavior**:
1. Lookup role in RBAC policy config
2. Return list of capabilities
3. Return `ErrRoleNotFound` if role doesn't exist

**Returns**:
- Array of capability strings
- Error if role not found

---

### ApplyBudget(ctx, tool) error

**Purpose**: Check if invocation is within tool's execution budget.

**Behavior**:
1. Lookup tool's budget (rate limit, concurrency, daily quota)
2. Check rate limit (requests per second + burst)
3. Check concurrency limit (current in-flight calls)
4. Check daily quota (calls today)
5. If within limits: Allow and increment counters
6. If exceeded: Return `ErrRateLimitExceeded` or `ErrConcurrencyExceeded`

**Rate Limiting Algorithm** (Token Bucket):
- Bucket refills at `RequestsPerSecond` rate
- Bucket capacity = `BurstSize`
- Allow if bucket has ≥1 token, consume 1 token

**Returns**:
- nil if within budget
- `ErrRateLimitExceeded` if rate limit hit
- `ErrConcurrencyExceeded` if concurrency limit hit
- `ErrDailyQuotaExceeded` if daily quota hit

**Acceptance Criteria**:
- Given tool with 10 req/s limit, when 11th request in same second, returns ErrRateLimitExceeded

---

### RecordInvocation(ctx, log) error

**Purpose**: Record invocation for audit trail and budget tracking.

**Behavior**:
1. Append log entry to audit log file (`~/.sdek/logs/mcp-invocations.jsonl`)
2. Update budget counters (daily quota, last invocation time)
3. Emit telemetry event (`mcp_invoked`)

**Returns**:
- nil on success
- Error if write fails

---

## RBAC Policy Configuration

**File**: `~/.sdek/config.yaml` or `./.sdek/config.yaml`

```yaml
mcp:
  rbac:
    roles:
      - role: evidence-collector
        capabilities:
          - github.read
          - github.commits.list
          - github.pr.list
          - jira.search
          - jira.tickets.list
          - aws.iam.list
      
      - role: security-auditor
        capabilities:
          - "*.*"  # Full access
      
      - role: read-only
        capabilities:
          - "*.read"  # Read-only access to all tools
```

---

## Budget Configuration

**File**: `~/.sdek/config.yaml` or `./.sdek/config.yaml`

```yaml
mcp:
  budgets:
    - tool: github
      rateLimit:
        requestsPerSecond: 10
        burstSize: 20
      concurrencyLimit: 5
      timeout: 30s
      dailyQuota: 5000
    
    - tool: aws
      rateLimit:
        requestsPerSecond: 5
        burstSize: 10
      concurrencyLimit: 3
      timeout: 60s
      dailyQuota: 1000
```

---

## Error Types

```go
var (
	ErrPermissionDenied      = errors.New("rbac: permission denied")
	ErrRoleNotFound          = errors.New("rbac: role not found in policy")
	ErrRateLimitExceeded     = errors.New("rbac: rate limit exceeded")
	ErrConcurrencyExceeded   = errors.New("rbac: concurrency limit exceeded")
	ErrDailyQuotaExceeded    = errors.New("rbac: daily quota exceeded")
)
```

---

## Testing Strategy

### Unit Tests
- `TestCheckPermission`: Exact match, wildcard match, denial
- `TestApplyBudget`: Rate limit, concurrency limit, daily quota
- `TestRecordInvocation`: Audit log append, budget counter update

### Integration Tests
- `TestRBACEndToEnd`: Agent invokes tool, RBAC enforced, audit logged

---

## Dependencies

- `internal/config`: Load RBAC policy and budgets from Viper
- `internal/store`: Persist daily quota counters
- Token bucket library or custom implementation

---

**Contract Status**: ✅ Defined, awaiting implementation
