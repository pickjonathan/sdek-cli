# Implementation Continuation Plan - MCP Native Agent

**Date**: October 19, 2025  
**Branch**: `004-mcp-native-agent`  
**Current Status**: 42/64 tasks complete (66%)  
**Last Commit**: `d357fc7` - AWS MCP configuration and test results

---

## Current State

### ✅ Completed Phases (M1-M3)
- **Phase 3.1: Setup & Foundation** (T001-T004) - 100% ✅
- **Phase 3.2: Tests First (TDD)** (T005-T020) - 100% ✅
- **Phase 3.3: Core Implementation**
  - Types & Schema (T021-T026) - 100% ✅
  - Transports (T027-T029) - 100% ✅
  - Registry & Orchestrator (T030-T038) - 100% ✅
  - RBAC & Budgets (T039-T042) - 100% ✅

### 🔄 Remaining Work (22 tasks, 34%)

#### Phase 3.3: Evidence Integration (M4) - 5 tasks
- [ ] **T043** - AgentInvoker interface (`internal/mcp/invoker.go`)
- [ ] **T044** - InvokeTool implementation with RBAC + budget orchestration
- [ ] **T045** - Integrate existing redaction (`internal/ai/redactor.go`)
- [ ] **T046** - Integrate existing caching (`internal/store/cache.go`)
- [ ] **T047** - normalizeEvidence helper (`internal/mcp/evidence.go`)

#### Phase 3.4: CLI Commands (M5) - 6 tasks
- [ ] **T048** - Parent `mcp` command (`cmd/mcp.go`)
- [ ] **T049** - `mcp list` command
- [ ] **T050** - `mcp validate` command
- [ ] **T051** - `mcp test` command
- [ ] **T052** - `mcp enable` command
- [ ] **T053** - `mcp disable` command

#### Phase 3.5: TUI Components (M5) - 4 tasks
- [ ] **T054** - MCP Tools panel Bubble Tea model
- [ ] **T055** - Status badge component with Lip Gloss
- [ ] **T056** - Integrate panel into main TUI
- [ ] **T057** - Quick-test action (key binding)

#### Phase 3.6: Integration & Polish (M5) - 7 tasks
- [ ] **T058** - Golden file test: `mcp list` output
- [ ] **T059** - Golden file test: `mcp validate` output
- [ ] **T060** - Golden file test: TUI rendering
- [ ] **T061** - Example configs (github.json, jira.json, aws.json)
- [ ] **T062** - Update `docs/CONNECTORS.md`
- [ ] **T063** - Update `README.md`
- [ ] **T064** - Final validation (all quickstart scenarios)

---

## Immediate Next Steps (Priority Order)

### 1. Evidence Integration (T043-T047) - HIGH PRIORITY
**Why first**: Required before CLI commands can work end-to-end

**Approach**:
1. **T043** - Define AgentInvoker interface
   ```go
   type AgentInvoker interface {
       InvokeTool(agentRole, toolName, method string, args map[string]interface{}) (*types.Evidence, error)
   }
   ```

2. **T044** - Implement orchestration flow:
   ```
   RBAC check → Budget check → Transport invoke → Audit log → Normalize evidence
   ```

3. **T045** - Wire up existing redaction (already implemented in `internal/ai/redactor.go`)

4. **T046** - Wire up existing caching (already implemented in `internal/store/cache.go`)

5. **T047** - Convert MCP response to Evidence entity

**Testing with AWS MCP**:
- Use `~/.sdek/mcp/aws.json` configuration
- Test with read-only AWS commands (e.g., `list-buckets`)
- Verify RBAC, budgets, audit logs, redaction, caching all working

**Estimated Time**: 1-2 days

---

### 2. CLI Commands (T048-T053) - MEDIUM PRIORITY
**Why next**: Provides user-facing interface for MCP operations

**Approach**:
1. **T048** - Create parent `mcp` command with Cobra
   - Add feature flag check: `mcp.enabled`
   - Wire up to Registry singleton

2. **T049-T053** - Implement 5 subcommands (can be parallel)
   - Each command: simple wrapper around Registry methods
   - Format output as tables (use existing UI styles)
   - Support `--format=json` flag

**Testing**:
- Run against AWS MCP configuration
- Test each command manually
- Verify error handling

**Estimated Time**: 1 day

---

### 3. TUI Components (T054-T057) - MEDIUM PRIORITY
**Why next**: Provides visual interface for MCP management

**Approach**:
1. **T054** - Create Bubble Tea model for MCP Tools panel
   - Display tools in table format
   - Show status badges (ready/degraded/offline)
   - Support arrow key navigation

2. **T055** - Create status badge component
   - Green for ready, yellow for degraded, red for offline
   - Use Lip Gloss for styling

3. **T056** - Integrate into main TUI
   - Add "MCP Tools" tab
   - Wire up to Registry
   - Poll status every 5s

4. **T057** - Add 't' key binding for quick-test

**Testing**:
- Launch TUI with AWS MCP tool configured
- Verify real-time status updates
- Test all key bindings

**Estimated Time**: 1 day

---

### 4. Integration & Polish (T058-T064) - LOW PRIORITY
**Why last**: Final validation and documentation

**Approach**:
1. **T058-T060** - Golden file tests
   - Capture CLI output for regression testing
   - Capture TUI screenshots for visual regression

2. **T061** - Create example configs
   - `github.json` - GitHub MCP server
   - `jira.json` - Jira connector example
   - `aws.json` - Copy from `~/.sdek/mcp/aws.json`

3. **T062-T063** - Documentation updates
   - Add MCP section to CONNECTORS.md
   - Add "Using MCP Tools" to README.md

4. **T064** - Final validation
   - Run all 6 quickstart scenarios (AC-01 to AC-06)
   - Verify all acceptance criteria
   - Performance check: <5s latency, 100/s throughput

**Estimated Time**: 1 day

---

## Execution Strategy

### Week 1 (Days 1-2): Evidence Integration
- **Day 1**: T043-T045 (Interface, InvokeTool, Redaction)
- **Day 2**: T046-T047 (Caching, Evidence normalization) + AWS MCP testing

### Week 2 (Days 3-4): CLI & TUI
- **Day 3**: T048-T053 (CLI commands)
- **Day 4**: T054-T057 (TUI components)

### Week 3 (Day 5): Polish & Validation
- **Day 5**: T058-T064 (Golden files, examples, docs, final validation)

**Total Estimated Time**: 5 working days

---

## Testing Strategy

### Unit Tests
- All tests already written (T005-T020) ✅
- Run after each implementation task
- Aim for >80% coverage in `internal/mcp/`

### Integration Tests
- Test with real AWS MCP server
- Verify end-to-end flow: config → registry → RBAC → transport → evidence
- Test circuit breaker and retry logic

### Manual Testing
- Use AWS MCP configuration from `~/.sdek/mcp/aws.json`
- Test CLI commands: `sdek mcp list`, `sdek mcp test aws-api`
- Test TUI: launch and navigate MCP Tools panel

### Golden File Tests
- Capture CLI output for regression
- Compare against fixtures in `tests/golden/`

---

## Risk Mitigation

### Known Risks

1. **File Creation Tool Corruption**
   - **Risk**: File creation tool duplicates content
   - **Mitigation**: Use Python script workaround for files >100 lines
   - **Example**: `generate_registry.py`, `generate_rbac.py`

2. **AWS MCP Server Changes**
   - **Risk**: Server API changes between versions
   - **Mitigation**: Pin to specific version if issues arise
   - **Current**: Using `@latest` for now

3. **Integration with Existing Code**
   - **Risk**: Redaction/caching integration may have breaking changes
   - **Mitigation**: Review `internal/ai/redactor.go` and `internal/store/cache.go` before T045-T046

4. **Performance**
   - **Risk**: MCP invocations may be slow
   - **Mitigation**: Implement caching (T046), circuit breaker already done (T036)

---

## Success Criteria

### Milestone 4 (Evidence Integration) Complete When:
- ✅ AgentInvoker interface defined and implemented
- ✅ End-to-end flow works: RBAC → Budget → Transport → Audit → Evidence
- ✅ Redaction integrated (PII stripped from responses)
- ✅ Caching integrated (duplicate requests served from cache)
- ✅ Can invoke AWS MCP tool and get Evidence entity back
- ✅ Integration test T020 passes

### Milestone 5 (CLI/TUI/Polish) Complete When:
- ✅ All 5 CLI commands work (`list`, `validate`, `test`, `enable`, `disable`)
- ✅ TUI shows MCP Tools panel with real-time status
- ✅ Golden file tests pass
- ✅ Example configs provided for 3 tools (GitHub, Jira, AWS)
- ✅ Documentation updated (CONNECTORS.md, README.md)
- ✅ All 6 quickstart scenarios validated (AC-01 to AC-06)

### Feature Complete When:
- ✅ All 64 tasks complete
- ✅ Test coverage >80% in `internal/mcp/`
- ✅ All acceptance criteria met
- ✅ Performance validated (<5s latency, 100/s throughput)
- ✅ PR ready for review and merge

---

## Implementation Patterns to Follow

### 1. Error Handling
```go
// Use typed errors from internal/mcp/errors.go
if err != nil {
    return nil, mcp.ErrToolNotFound
}
```

### 2. RBAC Checks
```go
// Always check permission before invocation
if !rbac.CheckPermission(agentRole, capability) {
    return nil, mcp.ErrPermissionDenied
}
```

### 3. Budget Enforcement
```go
// Check rate limits and concurrency
if err := rbac.ApplyBudget(toolName); err != nil {
    return nil, err // ErrRateLimited or ErrConcurrencyLimitExceeded
}
```

### 4. Audit Logging
```go
// Always log invocations
log := &types.MCPInvocationLog{
    ToolName: toolName,
    Method: method,
    Timestamp: time.Now(),
    // ...
}
rbac.RecordInvocation(log)
```

### 5. Evidence Normalization
```go
// Convert MCP response to Evidence
evidence := &types.Evidence{
    Source: types.Source{
        Type: "mcp",
        Name: toolName,
        // ...
    },
    // ...
}
```

---

## Commit Strategy

### After Each Task
```bash
git add -A
git commit -m "feat(mcp): implement T043 - AgentInvoker interface

- Define InvokeTool method signature
- Add orchestration flow documentation
- Tests: T018, T020 should start passing"

git push origin 004-mcp-native-agent
```

### After Each Milestone
```bash
git commit -m "feat(mcp): complete M4 - Evidence Integration (T043-T047)

Implemented:
- AgentInvoker interface and orchestration
- Redaction integration with internal/ai
- Caching integration with internal/store
- Evidence normalization from MCP responses

Progress: 47/64 tasks (73%)
Tests: T018, T020 now passing
Verified: AWS MCP tool integration working"
```

---

## Next Session Kickoff

When resuming work, start with:

1. **Review current state**:
   ```bash
   git log --oneline -5
   go test ./tests/unit/mcp_*
   ```

2. **Check AWS MCP config**:
   ```bash
   cat ~/.sdek/mcp/aws.json
   go run test_aws_mcp.go
   ```

3. **Start T043**:
   - Read existing redaction code: `internal/ai/redactor.go`
   - Read existing caching code: `internal/store/cache.go`
   - Define AgentInvoker interface in `internal/mcp/invoker.go`

---

## Questions for Consideration

1. **Should we support multiple agent roles?**
   - Currently: one role per invocation
   - Future: role hierarchy (admin > analyst > viewer)?

2. **How to handle MCP server version updates?**
   - Pin versions in config?
   - Auto-update with compatibility checks?

3. **Should we support custom transports?**
   - Plugin system for third-party transports?
   - Or keep stdio + HTTP only?

4. **Cache invalidation strategy?**
   - TTL-based (current approach)?
   - Event-based (on config change)?
   - Manual (via CLI command)?

---

**Ready to continue with T043-T047 (Evidence Integration)!**
