# Implementation Plan: MCP Pluggable Architecture

**Branch**: `006-mcp-pluggable-architecture` | **Date**: 2025-10-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-mcp-pluggable-architecture/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Transform sdek-cli into an MCP-pluggable system following kubectl-ai's architecture patterns. Enable zero-code evidence source addition through MCP server configuration, support multiple AI providers (OpenAI, Anthropic, Gemini, Bedrock, Ollama, llama.cpp), implement unified tool registry with three-tier safety validation, and maintain 100% backward compatibility with Feature 003 connectors. Phase 1 focuses on MCP client mode (consuming external MCP servers); Phase 2 will add MCP server mode (exposing sdek-cli capabilities to external AI clients).

## Technical Context

**Language/Version**: Go 1.23+ (latest stable, matching current codebase)
**Primary Dependencies**:
- Existing: Cobra (CLI), Viper (config), Bubble Tea (TUI)
- New: Native Go MCP client (custom implementation following spec)
- AI SDKs: OpenAI Go SDK, Anthropic SDK, Google AI SDK, AWS SDK (Bedrock)

**Storage**: File-based (JSON state in `~/.sdek/state.json`, YAML config in `~/.sdek/config.yaml`, no database changes)

**Testing**: Go standard testing (`go test`), test coverage target ≥80%, integration tests with real MCP servers (AWS API MCP, filesystem MCP)

**Target Platform**: macOS (primary), Linux (server), cross-platform Go binary

**Project Type**: Single CLI application (existing structure, no new projects)

**Performance Goals**:
- MCP tool discovery: <5 seconds for 10 configured servers
- Parallel evidence collection: 50% faster than sequential (Feature 003 baseline)
- Provider switching: <100ms overhead vs. direct API call
- Tool execution: <60 seconds timeout per tool (configurable)

**Constraints**:
- 100% backward compatibility with Feature 003 (existing configs must work unchanged)
- Zero breaking changes to public API (`pkg/types/`)
- MCP spec v1.0+ compliance (JSON-RPC 2.0 transport)
- Zero external dependencies for MCP client (implement natively in Go)
- Maintain <100MB binary size after MCP integration

**Scale/Scope**:
- Support ≥10 concurrent MCP servers
- Support ≥100 concurrent tool executions across all servers
- Handle 50+ evidence sources per autonomous plan
- 7+ AI providers (OpenAI, Anthropic, Gemini, Bedrock, Vertex AI, Ollama, llama.cpp)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

*Note: The project constitution is a template with placeholder sections. No specific gates are currently defined. The following principles are inferred from existing project patterns:*

**Inferred Principles:**
1. **Backward Compatibility**: ✅ PASS - Feature 003 configs auto-migrate, legacy connector API remains functional
2. **Test Coverage**: ✅ PASS - Target ≥80% coverage with unit + integration tests
3. **Zero Breaking Changes**: ✅ PASS - No changes to `pkg/types/` public API (only additions)
4. **Go Conventions**: ✅ PASS - Follows standard Go patterns (interfaces, factory pattern, init() registration)
5. **Documentation**: ✅ PASS - CLAUDE.md will be updated, user-facing docs in quickstart.md

**Constitution Compliance**: PASSED (no violations)

## Project Structure

### Documentation (this feature)

```text
specs/006-mcp-pluggable-architecture/
├── plan.md              # This file
├── research.md          # Architectural gap analysis, design decisions ✓ Complete
├── data-model.md        # Entity definitions, relationships ✓ Complete
├── quickstart.md        # User getting started guide ✓ Complete
├── contracts/           # API contract JSON schemas ✓ Complete
│   ├── mcp-config-schema.json
│   ├── tool-registry-interface.json
│   └── provider-interface.json
└── tasks.md             # Implementation task breakdown (via /speckit.tasks)
```

### Source Code (repository root)

**Existing Structure** (Feature 003):
```text
sdek-cli/
├── cmd/                    # CLI commands (Cobra)
├── internal/               # Private implementation
│   ├── ai/                # AI provider abstraction (Feature 003)
│   │   ├── engine.go      # MODIFY: Add MCP methods
│   │   ├── provider_factory.go  # REFACTOR: URL scheme-based
│   │   ├── providers/     # EXTEND: Add new providers
│   │   │   ├── openai.go
│   │   │   ├── anthropic.go
│   │   │   └── [NEW: gemini.go, bedrock.go, ollama.go, etc.]
│   │   └── connectors/    # KEEP: Legacy connectors (Feature 003)
│   │       ├── connector.go
│   │       └── registry.go
│   ├── analyze/           # Analysis engine
│   ├── config/            # Config loading (Viper)
│   ├── policy/            # Policy excerpt loading
│   └── store/             # State persistence
├── pkg/types/             # Public types (no breaking changes)
│   ├── config.go          # EXTEND: Add MCPConfig, ProviderConfig
│   ├── bundle.go
│   ├── context.go
│   └── plan.go
└── tests/
    ├── integration/
    └── unit/
```

**New Structure** (Feature 006 additions):
```text
sdek-cli/
├── internal/
│   ├── ai/
│   │   ├── factory/        # NEW: Provider factory with registration
│   │   │   ├── registry.go
│   │   │   └── factory.go
│   │   ├── providers/      # EXTEND: Add new providers
│   │   │   ├── gemini.go   # NEW
│   │   │   ├── bedrock.go  # NEW
│   │   │   ├── ollama.go   # NEW
│   │   │   ├── llamacpp.go # NEW
│   │   │   └── vertexai.go # NEW
│   │   └── session.go      # NEW: ChatSession implementation
│   │
│   ├── mcp/                # NEW: MCP client implementation
│   │   ├── manager.go      # MCP server orchestration
│   │   ├── client.go       # MCP client interface
│   │   ├── stdio_client.go # Stdio transport
│   │   ├── http_client.go  # HTTP transport
│   │   ├── transport.go    # Transport abstraction
│   │   ├── adapter.go      # Legacy connector → MCP adapter
│   │   └── jsonrpc.go      # JSON-RPC 2.0 implementation
│   │
│   ├── tools/              # NEW: Unified tool system
│   │   ├── registry.go     # Tool catalog
│   │   ├── tool.go         # Tool interface
│   │   ├── safety.go       # Three-tier safety validation
│   │   ├── executor.go     # Parallel tool execution
│   │   └── audit.go        # Audit logging
│   │
│   └── migration/          # NEW: Config migration
│       └── legacy.go       # Feature 003 → Feature 006 migration
│
└── pkg/types/
    ├── mcp.go              # NEW: MCP types (MCPConfig, MCPServerConfig, etc.)
    ├── provider.go         # NEW: Provider types (ProviderConfig, ChatSession, etc.)
    └── tool.go             # NEW: Tool types (Tool, ToolCall, ToolCallAnalysis, etc.)
```

**Structure Decision**: Single Go CLI application (Option 1 pattern). No new projects or major restructuring. New packages added under `internal/` for MCP, factory, and tool systems. Existing packages extended (AI providers, config types). Legacy code preserved with adapters for backward compatibility.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*No constitution violations detected. This section is not applicable.*

---

## Phase 0: Research & Architecture (Complete)

**Status**: ✅ COMPLETE

**Deliverables**:
- [x] kubectl-ai architecture research
- [x] MCP protocol analysis
- [x] Current sdek-cli architecture review
- [x] Architectural gap analysis
- [x] Design decisions documented
- [x] Migration strategy defined
- [x] Risk assessment completed

**Output**: `research.md` (complete)

---

## Phase 1: Design & Contracts (Complete)

**Status**: ✅ COMPLETE

**Deliverables**:
- [x] Data model entities defined (23 core entities)
- [x] API contracts created (JSON schemas for MCP config, tool registry, provider interface)
- [x] Entity relationships documented
- [x] Validation rules specified
- [x] User quickstart guide written

**Outputs**:
- `data-model.md` (complete)
- `contracts/mcp-config-schema.json` (complete)
- `contracts/tool-registry-interface.json` (complete)
- `contracts/provider-interface.json` (complete)
- `quickstart.md` (complete)

**Agent Context Update**: Run `.specify/scripts/bash/update-agent-context.sh claude` to update CLAUDE.md with new architecture details.

---

## Phase 2: Implementation Planning (Next)

**Status**: ⏭️ NEXT PHASE

**Command**: `/speckit.tasks`

This will generate `tasks.md` with:
- Dependency-ordered implementation tasks
- Test requirements per task
- Acceptance criteria
- Estimated complexity

**Implementation Order** (based on research decisions):
1. **Week 1-2**: AI provider factory refactoring
   - Extract registration system
   - Implement URL scheme parsing
   - Migrate OpenAI and Anthropic to factory pattern
   - Add new providers (Gemini, Bedrock, Ollama, llama.cpp)

2. **Week 3-4**: MCP client implementation
   - JSON-RPC 2.0 protocol
   - Stdio and HTTP transports
   - Tool discovery and execution
   - MCPManager orchestration

3. **Week 5**: Tool registry & safety
   - Unified tool catalog
   - Three-tier safety validation
   - Parallel execution engine
   - Audit logging

4. **Week 6**: Integration & migration
   - Legacy connector adapter
   - Config auto-migration
   - CLI commands (`sdek mcp list-servers`, `sdek mcp test`, etc.)
   - End-to-end testing

5. **Week 7**: Testing & validation
   - Unit tests (≥80% coverage)
   - Integration tests with real MCP servers
   - Performance benchmarks
   - Backward compatibility validation

6. **Week 8**: Documentation & release
   - Update CLAUDE.md
   - User migration guide
   - Release notes
   - Demo workflows

---

## Key Milestones

| Milestone | Deliverable | Target | Status |
|-----------|------------|--------|--------|
| Research Complete | research.md | Week 1 | ✅ DONE |
| Design Complete | data-model.md, contracts/ | Week 1 | ✅ DONE |
| Quickstart Written | quickstart.md | Week 1 | ✅ DONE |
| Tasks Generated | tasks.md | Week 2 | ⏭️ NEXT |
| Provider Factory | URL-based provider selection working | Week 3 | 🔜 Pending |
| MCP Client | AWS API MCP integration working | Week 5 | 🔜 Pending |
| Tool System | Safety validation + parallel execution | Week 6 | 🔜 Pending |
| Migration | Legacy configs auto-migrate | Week 6 | 🔜 Pending |
| Testing | ≥80% test coverage achieved | Week 7 | 🔜 Pending |
| Release | Feature 006 merged to main | Week 8 | 🔜 Pending |

---

## Success Criteria

**Must Have (Phase 1 - MCP Client Mode):**
- [x] Feature spec approved and documented
- [x] Research complete with design decisions
- [x] Data model with 23 core entities defined
- [x] API contracts (JSON schemas) created
- [x] User quickstart guide written
- [ ] 100% backward compatibility with Feature 003 configs
- [ ] 3+ MCP servers integrated without code changes (AWS API, filesystem, one community server)
- [ ] 7+ AI providers supported (OpenAI, Anthropic, Gemini, Bedrock, Vertex AI, Ollama, llama.cpp)
- [ ] 50% faster parallel evidence collection vs. sequential baseline
- [ ] Zero code deployments for new MCP servers (config-only)
- [ ] ≥80% test coverage (unit + integration)
- [ ] CLI commands: `sdek mcp list-servers`, `sdek mcp list-tools`, `sdek mcp test`, `sdek ai health`

**Nice to Have (Future - Phase 2 MCP Server Mode):**
- [ ] MCP server mode (expose sdek-cli to external AI clients)
- [ ] Dual-mode operation (client + server simultaneously)
- [ ] HTTP endpoint with authentication
- [ ] Session management for external clients

---

## Risk Mitigation

**High Risk: MCP Server Ecosystem Immaturity**
- Mitigation: Maintain backward compat with legacy connectors, provide adapter pattern
- Fallback: Legacy connector API remains fully functional

**Medium Risk: Performance Regression**
- Mitigation: Benchmark before/after, optimize JSON serialization, pool subprocess connections
- Target: 50% faster parallel execution offsets any per-tool overhead

**Low Risk: Configuration Complexity**
- Mitigation: Auto-migrate legacy configs, provide `sdek mcp init` wizard, include templates
- Validation: JSON schema validation with helpful error messages

---

## Open Questions for Implementation Phase

1. **MCP Server Priority**: Which MCP servers beyond AWS to integrate first?
   - Proposed: GitHub MCP (code evidence), Jira MCP (ticket evidence), filesystem MCP (local files)

2. **Provider Testing**: How to test 7 AI providers without incurring high costs?
   - Proposed: Use provider free tiers, mock expensive models, limit integration tests to OpenAI + Ollama

3. **Deprecation Timeline**: When to remove legacy connector API entirely?
   - Proposed: v2.0.0 (12 months from Feature 006 release), with clear deprecation warnings

4. **MCP Server Distribution**: Should sdek-cli bundle MCP servers or require manual install?
   - Proposed: Document manual installation (npm, pip, uvx), provide optional installer: `sdek mcp install aws-api`

---

## Next Steps

1. **Run** `/speckit.tasks` to generate implementation task breakdown
2. **Review** tasks.md with team for complexity estimates
3. **Assign** tasks to development sprints
4. **Begin** implementation starting with AI provider factory refactoring
5. **Iterate** with weekly reviews and architecture adjustments as needed

---

**Plan Status**: Phase 1 Complete, Ready for Task Generation
**Last Updated**: 2025-10-26
