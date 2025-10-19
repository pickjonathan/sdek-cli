# MCP Integration Status

**Date**: October 19, 2025  
**Context**: Testing AI evidence collection with AWS MCP tools  
**Branch**: 004-mcp-native-agent

## Summary

The AWS MCP server is **operational and ready**, but there's a **system integration gap** preventing its use in AI plan execution.

## Current State

### ✅ What's Working

1. **AWS MCP Server Configuration** (Fixed)
   - Config: `~/.sdek/mcp/aws.json`
   - Command: Docker (public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest)
   - Status: **READY** (handshake successful)
   - Latency: ~1.24s
   - Capabilities: `aws.call-aws`, `aws.suggest-aws-commands`

2. **MCP Registry** (Feature 004)
   - Registry initialization: ✅
   - Tool discovery: ✅
   - Health monitoring: ✅
   - Validation: ✅
   - Handshake protocol: ✅ (fixed to use MCP `initialize` method)

3. **AI Plan Generation**
   - GPT-4 generates plans with AWS evidence items
   - Plan includes: CloudTrail events, IAM users/policies
   - Signal scores: 1.00 (high confidence)
   - Auto-approval: ✅

### ❌ What's Broken

**AI Plan Execution**: Error: `"all MCP connector calls failed"`

**Root Cause**: The AI plan execution (`ExecutePlan`) uses the **legacy connector system from Feature 002/003**, which only has GitHub implemented. AWS, Jira, and Slack connectors are marked as TODOs.

```go
// internal/ai/engine.go:209-212
builder.RegisterFactory("github", connectors.NewGitHubConnector)
// TODO: Add more when implemented
// builder.RegisterFactory("jira", connectors.NewJiraConnector)
// builder.RegisterFactory("aws", connectors.NewAWSConnector)
// builder.RegisterFactory("slack", connectors.NewSlackConnector)
```

## Architecture Gap

There are **TWO SEPARATE SYSTEMS** that are not integrated:

### Legacy Connector System (Feature 002/003)
- **Location**: `internal/ai/connectors/`
- **Interface**: `MCPConnector.Collect(ctx, source, query) ([]types.EvidenceEvent, error)`
- **Implemented**: GitHub only
- **Status**: TODO for AWS/Jira/Slack
- **Used by**: `ai.Engine.ExecutePlan()`

### New MCP Registry (Feature 004)
- **Location**: `internal/mcp/`
- **Interface**: `AgentInvoker.InvokeTool(ctx, agentRole, toolName, method, args) (*types.Evidence, error)`
- **Implemented**: AWS (via Docker), GitHub (via config)
- **Status**: Operational but NOT used by AI plan execution
- **Features**: RBAC, budgets, caching, redaction, audit logs

## Fixes Applied

1. **MCP Handshake Protocol** (`internal/mcp/transport/stdio.go`)
   - Changed from `ping` to `initialize` method
   - Added proper MCP initialization parameters:
     ```go
     params := map[string]interface{}{
         "protocolVersion": "2024-11-05",
         "capabilities":    map[string]interface{}{},
         "clientInfo": map[string]interface{}{
             "name":    "sdek-cli",
             "version": "1.0.0",
         },
     }
     ```

2. **AWS MCP Configuration** (`~/.sdek/mcp/aws.json`)
   - Fixed command: `uvx` → `docker`
   - Fixed args: Proper Docker run command with AWS credentials
   - Fixed capabilities: `aws_call_aws` → `aws.call-aws` (hyphenated)

3. **Plan Approval** (`cmd/ai_plan.go:254`)
   - Set `plan.Status = types.PlanApproved` in auto-approve mode

## Test Results

### MCP Server Test
```bash
$ ./sdek mcp test aws-api
Tool Name:    aws-api
Status:       ready  ✅
Latency:      1.24s
Capabilities: 2
- aws.call-aws
- aws.suggest-aws-commands
```

### AI Plan Generation (Dry Run)
```bash
$ ./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file test_policy.json --dry-run
Plan Items:
  1. [pending] AWS: CloudTrail events (signal: 1.00)
  2. [pending] AWS: IAM users and policies (signal: 1.00)
  3. [pending] Github: type:commit message:access control (signal: 0.80)
  4. [pending] Github: type:pr label:security (signal: 0.90)
  5. [pending] Jira: project = SOC2 AND issuetype = Task (signal: 0.70)
  6. [pending] Slack: in:security channel 'access control' (signal: 0.60)
```

### AI Plan Execution (Failed)
```bash
$ ./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file test_policy.json --approve-all
MCP tools registered: 1
MCP tool available: aws-api (status: ready, enabled: true)
Legacy connectors enabled: [aws]
⚠️  Warning: Legacy connector system from Feature 003 - only GitHub is implemented
⚠️  For AWS/Jira/Slack, use MCP tools (Feature 004)
Error: failed to execute plan: ai: all MCP connector calls failed  ❌
```

## Solutions (Priority Order)

### Option 1: Bridge Adapter (Quick Fix)
**Effort**: Medium  
**Impact**: Immediate functionality  
**Approach**: Create an adapter in the legacy connector system that wraps the new MCP registry

```go
// Pseudocode
type MCPRegistryAdapter struct {
    registry *mcp.Registry
    invoker  mcp.AgentInvoker
}

func (a *MCPRegistryAdapter) Collect(ctx context.Context, source, query string) ([]types.EvidenceEvent, error) {
    // Map source name to tool name
    toolName := mapSourceToTool(source) // "AWS" → "aws-api"
    
    // Parse query to extract method and args
    method, args := parseQuery(query)
    
    // Invoke MCP tool
    evidence, err := a.invoker.InvokeTool(ctx, "autonomous-agent", toolName, method, args)
    
    // Convert Evidence → []EvidenceEvent
    return convertToEvents(evidence), nil
}
```

**Challenges**:
- Import cycle (cmd → internal/mcp → internal/ai)
- Different return types (Evidence vs EvidenceEvent)
- Query parsing complexity (natural language → structured args)

### Option 2: Refactor ExecutePlan (Proper Fix)
**Effort**: High  
**Impact**: Clean architecture  
**Approach**: Update `ai.Engine.ExecutePlan()` to use new MCP registry directly

**Changes Required**:
1. Update `ExecutePlan` signature or implementation
2. Replace `e.connector.Collect()` with MCP invoker calls
3. Map plan items to MCP capabilities
4. Handle Evidence → EvidenceEvent conversion
5. Update all tests

### Option 3: AI Query Translation (Advanced)
**Effort**: Very High  
**Impact**: Best user experience  
**Approach**: Use AI to translate natural language queries into MCP tool invocations

```go
// AI translates: "AWS: CloudTrail events" 
// Into: { tool: "aws-api", method: "call-aws", args: { command: "aws cloudtrail lookup-events --max-results 50" }}
```

## Recommended Next Steps

1. **Immediate**: Document the gap (this file ✅)
2. **Short-term**: Create minimal AWS connector wrapper for testing
3. **Medium-term**: Implement proper MCP registry integration in ExecutePlan
4. **Long-term**: Migrate all connectors to MCP system, deprecate legacy

## Related Files

- `internal/ai/engine.go` - ExecutePlan implementation
- `internal/ai/connectors/registry.go` - Legacy connector registry
- `internal/mcp/registry.go` - New MCP registry
- `internal/mcp/invoker.go` - MCP tool invocation with RBAC/budgets
- `cmd/ai_plan.go` - AI plan command with dual initialization

## Testing Commands

```bash
# Test MCP server directly
./sdek mcp test aws-api

# Preview AI plan (no execution)
./sdek ai plan --framework SOC2 --section CC6.1 \
    --excerpts-file test_policy.json \
    --config test_ai_config.yaml \
    --dry-run

# Attempt execution (currently fails)
./sdek ai plan --framework SOC2 --section CC6.1 \
    --excerpts-file test_policy.json \
    --config test_ai_config.yaml \
    --approve-all
```

## Conclusion

The AWS MCP infrastructure is **fully operational**, but there's a **missing bridge** between the new MCP system (Feature 004) and the AI plan execution (Feature 002/003). The quickest path forward is to create an adapter that wraps MCP tools in the legacy connector interface.

**Status**: 🔴 **Blocked** - Requires integration work to connect working systems
