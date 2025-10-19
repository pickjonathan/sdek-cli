# AWS MCP Evidence Collection Example

This example demonstrates how to configure and use the AWS MCP (Model Context Protocol) server for autonomous evidence collection in compliance audits.

## Overview

The AWS MCP integration allows SDEK to collect evidence from AWS services automatically using the `awslabs/aws-api-mcp-server`. The MCP server executes AWS CLI commands and returns structured evidence that can be analyzed by AI for compliance assessment.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Configuration](#configuration)
3. [Policy Excerpts](#policy-excerpts)
4. [AI Configuration](#ai-configuration)
5. [Running Evidence Collection](#running-evidence-collection)
6. [Viewing MCP Responses](#viewing-mcp-responses)
7. [Understanding the Output](#understanding-the-output)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### 1. Docker

The AWS MCP server runs in a Docker container:

```bash
# Check Docker is installed
docker --version

# Pull the AWS MCP server image
docker pull public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest
```

### 2. AWS Credentials

Ensure you have AWS credentials configured:

```bash
# Check AWS credentials
aws configure list

# Or set environment variables
export AWS_PROFILE=your-profile
export AWS_REGION=us-east-1
```

### 3. MCP Configuration Directory

Create the MCP configuration directory:

```bash
mkdir -p ~/.sdek/mcp
```

## Configuration

### Step 1: Create MCP Server Configuration

Create `~/.sdek/mcp/aws.json`:

```json
{
  "command": "docker",
  "args": [
    "run",
    "--rm",
    "--interactive",
    "--env", "AWS_REGION=us-east-1",
    "--env", "READ_OPERATIONS_ONLY=true",
    "--env", "AWS_PROFILE=default",
    "--volume", "/Users/YOUR_USERNAME/.aws:/app/.aws",
    "public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest"
  ],
  "transport": "stdio",
  "capabilities": [
    "aws.call-aws",
    "aws.suggest-aws-commands"
  ]
}
```

**Important**: Replace `YOUR_USERNAME` with your actual username, or use `$HOME`:

```json
"--volume", "$HOME/.aws:/app/.aws"
```

**Configuration Options**:

- `AWS_REGION`: AWS region for API calls (default: us-east-1)
- `READ_OPERATIONS_ONLY=true`: Safety measure - only allow read-only operations
- `AWS_PROFILE`: Which AWS credentials profile to use
- Volume mount: Maps your local AWS credentials into the container

### Step 2: Verify MCP Server Configuration

Test the MCP server manually:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
docker run --rm -i \
  --env AWS_REGION=us-east-1 \
  --env READ_OPERATIONS_ONLY=true \
  --env AWS_PROFILE=default \
  --volume $HOME/.aws:/app/.aws \
  public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest
```

Expected response:
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{},"serverInfo":{"name":"AWS-API-MCP","version":"1.11.0"}}}
```

## Policy Excerpts

Create a policy excerpt file that defines what you're auditing. Create `policy.json`:

```json
{
  "framework": "SOC2",
  "section": "CC6.1",
  "title": "Logical and Physical Access Controls",
  "excerpts": [
    {
      "id": "CC6.1-1",
      "text": "The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives.",
      "keywords": ["access control", "authentication", "authorization", "IAM", "security groups", "network ACLs"]
    },
    {
      "id": "CC6.1-2", 
      "text": "The entity uses encryption to protect data both at rest and in transit.",
      "keywords": ["encryption", "TLS", "SSL", "KMS", "encryption at rest", "encryption in transit"]
    },
    {
      "id": "CC6.1-3",
      "text": "The entity monitors and logs access to information assets.",
      "keywords": ["CloudTrail", "logging", "monitoring", "audit logs", "access logs"]
    }
  ]
}
```

## AI Configuration

Create `ai_config.yaml` to configure AI behavior:

```yaml
# AI Provider Configuration
provider: openai

# OpenAI Configuration
openai:
  api_key: ${OPENAI_API_KEY}  # Read from environment
  model: gpt-4-turbo-preview
  temperature: 0.1  # Low temperature for consistent, factual analysis
  max_tokens: 4000

# Evidence Collection Configuration  
evidence_collection:
  mode: autonomous  # autonomous | interactive
  auto_approve: true  # Auto-approve all evidence collection steps
  
# Analysis Configuration
analysis:
  confidence_threshold: 0.6  # Minimum confidence for findings
  enable_ai_analysis: true
  combine_with_heuristics: true
  
# Privacy & Redaction
privacy:
  redact_pii: true
  redact_secrets: true
  allowed_data_types:
    - configuration
    - metadata
    - logs
```

## Running Evidence Collection

### Basic Usage

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai_config.yaml \
  --approve-all
```

### With Debug Logging

To see the MCP responses and detailed execution:

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai_config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee evidence_collection.log
```

### Interactive Mode

To review each evidence collection step before execution:

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai_config.yaml
```

## Viewing MCP Responses

### Debug Logs

When running with `--log-level debug`, you'll see:

```json
{"level":"DEBUG","msg":"MCP tool response received","tool":"aws-api","method":"tools/call","has_result":true,"response_keys":["result","jsonrpc","id"]}
{"level":"DEBUG","msg":"MCP evidence details","tool":"aws-api","evidence_id":"51c189dd-fd50-4760-9bec-48811a9feec7","reasoning":"Evidence collected via MCP tool 'aws-api' using method 'tools/call'","keywords":["mcp-evidence"],"confidence_score":70,"analysis_method":"mcp-direct"}
```

### Evidence Events

The MCP responses are converted to `EvidenceEvent` objects. To view the raw events:

```bash
# If you have the TUI built
./sdek tui

# Or examine the findings JSON
cat findings.json | jq '.citations'
```

### Audit Logs

MCP invocations are logged in the audit trail:

```bash
# Check audit logs (if enabled)
cat ~/.sdek/audit/mcp_audit.log | jq '.'
```

## Understanding the Output

### Findings JSON

The `findings.json` file contains the complete analysis:

```json
{
  "id": "finding-1760898749",
  "control_id": "CC6.1",
  "framework_id": "SOC2",
  "title": "SOC2 CC6.1 Analysis",
  "severity": "medium",
  "confidence_score": 0.7,
  "summary": "The entity appears to have implemented logical access security controls...",
  "citations": [
    "AWS IAM/mcp-evidence",
    "AWS CloudTrail/mcp-evidence"
  ],
  "mode": "autonomous"
}
```

**Key Fields**:

- `confidence_score`: AI's confidence in the finding (0.0-1.0)
- `severity`: Risk level (low/medium/high)
- `citations`: Sources of evidence (shows MCP tools used)
- `summary`: AI-generated analysis of compliance
- `mode`: How evidence was collected (autonomous/interactive)

### Evidence in Summary

To verify the AI used MCP evidence in its analysis:

1. **Check citations**: Look for `/mcp-evidence` suffix in citations
2. **Review summary**: AI explains what evidence was found
3. **Confidence score**: Higher scores indicate stronger evidence
4. **Debug logs**: Show exact MCP responses used

Example debug output showing evidence flow:

```
INFO: Using MCP tool for evidence collection | tool=aws-api | source=AWS IAM
DEBUG: Invoking MCP tool | tool=aws-api | method=tools/call | args={...}
DEBUG: MCP tool response received | has_result=true
DEBUG: MCP evidence details | evidence_id=... | confidence_score=70
INFO: MCP tool invocation successful | evidence_id=...
INFO: Evidence collected | events=6
INFO: Analyzing collected evidence
```

## Troubleshooting

### Issue: "MCP tool not found"

**Solution**: Check MCP configuration exists:

```bash
ls -la ~/.sdek/mcp/aws.json
cat ~/.sdek/mcp/aws.json | jq '.'
```

### Issue: "Docker container fails to start"

**Solution**: Test Docker manually:

```bash
docker run --rm -i \
  --env AWS_REGION=us-east-1 \
  --env AWS_PROFILE=default \
  --volume $HOME/.aws:/app/.aws \
  public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest
```

### Issue: "AWS credentials not found"

**Solution**: Verify credentials are accessible:

```bash
# Check credentials file
cat ~/.aws/credentials

# Test AWS CLI
aws sts get-caller-identity --profile default
```

### Issue: "Response corruption (\x00 error)"

**Solution**: This was fixed in the latest version. Ensure you're using the concurrent-fix build:

```bash
make build
./sdek --version
```

### Issue: "No evidence collected"

**Solution**: Run with debug logging to see what the AI is attempting:

```bash
./sdek ai plan ... --log-level debug 2>&1 | grep -E "(Using MCP|MCP tool|evidence)"
```

### Issue: "AI not using MCP evidence in summary"

**Solution**: Check that evidence was actually collected:

```bash
# Look for "Evidence collected" message
grep "Evidence collected" evidence_collection.log

# Check citations in findings
cat findings.json | jq '.citations'
```

## Advanced Usage

### Custom AWS Commands

The AI will generate AWS CLI commands based on the policy excerpts. You can see what commands it plans to run in the plan output.

### Multiple Frameworks

Run evidence collection for multiple frameworks:

```bash
for framework in SOC2 ISO27001 NIST; do
  ./sdek ai plan \
    --framework $framework \
    --section AccessControl \
    --excerpts-file policies/${framework}_policy.json \
    --config ai_config.yaml \
    --approve-all \
    --output ${framework}_findings.json
done
```

### Integration with CI/CD

```bash
#!/bin/bash
# ci-compliance-check.sh

# Run evidence collection
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai_config.yaml \
  --approve-all \
  --output findings.json

# Check confidence score
CONFIDENCE=$(cat findings.json | jq -r '.confidence_score')
if (( $(echo "$CONFIDENCE < 0.7" | bc -l) )); then
  echo "❌ Compliance confidence too low: $CONFIDENCE"
  exit 1
fi

echo "✅ Compliance check passed: $CONFIDENCE"
```

## Next Steps

- Explore the [MCP Protocol Documentation](../../docs/MCP_CONCURRENT_FIX.md)
- Review [AI Workflow Architecture](../../docs/AI_WORKFLOW_ARCHITECTURE.md)
- See [Feature 004 Specification](../../specs/004-mcp-native-agent/spec.md)
