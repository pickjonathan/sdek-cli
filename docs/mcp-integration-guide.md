# MCP Integration Guide for sdek-cli

This guide documents how to integrate Model Context Protocol (MCP) servers with sdek-cli for automated evidence collection and compliance analysis.

## AWS API MCP Server Integration

### Overview

The AWS API MCP Server (from awslabs/mcp) enables sdek-cli to collect cloud infrastructure evidence automatically through AWS CLI commands. This is particularly useful for SOC2, ISO 27001, and PCI DSS compliance requirements that involve cloud infrastructure.

### Architecture

The AWS MCP Server provides three primary tools:

1. **call_aws**: Execute validated AWS CLI commands
2. **suggest_aws_commands**: Get command recommendations from natural language
3. **get_execution_plan**: Generate multi-step workflows (experimental)

### Security Model

The AWS MCP Server implements a four-layer security architecture:

#### Layer 1: IAM Credentials (Primary Control)
Use scoped-down AWS profiles for least-privilege access:

```bash
# Read-only access (recommended for evidence collection)
export AWS_API_MCP_PROFILE_NAME="readonly-profile"

# Administrative access (use with caution)
export AWS_API_MCP_PROFILE_NAME="admin-profile"
```

#### Layer 2: Application-Level Enforcement

**Read-Only Mode** (recommended for compliance evidence):
```bash
export READ_OPERATIONS_ONLY=true
```
Blocks all non-read operations regardless of IAM permissions.

**Mutation Consent** (for interactive workflows):
```bash
export REQUIRE_MUTATION_CONSENT=true
```
Prompts user for explicit approval before mutations.

#### Layer 3: Custom Security Policies

**Denylist Configuration**:
```bash
export AWS_API_MCP_DENY_LIST="emr ssh,emr socks,deploy install"
```

**Elicitation-Required Commands**:
```bash
export AWS_API_MCP_ELICIT_LIST="ec2 terminate-instances,rds delete-db-instance"
```

#### Layer 4: File System Restrictions

**Working Directory Constraint**:
```bash
export AWS_API_MCP_WORKING_DIR="/safe/workspace"
```

### Installation Methods

#### 1. Python pip
```bash
pip install aws-api-mcp-server
aws-api-mcp-server
```

#### 2. uv (recommended)
```bash
uvx aws-api-mcp-server
```

#### 3. Docker
```bash
docker run -e AWS_PROFILE=default \
  -v ~/.aws:/root/.aws \
  public.ecr.aws/awslabs/aws-api-mcp-server:latest
```

### Claude Desktop Configuration

To enable AWS MCP Server in Claude Desktop, add this to your Claude config:

**Location**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

```json
{
  "mcpServers": {
    "aws-api": {
      "command": "uvx",
      "args": ["aws-api-mcp-server"],
      "env": {
        "AWS_API_MCP_PROFILE_NAME": "readonly-profile",
        "READ_OPERATIONS_ONLY": "true",
        "AWS_REGION": "us-west-2"
      }
    }
  }
}
```

### sdek-cli Integration Configuration

For sdek-cli to use AWS MCP for evidence collection, configure in `~/.sdek/mcp-config.yaml`:

```yaml
servers:
  aws-api:
    command: "uvx"
    args: ["aws-api-mcp-server"]
    transport: "stdio"
    env:
      AWS_API_MCP_PROFILE_NAME: "compliance-readonly"
      READ_OPERATIONS_ONLY: "true"
      AWS_REGION: "us-east-1"
      # Optional: restrict file operations
      AWS_API_MCP_WORKING_DIR: "/tmp/sdek-evidence"
    capabilities:
      - "call_aws"
      - "suggest_aws_commands"
```

### Docker-based Configuration (Alternative)

If you prefer Docker isolation:

```yaml
servers:
  aws-api-docker:
    command: "docker"
    args:
      - "run"
      - "--rm"
      - "--interactive"
      - "--env"
      - "AWS_REGION=us-east-1"
      - "--env"
      - "READ_OPERATIONS_ONLY=true"
      - "--env"
      - "AWS_PROFILE=default"
      - "--volume"
      - "$HOME/.aws:/app/.aws"
      - "public.ecr.aws/awslabs/aws-api-mcp-server:latest"
    transport: "stdio"
    capabilities:
      - "call_aws"
      - "suggest_aws_commands"
```

### Usage Examples

#### Example 1: Collect IAM Policy Evidence

```bash
# Using sdek-cli with AWS MCP for SOC2 CC6.1 (Logical Access)
sdek analyze --framework soc2 --control CC6.1 \
  --evidence-source aws-mcp \
  --query "iam list-policies,iam list-roles"
```

The AWS MCP server will:
1. Execute `aws iam list-policies`
2. Execute `aws iam list-roles`
3. Return structured results to sdek-cli
4. sdek-cli normalizes into EvidenceEvent format
5. AI analysis runs with policy context injection

#### Example 2: EKS Security Audit

```bash
# Collect EKS cluster configuration for ISO 27001 A.9.4.2
sdek analyze --framework iso27001 --control A.9.4.2 \
  --evidence-source aws-mcp \
  --query "eks describe-cluster --name production,iam get-role --role-name eks-service-role"
```

#### Example 3: Autonomous Mode with AWS

```bash
# Let AI propose AWS evidence collection plan
sdek ai plan --framework pci-dss --requirement 8.2.3 \
  --autonomous \
  --sources aws-mcp,github,jira
```

The system will:
1. Analyze requirement 8.2.3 (Multi-factor authentication)
2. Propose AWS queries: IAM MFA policies, user MFA status
3. Propose GitHub queries: authentication code
4. Present plan for approval
5. Execute approved items via respective MCP servers
6. Run context mode analysis with collected evidence

### Compliance Use Cases

#### SOC2 CC6.1 - Logical Access Security
**AWS Evidence Required:**
- IAM user list and policies
- MFA enforcement status
- Access key rotation audit
- CloudTrail logging configuration

**MCP Queries:**
```bash
aws iam list-users
aws iam get-account-password-policy
aws iam generate-credential-report
aws cloudtrail describe-trails
```

#### ISO 27001 A.12.4.1 - Event Logging
**AWS Evidence Required:**
- CloudWatch logs configuration
- VPC Flow Logs status
- S3 access logging
- CloudTrail event history

**MCP Queries:**
```bash
aws logs describe-log-groups
aws ec2 describe-flow-logs
aws cloudtrail lookup-events --start-time <30-days-ago>
```

#### PCI DSS 8.2.3 - Multi-Factor Authentication
**AWS Evidence Required:**
- IAM users without MFA
- Root account MFA status
- Console access policies

**MCP Queries:**
```bash
aws iam get-credential-report
aws iam list-virtual-mfa-devices
aws iam get-account-summary
```

### Security Best Practices

1. **Always use read-only IAM profiles** for evidence collection
2. **Enable READ_OPERATIONS_ONLY=true** in environment
3. **Restrict AWS_API_MCP_WORKING_DIR** to prevent file system access
4. **Use AWS_API_MCP_DENY_LIST** to block dangerous operations
5. **Rotate AWS credentials** used by MCP server regularly
6. **Audit AWS CloudTrail logs** for MCP server activity
7. **Never expose HTTP mode** without proper authentication

### Troubleshooting

#### Issue: AWS credentials not found
```bash
# Verify AWS CLI is configured
aws sts get-caller-identity

# Check profile exists
cat ~/.aws/credentials | grep compliance-readonly

# Test MCP server directly
uvx aws-api-mcp-server
```

#### Issue: Permission denied errors
```bash
# Verify IAM permissions
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::ACCOUNT:role/compliance-readonly \
  --action-names iam:ListUsers iam:ListPolicies
```

#### Issue: Docker volume mount fails
```bash
# Ensure .aws directory exists and has correct permissions
ls -la ~/.aws
chmod 600 ~/.aws/credentials
```

### References

- **AWS MCP Server GitHub**: https://github.com/awslabs/mcp
- **AWS MCP Documentation**: https://awslabs.github.io/mcp/servers/aws-api-mcp-server
- **MCP Specification**: https://spec.modelcontextprotocol.io
- **sdek-cli MCP Architecture**: See `docs/architecture/mcp-integration.md` (after feature 006 implementation)

### Next Steps

After configuring AWS MCP Server:

1. Test connection: `sdek mcp test aws-api`
2. List available tools: `sdek mcp list-tools`
3. Run demo analysis: `sdek analyze --demo --evidence-source aws-mcp`
4. Configure autonomous mode: Edit `~/.sdek/config.yaml` (see feature 003 spec)

---

**Related Features:**
- Feature 003: AI Context Injection & Autonomous Evidence Collection
- Feature 004: MCP Native Agent (planned)
- Feature 006: MCP Pluggable Architecture (current)
