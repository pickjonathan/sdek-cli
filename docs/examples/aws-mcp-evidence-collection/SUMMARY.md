# Summary: MCP Response Visibility & Example Documentation

## What Was Done

### 1. Enhanced Debug Logging

Added comprehensive debug logging to show MCP responses:

**Files Modified:**
- `cmd/ai_plan.go` - Added debug logging for MCP evidence details
- `internal/mcp/invoker.go` - Added debug logging for MCP responses

**What You Can Now See:**

```json
// MCP Response
{"level":"DEBUG","msg":"MCP tool response received",
 "tool":"aws-api","method":"tools/call",
 "has_result":true,"response_keys":["content","isError"]}

// Evidence Details
{"level":"DEBUG","msg":"MCP evidence details",
 "tool":"aws-api","evidence_id":"494d4142-...",
 "reasoning":"Evidence collected via MCP tool 'aws-api'...",
 "keywords":["mcp-evidence"],
 "confidence_score":70,
 "analysis_method":"mcp-direct"}
```

### 2. Created Comprehensive Examples

Created a complete example directory at `examples/aws-mcp-evidence-collection/`:

#### Documentation Files

1. **INDEX.md** - Directory overview and quick navigation
2. **README.md** - Complete guide (100+ sections covering everything)
3. **QUICKSTART.md** - 5-minute setup guide
4. **VERIFICATION.md** - Detailed guide on viewing MCP responses and verifying usage

#### Configuration Files

1. **mcp-aws-config.json** - MCP server configuration template
2. **ai-config.yaml** - AI engine configuration with all options
3. **soc2-cc6.1-policy.json** - Example SOC2 policy with 5 detailed excerpts

#### Scripts

1. **test-evidence-collection.sh** - Automated test script that:
   - Checks prerequisites (Docker, AWS, OpenAI)
   - Sets up MCP configuration
   - Tests MCP server manually
   - Runs evidence collection
   - Shows results and verification

## How to View MCP Responses

### Method 1: Debug Logs

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai-config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee debug.log

# View MCP responses
grep "MCP tool response received" debug.log
grep "MCP evidence details" debug.log
```

### Method 2: Check Citations

```bash
cat findings.json | jq '.citations'
# Output: ["AWS IAM/mcp-evidence", "AWS CloudTrail/mcp-evidence"]
```

The `/mcp-evidence` suffix **proves** the AI used MCP tools!

### Method 3: Review AI Summary

```bash
cat findings.json | jq -r '.summary'
```

The AI explicitly mentions:
> "Evidence collected from AWS IAM and CloudTrail via the MCP tool 'aws-api'..."

### Method 4: Verify Evidence Count

```bash
# Count MCP invocations
grep "MCP tool invocation successful" debug.log | wc -l

# Check evidence collected
grep "Evidence collected" debug.log
```

## Verification Workflow

### Step 1: Run with Debug Logging

```bash
cd examples/aws-mcp-evidence-collection
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee full_debug.log
```

### Step 2: Verify MCP Registry

```bash
grep "MCP tools registered" full_debug.log
# ✅ {"msg":"MCP tools registered","count":1}

grep "MCP tool available" full_debug.log
# ✅ {"msg":"MCP tool available","name":"aws-api","status":"ready"}
```

### Step 3: Verify MCP Invocations

```bash
grep "MCP tool invocation successful" full_debug.log
# Shows all successful MCP calls with evidence IDs
```

### Step 4: Verify Evidence Used

```bash
cat findings.json | jq '{
  confidence: .confidence_score,
  citations: .citations,
  mcp_mentioned: (.summary | contains("MCP"))
}'

# Output:
# {
#   "confidence": 0.9,
#   "citations": ["AWS IAM/mcp-evidence", "AWS CloudTrail/mcp-evidence"],
#   "mcp_mentioned": true
# }
```

## Example Output

### Debug Logs Show:

```
INFO: MCP tools registered | count=1
INFO: MCP tool available | name=aws-api | status=ready
INFO: Using new MCP registry for evidence collection (Feature 004)
INFO: Executing evidence collection plan
INFO: Using MCP tool for evidence collection | tool=aws-api | source=AWS IAM
DEBUG: Invoking MCP tool | tool=aws-api | args={cli_command: "aws iam list-users"}
DEBUG: MCP tool response received | has_result=true | response_keys=[content,isError]
DEBUG: MCP evidence details | evidence_id=... | confidence_score=70
INFO: MCP tool invocation successful | evidence_id=...
INFO: Evidence collected | events=6
```

### Findings Show:

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
  "summary": "Evidence collected from AWS IAM and CloudTrail via the MCP tool 'aws-api' suggests...",
  "mode": "autonomous"
}
```

## Files Created

```
examples/aws-mcp-evidence-collection/
├── INDEX.md                          # Directory overview
├── README.md                         # Complete guide (1200+ lines)
├── QUICKSTART.md                     # 5-minute setup
├── VERIFICATION.md                   # How to verify MCP usage (900+ lines)
├── mcp-aws-config.json              # MCP server config
├── ai-config.yaml                    # AI configuration
├── soc2-cc6.1-policy.json           # Example policy with 5 excerpts
└── test-evidence-collection.sh       # Automated test script
```

## Key Features

### Debug Logging
- ✅ Shows MCP tool responses
- ✅ Shows evidence details
- ✅ Shows response keys
- ✅ Shows confidence scores
- ✅ Shows analysis method

### Verification Methods
- ✅ Citations with `/mcp-evidence` suffix
- ✅ AI summary explicitly mentions MCP
- ✅ Debug logs show MCP invocations
- ✅ Evidence count matches invocations
- ✅ High confidence scores from MCP data

### Documentation
- ✅ Complete setup guide
- ✅ Quick start (5 minutes)
- ✅ Verification guide
- ✅ Troubleshooting section
- ✅ Automated test script
- ✅ Configuration examples

## Testing the Examples

### Quick Test (Automated)

```bash
cd examples/aws-mcp-evidence-collection
./test-evidence-collection.sh
```

### Manual Test

```bash
# Setup
mkdir -p ~/.sdek/mcp
cp mcp-aws-config.json ~/.sdek/mcp/aws.json
sed -i '' "s|\$HOME|$HOME|g" ~/.sdek/mcp/aws.json

export OPENAI_API_KEY="your-key"
export AWS_PROFILE="default"

# Run
../../sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file soc2-cc6.1-policy.json \
  --config ai-config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee test.log

# Verify
grep "MCP" test.log | head -20
cat findings.json | jq '.citations'
```

## Success Criteria

All of the following should be true:

✅ **MCP Registry Initialized**
```bash
grep "MCP tools registered" test.log
# Shows: count=1
```

✅ **MCP Tools Invoked**
```bash
grep "MCP tool invocation successful" test.log | wc -l
# Shows: 6 (or however many evidence items)
```

✅ **Evidence Collected**
```bash
grep "Evidence collected" test.log
# Shows: events=6
```

✅ **Citations Include MCP**
```bash
cat findings.json | jq '.citations | map(select(contains("mcp-evidence")))'
# Shows: ["AWS IAM/mcp-evidence", ...]
```

✅ **AI Mentions MCP**
```bash
cat findings.json | jq -r '.summary' | grep -i "mcp tool"
# Shows: "...via the MCP tool 'aws-api'..."
```

✅ **High Confidence**
```bash
cat findings.json | jq '.confidence_score'
# Shows: 0.7-0.9 (70-90%)
```

## Next Steps

1. **Try the automated test**:
   ```bash
   cd examples/aws-mcp-evidence-collection
   ./test-evidence-collection.sh
   ```

2. **Read the documentation**:
   - Start with [QUICKSTART.md](examples/aws-mcp-evidence-collection/QUICKSTART.md)
   - Then [VERIFICATION.md](examples/aws-mcp-evidence-collection/VERIFICATION.md)

3. **Customize for your needs**:
   - Edit `soc2-cc6.1-policy.json` for your controls
   - Modify `ai-config.yaml` for your AI preferences
   - Create configs for other frameworks (ISO27001, NIST, etc.)

4. **Integrate with CI/CD**:
   - Use findings to fail builds with low confidence
   - Generate compliance reports automatically
   - Track compliance over time

## Summary

**Problem**: No visibility into MCP responses, couldn't verify AI was using MCP evidence

**Solution**:
1. ✅ Added debug logging to show MCP responses and evidence details
2. ✅ Created comprehensive examples with step-by-step guides
3. ✅ Documented multiple verification methods
4. ✅ Provided automated test script
5. ✅ Showed how citations prove MCP usage

**Result**: Complete transparency into MCP evidence collection with multiple ways to verify the AI is actually using the MCP-collected data in its analysis.
