# AWS MCP Evidence Collection - Complete Example

This directory contains a complete, working example of autonomous evidence collection using the AWS Model Context Protocol (MCP) server.

## 📁 Files in This Directory

### Documentation
- **[README.md](README.md)** - Complete guide with detailed explanations
- **[QUICKSTART.md](QUICKSTART.md)** - 5-minute setup and quick reference
- **[VERIFICATION.md](VERIFICATION.md)** - How to view MCP responses and verify evidence usage

### Configuration Files
- **[mcp-aws-config.json](mcp-aws-config.json)** - MCP server configuration for AWS
- **[ai-config.yaml](ai-config.yaml)** - AI engine configuration
- **[soc2-cc6.1-policy.json](soc2-cc6.1-policy.json)** - Example SOC2 policy excerpts

### Scripts
- **[test-evidence-collection.sh](test-evidence-collection.sh)** - Automated test script

## 🚀 Quick Start

### 1. One-Command Setup

```bash
# From the examples directory
./test-evidence-collection.sh
```

This will:
- ✅ Check all prerequisites
- ✅ Configure MCP server
- ✅ Test MCP connection
- ✅ Run evidence collection
- ✅ Show results

### 2. Manual Setup

```bash
# Install MCP configuration
mkdir -p ~/.sdek/mcp
cp mcp-aws-config.json ~/.sdek/mcp/aws.json
sed -i '' "s|\$HOME|$HOME|g" ~/.sdek/mcp/aws.json

# Set environment
export OPENAI_API_KEY="your-key"
export AWS_PROFILE="default"

# Run evidence collection
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --approve-all
```

## 📖 Documentation Guide

**New to MCP?** Start here:
1. [QUICKSTART.md](QUICKSTART.md) - Get running in 5 minutes
2. [README.md](README.md) - Understand how it works
3. [VERIFICATION.md](VERIFICATION.md) - Verify it's working correctly

**Want to customize?** See:
- [soc2-cc6.1-policy.json](soc2-cc6.1-policy.json) - Modify policy requirements
- [ai-config.yaml](ai-config.yaml) - Tune AI behavior
- [mcp-aws-config.json](mcp-aws-config.json) - Configure AWS access

## 🎯 What This Example Demonstrates

### MCP Integration
- ✅ Configuring AWS MCP server in Docker
- ✅ Initializing MCP registry
- ✅ Invoking MCP tools from AI
- ✅ Concurrent MCP requests
- ✅ Response handling and parsing

### Evidence Collection
- ✅ Autonomous plan generation
- ✅ AWS IAM evidence collection
- ✅ AWS CloudTrail evidence collection
- ✅ Multi-source evidence gathering
- ✅ Evidence normalization

### AI Analysis
- ✅ GPT-4 evidence analysis
- ✅ Confidence scoring
- ✅ Risk assessment
- ✅ Citation tracking
- ✅ Compliance finding generation

## 🔍 Verification Examples

### Check MCP is Working

```bash
# View debug logs
../../sdek ai plan --log-level debug ... 2>&1 | grep "MCP"

# Check citations
cat findings.json | jq '.citations'
# Should show: ["AWS IAM/mcp-evidence", "AWS CloudTrail/mcp-evidence"]

# Verify summary
cat findings.json | jq -r '.summary' | grep -i "mcp"
# Should mention "MCP tool 'aws-api'"
```

### Example Output

```json
{
  "control_id": "CC6.1",
  "framework_id": "SOC2",
  "confidence_score": 0.9,
  "severity": "low",
  "citations": [
    "AWS IAM/mcp-evidence",
    "AWS CloudTrail/mcp-evidence"
  ],
  "summary": "Evidence collected from AWS IAM and CloudTrail via the MCP tool 'aws-api' suggests that the entity is monitoring access controls...",
  "mode": "autonomous"
}
```

The `/mcp-evidence` citations **prove** MCP tools were used!

## 🛠️ Customization

### Different AWS Services

Edit `soc2-cc6.1-policy.json` to target different services:

```json
{
  "id": "custom-1",
  "text": "Check S3 bucket encryption",
  "keywords": ["S3", "encryption", "bucket", "KMS"]
}
```

The AI will automatically generate the appropriate AWS CLI commands.

### Different Compliance Frameworks

Create new policy files for other frameworks:

```bash
# ISO 27001
cp soc2-cc6.1-policy.json iso27001-a9.json
# Edit to use ISO 27001 controls

# NIST CSF
cp soc2-cc6.1-policy.json nist-csf-pr.json
# Edit to use NIST CSF requirements
```

### Multiple AWS Accounts

Configure different MCP profiles:

```bash
# Production account
cat mcp-aws-config.json | sed 's/default/production/' > ~/.sdek/mcp/aws-prod.json

# Staging account  
cat mcp-aws-config.json | sed 's/default/staging/' > ~/.sdek/mcp/aws-staging.json
```

## 📊 Sample Results

### Test Run Statistics

```
Framework:     SOC2 CC6.1
Evidence:      6 events collected
Plan Items:    6 approved, 6 executed
Confidence:    90.0% (high)
Severity:      low
Duration:      23.8s
MCP Calls:     6 successful
```

### Evidence Sources

- AWS IAM (users, roles, policies)
- AWS CloudTrail (access logs)
- AWS S3 (encryption settings)
- AWS KMS (key configurations)
- AWS VPC (network security)

## 🔗 Related Documentation

- [MCP Concurrent Fix](../../docs/MCP_CONCURRENT_FIX.md) - Technical details of the MCP implementation
- [AI Workflow Architecture](../../docs/AI_WORKFLOW_ARCHITECTURE.md) - Overall system design
- [Feature 004 Spec](../../specs/004-mcp-native-agent/spec.md) - MCP native agent specification
- [AWS MCP Server](https://github.com/awslabs/aws-mcp-server) - Official AWS MCP server docs

## 🐛 Troubleshooting

See [README.md#troubleshooting](README.md#troubleshooting) for common issues and solutions.

Quick fixes:

```bash
# MCP not working
ls -la ~/.sdek/mcp/aws.json

# Docker issues
docker ps
docker pull public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest

# AWS credentials
aws sts get-caller-identity

# Debug logs
../../sdek ai plan ... --log-level debug 2>&1 | tee debug.log
```

## 💡 Tips

1. **Start with dry-run**: Use `--dry-run` to see the plan first
2. **Use debug logging**: Always run with `--log-level debug` when testing
3. **Check citations**: The `/mcp-evidence` suffix confirms MCP usage
4. **Review summaries**: AI explicitly mentions MCP tools when used
5. **Verify counts**: Match MCP invocations to evidence collected

## 🤝 Contributing

Found an issue or have an improvement? Please update this example!

## 📝 License

See [LICENSE](../../LICENSE) in the project root.

---

**Questions?** Check [VERIFICATION.md](VERIFICATION.md) for detailed debugging steps.
