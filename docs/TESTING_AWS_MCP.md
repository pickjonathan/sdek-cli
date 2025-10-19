# Testing AWS MCP Tool Integration

This guide shows you how to test that the AI is actually using the AWS MCP tool to collect evidence.

## Prerequisites

1. **AWS MCP Server installed**:
   ```bash
   npm install -g @modelcontextprotocol/server-aws
   ```

2. **AWS credentials configured**:
   ```bash
   # Option 1: Environment variables
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   export AWS_REGION="us-east-1"
   
   # Option 2: AWS CLI profile
   aws configure
   ```

3. **SDEK CLI compiled**:
   ```bash
   make build
   ```

## Step 1: Create AWS MCP Configuration

Create a configuration file at `~/.sdek/mcp/aws.json`:

```bash
mkdir -p ~/.sdek/mcp
cat > ~/.sdek/mcp/aws.json <<'EOF'
{
  "schemaVersion": "1.0.0",
  "name": "aws-api",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-aws"],
  "transport": "stdio",
  "enabled": true,
  "rbac": {
    "allowedCapabilities": [
      "iam_list_users",
      "iam_get_user",
      "iam_list_policies",
      "cloudtrail_lookup_events",
      "ec2_describe_instances",
      "s3_list_buckets"
    ],
    "maxCallsPerMinute": 30,
    "maxTokensPerCall": 5000
  },
  "env": {
    "AWS_REGION": "${AWS_REGION}"
  },
  "description": "AWS API access for CloudTrail, IAM, EC2, and S3"
}
EOF
```

## Step 2: Validate the Configuration

```bash
# Validate the config file
./sdek mcp validate ~/.sdek/mcp/aws.json

# Test the MCP server (performs handshake)
./sdek mcp test aws-api

# List all MCP tools (should show aws-api as "Ready")
./sdek mcp list
```

Expected output from `sdek mcp list`:
```
╭─────────────────────────────────────────────────────────╮
│ MCP Tools                                                │
├─────────┬────────┬─────────┬────────┬──────────────────┤
│ NAME    │ STATUS │ LATENCY │ ERRORS │ CAPABILITIES     │
├─────────┼────────┼─────────┼────────┼──────────────────┤
│ aws-api │ 🟢 Ready│  123ms  │   0    │ 6 capabilities   │
╰─────────┴────────┴─────────┴────────┴──────────────────╯
```

## Step 3: Enable AI Autonomous Mode

Update your `config.yaml` to enable AI with AWS connector:

```yaml
ai:
  enabled: true
  provider: "openai"  # or "anthropic"
  openai_key: "${SDEK_OPENAI_KEY}"  # or set directly
  model: "gpt-4"
  
  mode: "autonomous"
  
  autonomous:
    enabled: true
    confidence_threshold: 0.7
    
    # Auto-approve AWS IAM events
    auto_approve:
      aws:
        - "iam:*"
        - "cloudtrail:*"
  
  # Enable AWS connector (legacy config - for Feature 002 backward compatibility)
  connectors:
    aws:
      enabled: true
```

Set your OpenAI API key:
```bash
export SDEK_OPENAI_KEY="sk-..."
```

## Step 4: Create Test Evidence Plan

Create a simple policy excerpt to test with:

```bash
cat > test_policy.json <<'EOF'
[
  {
    "framework": "SOC2",
    "version": "2017",
    "section": "CC6.1",
    "title": "Logical Access Controls",
    "text": "The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives. This includes controls to: (1) Identify and authenticate users; (2) Authorize users to access authorized resources; (3) Prevent unauthorized access.",
    "related_sections": ["CC6.2", "CC6.3"]
  }
]
EOF
```

## Step 5: Run AI Plan with Verbose Logging

Enable debug logging to see MCP tool calls:

```bash
# Set debug log level
export SDEK_LOG_LEVEL=debug

# Run AI plan generation
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --output test_finding.json
```

## Step 6: Verify AWS MCP Tool Usage

Look for these indicators in the log output:

### 6.1 MCP Registry Initialization
```
time=2025-10-19T... level=DEBUG msg="MCP registry initialized" tools=1
time=2025-10-19T... level=DEBUG msg="MCP tool registered" name=aws-api status=ready capabilities=6
```

### 6.2 Evidence Plan Generation
The AI should propose AWS sources in the plan:
```
time=2025-10-19T... level=INFO msg="AI proposed evidence plan" items=5
time=2025-10-19T... level=DEBUG msg="Plan item" source=aws query="iam:ListUsers" signal_strength=0.85
time=2025-10-19T... level=DEBUG msg="Plan item" source=aws query="cloudtrail:LookupEvents" signal_strength=0.90
```

### 6.3 MCP Tool Invocation
```
time=2025-10-19T... level=DEBUG msg="Executing MCP tool" name=aws-api operation=iam_list_users
time=2025-10-19T... level=DEBUG msg="MCP call successful" name=aws-api latency=234ms events=12
```

### 6.4 Evidence Collection
```
time=2025-10-19T... level=INFO msg="Evidence collected" source=aws events=12
time=2025-10-19T... level=DEBUG msg="Evidence normalized" type=iam_user source=aws timestamp=2025-10-19T...
```

## Step 7: Inspect the Output Finding

Check the generated finding for AWS evidence:

```bash
cat test_finding.json | jq '.evidence_sources[] | select(.type == "aws")'
```

Expected output:
```json
{
  "type": "aws",
  "query": "iam:ListUsers",
  "count": 12,
  "timestamp": "2025-10-19T10:30:45Z",
  "metadata": {
    "region": "us-east-1",
    "service": "iam",
    "operation": "ListUsers"
  }
}
```

## Step 8: Check MCP Audit Logs

The system logs all MCP tool calls for audit purposes:

```bash
# Check audit logs
cat ~/.sdek/mcp/audit.log | jq 'select(.tool_name == "aws-api")'
```

Expected audit log entries:
```json
{
  "timestamp": "2025-10-19T10:30:45Z",
  "tool_name": "aws-api",
  "operation": "iam_list_users",
  "user_role": "compliance_manager",
  "status": "success",
  "latency_ms": 234,
  "tokens_consumed": 1500,
  "rbac_check": "passed"
}
```

## Step 9: Use TUI to Monitor Real-Time

Start the TUI to see MCP tools in action:

```bash
./sdek
```

Navigate to the **MCP Tools** tab (usually Tab 5 or 6):
- You should see `aws-api` with status 🟢 Ready
- After running a plan, check the "Last Used" timestamp
- View error counts and latency metrics

## Step 10: Test with Dry Run

To see what the AI would do without actually executing:

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --dry-run
```

This will show you the proposed plan without invoking AWS MCP tools.

## Troubleshooting

### Issue: AWS MCP tool shows "Offline" status

**Solution**: Check the handshake:
```bash
# Test the MCP server directly
npx -y @modelcontextprotocol/server-aws

# Check AWS credentials
aws sts get-caller-identity

# Validate config
./sdek mcp validate ~/.sdek/mcp/aws.json
```

### Issue: AI doesn't propose AWS sources

**Solution**: 
1. Ensure the policy excerpt mentions AWS-related controls (IAM, access control, authentication)
2. Lower the confidence threshold in config: `autonomous.confidence_threshold: 0.5`
3. Check connector is enabled: `ai.connectors.aws.enabled: true`

### Issue: MCP calls timeout

**Solution**: Increase timeout in `aws.json`:
```json
{
  "rbac": {
    "maxCallsPerMinute": 10,
    "requestTimeout": "60s"
  }
}
```

### Issue: RBAC denies AWS operations

**Solution**: Add required capabilities to `allowedCapabilities`:
```json
{
  "rbac": {
    "allowedCapabilities": [
      "iam_list_users",
      "iam_get_user",
      "iam_list_policies",
      "iam_get_policy",
      "cloudtrail_lookup_events",
      "ec2_describe_instances"
    ]
  }
}
```

## Advanced Testing

### Test Specific AWS Operations

Create a custom test script:

```bash
cat > test_aws_mcp.sh <<'EOF'
#!/bin/bash

# Test IAM list users
echo "Testing IAM ListUsers..."
./sdek mcp test aws-api --operation iam_list_users

# Test CloudTrail events
echo "Testing CloudTrail LookupEvents..."
./sdek mcp test aws-api --operation cloudtrail_lookup_events

# Test EC2 instances
echo "Testing EC2 DescribeInstances..."
./sdek mcp test aws-api --operation ec2_describe_instances

# Check audit log
echo "Checking audit log..."
cat ~/.sdek/mcp/audit.log | jq -r '.timestamp + " " + .tool_name + " " + .operation + " " + .status'
EOF

chmod +x test_aws_mcp.sh
./test_aws_mcp.sh
```

### Monitor MCP Performance

```bash
# Watch MCP list output in real-time
watch -n 2 './sdek mcp list'

# Monitor audit log
tail -f ~/.sdek/mcp/audit.log | jq
```

### Test Auto-Approval Rules

Verify auto-approval works for AWS IAM events:

```bash
# Should auto-approve without TUI prompt
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --approve-all

# Check the plan JSON for approval status
cat test_finding.json | jq '.evidence_plan.items[] | select(.source == "aws") | {source, query, approval_status}'
```

Expected output shows `"approval_status": "auto_approved"`:
```json
{
  "source": "aws",
  "query": "iam:ListUsers",
  "approval_status": "auto_approved"
}
```

## Success Criteria

You've successfully verified AWS MCP tool usage when you see:

1. ✅ `sdek mcp list` shows `aws-api` with 🟢 Ready status
2. ✅ Debug logs show "MCP tool registered" for aws-api
3. ✅ AI proposes AWS sources in evidence plan
4. ✅ Debug logs show "Executing MCP tool" with aws-api
5. ✅ Output finding contains AWS evidence sources
6. ✅ Audit log records AWS MCP operations
7. ✅ TUI shows aws-api with recent "Last Used" timestamp

## Next Steps

- Add more AWS-specific auto-approval rules
- Configure additional AWS services in RBAC capabilities
- Set up CloudWatch integration for MCP monitoring
- Create custom policy excerpts that target specific AWS controls
