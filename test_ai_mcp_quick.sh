#!/bin/bash
# Quick test for AI + AWS MCP integration

set -e

echo "🧪 Testing AI + AWS MCP Integration"
echo "======================================"
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

if [ -z "$SDEK_OPENAI_KEY" ]; then
    echo "❌ SDEK_OPENAI_KEY not set"
    echo ""
    echo "Set your OpenAI API key first:"
    echo "  export SDEK_OPENAI_KEY='sk-...'"
    echo ""
    exit 1
fi
echo "✅ OpenAI API key found"

if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured"
    echo ""
    echo "Configure AWS credentials:"
    echo "  aws configure"
    echo ""
    exit 1
fi
echo "✅ AWS credentials configured"

if [ ! -f "./sdek" ]; then
    echo "❌ sdek binary not found"
    echo ""
    echo "Build the binary first:"
    echo "  make build"
    echo ""
    exit 1
fi
echo "✅ sdek binary found"

echo ""
echo "📝 Test policy excerpt: test_policy.json"
echo "⚙️  Test config: test_ai_config.yaml"
echo ""

# Test 1: Dry run to see the plan
echo "===================================="
echo "Test 1: Generate AI plan (dry-run)"
echo "===================================="
echo ""
echo "This will show you what the AI proposes without executing..."
echo ""

./sdek --config test_ai_config.yaml ai plan \
    --framework SOC2 \
    --section CC6.1 \
    --excerpts-file test_policy.json \
    --dry-run

echo ""
echo "✅ Dry run complete!"
echo ""
read -p "Did you see AWS sources in the plan? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "⚠️  The AI didn't propose AWS sources."
    echo ""
    echo "Possible reasons:"
    echo "1. The policy excerpt doesn't mention AWS-related controls"
    echo "2. Confidence threshold is too high (try lowering to 0.5)"
    echo "3. AWS connector not properly configured"
    echo ""
    echo "Try editing test_policy.json to explicitly mention:"
    echo "  - AWS IAM"
    echo "  - CloudTrail"
    echo "  - Access control monitoring"
    echo ""
    exit 0
fi

echo ""
echo "===================================="
echo "Test 2: Execute with real MCP calls"
echo "===================================="
echo ""
echo "This will actually call AWS MCP tools..."
echo ""
read -p "Ready to execute? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipped execution test."
    exit 0
fi

echo ""
echo "Running with debug logging..."
echo ""

./sdek --config test_ai_config.yaml ai plan \
    --framework SOC2 \
    --section CC6.1 \
    --excerpts-file test_policy.json \
    --approve-all \
    --output test_finding.json

echo ""
echo "===================================="
echo "Test 3: Verify AWS evidence collected"
echo "===================================="
echo ""

if [ -f "test_finding.json" ]; then
    echo "📄 Finding saved to: test_finding.json"
    echo ""
    
    if command -v jq &> /dev/null; then
        echo "AWS Evidence Sources:"
        cat test_finding.json | jq '.evidence_sources[] | select(.type == "aws")'
        echo ""
    else
        echo "Install jq to parse the JSON output:"
        echo "  brew install jq"
        echo ""
        echo "Or view manually:"
        echo "  cat test_finding.json"
    fi
else
    echo "⚠️  No finding file created"
fi

echo ""
echo "===================================="
echo "Test 4: Check MCP audit logs"
echo "===================================="
echo ""

if [ -f "$HOME/.sdek/mcp/audit.log" ]; then
    echo "📋 Recent MCP calls to aws-api:"
    if command -v jq &> /dev/null; then
        cat ~/.sdek/mcp/audit.log | jq 'select(.tool_name == "aws-api")'
    else
        grep "aws-api" ~/.sdek/mcp/audit.log || echo "No aws-api calls found"
    fi
else
    echo "⚠️  No audit log found at ~/.sdek/mcp/audit.log"
fi

echo ""
echo "✅ Testing complete!"
echo ""
echo "Summary:"
echo "- Check test_finding.json for AWS evidence"
echo "- Review ~/.sdek/mcp/audit.log for MCP calls"
echo "- Look for 'Executing MCP tool' in the output above"
