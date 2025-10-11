# Research: Create sdek

**Feature**: 001-create-sdek  
**Date**: 2025-10-11

## Technology Stack Decisions

### Decision: Go 1.23+ for CLI Development
**Rationale**: 
- Fast compilation and single-binary distribution
- Excellent cross-platform support (Linux, macOS, Windows)
- Strong standard library for file I/O and JSON handling
- Native support for concurrent operations (goroutines) for future phases
- Active ecosystem for CLI tools (Cobra, Viper widely adopted)

**Alternatives Considered**:
- **Rust**: Excellent performance but steeper learning curve, smaller CLI ecosystem
- **Python**: Easier prototyping but slower startup time, binary distribution challenges
- **TypeScript/Node**: Good for rapid development but larger runtime footprint, slower startup

### Decision: Cobra for Command Structure
**Rationale**:
- Industry-standard CLI framework in Go ecosystem (used by kubectl, hugo, gh)
- Built-in support for subcommands, flags, and help generation
- PreRun/PostRun hooks align with constitution validation requirements
- Automatic completion generation for bash/zsh/fish
- Excellent documentation and examples

**Alternatives Considered**:
- **urfave/cli**: Simpler but less feature-rich, no command tree structure
- **Standard flag package**: Too low-level, would require manual help/completion
- **pflag**: Good flags but no command structure

### Decision: Viper for Configuration Management
**Rationale**:
- De facto standard for Go configuration (pairs naturally with Cobra)
- Supports all required precedence levels (flags → env → file → defaults)
- Built-in support for YAML, JSON, TOML
- Watch configuration files for changes
- Type-safe access to config values

**Alternatives Considered**:
- **envconfig**: Environment variables only, no file or flag support
- **Manual implementation**: High maintenance burden, error-prone

### Decision: Bubble Tea for Terminal UI
**Rationale**:
- Modern, composable TUI framework based on Elm architecture
- Excellent performance and responsiveness (60fps capable)
- Built-in support for keyboard navigation and terminal resizing
- Strong ecosystem (Lip Gloss for styling, Bubbles for components)
- Active development and good documentation

**Alternatives Considered**:
- **tview**: More batteries-included but less flexible, harder to customize
- **termui**: Dashboard-focused, not ideal for interactive navigation
- **gocui**: Lower-level, more manual work for layouts

### Decision: Lip Gloss + Bubbles for UI Components
**Rationale**:
- Lip Gloss provides CSS-like styling (colors, borders, padding, alignment)
- Bubbles offers pre-built components (lists, spinners, text inputs)
- Same authors as Bubble Tea, guaranteed compatibility
- Declarative styling approach easier to maintain than manual ANSI codes

**Alternatives Considered**:
- **Manual ANSI codes**: Error-prone, hard to maintain, poor cross-platform
- **fatih/color**: Good for simple coloring but no layout capabilities

### Decision: log/slog for Structured Logging
**Rationale**:
- Built into Go 1.21+ standard library (no external dependency)
- Structured logging with key-value pairs
- Multiple output formats (JSON, text)
- Log level filtering (debug, info, warn, error)
- Zero-allocation in hot paths

**Alternatives Considered**:
- **zerolog**: Excellent performance but external dependency
- **zap**: Very fast but more complex API
- **logrus**: Popular but slower, less actively maintained

## Data Model & State Management

### Decision: Local JSON for State Persistence
**Rationale**:
- Simple, human-readable format for debugging
- Go standard library encoding/json is robust and fast
- Easy to inspect/modify for development and troubleshooting
- Sufficient performance for 10,000 evidence items
- No database setup required for local-only phase

**Alternatives Considered**:
- **SQLite**: Overkill for single-user local storage, adds dependency
- **BoltDB/BadgerDB**: Good for key-value but JSON simpler for structured data
- **YAML**: Human-friendly but slower parsing, no advantage over JSON here

### Decision: In-Memory Cache with Auto-Save
**Rationale**:
- Fast reads for TUI responsiveness
- Write-through caching on command completion
- Prevents data loss from crashes (save after each command)
- Simple implementation with sync.RWMutex for safe concurrent access

**Pattern**: Load on startup → Modify in memory → Save after command/action

### Decision: Deterministic Seeding for Simulated Data
**Rationale**:
- Reproducible demos and testing
- Use fixed seed for default data, configurable seed for variation
- Ensures framework guarantees (each has green/yellow/red controls)
- Simplifies golden file tests (consistent output)

**Implementation**: `math/rand` with configurable seed, generate 10-50 events per source with uniform distribution

## Evidence Mapping Strategy

### Decision: Rule-Based Heuristics for Mapping
**Rationale**:
- No AI inference required (deterministic, testable)
- Pattern matching on event content (keywords, metadata)
- Score-based confidence assignment (Low <30%, Medium 30-70%, High >70%)
- Framework-specific rule sets for SOC2/ISO/PCI

**Pattern**:
```
Event → Extract keywords → Match against control keywords → Calculate score → Assign confidence tier
```

**Example Rules**:
- Git commits with "security" keyword → SOC2 CC6.1 (Logical and Physical Access Controls)
- Jira tickets with "bug" + "critical" → ISO 27001 A.12.6.1 (Technical Vulnerability Management)
- CI/CD events with "test" + "passed" → PCI DSS 6.3.2 (Code Review)

### Decision: Severity-Based Risk Scoring
**Rationale**:
- Weighted scoring system (3 high = 1 critical, 6 medium = 1 critical, 12 low = 1 critical)
- Prevents green status with any critical-equivalent issues
- Clear visual risk communication (green/yellow/red)
- Aligns with compliance audit priorities

**Implementation**: Aggregate issues by severity → Calculate critical-equivalents → Assign color

## Testing Strategy

### Decision: Three-Tier Testing Approach
**Rationale**:
- **Unit tests**: Fast feedback, test business logic in isolation
- **Integration tests**: Validate command flows (ingest → analyze → report)
- **Golden file tests**: Catch TUI rendering regressions

**Test Coverage Goals**:
- Unit: 80%+ coverage for internal packages
- Integration: All CLI command workflows
- Golden: All major TUI screens (home, sources, frameworks, controls, evidence)

### Decision: Table-Driven Tests for Data Generation
**Rationale**:
- Test multiple scenarios with single test function
- Easy to add new test cases
- Clear documentation of expected behavior

**Pattern**:
```go
tests := []struct {
    name       string
    source     string
    wantEvents int
    wantMin    int
    wantMax    int
}{
    {"Git source", "git", 25, 10, 50},
    {"Jira source", "jira", 30, 10, 50},
    // ...
}
```

### Decision: Golden File Testing with go-golden
**Rationale**:
- Capture TUI output as text snapshots
- Compare against approved baselines
- Update mode for intentional changes
- Version control golden files for review

**Tool**: `github.com/sebdah/goldie` or custom implementation

## Performance Considerations

### Decision: Lazy Loading for TUI Data
**Rationale**:
- Load only visible data initially
- Paginate large lists (sources, controls, evidence)
- Render only viewport contents (virtual scrolling)
- Keeps memory footprint low for large datasets

### Decision: Pre-Calculate Risk Scores on State Load
**Rationale**:
- Compute scores once on load, cache in memory
- Avoid recalculating on every render
- Invalidate and recompute only on data changes (ingest, analyze)

## Cross-Platform Considerations

### Decision: Detect Terminal Capabilities at Runtime
**Rationale**:
- Check terminal size on startup and resize events
- Detect color support (256-color, true color, none)
- Graceful degradation on limited terminals
- Platform-specific path handling via filepath package

**Fallback Strategy**:
- Minimum 80×24 size check with clear error
- ASCII-only mode for terminals without color
- Non-interactive mode for pipe/redirect detection

## Configuration Schema

### Decision: Hierarchical YAML Configuration
**Rationale**:
- Human-readable and editable
- Supports nested structures (frameworks, sources, etc.)
- Comments preserved by Viper where possible

**Schema**:
```yaml
# $HOME/.sdek/config.yaml
data_dir: "$HOME/.sdek"
log_level: "info"
theme: "dark"  # dark, light, custom
user_role: "compliance_manager"  # compliance_manager, engineer

export:
  default_path: "$HOME/sdek/reports"
  format: "json"

frameworks:
  enabled:
    - soc2
    - iso27001
    - pci_dss

sources:
  enabled:
    - git
    - jira
    - slack
    - cicd
    - docs
```

## Unknowns Resolved

All technical context items have been clarified:
- ✅ Language/Version: Go 1.23+
- ✅ Dependencies: Cobra, Viper, Bubble Tea, Lip Gloss, Bubbles, log/slog
- ✅ Storage: Local JSON files
- ✅ Testing: Go test, golden files, integration tests
- ✅ Performance: <100ms startup, 60fps TUI, 10K items
- ✅ Constraints: Local-only, no network, deterministic
- ✅ Project structure: Single project with cmd/internal/ui/pkg layout

No NEEDS CLARIFICATION items remain.
