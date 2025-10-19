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
```

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
