# How to Test AWS MCP Tool Integration - Quick Summary

## TL;DR - Quick Test

```bash
# 1. Run the automated test script
./test_aws_mcp_integration.sh

# 2. If you have an OpenAI API key, run a real test
export SDEK_OPENAI_KEY="sk-..."
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --dry-run
```

## What You Need

1. **AWS Credentials**: `aws configure` or environment variables
2. **Node.js**: For running the MCP server (`npx`)
3. **SDEK Binary**: `make build`
4. **AI API Key**: OpenAI or Anthropic (optional for basic tests)

## 5 Key Indicators the AI is Using AWS MCP

### ✅ 1. MCP Tool Registration
```bash
./sdek mcp list
```
Look for: `aws-api` with 🟢 Ready status

### ✅ 2. Config Validation
```bash
./sdek mcp validate ~/.sdek/mcp/aws.json
```
Look for: `✓ Valid MCP configuration`

### ✅ 3. Handshake Test
```bash
./sdek mcp test aws-api
```
Look for: Success message with capabilities listed

### ✅ 4. Debug Logs (Most Important!)
```bash
export SDEK_LOG_LEVEL=debug
./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file test_policy.json
```

**Look for these log lines:**
```
level=DEBUG msg="MCP tool registered" name=aws-api
level=DEBUG msg="Executing MCP tool" name=aws-api operation=iam_list_users
level=DEBUG msg="MCP call successful" name=aws-api latency=234ms events=12
level=INFO msg="Evidence collected" source=aws events=12
```

### ✅ 5. Output Finding Inspection
```bash
cat test_finding.json | jq '.evidence_sources[] | select(.type == "aws")'
```

Look for AWS evidence sources in the output:
```json
{
  "type": "aws",
  "query": "iam:ListUsers",
  "count": 12,
  "timestamp": "2025-10-19T...",
  "metadata": {
    "region": "us-east-1",
    "service": "iam"
  }
}
```

## Troubleshooting Quick Fixes

| Issue | Quick Fix |
|-------|-----------|
| `aws-api` shows "Offline" | Run `./sdek mcp test aws-api` and check error |
| AI doesn't propose AWS sources | Lower confidence: `autonomous.confidence_threshold: 0.5` |
| MCP calls timeout | Increase timeout in `aws.json`: `"requestTimeout": "60s"` |
| RBAC denies operations | Add to `allowedCapabilities` in `aws.json` |
| No AWS credentials | Run `aws configure` or set `AWS_ACCESS_KEY_ID` |

## Directory Structure After Setup

```
~/.sdek/
├── mcp/
│   ├── aws.json           # AWS MCP config
│   ├── github.json        # Optional: GitHub config
│   ├── slack.json         # Optional: Slack config
│   └── audit.log          # MCP audit logs
└── cache/
    └── ai/                # AI response cache
```

## Real-World Testing Workflow

```bash
# 1. Setup AWS MCP config (one time)
mkdir -p ~/.sdek/mcp
cp docs/examples/mcp/github.json ~/.sdek/mcp/
# Edit ~/.sdek/mcp/aws.json to add your config

# 2. Validate and test
./sdek mcp validate ~/.sdek/mcp/aws.json
./sdek mcp test aws-api

# 3. Enable AI in config.yaml
ai:
  enabled: true
  provider: "openai"
  mode: "autonomous"
  autonomous:
    enabled: true

# 4. Set API key
export SDEK_OPENAI_KEY="sk-..."

# 5. Run with debug logging
export SDEK_LOG_LEVEL=debug

# 6. Test with dry-run first
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --dry-run

# 7. If dry-run looks good, run for real
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --output finding.json

# 8. Inspect the results
cat finding.json | jq '.evidence_sources'
cat ~/.sdek/mcp/audit.log | jq 'select(.tool_name == "aws-api")'
```

## What Success Looks Like

When everything is working, you'll see:

1. **Console output** showing MCP tool status
2. **Debug logs** with "Executing MCP tool" messages
3. **Finding JSON** containing AWS evidence sources
4. **Audit log** recording AWS API calls
5. **TUI** showing aws-api with "Last Used" timestamp

## Advanced: Monitor in Real-Time

Open two terminals:

**Terminal 1** - Run SDEK:
```bash
export SDEK_LOG_LEVEL=debug
./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file test_policy.json
```

**Terminal 2** - Watch MCP activity:
```bash
# Watch MCP list
watch -n 2 './sdek mcp list'

# OR tail audit log
tail -f ~/.sdek/mcp/audit.log | jq
```

## Full Documentation

- **Comprehensive Guide**: [docs/TESTING_AWS_MCP.md](./docs/TESTING_AWS_MCP.md)
- **MCP Commands Reference**: [docs/commands.md#sdek-mcp](./docs/commands.md#sdek-mcp)
- **Example Configs**: [docs/examples/mcp/](./docs/examples/mcp/)
- **Automated Test Script**: [test_aws_mcp_integration.sh](./test_aws_mcp_integration.sh)

## Getting Help

1. Run automated test first: `./test_aws_mcp_integration.sh`
2. Check logs with debug enabled: `export SDEK_LOG_LEVEL=debug`
3. Validate config: `./sdek mcp validate ~/.sdek/mcp/aws.json`
4. Test handshake: `./sdek mcp test aws-api`
5. Review full guide: [docs/TESTING_AWS_MCP.md](./docs/TESTING_AWS_MCP.md)
