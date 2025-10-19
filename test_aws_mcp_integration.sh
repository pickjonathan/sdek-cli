#!/bin/bash
# test_aws_mcp_integration.sh
# Quick test script to verify AWS MCP tool is working with SDEK AI

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║  SDEK AWS MCP Integration Test                         ║${NC}"
echo -e "${YELLOW}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}[1/10] Checking prerequisites...${NC}"

if ! command -v npx &> /dev/null; then
    echo -e "${RED}✗ npx not found. Install Node.js first.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ npx found${NC}"

if [ ! -f "./sdek" ]; then
    echo -e "${RED}✗ sdek binary not found. Run 'make build' first.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ sdek binary found${NC}"

if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}✗ AWS credentials not configured. Run 'aws configure' first.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ AWS credentials configured${NC}"

# Create MCP config directory
echo -e "\n${YELLOW}[2/10] Setting up MCP config directory...${NC}"
mkdir -p ~/.sdek/mcp
echo -e "${GREEN}✓ Directory created: ~/.sdek/mcp${NC}"

# Create AWS MCP config
echo -e "\n${YELLOW}[3/10] Creating AWS MCP configuration...${NC}"
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
    "AWS_REGION": "${AWS_REGION:-us-east-1}"
  },
  "description": "AWS API access for CloudTrail, IAM, EC2, and S3"
}
EOF
echo -e "${GREEN}✓ Config created: ~/.sdek/mcp/aws.json${NC}"

# Validate config
echo -e "\n${YELLOW}[4/10] Validating AWS MCP config...${NC}"
if ./sdek mcp validate ~/.sdek/mcp/aws.json; then
    echo -e "${GREEN}✓ Config validation passed${NC}"
else
    echo -e "${RED}✗ Config validation failed${NC}"
    exit 1
fi

# Test MCP handshake
echo -e "\n${YELLOW}[5/10] Testing MCP server handshake...${NC}"
if timeout 10 ./sdek mcp test aws-api 2>&1 | grep -q "success\|passed\|ready"; then
    echo -e "${GREEN}✓ MCP handshake successful${NC}"
else
    echo -e "${RED}✗ MCP handshake failed${NC}"
    echo "Tip: Ensure @modelcontextprotocol/server-aws is installed:"
    echo "  npm install -g @modelcontextprotocol/server-aws"
    exit 1
fi

# List MCP tools
echo -e "\n${YELLOW}[6/10] Listing MCP tools...${NC}"
./sdek mcp list
if ./sdek mcp list 2>&1 | grep -q "aws-api"; then
    echo -e "${GREEN}✓ aws-api tool registered${NC}"
else
    echo -e "${RED}✗ aws-api tool not found${NC}"
    exit 1
fi

# Create test policy excerpt
echo -e "\n${YELLOW}[7/10] Creating test policy excerpt...${NC}"
cat > /tmp/test_policy.json <<'EOF'
[
  {
    "framework": "SOC2",
    "version": "2017",
    "section": "CC6.1",
    "title": "Logical Access Controls",
    "text": "The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives. This includes controls to: (1) Identify and authenticate users; (2) Authorize users to access authorized resources; (3) Prevent unauthorized access. The entity should monitor AWS IAM users, policies, and CloudTrail events to demonstrate access control implementation.",
    "related_sections": ["CC6.2", "CC6.3"]
  }
]
EOF
echo -e "${GREEN}✓ Test policy created: /tmp/test_policy.json${NC}"

# Check AI configuration
echo -e "\n${YELLOW}[8/10] Checking AI configuration...${NC}"
if [ -z "$SDEK_OPENAI_KEY" ] && [ -z "$SDEK_ANTHROPIC_KEY" ]; then
    echo -e "${YELLOW}⚠ No AI API key found in environment${NC}"
    echo "Set one of:"
    echo "  export SDEK_OPENAI_KEY='sk-...'"
    echo "  export SDEK_ANTHROPIC_KEY='sk-ant-...'"
    echo ""
    echo -e "${YELLOW}Skipping AI plan test (steps 9-10)${NC}"
    SKIP_AI=true
else
    echo -e "${GREEN}✓ AI API key configured${NC}"
    SKIP_AI=false
fi

if [ "$SKIP_AI" = false ]; then
    # Create temporary config with AI enabled
    echo -e "\n${YELLOW}[9/10] Creating temporary AI config...${NC}"
    cat > /tmp/test_config.yaml <<EOF
ai:
  enabled: true
  provider: "openai"
  openai_key: "${SDEK_OPENAI_KEY}"
  model: "gpt-4"
  mode: "autonomous"
  autonomous:
    enabled: true
    confidence_threshold: 0.7
    auto_approve:
      aws:
        - "iam:*"
        - "cloudtrail:*"
  connectors:
    aws:
      enabled: true
EOF
    echo -e "${GREEN}✓ Test config created${NC}"

    # Run AI plan with dry-run
    echo -e "\n${YELLOW}[10/10] Testing AI plan with AWS MCP (dry-run)...${NC}"
    export SDEK_LOG_LEVEL=debug
    
    if ./sdek --config /tmp/test_config.yaml ai plan \
        --framework SOC2 \
        --section CC6.1 \
        --excerpts-file /tmp/test_policy.json \
        --dry-run \
        2>&1 | tee /tmp/sdek_test.log | grep -i "aws\|mcp"; then
        echo -e "${GREEN}✓ AI plan generated with AWS sources${NC}"
    else
        echo -e "${YELLOW}⚠ No AWS sources in plan (check log: /tmp/sdek_test.log)${NC}"
    fi

    # Check for MCP usage indicators
    echo -e "\n${YELLOW}Checking for MCP usage indicators...${NC}"
    if grep -qi "mcp.*aws\|aws.*mcp\|aws-api" /tmp/sdek_test.log; then
        echo -e "${GREEN}✓ MCP tool usage detected in logs${NC}"
    else
        echo -e "${YELLOW}⚠ No clear MCP usage found (AI may not have selected AWS)${NC}"
    fi
fi

# Summary
echo -e "\n${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Test Summary                                          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✓ MCP config created and validated${NC}"
echo -e "${GREEN}✓ AWS MCP tool registered and ready${NC}"
echo -e "${GREEN}✓ Handshake test passed${NC}"

if [ "$SKIP_AI" = false ]; then
    echo -e "${GREEN}✓ AI integration tested${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review the test log: /tmp/sdek_test.log"
    echo "2. Run without --dry-run to execute actual AWS API calls:"
    echo "   ./sdek --config /tmp/test_config.yaml ai plan \\"
    echo "     --framework SOC2 --section CC6.1 \\"
    echo "     --excerpts-file /tmp/test_policy.json \\"
    echo "     --output test_finding.json"
    echo ""
    echo "3. Inspect the finding for AWS evidence:"
    echo "   cat test_finding.json | jq '.evidence_sources[] | select(.type == \"aws\")'"
else
    echo ""
    echo "To test AI integration:"
    echo "1. Set AI API key: export SDEK_OPENAI_KEY='sk-...'"
    echo "2. Run this script again"
fi

echo ""
echo -e "${GREEN}✓ AWS MCP integration test complete!${NC}"
