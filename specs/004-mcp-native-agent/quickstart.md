# Quickstart: MCP-Native Agent Orchestrator & Tooling Config

**Feature**: 004-mcp-native-agent  
**Date**: 2025-10-19  
**Purpose**: Step-by-step validation scenarios for testing the MCP feature implementation

---

## Prerequisites

- sdek-cli installed and configured
- Access to test MCP servers or mock servers (provided in `testdata/mcp/mock_server/`)
- Environment variables set (if using real tools):
  - `GITHUB_TOKEN` (for GitHub MCP)
  - `JIRA_API_TOKEN` (for Jira MCP)
  - `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` (for AWS MCP)

---

## Scenario 1: Configuration Discovery and Loading (AC-01)

**Objective**: Verify that valid MCP configs are discovered and loaded automatically.

### Steps

1. **Create a sample GitHub MCP config**:
   ```bash
   mkdir -p ~/.sdek/mcp
   cat > ~/.sdek/mcp/github.json <<EOF
   {
     "name": "github",
     "command": "/usr/local/bin/mcp-github",
     "args": ["--verbose"],
     "env": {
       "GITHUB_TOKEN": "${GITHUB_TOKEN}"
     },
     "transport": "stdio",
     "capabilities": ["read", "commits.list", "pr.list"],
     "timeout": "30s",
     "schemaVersion": "1.0.0"
   }
   EOF
   ```

2. **Start sdek-cli and list MCP tools**:
   ```bash
   sdek mcp list
   ```

3. **Expected Output**:
   ```
   MCP Tools:
   
   NAME     STATUS    LATENCY  CAPABILITIES           ERRORS
   github   ready     45ms     read, commits.list...  -
   ```

4. **Verify in TUI**:
   ```bash
   sdek tui
   # Navigate to "MCP Tools" panel
   # Should see github with green "ready" status
   ```

### Success Criteria
- ✅ Tool appears in `sdek mcp list` with status "ready"
- ✅ Tool is available for agent invocation
- ✅ Handshake latency is displayed

---

## Scenario 2: Schema Validation (AC-02)

**Objective**: Verify that invalid configs are rejected with detailed errors.

### Steps

1. **Create an invalid MCP config** (missing required "command" field):
   ```bash
   cat > ~/.sdek/mcp/invalid.json <<EOF
   {
     "name": "invalid-tool",
     "transport": "stdio",
     "capabilities": ["read"],
     "schemaVersion": "1.0.0"
   }
   EOF
   ```

2. **Validate the config**:
   ```bash
   sdek mcp validate ~/.sdek/mcp/invalid.json
   ```

3. **Expected Output**:
   ```
   Validation Errors:
   
   File: /Users/you/.sdek/mcp/invalid.json
   Line: 1
   Path: /command
   Error: missing required property 'command'
   ```

4. **Verify tool is not loaded**:
   ```bash
   sdek mcp list | grep invalid-tool
   # Should return empty (tool not loaded)
   ```

### Success Criteria
- ✅ Validation error specifies file, line, and missing property
- ✅ Invalid tool is not loaded into registry
- ✅ Other valid tools remain operational

---

## Scenario 3: Orchestrator Resilience (AC-03)

**Objective**: Verify that the orchestrator handles unavailable MCP servers gracefully.

### Steps

1. **Create a config pointing to a non-existent server**:
   ```bash
   cat > ~/.sdek/mcp/unreachable.json <<EOF
   {
     "name": "unreachable",
     "command": "/tmp/nonexistent-mcp-server",
     "transport": "stdio",
     "capabilities": ["test"],
     "timeout": "5s",
     "schemaVersion": "1.0.0"
   }
   EOF
   ```

2. **Start sdek-cli (or trigger reload)**:
   ```bash
   sdek mcp list
   ```

3. **Expected Output**:
   ```
   MCP Tools:
   
   NAME          STATUS     LATENCY  CAPABILITIES  ERRORS
   unreachable   degraded   -        test          exec: "/tmp/nonexistent...": file not found
   ```

4. **Observe retry behavior** (check logs):
   ```bash
   sdek mcp list --verbose
   # Should show retry attempts with exponential backoff
   ```

5. **Fix the server** (simulate recovery):
   ```bash
   # Copy a working mock server to /tmp/nonexistent-mcp-server
   cp testdata/mcp/mock_server/main.go /tmp/nonexistent-mcp-server
   chmod +x /tmp/nonexistent-mcp-server
   ```

6. **Verify transition to ready** (without restart):
   ```bash
   # Wait for health check interval (default: 30s) or trigger manually:
   sdek mcp test unreachable
   
   # Check status:
   sdek mcp list | grep unreachable
   # Should now show "ready"
   ```

### Success Criteria
- ✅ Tool marked as "degraded" when server unavailable
- ✅ Retry attempts logged with exponential backoff
- ✅ Tool transitions to "ready" when server becomes available (no restart)
- ✅ Other tools remain operational

---

## Scenario 4: RBAC Enforcement (AC-04)

**Objective**: Verify that agents without permission are denied access.

### Steps

1. **Configure RBAC policies**:
   ```bash
   cat >> ~/.sdek/config.yaml <<EOF
   mcp:
     rbac:
       roles:
         - role: evidence-collector
           capabilities:
             - github.read
             - github.commits.list
         - role: read-only
           capabilities:
             - github.read
   EOF
   ```

2. **Attempt unauthorized call** (as read-only agent):
   ```bash
   # Simulate an agent call (requires integration test or direct API call)
   # This is best tested via integration test, but CLI example:
   
   sdek analyze --agent-role=read-only --mcp-tool=github --mcp-method=pr.create
   ```

3. **Expected Output**:
   ```
   Error: permission denied
   Agent role 'read-only' does not have capability 'github.pr.create'
   Required capability: github.pr.create
   ```

4. **Verify audit log**:
   ```bash
   cat ~/.sdek/logs/mcp-invocations.jsonl | tail -1 | jq .
   ```

   **Expected**:
   ```json
   {
     "id": "uuid",
     "timestamp": "2025-10-19T12:00:00Z",
     "agentRole": "read-only",
     "toolName": "github",
     "method": "pr.create",
     "status": "permission_denied",
     "errorMessage": "Agent role 'read-only' does not have capability 'github.pr.create'"
   }
   ```

### Success Criteria
- ✅ Unauthorized call is denied immediately
- ✅ Audit log contains `permission_denied` entry
- ✅ Error message indicates required capability

---

## Scenario 5: Evidence Collection via MCP (AC-05)

**Objective**: Verify end-to-end evidence collection via MCP tools.

### Steps

1. **Configure MCP tools** (GitHub, Jira):
   ```bash
   # (Already configured in Scenario 1)
   # Add Jira config:
   cat > ~/.sdek/mcp/jira.json <<EOF
   {
     "name": "jira",
     "command": "npx",
     "args": ["@jira/mcp-server"],
     "env": {
       "JIRA_API_TOKEN": "${JIRA_API_TOKEN}",
       "JIRA_URL": "https://yourcompany.atlassian.net"
     },
     "transport": "stdio",
     "capabilities": ["search", "tickets.list"],
     "timeout": "45s",
     "schemaVersion": "1.0.0"
   }
   EOF
   ```

2. **Run a compliance analysis** (SOC2 example):
   ```bash
   sdek analyze --framework=soc2 --controls=CC6.1 --evidence-sources=mcp
   ```

3. **Expected Behavior**:
   - sdek-cli invokes `github.commits.list` to collect code changes
   - sdek-cli invokes `jira.tickets.list` to collect change tickets
   - Evidence is normalized into the evidence graph
   - Redaction policies are applied
   - Analysis report generated

4. **Verify evidence provenance**:
   ```bash
   sdek report --format=json | jq '.findings[0].evidence[] | select(.sourceType == "mcp")'
   ```

   **Expected**:
   ```json
   {
     "id": "uuid",
     "sourceType": "mcp",
     "mcpToolName": "github",
     "mcpMethod": "commits.list",
     "timestamp": "2025-10-19T12:00:00Z",
     "data": {...},
     "redacted": true
   }
   ```

5. **Verify audit trail**:
   ```bash
   cat ~/.sdek/logs/mcp-invocations.jsonl | jq 'select(.toolName == "github" or .toolName == "jira")'
   ```

   **Expected**: Log entries for both GitHub and Jira invocations with:
   - Tool name, method, duration
   - Redaction applied flag
   - Status: "success"

### Success Criteria
- ✅ Evidence collected via MCP tools
- ✅ Evidence normalized into evidence graph
- ✅ Redaction policies applied
- ✅ Audit logs include tool name, method, redaction status, duration
- ✅ Analysis report generated successfully

---

## Scenario 6: CLI and TUI Operations (AC-06)

**Objective**: Verify CLI/TUI operations for managing MCP tools.

### CLI Operations

1. **List all tools**:
   ```bash
   sdek mcp list
   ```
   Expected: Table with tool name, status, latency, capabilities, errors

2. **Test a tool**:
   ```bash
   sdek mcp test github
   ```
   Expected: Health check report with handshake status, latency, capabilities

3. **Disable a tool**:
   ```bash
   sdek mcp disable github
   ```
   Expected: `Tool 'github' disabled successfully`
   
   Verify:
   ```bash
   sdek mcp list | grep github
   # Status should be "offline" (administratively disabled)
   ```

4. **Enable a tool**:
   ```bash
   sdek mcp enable github
   ```
   Expected: `Tool 'github' enabled successfully`
   
   Verify:
   ```bash
   sdek mcp list | grep github
   # Status should transition to "ready"
   ```

### TUI Operations

1. **Open TUI**:
   ```bash
   sdek tui
   ```

2. **Navigate to MCP Tools panel**:
   - Use arrow keys to select "MCP Tools" from main menu
   - Press Enter

3. **View tool details**:
   - Select a tool (e.g., "github")
   - Press Enter to view details:
     - Status badge (green=ready, yellow=degraded, red=offline)
     - Latency metrics (handshake, average invocation)
     - Error details (if any)
     - Last health check timestamp

4. **Toggle tool on/off**:
   - Select tool
   - Press Space to toggle enable/disable
   - Observe status change in real-time

5. **Quick test**:
   - Select tool
   - Press 't' to trigger health check
   - Observe results inline (status, latency)

### Success Criteria
- ✅ `sdek mcp list` displays all tools with metrics
- ✅ `sdek mcp test <tool>` returns diagnostics
- ✅ `sdek mcp enable/disable` toggles tool availability
- ✅ State persists across restarts
- ✅ TUI panel shows real-time status updates
- ✅ TUI quick actions (test, toggle) work correctly

---

## Cleanup

After testing, remove test configs:
```bash
rm ~/.sdek/mcp/invalid.json
rm ~/.sdek/mcp/unreachable.json
# Keep github.json and jira.json for future testing
```

---

## Troubleshooting

### Tool shows "degraded" but server is running
- Check tool logs: `sdek mcp test <tool> --verbose`
- Verify environment variables are set correctly
- Check server logs (if accessible)

### Permission denied errors
- Verify RBAC policy in `~/.sdek/config.yaml`
- Check agent role has required capabilities
- Review audit log: `~/.sdek/logs/mcp-invocations.jsonl`

### Hot-reload not working
- Verify `mcp.hotReload` feature flag is enabled (default: true)
- Check file watcher logs: `sdek mcp list --verbose`
- Manually trigger reload: `sdek mcp reload` (if implemented)

---

## Summary

This quickstart validates all 6 acceptance criteria (AC-01 through AC-06) with concrete, executable scenarios. Each scenario includes:
- Clear steps
- Expected outputs
- Success criteria

**Status**: ✅ Quickstart defined, ready for implementation validation
