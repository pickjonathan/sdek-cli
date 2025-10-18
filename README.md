# sdek-cli

**S**ecurity **D**ocumentation **E**vidence **K**it - A CLI and TUI tool for compliance evidence mapping.

## Overview

sdek-cli automates compliance evidence mapping by ingesting data from multiple sources (Git, Jira, Slack, CI/CD, Docs), mapping them to compliance frameworks (SOC2, ISO 27001, PCI DSS), and providing interactive visualization with export capabilities.

## Features

- 🔄 **Multi-source ingestion**: Git commits, Jira tickets, Slack messages, CI/CD pipelines, Documentation
- 📊 **Framework mapping**: SOC 2, ISO 27001, PCI DSS with 124 controls
- 🎯 **Evidence analysis**: Automatic evidence-to-control mapping with confidence scores
- ⚠️ **Risk scoring**: Severity-weighted risk calculation and finding generation
- 📑 **Report export**: JSON compliance reports with role-based filtering
- 🌐 **HTML dashboards**: Interactive web-based compliance visualization
- 🖥️ **Interactive TUI**: Terminal UI for exploring compliance data
- ⚙️ **Configuration management**: Flexible config via files, environment variables, and flags

## Installation

### Prerequisites

- Go 1.23 or later

### Build from source

```bash
git clone https://github.com/pickjonathan/sdek-cli.git
cd sdek-cli
make build
```

Or simply:

```bash
go build -o sdek
```

## Quick Start

### 1. Generate demo data

```bash
./sdek seed --demo
```

This creates:
- 5 data sources (Git, Jira, Slack, CI/CD, Docs)
- ~130 events across all sources
- 3 compliance frameworks with 124 controls
- ~565 evidence mappings
- ~124 findings

### 2. Launch the TUI

```bash
./sdek tui
```

Navigate with:
- `Tab` - Switch between sections
- `↑/↓` - Navigate lists
- `Enter` - Select item
- `q` - Quit

### 3. Analyze evidence (CLI)

```bash
# Ingest from specific source
./sdek ingest --source git --events 50

# Analyze and map evidence
./sdek analyze

# Export compliance report
./sdek report --output ~/compliance-report.json

# Generate interactive HTML dashboard
./sdek html --input ~/compliance-report.json --output ~/dashboard.html
```

## Commands

### `sdek seed`
Generate demo data for testing and development.

```bash
sdek seed --demo [--seed 12345] [--reset]
```

### `sdek ingest`
Ingest events from specific sources.

```bash
sdek ingest --source git --events 30 [--seed 42]
```

Supported sources: `git`, `jira`, `slack`, `cicd`, `docs`

### `sdek analyze`
Map events to controls and calculate risk scores.

```bash
sdek analyze
```

### `sdek report`
Export compliance report to JSON.

```bash
sdek report [--output ~/report.json] [--role manager|engineer]
```

### `sdek html`
Generate an interactive HTML compliance dashboard from a JSON report.

```bash
# Generate HTML from default report location
sdek html

# Specify input and output files
sdek html --input ~/sdek-report.json --output ~/compliance-dashboard.html

# Use short flags
sdek html -i report.json -o dashboard.html
```

The HTML report provides:
- 📊 Visual compliance dashboard with charts and gauges
- 🔍 Interactive framework and control exploration
- 🤖 Filterable evidence with AI enhancement indicators
- ⚠️ Detailed findings analysis with severity indicators
- 📋 Expandable control details with full context
- 🌐 Self-contained file that works offline

### `sdek config`
Manage configuration.

```bash
sdek config init                    # Create default config
sdek config get log.level           # Get config value
sdek config set log.level debug     # Set config value
sdek config list                    # List all settings
sdek config validate                # Validate configuration
```

### `sdek tui`
Launch interactive terminal UI.

```bash
sdek tui [--role manager|engineer]
```

## Configuration

Configuration precedence (highest to lowest):
1. Command-line flags
2. Environment variables (prefix: `SDEK_`)
3. Config file (`~/.sdek/config.yaml`)
4. Default values

### Example config file

```yaml
log:
  level: info

export:
  enabled: true
  path: ~/sdek/reports

data:
  dir: ~/.sdek

frameworks:
  enabled:
    - soc2
    - iso27001
    - pcidss

sources:
  enabled:
    - git
    - jira
    - slack
    - cicd
    - docs

# AI-enhanced evidence analysis (optional)
ai:
  enabled: true
  provider: openai  # openai | anthropic | none
  model: gpt-4-turbo-preview
  max_tokens: 4096
  temperature: 0.3
  timeout: 60
  rate_limit: 10
  # API keys (also via env: SDEK_AI_OPENAI_KEY, SDEK_AI_ANTHROPIC_KEY)
  # openai_key: sk-...
  # anthropic_key: sk-ant-...
```

### AI-Enhanced Evidence Analysis

sdek-cli supports optional AI-powered evidence analysis using OpenAI or Anthropic to enhance compliance control mapping with natural language understanding.

#### Features

- **Multi-provider support**: OpenAI (GPT-4) or Anthropic (Claude 3) with unified abstraction
- **Hybrid confidence scoring**: Weighted average (70% AI + 30% heuristic) for balanced accuracy
- **Privacy-first**: Automatic PII/secret redaction before AI transmission
- **Intelligent caching**: Event-driven cache invalidation reduces redundant API calls
- **Graceful fallback**: Continues with heuristic analysis if AI fails
- **Enhanced reporting**: AI justifications, confidence scores, and residual risk notes

#### Enabling AI Analysis

**Option 1: OpenAI**

```bash
# Set API key
export SDEK_AI_OPENAI_KEY="sk-..."

# Configure provider
./sdek config set ai.provider openai
./sdek config set ai.enabled true
./sdek config set ai.model gpt-4-turbo-preview

# Run analysis with AI
./sdek analyze
```

**Option 2: Anthropic**

```bash
# Set API key
export SDEK_AI_ANTHROPIC_KEY="sk-ant-..."

# Configure provider
./sdek config set ai.provider anthropic
./sdek config set ai.enabled true
./sdek config set ai.model claude-3-opus-20240229

# Run analysis with AI
./sdek analyze
```

**Option 3: Command-line flag**

```bash
# Use specific provider for single run
./sdek analyze --ai-provider openai

# Disable AI for CI/CD environments
./sdek analyze --ai-provider none
# or
./sdek analyze --no-ai
```

#### Switching Providers

```bash
# Start with OpenAI
./sdek config set ai.provider openai
./sdek analyze

# Switch to Anthropic (cache will be invalidated)
./sdek config set ai.provider anthropic
./sdek analyze
```

#### Disabling AI for CI/CD

For continuous integration or offline environments:

```bash
# Disable AI via configuration
./sdek config set ai.enabled false

# Or via flag
./sdek analyze --no-ai

# Or via environment variable
export SDEK_AI_ENABLED=false
./sdek analyze
```

When AI is disabled, sdek-cli uses deterministic heuristic-only analysis, ensuring reproducible results in automated pipelines.

#### AI Configuration Options

| Setting | Default | Description |
|---------|---------|-------------|
| `ai.enabled` | `false` | Master switch for AI analysis |
| `ai.provider` | `none` | AI provider: `openai`, `anthropic`, or `none` |
| `ai.model` | (varies) | Model identifier (e.g., `gpt-4-turbo-preview`, `claude-3-opus-20240229`) |
| `ai.max_tokens` | `4096` | Maximum tokens per request (0-32768) |
| `ai.temperature` | `0.3` | Randomness (0.0-1.0, lower = more deterministic) |
| `ai.timeout` | `60` | Request timeout in seconds (0-300) |
| `ai.rate_limit` | `10` | Maximum requests per minute |

#### Privacy & Security

AI analysis includes automatic redaction of:
- Email addresses (`<EMAIL_REDACTED>`)
- Phone numbers (`<PHONE_REDACTED>`)
- API keys and tokens (`<API_KEY_REDACTED>`)
- Credit card numbers (`<CREDIT_CARD_REDACTED>`)
- Social Security Numbers (`<SSN_REDACTED>`)
- Private keys and passwords

**Original events are never modified** - redaction applies only to AI requests. All PII remains intact in your local state files.

#### Performance & Caching

- **First analysis**: AI calls made for each control (~60s for 124 controls)
- **Subsequent runs**: Cache reuse provides instant results (>70% hit rate)
- **Event changes**: Only affected controls are re-analyzed
- **Provider switching**: Cache invalidated to ensure fresh analysis

Cache stored in: `~/.cache/sdek/ai-cache/`

#### Cost Estimation

Based on typical usage (100 events, 124 controls):

- **OpenAI GPT-4 Turbo**: ~$0.15-0.30 per analysis run
- **Anthropic Claude 3 Opus**: ~$0.20-0.40 per analysis run
- **Cache hit rate >70%**: Subsequent runs cost <$0.10

**Note**: Costs vary based on event count and control complexity.

## Development

### Project Structure

```
sdek-cli/
├── cmd/              # CLI commands (Cobra)
├── internal/         # Internal packages
│   ├── analyze/      # Evidence mapping & risk scoring
│   ├── config/       # Configuration management
│   ├── ingest/       # Data generators
│   ├── report/       # Report export
│   └── store/        # State persistence
├── pkg/types/        # Public types
├── ui/               # TUI implementation (Bubble Tea)
│   ├── components/   # Reusable UI components
│   ├── models/       # Screen models
│   └── styles/       # Lip Gloss styles
└── tests/            # Integration & E2E tests
```

### Build

```bash
make build          # Build binary
make test           # Run tests
make coverage       # Generate coverage report
make clean          # Clean build artifacts
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./cmd/...
```

## Architecture

### Data Flow

```
Sources (Git, Jira, etc.) 
  ↓ ingest
Events (timestamped actions)
  ↓ analyze
Evidence (event-to-control mappings)
  ↓ score
Findings (risk assessments)
  ↓ report
JSON Export
```

### Evidence Mapping

Events are mapped to controls using keyword-based heuristics:

- **Confidence calculation**: Based on keyword matches, event recency, and source type
- **Risk scoring**: Severity-weighted formula (3 High = 1 Critical, 6 Medium = 1 Critical, 12 Low = 1 Critical)
- **Status determination**: Green (low risk), Yellow (medium risk), Red (high risk)

## Technology Stack

- **Language**: Go 1.23+
- **CLI Framework**: Cobra v1.10
- **Config**: Viper v1.21
- **TUI**: Bubble Tea v0.27
- **Styling**: Lip Gloss v0.13
- **Logging**: log/slog (structured JSON)
- **Storage**: JSON file-based state (~/.sdek/state.json)

## Roadmap

- [x] Core CLI commands (seed, ingest, analyze, report, config)
- [x] Command tests
- [x] TUI application structure
- [x] Interactive HTML compliance dashboards
- [ ] Full interactive TUI with Bubble Tea
- [ ] Integration tests
- [ ] Performance optimization (<100ms startup, 60fps TUI)
- [ ] Multi-format export (PDF, Markdown)
- [ ] Real-time data ingestion
- [ ] API endpoints for automation

## Contributing

Contributions are welcome! Please follow the development guidelines in `.github/copilot-instructions.md`.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Project Status

**Current Progress**: 79% complete (45/57 tasks)

This is an active development project implementing the specification in `specs/001-create-sdek/`.

## Contact

- **Author**: Jonathan Pick
- **Repository**: https://github.com/pickjonathan/sdek-cli
- **Issues**: https://github.com/pickjonathan/sdek-cli/issues
