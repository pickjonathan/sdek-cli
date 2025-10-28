# Quickstart: MCP Pluggable Architecture

**Feature**: 006-mcp-pluggable-architecture
**Target Users**: Compliance Managers, Security Engineers, DevOps Engineers
**Time to Complete**: 15-30 minutes

This guide walks you through the new MCP pluggable architecture in sdek-cli, showing you how to configure MCP servers, switch AI providers, and run multi-system compliance workflows.

---

## Prerequisites

- sdek-cli v1.x+ installed (with Feature 006 support)
- Basic familiarity with sdek-cli commands (`sdek analyze`, `sdek ai plan`)
- (Optional) AWS CLI configured for AWS MCP integration
- (Optional) Ollama installed for local AI models

---

## Quick Start: 5 Minutes

### 1. Check MCP Status

```bash
# Verify MCP integration is available
sdek mcp --version

# List configured MCP servers (initially empty)
sdek mcp list-servers
```

**Expected Output:**
```
MCP Integration: Enabled
Configured Servers: 0
```

### 2. Configure Your First MCP Server (AWS API)

Create or edit `~/.sdek/mcp-config.yaml`:

```yaml
mcp:
  enabled: true
  prefer_mcp: true

  servers:
    aws-api:
      command: "uvx"
      args: ["aws-api-mcp-server"]
      transport: "stdio"
      env:
        AWS_PROFILE: "default"
        READ_OPERATIONS_ONLY: "true"
```

### 3. Test MCP Server Connection

```bash
# Test AWS MCP server
sdek mcp test aws-api
```

**Expected Output:**
```
✓ aws-api: Connected successfully
✓ Tools discovered: call_aws, suggest_aws_commands
✓ Health check: PASSED
```

### 4. Run Your First MCP-Powered Analysis

```bash
# Analyze with AWS evidence collection
sdek ai plan --framework pci-dss --requirement 8.2.3 \
  --sources aws-api \
  --autonomous
```

**What Happens:**
1. AI analyzes PCI-DSS 8.2.3 (Multi-Factor Authentication)
2. AI proposes AWS evidence queries (IAM MFA policies, user MFA status)
3. You approve the plan
4. System executes queries via AWS MCP server
5. AI analyzes collected evidence with policy context
6. You get a compliance finding report

---

## Common Workflows

### Workflow 1: Multi-Source Evidence Collection

**Goal**: Collect evidence from GitHub, AWS, and Jira simultaneously

**Step 1**: Configure multiple MCP servers

```yaml
mcp:
  enabled: true
  max_concurrent: 10  # Collect from 10 sources in parallel

  servers:
    aws-api:
      command: "uvx"
      args: ["aws-api-mcp-server"]
      transport: "stdio"
      env:
        AWS_PROFILE: "readonly"
        READ_OPERATIONS_ONLY: "true"

    github-mcp:
      command: "npx"
      args: ["-y", "@github/github-mcp-server"]
      transport: "stdio"
      env:
        GITHUB_TOKEN: "${GITHUB_TOKEN}"

    jira-mcp:
      command: "node"
      args: ["jira-mcp-server/build/index.js"]
      transport: "stdio"
      env:
        JIRA_API_TOKEN: "${JIRA_API_TOKEN}"
        JIRA_HOST: "https://yourcompany.atlassian.net"
```

**Step 2**: Run autonomous mode with all sources

```bash
sdek ai plan --framework soc2 --control CC6.1 \
  --sources all \
  --autonomous
```

**Step 3**: Review and approve plan

```
Evidence Collection Plan for SOC2 CC6.1:

 Source       Tool              Query                Signal
──────────────────────────────────────────────────────────
✓ github-mcp  search_code       auth, login, MFA     0.85
✓ aws-api     call_aws          iam list-users       0.90
✓ jira-mcp    search_issues     project=SECURITY     0.75

Approve this plan? [y/n]: y
```

**Step 4**: Watch parallel execution

```
Collecting evidence...
✓ aws-api: 15 events (2.3s)
✓ github-mcp: 32 events (3.1s)
✓ jira-mcp: 8 events (1.8s)

Evidence collected: 55 events from 3 sources
Running AI analysis...
```

---

### Workflow 2: Switch AI Providers

**Goal**: Use local Ollama model for development, GPT-4 for production

**Development Setup:**

```bash
# Configure Ollama provider
sdek config set ai.provider_url "ollama://localhost:11434"
sdek config set ai.model "gemma3:12b"

# Verify health
sdek ai health
```

**Expected Output:**
```
AI Provider: Ollama (localhost:11434)
Model: gemma3:12b
Status: ✓ Healthy
```

**Run Test Analysis:**

```bash
sdek analyze --demo --control CC6.1
```

**Production Setup:**

```bash
# Switch to OpenAI
sdek config set ai.provider_url "openai://api.openai.com"
sdek config set ai.model "gpt-4o"
sdek config set ai.providers.openai.api_key "${SDEK_AI_OPENAI_KEY}"

# Verify health
sdek ai health
```

**Expected Output:**
```
AI Provider: OpenAI (api.openai.com)
Model: gpt-4o
Status: ✓ Healthy
Rate Limit: 10 req/min
```

---

### Workflow 3: Configure Auto-Approval Policies

**Goal**: Auto-approve safe AWS queries, require manual approval for others

**Edit** `~/.sdek/config.yaml`:

```yaml
ai:
  autonomous:
    enabled: true
    auto_approve:
      # Auto-approve read-only AWS operations
      aws-api:
        - "iam:List*"
        - "iam:Get*"
        - "cloudtrail:Describe*"
        - "s3:ListBucket"

      # Auto-approve GitHub code searches
      github-mcp:
        - "auth*"
        - "login*"
        - "security*"

      # Require manual approval for Jira
      # (no patterns listed = manual approval)
```

**Test Auto-Approval:**

```bash
sdek ai plan --framework iso27001 --section A.9.4 \
  --sources aws-api,github-mcp,jira-mcp \
  --autonomous
```

**Expected Behavior:**
```
Evidence Collection Plan:

 Source       Tool           Query              Approval
──────────────────────────────────────────────────────────
✓ aws-api     call_aws       iam list-users     AUTO
✓ github-mcp  search_code    auth*              AUTO
⚠ jira-mcp    search_issues  project=SEC        MANUAL

Auto-approved: 2 items
Require approval: 1 item

Approve jira-mcp search? [y/n]:
```

---

### Workflow 4: Health Monitoring & Troubleshooting

**Check All MCP Server Health:**

```bash
sdek mcp health
```

**Expected Output:**
```
MCP Server Health:

 Server        Status      Tools  Last Check         Error Rate
───────────────────────────────────────────────────────────────
✓ aws-api      Healthy     2      2025-10-26 10:30   0.0%
✓ github-mcp   Healthy     5      2025-10-26 10:30   0.0%
✗ jira-mcp     Down        0      2025-10-26 10:28   100.0%
  └─ Error: connection timeout after 30s

Overall: 2/3 servers healthy (66.7%)
```

**Troubleshoot Failed Server:**

```bash
# Test specific server with verbose output
sdek mcp test jira-mcp --verbose
```

**Common Issues & Solutions:**

| Issue | Likely Cause | Solution |
|-------|-------------|----------|
| Connection timeout | MCP server not running | Check `command` path, try running manually |
| Authentication failure | Invalid API key | Verify environment variable: `echo $JIRA_API_TOKEN` |
| Tool discovery fails | Wrong transport type | Check `transport: stdio` vs `transport: http` |
| Rate limit errors | Too many requests | Increase `rate_limit` in config or reduce `max_concurrent` |

---

## Advanced Configurations

### HTTP Transport (Remote MCP Servers)

```yaml
mcp:
  servers:
    remote-compliance-api:
      url: "https://compliance-mcp.example.com/api"
      transport: "http"
      timeout: 30
      headers:
        Authorization: "Bearer ${COMPLIANCE_API_TOKEN}"
        Content-Type: "application/json"
      health_url: "https://compliance-mcp.example.com/health"
```

### Retry Configuration

```yaml
mcp:
  retry:
    max_attempts: 3
    backoff: "exponential"  # exponential | linear | constant
    initial_delay_ms: 1000
    max_delay_ms: 30000
```

### Provider Fallback Chain

```yaml
ai:
  provider_url: "openai://api.openai.com"
  model: "gpt-4o"

  fallback:
    enabled: true
    providers:
      - name: "gemini"
        url: "gemini://generativelanguage.googleapis.com"
        model: "gemini-2.5-flash"
      - name: "ollama"
        url: "ollama://localhost:11434"
        model: "gemma3:12b"
```

**Behavior:**
1. Try OpenAI GPT-4
2. If fails → try Gemini Flash
3. If fails → try local Ollama
4. If all fail → error

---

## Performance Tuning

### Optimize for Speed

```yaml
mcp:
  max_concurrent: 20  # More parallel connections

  servers:
    aws-api:
      timeout: 30      # Shorter timeout
      rate_limit: 60   # Higher rate limit
```

### Optimize for Reliability

```yaml
mcp:
  max_concurrent: 5   # Fewer parallel connections

  retry:
    max_attempts: 5   # More retries
    backoff: "exponential"

  servers:
    aws-api:
      timeout: 120    # Longer timeout
      rate_limit: 10  # Conservative rate limit
```

### Optimize for Cost

```yaml
ai:
  provider_url: "ollama://localhost:11434"  # Use local model
  model: "gemma3:12b"

  budgets:
    max_api_calls: 100    # Limit API calls
    max_tokens: 50000     # Limit tokens
```

---

## Migration from Feature 003

If you have existing connector configurations, sdek-cli will auto-migrate them on first run:

**Old Config** (`~/.sdek/config.yaml` from Feature 003):
```yaml
ai:
  connectors:
    github:
      enabled: true
      api_key: "${GITHUB_TOKEN}"
    aws:
      enabled: true
      endpoint: "https://aws.amazonaws.com"
```

**Auto-Migrated Config** (Feature 006):
```yaml
mcp:
  enabled: true
  servers:
    github-mcp:
      command: "npx"
      args: ["-y", "@github/github-mcp-server"]
      transport: "stdio"
      env:
        GITHUB_TOKEN: "${GITHUB_TOKEN}"

    aws-api:
      command: "uvx"
      args: ["aws-api-mcp-server"]
      transport: "stdio"
      env:
        AWS_REGION: "us-east-1"
```

**Migration Log:**
```
$ sdek analyze

INFO: Legacy connector config detected, auto-migrating to MCP format
INFO: Migrated 2 connectors: github → github-mcp, aws → aws-api
INFO: Review new config: ~/.sdek/config.yaml (section: mcp.servers)
INFO: Backup created: ~/.sdek/config.yaml.backup.2025-10-26
WARN: Legacy connector API deprecated, will be removed in v2.0.0
```

---

## Security Best Practices

### 1. Use Read-Only Credentials

```yaml
mcp:
  servers:
    aws-api:
      env:
        AWS_PROFILE: "compliance-readonly"  # ✓ Read-only IAM profile
        READ_OPERATIONS_ONLY: "true"        # ✓ Application-level enforcement
```

### 2. Restrict File System Access

```yaml
mcp:
  servers:
    filesystem:
      command: "npx"
      args: ["@modelcontextprotocol/server-filesystem", "/compliance-evidence"]
      # ✓ Restricted to specific directory
```

### 3. Never Hardcode Secrets

```yaml
# ❌ BAD: Hardcoded secret
servers:
  jira-mcp:
    env:
      JIRA_API_TOKEN: "abc123..."

# ✓ GOOD: Environment variable
servers:
  jira-mcp:
    env:
      JIRA_API_TOKEN: "${JIRA_API_TOKEN}"
```

### 4. Enable Audit Logging

```yaml
ai:
  safety:
    audit_all_executions: true
    audit_log: "/var/log/sdek/tool-audit.log"
```

---

## Next Steps

- **Explore MCP Servers**: Browse [modelcontextprotocol.io/servers](https://modelcontextprotocol.io/servers) for community MCP servers
- **Custom MCP Servers**: Learn to build custom MCP servers for your organization
- **Advanced Workflows**: Set up scheduled autonomous evidence collection
- **Integration**: Connect sdek-cli to CI/CD pipelines for continuous compliance

---

## Getting Help

- **Documentation**: `sdek mcp --help`
- **Troubleshooting**: `sdek mcp health --verbose`
- **Examples**: `~/.sdek/examples/` directory
- **Community**: [GitHub Discussions](https://github.com/pickjonathan/sdek-cli/discussions)

---

**Version**: Feature 006 (MCP Pluggable Architecture)
**Last Updated**: 2025-10-26
