# Viewing MCP Responses and Verifying Evidence Usage

## How to View MCP Responses

### 1. Enable Debug Logging

Run your evidence collection with `--log-level debug`:

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file policy.json \
  --config ai_config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee debug.log
```

### 2. View MCP Response Logs

The debug logs show detailed information about each MCP invocation:

```bash
# View all MCP-related logs
grep "MCP" debug.log

# View just MCP responses
grep "MCP tool response received" debug.log

# View evidence details
grep "MCP evidence details" debug.log
```

**Example Output:**

```json
{
  "time": "2025-10-19T21:42:03.621731+03:00",
  "level": "DEBUG",
  "msg": "MCP tool response received",
  "tool": "aws-api",
  "method": "tools/call",
  "has_result": false,
  "response_keys": ["content", "isError"]
}

{
  "time": "2025-10-19T21:42:03.621792+03:00",
  "level": "DEBUG",
  "msg": "MCP evidence details",
  "tool": "aws-api",
  "evidence_id": "494d4142-d5a7-4d9e-a0d3-ee3e59f849f9",
  "reasoning": "Evidence collected via MCP tool 'aws-api' using method 'tools/call'",
  "keywords": ["mcp-evidence"],
  "confidence_score": 70,
  "analysis_method": "mcp-direct"
}
```

### 3. Extract MCP Invocations

```bash
# Count successful MCP invocations
grep "MCP tool invocation successful" debug.log | wc -l

# Show all successful invocations with evidence IDs
grep "MCP tool invocation successful" debug.log | jq -r '.evidence_id'

# Show what sources were queried
grep "Using MCP tool for evidence collection" debug.log | jq -r '.source'
```

## How to Verify AI Used MCP Evidence

### 1. Check Citations in Findings

The AI's findings include citations that explicitly show MCP evidence was used:

```bash
cat findings.json | jq '.citations'
```

**Output:**
```json
[
  "AWS IAM/mcp-evidence",
  "AWS CloudTrail/mcp-evidence"
]
```

The `/mcp-evidence` suffix **proves** the evidence came from MCP tools!

### 2. Check AI Summary

The AI explicitly mentions the MCP tool in its analysis:

```bash
cat findings.json | jq -r '.summary'
```

**Example Output:**
```
The entity appears to have implemented logical access security measures in line 
with SOC2 CC6.1. Evidence collected from AWS IAM and CloudTrail via the MCP tool 
'aws-api' suggests that the entity is monitoring access controls...
```

Look for phrases like:
- "Evidence collected from ... via the MCP tool"
- "MCP tool 'aws-api' suggests"
- "Data from MCP sources indicates"

### 3. Compare Evidence Count

Verify the number of MCP invocations matches evidence collected:

```bash
# MCP invocations
grep "MCP tool invocation successful" debug.log | wc -l
# Output: 6

# Evidence collected
cat findings.json | jq '.evidence_count // "N/A"'
# Or check the summary
grep "Evidence collected" debug.log
# Output: {"msg":"Evidence collected","events":6}
```

### 4. Check Confidence Score

MCP-collected evidence typically has higher confidence scores:

```bash
cat findings.json | jq '{
  confidence: .confidence_score,
  severity: .severity,
  evidence_sources: .citations
}'
```

**Output:**
```json
{
  "confidence": 0.9,
  "severity": "low",
  "evidence_sources": [
    "AWS IAM/mcp-evidence",
    "AWS CloudTrail/mcp-evidence"
  ]
}
```

Confidence of 0.9 (90%) indicates strong evidence from MCP sources!

## Complete Verification Workflow

### Step 1: Run with Debug Logging

```bash
./sdek ai plan \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file test_policy.json \
  --config test_ai_config.yaml \
  --approve-all \
  --log-level debug 2>&1 | tee full_debug.log
```

### Step 2: Verify MCP Registry Initialized

```bash
grep "MCP tools registered" full_debug.log
# Output: {"msg":"MCP tools registered","count":1}

grep "MCP tool available" full_debug.log
# Output: {"msg":"MCP tool available","name":"aws-api","status":"ready","enabled":true}
```

✅ This confirms MCP system is active!

### Step 3: Verify MCP Tools Were Called

```bash
# Count invocations
INVOCATIONS=$(grep "MCP tool invocation successful" full_debug.log | wc -l | tr -d ' ')
echo "Total MCP invocations: $INVOCATIONS"

# Show evidence IDs
echo "Evidence IDs collected:"
grep "MCP tool invocation successful" full_debug.log | jq -r '.evidence_id'
```

✅ This shows MCP tools were actually executed!

### Step 4: Verify Evidence Was Collected

```bash
grep "Evidence collected" full_debug.log
# Output: {"msg":"Evidence collected","events":6}
```

✅ This confirms evidence was returned!

### Step 5: Verify AI Used the Evidence

```bash
# Check citations
echo "Citations in findings:"
cat findings.json | jq '.citations'

# Check summary mentions MCP
echo -e "\nAI Summary excerpt:"
cat findings.json | jq -r '.summary' | head -c 300

# Check confidence
echo -e "\nConfidence Score:"
cat findings.json | jq '.confidence_score'
```

**Expected Output:**
```
Citations in findings:
[
  "AWS IAM/mcp-evidence",
  "AWS CloudTrail/mcp-evidence"
]

AI Summary excerpt:
The entity appears to have implemented logical access security measures in line with SOC2 CC6.1. Evidence collected from AWS IAM and CloudTrail via the MCP tool 'aws-api' suggests that the entity is monitoring access controls and maintaining audit logs...

Confidence Score:
0.9
```

✅ This **proves** the AI used MCP evidence in its analysis!

## Understanding the Evidence Flow

```
┌─────────────────────┐
│  AI Plan Generation │
│  (OpenAI GPT-4)     │
└──────────┬──────────┘
           │
           │ Generates plan:
           │ "Collect IAM users from AWS"
           │
           ▼
┌─────────────────────┐
│ MCP Registry        │
│ (aws-api tool)      │
└──────────┬──────────┘
           │
           │ Invokes: tools/call
           │ Args: {cli_command: "aws iam list-users"}
           │
           ▼
┌─────────────────────┐
│ AWS MCP Server      │
│ (Docker container)  │
└──────────┬──────────┘
           │
           │ Executes AWS CLI
           │ Returns: JSON response
           │
           ▼
┌─────────────────────┐
│ Evidence Normalizer │
│ (normalizeEvidence) │
└──────────┬──────────┘
           │
           │ Creates Evidence object
           │ ID: uuid
           │ Keywords: ["mcp-evidence"]
           │ Confidence: 70
           │
           ▼
┌─────────────────────┐
│ Evidence Event      │
│ (convertToEvents)   │
└──────────┬──────────┘
           │
           │ Source: "AWS IAM"
           │ Type: "mcp-evidence"
           │ Content: reasoning
           │
           ▼
┌─────────────────────┐
│ AI Analysis         │
│ (OpenAI GPT-4)      │
└──────────┬──────────┘
           │
           │ Analyzes all evidence
           │ Generates summary
           │ Assigns confidence
           │
           ▼
┌─────────────────────┐
│ Finding Output      │
│ (findings.json)     │
└─────────────────────┘
  Citations:
  - "AWS IAM/mcp-evidence" ← PROOF!
```

## Debug Log Analysis Examples

### Example 1: Successful MCP Flow

```json
// 1. MCP tool registered
{"msg":"MCP tools registered","count":1}
{"msg":"MCP tool available","name":"aws-api","status":"ready"}

// 2. AI plans to use MCP
{"msg":"Using new MCP registry for evidence collection (Feature 004)"}

// 3. Evidence collection starts
{"msg":"Executing evidence collection plan"}
{"msg":"Using MCP tool for evidence collection","tool":"aws-api","source":"AWS IAM"}

// 4. MCP tool invoked
{"level":"DEBUG","msg":"Invoking MCP tool","tool":"aws-api","args":{...}}

// 5. Response received
{"level":"DEBUG","msg":"MCP tool response received","tool":"aws-api","has_result":true}

// 6. Evidence normalized
{"level":"DEBUG","msg":"MCP evidence details","evidence_id":"...","confidence_score":70}

// 7. Success
{"msg":"MCP tool invocation successful","tool":"aws-api","evidence_id":"..."}
{"msg":"Evidence collected","events":6}

// 8. AI analyzes
{"msg":"Analyzing collected evidence"}
```

### Example 2: Verifying No Fallback to Legacy Connectors

If you see this, it means legacy connectors were used (BAD):

```json
{"msg":"Using legacy connectors for evidence collection"}  // ❌ WRONG!
```

You should see this instead:

```json
{"msg":"Using new MCP registry for evidence collection (Feature 004)"}  // ✅ CORRECT!
```

## Automated Verification Script

Create `verify-mcp-evidence.sh`:

```bash
#!/bin/bash
# Verify MCP evidence was collected and used

LOG_FILE="$1"
FINDINGS_FILE="${2:-findings.json}"

if [ ! -f "$LOG_FILE" ]; then
    echo "❌ Log file not found: $LOG_FILE"
    exit 1
fi

if [ ! -f "$FINDINGS_FILE" ]; then
    echo "❌ Findings file not found: $FINDINGS_FILE"
    exit 1
fi

echo "Verifying MCP Evidence Collection..."
echo ""

# Check 1: MCP Registry
if grep -q "MCP tools registered" "$LOG_FILE"; then
    COUNT=$(grep "MCP tools registered" "$LOG_FILE" | jq -r '.count')
    echo "✅ MCP Registry initialized with $COUNT tool(s)"
else
    echo "❌ MCP Registry not initialized"
    exit 1
fi

# Check 2: MCP Invocations
INVOCATIONS=$(grep "MCP tool invocation successful" "$LOG_FILE" | wc -l | tr -d ' ')
if [ "$INVOCATIONS" -gt 0 ]; then
    echo "✅ MCP tools invoked $INVOCATIONS time(s)"
else
    echo "❌ No MCP tool invocations found"
    exit 1
fi

# Check 3: Evidence Collection
if grep -q "Evidence collected" "$LOG_FILE"; then
    EVENTS=$(grep "Evidence collected" "$LOG_FILE" | jq -r '.events')
    echo "✅ Evidence collected: $EVENTS event(s)"
else
    echo "❌ No evidence collected"
    exit 1
fi

# Check 4: MCP Citations
MCP_CITATIONS=$(cat "$FINDINGS_FILE" | jq -r '.citations[]' | grep "mcp-evidence" | wc -l | tr -d ' ')
if [ "$MCP_CITATIONS" -gt 0 ]; then
    echo "✅ AI used $MCP_CITATIONS MCP evidence source(s)"
    cat "$FINDINGS_FILE" | jq -r '.citations[]' | grep "mcp-evidence" | sed 's/^/   - /'
else
    echo "❌ AI did not use MCP evidence (no mcp-evidence citations)"
    exit 1
fi

# Check 5: Confidence Score
CONFIDENCE=$(cat "$FINDINGS_FILE" | jq -r '.confidence_score')
echo "✅ Confidence Score: $CONFIDENCE"

echo ""
echo "🎉 All checks passed! MCP evidence was collected and used."
```

Usage:

```bash
chmod +x verify-mcp-evidence.sh
./verify-mcp-evidence.sh debug.log findings.json
```

## Summary

To verify MCP evidence collection and usage:

1. ✅ **Run with debug logging**: `--log-level debug`
2. ✅ **Check MCP registry**: `grep "MCP tools registered"`
3. ✅ **Verify invocations**: `grep "MCP tool invocation successful"`
4. ✅ **Check citations**: `cat findings.json | jq '.citations'`
5. ✅ **Verify summary**: AI mentions "MCP tool" in analysis
6. ✅ **Confirm confidence**: Higher scores indicate good evidence

The **definitive proof** is the `/mcp-evidence` suffix in citations!
