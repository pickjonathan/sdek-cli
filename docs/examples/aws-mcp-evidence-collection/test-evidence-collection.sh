#!/bin/bash
# Test script for AWS MCP evidence collection
# This script demonstrates how to configure, test, and run evidence collection

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
EXAMPLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDEK_CLI="${EXAMPLE_DIR}/../../sdek"
MCP_CONFIG_DIR="${HOME}/.sdek/mcp"
MCP_CONFIG_FILE="${MCP_CONFIG_DIR}/aws.json"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}AWS MCP Evidence Collection Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Check prerequisites
echo -e "${YELLOW}Step 1: Checking prerequisites...${NC}"

# Check if sdek CLI exists
if [ ! -f "${SDEK_CLI}" ]; then
    echo -e "${RED}❌ SDEK CLI not found at ${SDEK_CLI}${NC}"
    echo -e "${YELLOW}   Run 'make build' in the project root${NC}"
    exit 1
fi
echo -e "${GREEN}✓ SDEK CLI found${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker not found${NC}"
    echo -e "${YELLOW}   Install Docker: https://docs.docker.com/get-docker/${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker installed${NC}"

# Check Docker is running
if ! docker ps &> /dev/null; then
    echo -e "${RED}❌ Docker is not running${NC}"
    echo -e "${YELLOW}   Start Docker Desktop or Docker daemon${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"

# Check AWS credentials
if [ ! -f "${HOME}/.aws/credentials" ]; then
    echo -e "${RED}❌ AWS credentials not found${NC}"
    echo -e "${YELLOW}   Run 'aws configure' to set up credentials${NC}"
    exit 1
fi
echo -e "${GREEN}✓ AWS credentials configured${NC}"

# Check OpenAI API key
if [ -z "${OPENAI_API_KEY}" ]; then
    echo -e "${RED}❌ OPENAI_API_KEY environment variable not set${NC}"
    echo -e "${YELLOW}   Export your OpenAI API key: export OPENAI_API_KEY=sk-...${NC}"
    exit 1
fi
echo -e "${GREEN}✓ OpenAI API key configured${NC}"

echo ""

# Step 2: Set up MCP configuration
echo -e "${YELLOW}Step 2: Setting up MCP configuration...${NC}"

# Create MCP config directory
mkdir -p "${MCP_CONFIG_DIR}"
echo -e "${GREEN}✓ MCP config directory created: ${MCP_CONFIG_DIR}${NC}"

# Copy MCP configuration (with variable expansion for HOME)
cat "${EXAMPLE_DIR}/mcp-aws-config.json" | \
  sed "s|\$HOME|${HOME}|g" > "${MCP_CONFIG_FILE}"
echo -e "${GREEN}✓ MCP configuration installed: ${MCP_CONFIG_FILE}${NC}"

# Show configuration
echo -e "${BLUE}MCP Configuration:${NC}"
cat "${MCP_CONFIG_FILE}" | jq '.'
echo ""

# Step 3: Test MCP server manually
echo -e "${YELLOW}Step 3: Testing MCP server...${NC}"

# Test initialize
echo -e "${BLUE}Testing MCP initialize...${NC}"
INIT_RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
  docker run --rm -i \
    --env AWS_REGION=us-east-1 \
    --env READ_OPERATIONS_ONLY=true \
    --env AWS_PROFILE=default \
    --volume "${HOME}/.aws:/app/.aws" \
    public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest)

if echo "${INIT_RESPONSE}" | jq -e '.result.serverInfo.name' > /dev/null 2>&1; then
    SERVER_NAME=$(echo "${INIT_RESPONSE}" | jq -r '.result.serverInfo.name')
    SERVER_VERSION=$(echo "${INIT_RESPONSE}" | jq -r '.result.serverInfo.version')
    echo -e "${GREEN}✓ MCP server responded: ${SERVER_NAME} v${SERVER_VERSION}${NC}"
else
    echo -e "${RED}❌ MCP server initialization failed${NC}"
    echo "${INIT_RESPONSE}"
    exit 1
fi

# Test AWS CLI execution
echo -e "${BLUE}Testing AWS CLI execution...${NC}"
CALLER_RESPONSE=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"call_aws","arguments":{"cli_command":"aws sts get-caller-identity"}}}' | \
  docker run --rm -i \
    --env AWS_REGION=us-east-1 \
    --env READ_OPERATIONS_ONLY=true \
    --env AWS_PROFILE=default \
    --volume "${HOME}/.aws:/app/.aws" \
    public.ecr.aws/awslabs-mcp/awslabs/aws-api-mcp-server:latest)

if echo "${CALLER_RESPONSE}" | jq -e '.result' > /dev/null 2>&1; then
    ACCOUNT_ID=$(echo "${CALLER_RESPONSE}" | jq -r '.result.response.json | fromjson | .Account')
    echo -e "${GREEN}✓ AWS CLI working - Account ID: ${ACCOUNT_ID}${NC}"
else
    echo -e "${RED}❌ AWS CLI execution failed${NC}"
    echo "${CALLER_RESPONSE}"
    exit 1
fi

echo ""

# Step 4: Run evidence collection (dry-run)
echo -e "${YELLOW}Step 4: Running evidence collection (plan only)...${NC}"

echo -e "${BLUE}Generating evidence collection plan...${NC}"
"${SDEK_CLI}" ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file "${EXAMPLE_DIR}/soc2-cc6.1-policy.json" \
  --config "${EXAMPLE_DIR}/ai-config.yaml" \
  --dry-run \
  --log-level info

echo ""
echo -e "${GREEN}✓ Evidence collection plan generated${NC}"
echo -e "${YELLOW}Review the plan above. To execute, run without --dry-run${NC}"
echo ""

# Step 5: Optionally run full collection
read -p "$(echo -e ${YELLOW}Do you want to run the full evidence collection? [y/N]: ${NC})" -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Step 5: Running full evidence collection...${NC}"
    
    OUTPUT_FILE="${EXAMPLE_DIR}/findings-$(date +%Y%m%d-%H%M%S).json"
    LOG_FILE="${EXAMPLE_DIR}/collection-$(date +%Y%m%d-%H%M%S).log"
    
    echo -e "${BLUE}Starting evidence collection with debug logging...${NC}"
    "${SDEK_CLI}" ai plan \
      --framework SOC2 \
      --section CC6.1 \
      --excerpts-file "${EXAMPLE_DIR}/soc2-cc6.1-policy.json" \
      --config "${EXAMPLE_DIR}/ai-config.yaml" \
      --approve-all \
      --log-level debug \
      --output "${OUTPUT_FILE}" \
      2>&1 | tee "${LOG_FILE}"
    
    echo ""
    echo -e "${GREEN}✓ Evidence collection complete!${NC}"
    echo -e "${BLUE}Findings saved to: ${OUTPUT_FILE}${NC}"
    echo -e "${BLUE}Logs saved to: ${LOG_FILE}${NC}"
    echo ""
    
    # Show summary
    echo -e "${YELLOW}Finding Summary:${NC}"
    cat "${OUTPUT_FILE}" | jq '{
      control: .control_id,
      framework: .framework_id,
      confidence: .confidence_score,
      severity: .severity,
      citations: .citations
    }'
    
    echo ""
    echo -e "${YELLOW}MCP Evidence in Logs:${NC}"
    echo -e "${BLUE}Evidence collection events:${NC}"
    grep "MCP tool invocation successful" "${LOG_FILE}" | wc -l | xargs echo "  Total successful invocations:"
    
    echo -e "${BLUE}Debug details:${NC}"
    grep "MCP evidence details" "${LOG_FILE}" | head -3
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Test Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Review findings: cat ${OUTPUT_FILE} | jq ."
    echo -e "  2. View debug logs: cat ${LOG_FILE} | grep 'MCP'"
    echo -e "  3. Generate report: ${SDEK_CLI} report --input ${OUTPUT_FILE}"
else
    echo -e "${YELLOW}Skipping full collection. Re-run this script to execute.${NC}"
fi

echo ""
echo -e "${BLUE}For more information, see README.md${NC}"
