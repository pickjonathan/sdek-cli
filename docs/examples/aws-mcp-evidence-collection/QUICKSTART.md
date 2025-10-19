# Quick Start Guide - AWS MCP Evidence Collection

## 5-Minute Setup

### 1. Install MCP Configuration

```bash
# Copy configuration to MCP directory
mkdir -p ~/.sdek/mcp
cp examples/aws-mcp-evidence-collection/mcp-aws-config.json ~/.sdek/mcp/aws.json

# Update with your username (or use $HOME which auto-expands)
sed -i '' "s|\$HOME|$HOME|g" ~/.sdek/mcp/aws.json
```

### 2. Set Environment Variables

```bash
export OPENAI_API_KEY="sk-your-key-here"
export AWS_PROFILE="default"  # or your AWS profile name
export AWS_REGION="us-east-1"
```

### 3. Run Evidence Collection

```bash
cd examples/aws-mcp-evidence-collection

# Quick test (plan only, no execution)
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --dry-run

# Full execution
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --approve-all
```

### 4. View Results

```bash
# View findings
cat findings.json | jq '.'

# Check what evidence was collected
cat findings.json | jq '.citations'

# See confidence score
cat findings.json | jq '{confidence: .confidence_score, severity: .severity}'
```

## Viewing MCP Responses

### Method 1: Debug Logs

Run with `--log-level debug` to see MCP responses:

```bash
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee debug.log

# View MCP-related logs
grep "MCP" debug.log

# View evidence details
grep "MCP evidence details" debug.log | jq -R 'fromjson'

# View response data
grep "MCP tool response received" debug.log
```

### Method 2: Check Citations

The AI summary includes citations showing which MCP tools were used:

```bash
cat findings.json | jq '.citations'
# Output: ["AWS IAM/mcp-evidence", "AWS CloudTrail/mcp-evidence"]
```

The `/mcp-evidence` suffix confirms the AI used data from MCP tools.

### Method 3: Summary Analysis

The AI's summary explicitly mentions the evidence source:

```bash
cat findings.json | jq -r '.summary'
```

Look for phrases like:
- "Evidence collected via the MCP tool 'aws-api'"
- "AWS IAM users, policies, and CloudTrail events are being monitored"
- "MCP tool invocation suggests..."

### Method 4: Evidence Events

View the raw evidence events (if enabled in config):

```bash
cat findings.json | jq '.evidence_events[]'
```

## Verification Checklist

✅ **MCP Configuration Working**:
```bash
# Should return server info
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
docker run --rm -i \
  --env AWS_REGION=us-east-1 \
  --env READ_OPERATIONS_ONLY=true \
  --env AWS_PROFILE=default \
  --volume $HOME/.aws:/app/.aws \
  public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest
```

✅ **MCP Registry Initialized**:
```bash
# Look for this in logs
grep "MCP tools registered" debug.log
# Output: {"level":"INFO","msg":"MCP tools registered","count":1}

grep "MCP tool available" debug.log  
# Output: {"level":"INFO","msg":"MCP tool available","name":"aws-api","status":"ready","enabled":true}
```

✅ **Evidence Collected via MCP**:
```bash
# Count successful MCP invocations
grep "MCP tool invocation successful" debug.log | wc -l
# Should match number of evidence items collected
```

✅ **AI Used MCP Evidence**:
```bash
# Check citations
cat findings.json | jq '.citations | map(select(contains("mcp-evidence")))'
# Should return array with MCP evidence citations

# Check summary mentions MCP
cat findings.json | jq -r '.summary' | grep -i "mcp\|tool"
```

## Common Commands

### Run with Specific AWS Profile

```bash
AWS_PROFILE=my-profile ../../sdek ai plan ...
```

### Test Single Evidence Item

Create a minimal policy:

```bash
cat > minimal-policy.json << 'EOF'
{
  "framework": "SOC2",
  "section": "CC6.1",
  "title": "IAM Test",
  "excerpts": [{
    "id": "test-1",
    "text": "Check IAM users exist",
    "keywords": ["IAM", "users"]
  }]
}
EOF

../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file minimal-policy.json \
  --config ai-config.yaml \
  --approve-all \
  --log-level debug
```

### Run Automated Test Script

```bash
./test-evidence-collection.sh
```

This script will:
1. ✓ Check all prerequisites
2. ✓ Set up MCP configuration
3. ✓ Test MCP server manually
4. ✓ Generate evidence collection plan
5. ✓ Optionally run full collection
6. ✓ Show results and summary

## Troubleshooting Quick Fixes

### "MCP tool not found"
```bash
ls -la ~/.sdek/mcp/aws.json
# If missing: cp examples/aws-mcp-evidence-collection/mcp-aws-config.json ~/.sdek/mcp/aws.json
```

### "Docker permission denied"
```bash
# Add your user to docker group (Linux)
sudo usermod -aG docker $USER
# Then logout and login

# Or use Docker Desktop (macOS/Windows)
```

### "AWS credentials not accessible"
```bash
# Verify credentials file
cat ~/.aws/credentials

# Test AWS CLI
aws sts get-caller-identity
```

### "No evidence collected"
```bash
# Run with debug logging
../../sdek ai plan ... --log-level debug 2>&1 | tee debug.log

# Check what AI planned
grep "Using MCP tool" debug.log

# Check for errors
grep -i "error\|failed" debug.log
```

## Understanding the Output

### Finding Structure

```json
{
  "control_id": "CC6.1",           // What you're auditing
  "framework_id": "SOC2",           // Compliance framework
  "confidence_score": 0.7,          // AI's confidence (0-1)
  "severity": "medium",             // Risk level
  "citations": [                    // Evidence sources
    "AWS IAM/mcp-evidence",         // ← MCP tool used!
    "AWS CloudTrail/mcp-evidence"   // ← MCP tool used!
  ],
  "summary": "...",                 // AI analysis
  "mode": "autonomous"              // How it was collected
}
```

### Confidence Scores

- **0.9-1.0 (High)**: Strong evidence, clear compliance
- **0.7-0.9 (Medium-High)**: Good evidence, likely compliant
- **0.5-0.7 (Medium)**: Moderate evidence, needs review
- **<0.5 (Low)**: Weak evidence, significant gaps

### Citations Format

```
{SOURCE_NAME}/{EVIDENCE_TYPE}
```

Examples:
- `AWS IAM/mcp-evidence` - IAM data from MCP
- `AWS CloudTrail/mcp-evidence` - CloudTrail from MCP  
- `GitHub/api` - Direct GitHub API (not MCP)

The `/mcp-evidence` suffix confirms MCP was used.

## Next Steps

1. **Review full documentation**: See [README.md](README.md)
2. **Customize policy**: Edit `soc2-cc6.1-policy.json` for your needs
3. **Add more frameworks**: Create policy files for ISO27001, NIST, etc.
4. **Integrate with CI/CD**: Use findings in automated compliance checks
5. **Explore other MCP servers**: GitHub, Jira, Slack connectors

## Resources

- [MCP Protocol Documentation](../../docs/MCP_CONCURRENT_FIX.md)
- [AI Workflow Architecture](../../docs/AI_WORKFLOW_ARCHITECTURE.md)
- [Feature 004 Spec](../../specs/004-mcp-native-agent/spec.md)
- [AWS MCP Server](https://github.com/awslabs/aws-mcp-server)
