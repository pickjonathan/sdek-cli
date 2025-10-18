# Quickstart: AI Context Injection & Autonomous Evidence Collection

**Feature**: 003-ai-context-injection  
**Purpose**: Step-by-step guide to using context mode and autonomous mode

---

## Prerequisites

1. **Configure AI Provider** (`config.yaml`):
   ```yaml
   ai:
     provider: anthropic  # or openai
     apiKey: ${ANTHROPIC_API_KEY}  # or ${OPENAI_API_KEY}
     mode: context  # disabled|context|autonomous
     
     concurrency:
       maxAnalyses: 25
     
     budgets:
       maxSources: 50
       maxAPICalls: 500
       maxTokens: 250000
     
     autonomous:
       enabled: true
       autoApprove:
         enabled: false  # Set true to enable auto-approval
         rules:
           github: ["auth*", "*login*"]
           aws: ["iam*", "security*"]
   ```

2. **Load Framework Excerpts**:
   - Ensure `testdata/ai/policies/soc2_excerpts.json` exists
   - Or create custom excerpt file

3. **Set Up MCP Connectors** (for autonomous mode):
   - GitHub: `GITHUB_TOKEN` environment variable
   - AWS: AWS credentials configured
   - Jira: `JIRA_TOKEN` and `JIRA_URL` set

---

## Scenario 1: Context Mode Analysis (Phase 1)

**Goal**: Analyze SOC 2 CC6.1 with explicit framework context

### Step 1: Prepare Evidence
```bash
# Collect evidence manually (or use existing)
$ ls evidence/
github_auth_code.json
aws_iam_policies.json
jira_security_tickets.json
```

### Step 2: Run Context Mode Analysis
```bash
$ sdek ai analyze \
    --framework SOC2 \
    --section CC6.1 \
    --excerpts-file ./testdata/ai/policies/soc2_excerpts.json \
    --evidence-path ./evidence/*.json \
    --mode context \
    --output findings.json
```

**Expected Output**:
```
🔍 Analyzing SOC 2 CC6.1: Logical and Physical Access Controls
📋 Framework: SOC2 (2017)
📄 Excerpt: The entity restricts logical access...

🔐 Redaction: 12 PII instances, 3 secrets removed
💾 Cache: MISS (first run)
⏱️  Duration: 18.3s

✅ Finding Generated:
   Confidence: 0.82 (High)
   Residual Risk: Low
   Citations: 15 evidence events
   
📊 Export: findings.json (JSON)
```

### Step 3: Verify Finding
```bash
$ cat findings.json | jq '.confidence_score'
0.82

$ cat findings.json | jq '.review_required'
false

$ cat findings.json | jq '.mode'
"ai"
```

### Step 4: Re-run (Cache Hit)
```bash
$ sdek ai analyze \
    --framework SOC2 \
    --section CC6.1 \
    --excerpts-file ./testdata/ai/policies/soc2_excerpts.json \
    --evidence-path ./evidence/*.json \
    --mode context

# Output shows:
# 💾 Cache: HIT
# ⏱️  Duration: 0.08s
```

### Step 5: Bypass Cache (If Needed)
```bash
$ sdek ai analyze ... --no-cache
```

---

## Scenario 2: Autonomous Mode (Phase 2)

**Goal**: Generate evidence plan, approve it, collect evidence, analyze

### Step 1: Generate Evidence Plan
```bash
$ sdek ai plan \
    --framework ISO27001 \
    --section A.9.4.2 \
    --excerpts-file ./testdata/ai/policies/iso27001_excerpts.json \
    --print
```

**Expected Output** (TUI):
```
┌─ Evidence Plan: ISO 27001 A.9.4.2 ────────────────────────┐
│                                                            │
│  ✓ github: auth* (auto-approved) ─────────── Signal: 0.92 │
│  ? aws: iam* (pending) ──────────────────── Signal: 0.85  │
│  ? jira: INFOSEC-* (pending) ────────────── Signal: 0.78  │
│  ? slack: #security* (pending) ───────────── Signal: 0.65 │
│                                                            │
│  Budget: 4 sources, ~40 API calls, ~35K tokens             │
│                                                            │
│  [a] Approve All  [d] Deny  [↑↓] Navigate  [Enter] Toggle │
└────────────────────────────────────────────────────────────┘
```

### Step 2: Approve Plan Items
- Press `↑`/`↓` to navigate
- Press `Enter` to toggle approval
- Press `a` to approve all
- Press `q` to confirm and proceed

### Step 3: Execute Plan
```bash
# Plan execution starts automatically after approval
🔄 Executing Evidence Plan...

✓ github: auth* (collected 23 events)
✓ aws: iam* (collected 15 events)
✓ jira: INFOSEC-* (collected 8 events)
⏭ slack: #security* (denied by user)

📦 Total: 46 evidence events collected
⏱️  Duration: 2m 34s
```

### Step 4: Automatic Analysis
```bash
# After execution, analysis runs automatically

🔍 Analyzing ISO 27001 A.9.4.2 with collected evidence...
✅ Finding Generated:
   Confidence: 0.74 (Medium)
   Residual Risk: Medium
   Citations: 12 evidence events
   Provenance:
     - github: 12 events (contribution: 0.52)
     - aws: 8 events (contribution: 0.35)
     - jira: 4 events (contribution: 0.13)
```

### Step 5: Review Finding with Provenance
```bash
$ cat findings.json | jq '.provenance'
[
  {
    "source": "github",
    "query": "auth*",
    "events_used": 12,
    "contribution": 0.52
  },
  {
    "source": "aws",
    "query": "iam*",
    "events_used": 8,
    "contribution": 0.35
  },
  {
    "source": "jira",
    "query": "INFOSEC-*",
    "events_used": 4,
    "contribution": 0.13
  }
]
```

---

## Scenario 3: Dry-Run Mode (Preview Plan)

**Goal**: Generate and preview plan without executing

```bash
$ sdek ai plan \
    --framework PCI-DSS \
    --section 8.2.3 \
    --excerpts-file ./testdata/ai/policies/pci_excerpts.json \
    --dry-run \
    --output plan.json
```

**Output**:
```json
{
  "id": "plan-uuid",
  "framework": "PCI-DSS",
  "section": "8.2.3",
  "items": [
    {
      "source": "github",
      "query": "password*",
      "signal_strength": 0.88,
      "approval_status": "pending",
      "auto_approved": false
    },
    {
      "source": "aws",
      "query": "iam*",
      "signal_strength": 0.82,
      "approval_status": "pending",
      "auto_approved": false
    }
  ],
  "estimated_sources": 2,
  "estimated_calls": 25,
  "estimated_tokens": 18000,
  "status": "pending"
}
```

---

## Scenario 4: Low Confidence Finding (Review Required)

**Goal**: Handle findings with confidence <0.6

### Step 1: Run Analysis (Results in Low Confidence)
```bash
$ sdek ai analyze \
    --framework SOC2 \
    --section CC7.2 \
    --evidence-path ./sparse_evidence/*.json  # Minimal evidence
```

**Output**:
```
✅ Finding Generated:
   Confidence: 0.52 (Low)
   ⚠️  REVIEW REQUIRED (confidence < 0.6)
   Residual Risk: High
   Citations: 3 evidence events
```

### Step 2: Review in TUI
```
┌─ Finding: SOC 2 CC7.2 ─────────────────────────────────────┐
│                                                            │
│  ⚠️  Review Required (Confidence: 0.52)                     │
│                                                            │
│  Summary: Limited evidence suggests controls exist but     │
│  coverage is incomplete. Manual validation recommended.    │
│                                                            │
│  Residual Risk: HIGH                                       │
│  Citations: 3 events (below recommended 5+)                │
│                                                            │
│  [r] Mark Reviewed  [e] Export  [q] Quit                   │
└────────────────────────────────────────────────────────────┘
```

### Step 3: Collect More Evidence (Autonomous Mode)
```bash
# Generate plan to collect more evidence
$ sdek ai plan --framework SOC2 --section CC7.2
```

---

## Scenario 5: Fallback to Heuristics Mode

**Goal**: Handle AI provider failures gracefully

### Step 1: Simulate AI Failure
```bash
# Set invalid API key to simulate failure
$ export ANTHROPIC_API_KEY=invalid

$ sdek ai analyze \
    --framework SOC2 \
    --section CC6.1 \
    --evidence-path ./evidence/*.json
```

**Output**:
```
❌ AI Provider Error: Invalid API key

⚙️  Falling back to heuristics mode...

✅ Finding Generated (Heuristics):
   Confidence: 0.45 (Low - heuristics mode)
   ⚠️  REVIEW REQUIRED
   Mode: heuristics
   Residual Risk: Unknown
```

### Step 2: Verify Mode
```bash
$ cat findings.json | jq '.mode'
"heuristics"
```

---

## Scenario 6: Concurrent Analysis (Bulk Mode)

**Goal**: Analyze multiple controls concurrently

```bash
$ sdek ai analyze \
    --framework SOC2 \
    --sections CC6.1,CC6.2,CC6.3,CC7.1,CC7.2 \
    --evidence-path ./evidence/*.json \
    --concurrent 5 \
    --output-dir ./findings/
```

**Output**:
```
🔄 Analyzing 5 controls concurrently (limit: 5)...

✓ CC6.1 (18.2s, confidence: 0.82)
✓ CC6.2 (21.4s, confidence: 0.78)
✓ CC6.3 (19.8s, confidence: 0.85)
✓ CC7.1 (22.1s, confidence: 0.71)
✓ CC7.2 (17.9s, confidence: 0.52) ⚠️

📊 Total: 5 findings in 22.1s (parallel execution)
📁 Output: ./findings/
```

---

## Troubleshooting

### Issue: "Excerpt required for [framework] section [X]"
**Solution**: Ensure excerpt file exists and contains the section:
```bash
$ cat testdata/ai/policies/soc2_excerpts.json | jq '.CC6_1'
```

### Issue: "Budget exceeded"
**Solution**: Adjust budget limits in config.yaml:
```yaml
ai:
  budgets:
    maxSources: 100  # Increase from 50
```

### Issue: "Evidence heavily redacted, confidence may be low"
**Solution**: Review redaction patterns or provide less sensitive evidence

### Issue: Cache directory full
**Solution**: Clear old cache files:
```bash
$ rm -rf ~/.sdek/cache/ai/*.json
```

---

## Summary

This quickstart covered:
1. **Context Mode**: Analyze with framework grounding
2. **Autonomous Mode**: Generate + approve + execute + analyze
3. **Dry-Run**: Preview plans without execution
4. **Low Confidence**: Handle review-required findings
5. **Fallback**: Graceful degradation to heuristics
6. **Concurrent**: Bulk analysis with configurable limits

Ready for production use!
