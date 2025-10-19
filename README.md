# sdek-cli

**S**ecurity **D**ocumentation **E**vidence **K**it - A CLI and TUI tool for compliance evidence mapping.

## Overview

sdek-cli automates compliance evidence mapping by ingesting data from multiple sources (Git, Jira, Slack, CI/CD, Docs), mapping them to compliance frameworks (SOC2, ISO 27001, PCI DSS), and providing interactive visualization with export capabilities.

## AI-Powered Compliance Analysis Workflow

The `sdek ai` commands provide intelligent, context-aware compliance analysis by injecting policy requirements directly into AI prompts. This approach delivers more accurate, policy-grounded findings compared to generic AI analysis.

### Workflow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SDEK AI ANALYSIS WORKFLOW                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────┐
│   1. INPUT   │  User provides compliance context
└──────┬───────┘
       │
       │  • Framework (SOC2, ISO27001, PCI-DSS)
       │  • Section/Control ID (e.g., CC6.1, A.9.4.2)
       │  • Policy Excerpts JSON file
       │  • Evidence files (GitHub, Jira, AWS, etc.)
       │
       ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   2. LOAD POLICY CONTEXT                                                 │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  Policy Excerpts File (excerpts.json)                      │        │
│   │  {                                                          │        │
│   │    "framework": "SOC2",                                     │        │
│   │    "version": "2023",                                       │        │
│   │    "excerpts": [{                                           │        │
│   │      "section": "CC6.1",                                    │        │
│   │      "excerpt": "The entity implements logical access...",  │        │
│   │      "control_ids": ["CC6.1", "CC6.2"]                     │        │
│   │    }]                                                        │        │
│   │  }                                                          │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   3. LOAD EVIDENCE EVENTS                                                │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  Evidence Bundle (*.json files via glob)                   │        │
│   │  [{                                                         │        │
│   │    "id": "event-001",                                       │        │
│   │    "source": "github",                                      │        │
│   │    "type": "commit",                                        │        │
│   │    "timestamp": "2025-10-15T10:30:00Z",                    │        │
│   │    "content": "Added MFA support...",                       │        │
│   │    "metadata": {"repo": "auth-service", ...}               │        │
│   │  }]                                                         │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   4. BUILD CONTEXT PREAMBLE                                              │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  ContextPreamble {                                          │        │
│   │    Framework:   "SOC2"                                      │        │
│   │    Section:     "CC6.1"                                     │        │
│   │    Excerpt:     "The entity implements..."  (policy text)  │        │
│   │    ControlIDs:  ["CC6.1", "CC6.2"]                         │        │
│   │  }                                                          │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   5. INTERACTIVE TUI PREVIEW (Optional - can be skipped)                │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  🔍 AI Context Preview                                      │        │
│   │                                                             │        │
│   │  Framework: SOC2 2023                                       │        │
│   │  Section:   CC6.1                                           │        │
│   │  Evidence:  3 events                                        │        │
│   │                                                             │        │
│   │  Policy Excerpt:                                            │        │
│   │  ┌───────────────────────────────────────────────┐         │        │
│   │  │ The entity implements logical access security │         │        │
│   │  │ software, infrastructure, and architectures   │         │        │
│   │  │ over protected information assets...          │         │        │
│   │  └───────────────────────────────────────────────┘         │        │
│   │                                                             │        │
│   │  [Proceed]  [Cancel]                                        │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   6. CONSTRUCT AI PROMPT WITH CONTEXT INJECTION                          │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  SYSTEM PROMPT:                                             │        │
│   │  "You are an expert compliance analyst..."                 │        │
│   │                                                             │        │
│   │  USER PROMPT (Context-Injected):                            │        │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │        │
│   │  Analyze evidence for SOC2 CC6.1:                          │        │
│   │                                                             │        │
│   │  POLICY CONTEXT:                                            │        │
│   │  "The entity implements logical access security            │        │
│   │   software, infrastructure, and architectures over         │        │
│   │   protected information assets..."                          │        │
│   │                                                             │        │
│   │  EVIDENCE EVENTS:                                           │        │
│   │  1. [github/commit] 2025-10-15T10:30:00Z                  │        │
│   │     ID: event-001                                           │        │
│   │     Content: Added MFA support with TOTP...                │        │
│   │                                                             │        │
│   │  2. [github/commit] 2025-10-16T14:20:00Z                  │        │
│   │     ID: event-002                                           │        │
│   │     Content: Updated password policy...                     │        │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   7. AI PROVIDER PROCESSING                                              │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  ┌──────────────┐         ┌──────────────┐                 │        │
│   │  │   OpenAI     │   OR    │  Anthropic   │                 │        │
│   │  │   GPT-4      │         │  Claude 3    │                 │        │
│   │  └──────────────┘         └──────────────┘                 │        │
│   │                                                             │        │
│   │  • Function Calling (OpenAI)                                │        │
│   │  • Tool Use (Anthropic)                                     │        │
│   │  • Structured JSON Schema Output                            │        │
│   │  • Rate Limiting & Timeout Handling                         │        │
│   │  • Automatic PII/Secret Redaction                           │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   8. PARSE STRUCTURED AI RESPONSE                                        │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  AI Response (Structured JSON):                             │        │
│   │  {                                                          │        │
│   │    "title": "Partial Compliance with SOC2 CC6.1",          │        │
│   │    "summary": "MFA, password policy, session timeout...",  │        │
│   │    "justification": "Evidence demonstrates MFA...",         │        │
│   │    "confidence_score": 0.9,                                 │        │
│   │    "residual_risk": "Missing user registration...",         │        │
│   │    "mapped_controls": ["CC6.1"],                            │        │
│   │    "citations": ["event-001", "event-002", "event-003"],   │        │
│   │    "severity": "medium"                                     │        │
│   │  }                                                          │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   9. BUILD COMPLIANCE FINDING                                            │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  Finding {                                                  │        │
│   │    ID:              "finding-CC6.1-1760779788"              │        │
│   │    ControlID:       "CC6.1"                                 │        │
│   │    FrameworkID:     "SOC2"                                  │        │
│   │    Title:           "Partial Compliance with SOC2 CC6.1"   │        │
│   │    ConfidenceScore: 0.9                                     │        │
│   │    Severity:        "medium"                                │        │
│   │    Status:          "open"                                  │        │
│   │    ReviewRequired:  false  (confidence >= 70%)             │        │
│   │    Mode:            "ai"                                    │        │
│   │    Provenance: [                                            │        │
│   │      {Source: "github", EventsUsed: 3}                     │        │
│   │    ]                                                        │        │
│   │    CreatedAt:       "2025-10-18T12:29:48Z"                 │        │
│   │  }                                                          │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   10. EXPORT COMPLIANCE REPORT                                           │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  findings.json                                              │        │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │        │
│   │  • Complete Finding object with all metadata               │        │
│   │  • Citations linking back to evidence events               │        │
│   │  • Provenance tracking (source → event count)              │        │
│   │  • Confidence score for quality assessment                 │        │
│   │  • Residual risk for gap identification                    │        │
│   │  • Machine-readable JSON format                             │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│   11. DISPLAY SUMMARY                                                    │
│   ┌────────────────────────────────────────────────────────────┐        │
│   │  ✅ Analysis Complete!                                      │        │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │        │
│   │  Framework:       SOC2                                      │        │
│   │  Control:         CC6.1                                     │        │
│   │  Confidence:      90.0%                                     │        │
│   │  Residual Risk:   Missing user registration procedures     │        │
│   │  Mapped Controls: 1                                         │        │
│   │    • CC6.1                                                  │        │
│   │  Citations:       3                                         │        │
│   │    • event-001                                              │        │
│   │    • event-002                                              │        │
│   │    • event-003                                              │        │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │        │
│   │  📄 Finding saved to: findings.json                         │        │
│   └────────────────────────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────────────────────────┘
```

### Key Workflow Features

#### 1. **Context Injection**
Instead of generic AI prompts, SDEK injects specific policy requirements directly into the AI prompt:
- ✅ Framework-specific compliance language
- ✅ Control IDs and section references
- ✅ Exact policy excerpt text
- ✅ Related control mappings

**Result:** AI understands the exact compliance requirements and provides policy-grounded analysis.

#### 2. **Evidence Bundle**
Multiple evidence sources are normalized into a standard format:
- **Sources:** GitHub commits, Jira tickets, AWS CloudTrail logs, CI/CD pipelines, documentation
- **Metadata:** Timestamps, authors, repositories, tags
- **Content:** Full text with context

**Result:** AI has complete visibility into all relevant evidence across your infrastructure.

#### 3. **Structured Output**
AI responses follow a strict schema enforced through:
- **OpenAI:** Function Calling with JSON schema
- **Anthropic:** Tool Use with input schema
- **Required fields:** title, summary, justification, confidence_score, mapped_controls
- **Optional fields:** residual_risk, citations, severity

**Result:** Consistent, machine-readable findings that can be tracked, reported, and audited.

#### 4. **Quality Assurance**
Every finding includes quality metrics:
- **Confidence Score (0.0-1.0):** AI's certainty in its analysis
- **Review Required Flag:** Auto-flagged when confidence < 70%
- **Citations:** Links back to specific evidence events
- **Provenance:** Tracks which sources contributed how many events
- **Residual Risk:** Identifies gaps and remaining concerns

**Result:** Transparent, auditable compliance analysis with clear quality indicators.

#### 5. **Privacy & Security**
Before sending to AI providers:
- ✅ Automatic PII redaction (emails, SSNs, credit cards)
- ✅ Secret detection (API keys, passwords, tokens)
- ✅ Configurable redaction rules
- ✅ Local caching to minimize API calls
- ✅ Rate limiting and timeout handling

**Result:** Compliance analysis that respects data privacy and security.

### Example Usage

```bash
# Analyze evidence with AI context injection
./sdek ai analyze \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file ./policies/soc2_excerpts.json \
  --evidence-path ./evidence/github_*.json \
  --evidence-path ./evidence/jira_*.json \
  --output ./findings/cc61_finding.json

# Generate and execute evidence collection plan
./sdek ai plan \
  --framework ISO27001 \
  --section A.9.4.2 \
  --excerpts-file ./policies/iso_excerpts.json
```

See [AI-Enhanced Evidence Analysis](#ai-enhanced-evidence-analysis) below for configuration details.

## Features

- 🔄 **Multi-source ingestion**: Git commits, Jira tickets, Slack messages, CI/CD pipelines, Documentation
- 📊 **Framework mapping**: SOC 2, ISO 27001, PCI DSS with 124 controls
- 🎯 **Evidence analysis**: Automatic evidence-to-control mapping with confidence scores
- 🔌 **MCP Integration**: Model Context Protocol support for extensible evidence collection
  - **Unified Protocol**: Connect to any MCP-compatible tool (GitHub, Slack, Jira, AWS, etc.)
  - **Auto-discovery**: Load configurations from project, user, or custom directories
  - **Health Monitoring**: Real-time status tracking with latency and error metrics
  - **TUI Management**: Visual panel for testing and managing MCP tools
  - **RBAC & Budgets**: Fine-grained permission controls and rate limiting per tool
- 🤖 **AI-Powered Analysis**: Context injection for policy-grounded compliance insights
  - **Context Injection**: Ground AI analysis in exact framework control language
  - **Privacy-First**: Mandatory PII/secret redaction before sending to AI providers
  - **Intelligent Caching**: SHA256-based prompt/response caching for efficiency
  - **Confidence Scoring**: 0-100 scale with automatic low-confidence flagging
- ⚠️ **Risk scoring**: Severity-weighted risk calculation and finding generation
- 📑 **Report export**: JSON compliance reports with role-based filtering
- 🌐 **HTML dashboards**: Interactive web-based compliance visualization
- 🖥️ **Interactive TUI**: Terminal UI for exploring compliance data
- ⚙️ **Configuration management**: Flexible config via files, environment variables, and flags

See the [AI Workflow Architecture](#ai-powered-compliance-analysis-workflow) and [Quickstart Guide](./specs/003-ai-context-injection/quickstart.md) for AI analysis examples.

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

### 4. Using MCP Tools (Optional)

MCP (Model Context Protocol) tools extend SDEK's evidence collection capabilities to external services.

**Setup:**
```bash
# Copy an example configuration
mkdir -p ~/.sdek/mcp
cp docs/examples/mcp/github.json ~/.sdek/mcp/

# Set required environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_OWNER="your-org"
export GITHUB_REPO="your-repo"

# Validate configuration
./sdek mcp validate ~/.sdek/mcp/github.json

# Test connection
./sdek mcp test github
```

**View in TUI:**
```bash
./sdek tui
# Press '5' to navigate to MCP Tools panel
# View real-time status of all configured tools
# Press 't' to test a selected tool
```

**Available Examples:**
- `github.json` - GitHub commits, PRs, and issues
- `slack.json` - Slack messages and channels
- `jira.json` - Jira issues and projects

**Testing AWS MCP Integration:**

To verify that the AI is actually using AWS MCP tools for evidence collection:

```bash
# Run automated integration test
./test_aws_mcp_integration.sh
```

**Testing Guides:**
- **Quick Start**: [How to Verify AWS MCP Usage](./docs/examples/mcp/HOW_TO_VERIFY_AWS_MCP_USAGE.md) - 5-minute verification guide
- **Comprehensive**: [Testing AWS MCP](./docs/TESTING_AWS_MCP.md) - Full testing documentation with troubleshooting

See [MCP Commands](./docs/commands.md#sdek-mcp) for detailed usage and configuration options.

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

### Autonomous Evidence Collection (Experimental)

sdek-cli includes an **experimental autonomous mode** that uses AI to intelligently plan and execute evidence collection across multiple data sources via Model Context Protocol (MCP) connectors.

#### Overview

Instead of manually specifying which evidence to analyze, autonomous mode:

1. **AI Planning**: Analyzes framework requirements and generates an evidence collection plan
2. **Smart Queries**: Identifies relevant data sources and query patterns
3. **Automated Execution**: Fetches evidence from GitHub, Jira, AWS, Slack automatically
4. **Quality Filtering**: Validates signal strength and filters low-quality data
5. **AI Analysis**: Analyzes collected evidence with context injection

```bash
# Generate and execute autonomous evidence collection plan
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file ./policies/soc2_excerpts.json \
  --auto-approve

# Review plan before execution (default behavior)
./sdek ai plan \
  --framework ISO27001 \
  --section A.9.4.2 \
  --excerpts-file ./policies/iso_excerpts.json
```

#### MCP Connectors

Autonomous mode leverages **Model Context Protocol (MCP)** connectors to fetch evidence:

| Connector | Status | Capabilities |
|-----------|--------|--------------|
| **GitHub** | ✅ Implemented | Commits, PRs, issues, releases, code changes |
| **Jira** | 🔨 Planned | Tickets, comments, transitions, JQL queries |
| **AWS** | 🔨 Planned | CloudTrail logs, IAM events, Config changes |
| **Slack** | 🔨 Planned | Messages, threads, channel history |

#### Configuration

Add connector configuration to your config file:

```yaml
ai:
  enabled: true
  provider: openai
  autonomous:
    enabled: true
    auto_approve: false  # Require manual approval before execution
  
  # MCP Connector configuration
  connectors:
    github:
      enabled: true
      api_key: ${GITHUB_TOKEN}  # or use env var
      endpoint: https://api.github.com
      rate_limit: 5000  # requests per hour
      timeout: 30       # seconds
      extra:
        owner: your-org
        default_repos:
          - backend
          - frontend
    
    jira:
      enabled: false
      api_key: ${JIRA_API_TOKEN}
      endpoint: https://your-domain.atlassian.net
      rate_limit: 100
      timeout: 30
      extra:
        project_key: PROJ
    
    aws:
      enabled: false
      # Uses AWS SDK credentials (~/.aws/credentials)
      extra:
        regions:
          - us-east-1
          - us-west-2
    
    slack:
      enabled: false
      api_key: ${SLACK_BOT_TOKEN}
      endpoint: https://slack.com/api
      rate_limit: 50
      timeout: 30
```

**Environment Variables:**

```bash
# GitHub Personal Access Token
export GITHUB_TOKEN="ghp_..."

# Jira API Token
export JIRA_API_TOKEN="..."

# AWS Credentials (standard AWS SDK)
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
export AWS_REGION="us-east-1"

# Slack Bot Token
export SLACK_BOT_TOKEN="xoxb-..."
```

#### How It Works

**Step 1: AI generates evidence collection plan**

```bash
./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file policies.json
```

The AI analyzes the framework requirements and creates a plan:

```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 Evidence Collection Plan                                     │
│                                                                  │
│ Framework: SOC2 | Section: CC6.1                                │
│ Control: Logical access security controls                       │
│                                                                  │
│ Proposed Evidence Sources:                                      │
│                                                                  │
│ 1. GitHub Commits [High Signal - 85%]                          │
│    Query: "authentication OR MFA OR password policy"            │
│    Filters: last 90 days, repos: auth-service, user-mgmt       │
│    Rationale: Direct evidence of access control implementations │
│                                                                  │
│ 2. Jira Tickets [Medium Signal - 70%]                          │
│    Query: "project=SEC AND labels=access-control"               │
│    Filters: last 180 days, status: Done, Closed                │
│    Rationale: Planned and completed access control work        │
│                                                                  │
│ 3. AWS CloudTrail [High Signal - 90%]                          │
│    Query: "IAM policy changes AND authentication events"        │
│    Filters: last 90 days, eventName: CreatePolicy, AttachUser  │
│    Rationale: Infrastructure-level access control changes       │
│                                                                  │
│ [A]pprove | [R]eject | [M]odify                                │
└─────────────────────────────────────────────────────────────────┘
```

**Step 2: User approval (unless --auto-approve)**

Review and approve/modify the plan interactively.

**Step 3: Execute evidence collection**

sdek fetches evidence from all enabled connectors in parallel:

```
Fetching from GitHub... ████████████████ 23 commits found
Fetching from Jira...   ████████████████ 12 tickets found
Fetching from AWS...    ████████████████ 8 IAM events found
```

**Step 4: AI analyzes collected evidence**

The collected evidence is automatically analyzed with context injection:

```
Analyzing 43 events for SOC2 CC6.1...
✅ Analysis complete! Confidence: 88% | Finding saved to findings.json
```

#### Auto-Approve Mode

For CI/CD pipelines or trusted environments:

```bash
# Skip manual approval step
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policies.json \
  --auto-approve
```

**⚠️ Warning**: Auto-approve mode will execute queries against all enabled connectors without confirmation. Ensure your rate limits and API quotas are appropriately configured.

#### Best Practices

1. **Start Small**: Enable one connector at a time to understand query patterns
2. **Set Rate Limits**: Configure conservative rate limits to avoid API quota exhaustion
3. **Review Plans**: Don't use `--auto-approve` until you trust the AI's query generation
4. **Monitor Costs**: Track API usage across connectors (especially AWS CloudTrail)
5. **Use Timeouts**: Set reasonable timeouts to prevent long-running queries
6. **Test Queries**: Validate connector queries manually before autonomous execution

#### Limitations

- **Experimental Feature**: Autonomous mode is under active development
- **Provider Required**: Requires OpenAI or Anthropic API access
- **Connector Availability**: Only GitHub fully implemented (Jira/AWS/Slack planned)
- **Query Validation**: AI-generated queries may need refinement for specific use cases
- **Cost Considerations**: Multiple API calls (AI + connectors) can accumulate costs

See [docs/CONNECTORS.md](./docs/CONNECTORS.md) for detailed connector setup guides.

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
