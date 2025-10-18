# sdek Command Reference

Complete reference for all `sdek` commands.

## Global Flags

These flags are available for all commands:

- `--config string` - Config file path (default: `$HOME/.sdek/config.yaml`)
- `--data-dir string` - Data directory path (default: `$HOME/.sdek`)
- `--log-level string` - Log level: debug, info, warn, error (default: `info`)
- `-v, --verbose` - Enable verbose output
- `-h, --help` - Show help for any command

## Commands

### sdek tui

Launch the interactive terminal UI for visual compliance management.

**Usage:**
```bash
sdek tui
```

**Description:**
The TUI provides an interactive dashboard with:
- Home screen with compliance overview
- Sources management view
- Controls browsing with filtering
- Evidence listing grouped by confidence level
- Frameworks comparison

**Keyboard Shortcuts:**
- `1-4` or `←/→` (h/l) - Navigate between views
- `/` - Enter search mode
- `r` - Refresh data from disk
- `e` - Export current view
- `q` or `Ctrl+C` - Quit
- `?` - Show help

---

### sdek seed

Generate demo data for testing and development.

**Usage:**
```bash
sdek seed [flags]
```

**Flags:**
- `--demo` - Generate demo data with all sources
- `--sources int` - Number of sources to generate (default: 5)
- `--events int` - Number of events per source (default: 100)
- `--frameworks string` - Comma-separated framework IDs (default: "soc2,iso27001,pcidss")

**Examples:**
```bash
# Generate default demo data
sdek seed --demo

# Custom data generation
sdek seed --sources 3 --events 50 --frameworks soc2,iso27001

# Generate only SOC 2 data
sdek seed --frameworks soc2 --events 200
```

**Description:**
Creates sample compliance data including:
- Data sources (Git, Jira, Slack, CI/CD, Docs)
- Events from each source type
- Framework definitions and controls
- Initial state file at `$HOME/.sdek/state.json`

---

### sdek ingest

Ingest events from a specific data source.

**Usage:**
```bash
sdek ingest [flags]
```

**Flags:**
- `--source string` - Source type: git, jira, slack, cicd, docs
- `--events int` - Number of events to ingest (default: 50)
- `--source-id string` - Existing source ID to ingest into

**Examples:**
```bash
# Ingest 100 Git commits
sdek ingest --source git --events 100

# Ingest Jira tickets into existing source
sdek ingest --source jira --source-id src-123 --events 30

# Ingest Slack messages
sdek ingest --source slack --events 200
```

**Description:**
Generates events from specified source types and adds them to the state.  
Each source type produces realistic sample data:
- **git**: Commits with messages, authors, timestamps
- **jira**: Tickets with status, priority, assignees
- **slack**: Messages with channels, threads, reactions
- **cicd**: Pipeline runs, deployments, test results
- **docs**: Documentation changes, reviews, approvals

---

### sdek analyze

Analyze events and map them to compliance controls.

**Usage:**
```bash
sdek analyze [flags]
```

**Flags:**
- `--framework string` - Analyze specific framework (default: all)
- `--confidence string` - Minimum confidence level: low, medium, high (default: low)
- `--ai-provider string` - AI provider for enhanced analysis: openai, anthropic, none (default: from config)
- `--no-ai` - Disable AI analysis (use heuristics only)

**Examples:**
```bash
# Analyze all frameworks
sdek analyze

# Analyze only SOC 2 controls
sdek analyze --framework soc2

# Only map high-confidence evidence
sdek analyze --confidence high

# Use AI-enhanced analysis with OpenAI
sdek analyze --ai-provider openai

# Use Anthropic for this analysis
sdek analyze --ai-provider anthropic

# Disable AI for CI/CD environments (heuristics only)
sdek analyze --no-ai
```

**Description:**
Performs compliance evidence mapping:
1. Maps events to framework controls using keyword matching
2. (Optional) Enhances mapping with AI natural language understanding
3. Calculates hybrid confidence scores (70% AI + 30% heuristic when AI enabled)
4. Generates compliance findings based on evidence gaps
5. Updates risk status for each control (green/yellow/red)
6. Calculates framework compliance percentages

**Analysis Process:**
- **Keyword Matching**: Events matched to controls via predefined keywords (heuristic confidence)
- **AI Enhancement** (when enabled): Natural language analysis with justifications and residual risk notes
- **Confidence Scoring**: 
  - Heuristic-only: Based on keyword relevance and event metadata
  - AI-enhanced: Weighted average (70% AI + 30% heuristic)
- **Risk Calculation**: Formula: 3 High = 1 Critical, 6 Medium = 1 Critical, 12 Low = 1 Critical
- **Compliance**: Percentage of controls with green risk status

**AI-Enhanced Analysis:**

When AI is enabled (via config or `--ai-provider` flag), evidence mapping is enhanced with:
- **Natural language understanding**: AI interprets event context beyond keyword matching
- **Structured justifications**: Explains why each event supports the control
- **Confidence scores**: 0-100 scale based on semantic relevance
- **Residual risk notes**: Identifies gaps or concerns not addressed by evidence
- **Privacy protection**: Automatic PII/secret redaction before AI transmission
- **Intelligent caching**: Reuses previous AI analysis for unchanged events

AI providers:
- **OpenAI**: GPT-4 Turbo (`gpt-4-turbo-preview`) - General-purpose, fast
- **Anthropic**: Claude 3 Opus (`claude-3-opus-20240229`) - Long context, detailed
- **none**: Disable AI, use heuristics only (default, reproducible)

---

### sdek report

Export compliance report to JSON or YAML format.

**Usage:**
```bash
sdek report [flags]
```

**Flags:**
- `--output string` - Output file path (default: `./compliance-report.json`)
- `--format string` - Output format: json, yaml (default: json)
- `--role string` - Filter by user role: compliance-manager, engineer
- `--framework string` - Include specific framework only

**Examples:**
```bash
# Export JSON report
sdek report --output ~/reports/compliance.json

# Export YAML report for compliance managers
sdek report --format yaml --role compliance-manager --output report.yaml

# Export only SOC 2 framework
sdek report --framework soc2 --output soc2-report.json
```

**Description:**
Generates comprehensive compliance reports including:
- **Metadata**: Generation timestamp, version
- **Summary**: Counts for all entities, compliance percentages
- **Frameworks**: Detailed breakdown by framework
- **Controls**: Per-control evidence and findings
- **Evidence**: All mapped evidence items
- **Findings**: Risk findings and recommendations

**Report Structure:**
```json
{
  "metadata": {
    "generated_at": "2025-10-16T...",
    "version": "1.0"
  },
  "summary": {
    "total_sources": 5,
    "total_events": 500,
    "total_evidence": 150,
    "overall_compliance_percentage": 75.5
  },
  "frameworks": [...]
}
```

---

### sdek html

Generate an interactive HTML compliance dashboard from a JSON report.

**Usage:**
```bash
sdek html [flags]
```

**Flags:**
- `-i, --input string` - Input JSON report file (default: `~/sdek-report.json`)
- `-o, --output string` - Output HTML file (default: `~/sdek-report.html`)

**Examples:**
```bash
# Generate HTML from default report location
sdek html

# Specify input and output files
sdek html --input ~/sdek-report.json --output ~/compliance-dashboard.html

# Use short flags
sdek html -i report.json -o dashboard.html
```

**Description:**
Creates a standalone, interactive HTML dashboard from a JSON compliance report. The HTML file includes:
- **Visual Dashboard**: Summary cards with key metrics and compliance scores
- **Interactive Tabs**: Navigate between Overview, Frameworks, Findings, and Evidence
- **Framework View**: Expandable framework sections with risk breakdown
- **Control Cards**: Color-coded risk indicators with evidence and finding counts
- **Evidence Filtering**: Filter by AI-enhanced vs heuristic evidence
- **AI Indicators**: Special badges and styling for AI-analyzed evidence
- **Finding Details**: Severity-colored findings with recommendations
- **Modal Details**: Click controls for detailed evidence inspection
- **Search**: Real-time search across controls and evidence
- **Responsive Design**: Works on desktop, tablet, and mobile devices

**HTML Report Features:**
- 📊 **Self-contained**: All CSS and JavaScript embedded, no external dependencies
- 🌐 **Offline Ready**: Works without internet connection
- 📤 **Shareable**: Can be emailed or hosted on any web server
- 🔒 **Audit Trail**: Perfect for compliance audits and stakeholder reviews
- 🎨 **Modern Design**: Purple gradient theme with smooth animations
- 🤖 **AI Highlighting**: Green borders for AI-enhanced evidence

**File Structure:**
The output HTML file embeds the entire JSON report and renders it dynamically using JavaScript. File size is typically ~800KB for a full report with AI analysis.

**Browser Compatibility:**
- ✅ Chrome/Edge 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Mobile browsers (iOS Safari, Chrome Mobile)

---

### sdek config

Manage sdek configuration settings.

**Usage:**
```bash
sdek config [command] [flags]
```

**Subcommands:**
- `get <key>` - Get configuration value
- `set <key> <value>` - Set configuration value
- `reset` - Reset to default configuration
- `show` - Show all configuration

**Configuration Keys:**
- `export.enabled` - Enable/disable report export
- `export.format` - Default export format (json/yaml)
- `analysis.min_confidence` - Minimum confidence level
- `tui.theme` - TUI color theme
- `tui.refresh_interval` - Auto-refresh interval (seconds)

**Examples:**
```bash
# View all configuration
sdek config show

# Get specific value
sdek config get export.format

# Set configuration
sdek config set analysis.min_confidence medium
sdek config set export.format yaml

# Reset to defaults
sdek config reset
```

**Description:**
Configuration is stored in `$HOME/.sdek/config.yaml`. Changes persist across sessions.

---

### sdek version

Print version information.

**Usage:**
```bash
sdek version
```

**Example Output:**
```
sdek version v1.0.0
```

**Description:**
Displays the current version of sdek, including:
- Version tag (from git tags)
- Git commit hash
- Build date

---

### sdek completion

Generate shell completion scripts.

**Usage:**
```bash
sdek completion [bash|zsh|fish|powershell]
```

**Examples:**
```bash
# Bash
sdek completion bash > /etc/bash_completion.d/sdek

# Zsh  
sdek completion zsh > "${fpath[1]}/_sdek"

# Fish
sdek completion fish > ~/.config/fish/completions/sdek.fish

# PowerShell
sdek completion powershell > sdek.ps1
```

**Description:**
Generates shell-specific completion scripts for command and flag autocompletion.

---

## Workflow Examples

### Quick Start - Demo Data
```bash
# 1. Generate demo data
sdek seed --demo

# 2. Run analysis
sdek analyze

# 3. Launch TUI
sdek tui

# 4. Export report
sdek report --output compliance-report.json
```

### Custom Compliance Assessment
```bash
# 1. Start fresh
rm -rf ~/.sdek

# 2. Generate custom data
sdek seed --sources 3 --events 200 --frameworks soc2

# 3. Ingest additional Git commits
sdek ingest --source git --events 100

# 4. Analyze with high confidence only
sdek analyze --confidence high

# 5. Export SOC 2 report for compliance team
sdek report --framework soc2 --role compliance-manager --output soc2.json
```

### Development Workflow
```bash
# Generate test data
sdek seed --events 50

# Iterate: modify code, then re-analyze
sdek analyze

# Check in TUI
sdek tui

# Run analysis again
sdek analyze --framework soc2

# Export and review
sdek report --output test-report.json
```

---

## AI Commands (Feature 003: Context Injection)

The `sdek ai` command group provides AI-powered compliance analysis with policy context injection. These commands offer deeper, policy-grounded insights compared to the standard event-to-control mapping workflow.

### sdek ai analyze

Perform AI-enhanced compliance analysis by injecting policy context (framework excerpts, control descriptions) into the AI prompt.

**Usage:**
```bash
sdek ai analyze --framework <framework> --section <section> \
    --excerpts-file <path> --evidence-path <glob> [flags]
```

**Required Flags:**
- `--framework string` - Framework name (e.g., SOC2, ISO27001, PCI-DSS)
- `--section string` - Section ID (e.g., CC6.1, A.9.4.2)
- `--excerpts-file string` - Path to policy excerpts JSON file
- `--evidence-path strings` - Evidence file paths (supports globs, can be specified multiple times)

**Optional Flags:**
- `--no-cache` - Bypass cache and perform fresh analysis
- `--output string` - Output file for finding results (default: `findings.json`)

**Examples:**
```bash
# Basic context mode analysis for SOC2 CC6.1 (Access Controls)
sdek ai analyze --framework SOC2 --section CC6.1 \
    --excerpts-file ./policies/soc2_excerpts.json \
    --evidence-path ./evidence/github_*.json \
    --evidence-path ./evidence/jira_*.json

# Analyze ISO 27001 section with single evidence source
sdek ai analyze --framework ISO27001 --section A.9.4.2 \
    --excerpts-file ./policies/iso_excerpts.json \
    --evidence-path ./evidence/audit_logs.json

# Bypass cache for fresh analysis (useful for testing policy changes)
sdek ai analyze --framework SOC2 --section CC6.1 \
    --excerpts-file ./policies/soc2_excerpts.json \
    --evidence-path ./evidence/*.json \
    --no-cache

# Multiple evidence paths from different sources
sdek ai analyze --framework PCI-DSS --section 8.2.4 \
    --excerpts-file ./policies/pci_excerpts.json \
    --evidence-path ./evidence/github/*.json \
    --evidence-path ./evidence/jira/*.json \
    --evidence-path ./evidence/slack/*.json

# Specify custom output file for finding results
sdek ai analyze --framework ISO27001 --section A.9.4.2 \
    --excerpts-file ./policies/iso_excerpts.json \
    --evidence-path ./evidence/*.json \
    --output ./findings/iso_a942_finding.json
```

**Description:**

Context injection mode provides policy-grounded AI analysis that goes beyond generic event-to-control mapping. Key features:

- **Context Injection**: Policy excerpts and control descriptions guide AI analysis for more accurate, policy-aligned findings
- **PII/Secret Redaction**: Automatic redaction of sensitive data before sending to AI provider
- **Response Caching**: Reuses previous AI analysis for identical context/evidence combinations
- **Confidence Scoring**: 0-100 scale with automatic low-confidence flagging
- **Detailed Findings**: Structured findings with citations, justifications, and residual risk assessment

**Differences from `sdek analyze --ai`:**

| Feature | `sdek analyze --ai` | `sdek ai analyze` |
|---------|---------------------|-------------------|
| **Purpose** | Event-to-control mapping enhancement | Policy-grounded compliance analysis |
| **Context** | Generic control descriptions | Detailed policy excerpts |
| **Workflow** | Enhances existing mapping | Specialized analysis mode |
| **Output** | Enhanced evidence items | Detailed findings with citations |
| **Use Case** | Quick analysis across all controls | Deep dive on specific policy sections |

**Policy Excerpts File Format:**

The excerpts file should contain policy text relevant to the framework and section:

```json
[
  {
    "framework": "SOC2",
    "version": "2017",
    "section": "CC6.1",
    "text": "The entity implements logical access security software...",
    "related_sections": ["CC6.2", "CC6.3"]
  }
]
```

**Configuration:**

Confidence thresholds and redaction patterns are configured in `config.yaml`:

```yaml
ai:
  context_injection:
    confidence_threshold: 70  # Flag findings below this score
    auto_approve_rules:
      - "github:repo:myorg/security-*"
      - "jira:project:INFOSEC"
```

---

## State Management

sdek stores all data in `$HOME/.sdek/state.json`. This file contains:
- Sources and their configurations
- All ingested events
- Framework definitions and controls
- Mapped evidence items
- Generated findings
- Configuration settings

**State Operations:**
```bash
# Backup state
cp ~/.sdek/state.json ~/.sdek/state.backup.json

# Reset state (warning: deletes all data)
rm ~/.sdek/state.json
sdek seed --demo

# View state size
du -h ~/.sdek/state.json
```

---

## Performance

**Benchmarks** (target performance):
- Cold start: <100ms
- TUI rendering: 60fps
- Analysis (1000 events): <2s
- Report export (1000 events): <500ms

**Optimization Tips:**
- Use `--confidence medium` or `high` to reduce analysis time
- Limit events with `--events` flag during ingestion
- Export smaller reports with `--framework` filter

---

## Troubleshooting

### State File Corrupted
```bash
rm ~/.sdek/state.json
sdek seed --demo
```

### TUI Not Rendering Correctly
```bash
# Check terminal compatibility
echo $TERM

# Use verbose mode for debugging
sdek tui --verbose --log-level debug
```

### Analysis Takes Too Long
```bash
# Analyze specific framework
sdek analyze --framework soc2

# Use higher confidence threshold
sdek analyze --confidence high
```

### Cannot Find Command
```bash
# Check installation
which sdek

# Reinstall
make install
```

---

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - State file error
- `4` - Analysis error
- `5` - Export error

---

## See Also

- [README.md](../README.md) - Project overview and installation
- [Quickstart Guide](../specs/001-create-sdek/quickstart.md) - Step-by-step tutorial
- [Data Model](../specs/001-create-sdek/data-model.md) - Entity specifications
- [CLI Contracts](../specs/001-create-sdek/contracts/cli-commands.md) - Command specifications
