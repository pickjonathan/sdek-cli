# Quickstart: AI Evidence Analysis

**Date**: 2025-10-11  
**Purpose**: Integration test scenarios and user validation steps

## Prerequisites

1. **SDEK CLI installed** with existing functionality (Phases 3.1-3.8 complete)
2. **AI provider API key** (OpenAI or Anthropic) exported as environment variable
3. **Demo data generated**: `sdek seed --demo` to create sample events

---

## Scenario 1: Enable AI Analysis with OpenAI

### Setup

```bash
# Export OpenAI API key
export SDEK_AI_OPENAI_KEY="sk-..."

# Configure AI provider
sdek config set ai.provider openai
sdek config set ai.enabled true
sdek config set ai.model gpt-4-turbo-preview
```

### Execute

```bash
# Run analysis with AI enabled
sdek analyze --verbose

# Expected output:
# ✓ Loaded 130 events from 5 sources
# ✓ Analyzing 124 controls across 3 frameworks
# ⚙ AI provider: openai (gpt-4-turbo-preview)
# ⚙ Analyzing control SOC2-CC1.1 with AI...
# ⚙ AI confidence: 87%, heuristic: 65%, combined: 80%
# ⚙ Analyzing control SOC2-CC1.2 with AI...
# ... (progress for each control)
# ✓ Analysis complete: 124 controls, 18 findings
# ✓ AI cache hit rate: 0% (first run)
```

### Validate

```bash
# Check report includes AI metadata
sdek report --output report.json
cat report.json | jq '.frameworks[0].controls[0]'

# Expected fields:
# {
#   "control": { ... },
#   "evidence": [
#     {
#       "id": "...",
#       "ai_analyzed": true,
#       "ai_justification": "Event demonstrates access control review process",
#       "ai_confidence": 87,
#       "heuristic_confidence": 65,
#       "combined_confidence": 80,
#       "ai_residual_risk": "No evidence of automated enforcement",
#       "analysis_method": "ai+heuristic"
#     }
#   ]
# }
```

**Success Criteria**:
- ✅ AI provider connected successfully
- ✅ All controls analyzed with AI (no fallback to heuristics)
- ✅ Combined confidence scores present (70% AI + 30% heuristic)
- ✅ Justifications and residual risk notes populated
- ✅ Report indicates "ai+heuristic" analysis method

---

## Scenario 2: AI Fallback on Provider Error

### Setup

```bash
# Use invalid API key to simulate auth failure
export SDEK_AI_OPENAI_KEY="invalid-key"
sdek config set ai.provider openai
sdek config set ai.enabled true
```

### Execute

```bash
sdek analyze --verbose

# Expected output:
# ✓ Loaded 130 events from 5 sources
# ✓ Analyzing 124 controls across 3 frameworks
# ⚠ AI provider health check failed: authentication error
# ⚠ Falling back to heuristic-only analysis
# ⚙ Analyzing control SOC2-CC1.1 with heuristics...
# ... (heuristic analysis for all controls)
# ✓ Analysis complete: 124 controls, 18 findings
# ⚠ AI was disabled due to provider errors
```

### Validate

```bash
# Check report shows heuristic-only analysis
cat report.json | jq '.frameworks[0].controls[0].evidence[0]'

# Expected fields:
# {
#   "id": "...",
#   "ai_analyzed": false,
#   "analysis_method": "heuristic-only",
#   "heuristic_confidence": 65,
#   "combined_confidence": 65  // Same as heuristic (no AI)
# }
```

**Success Criteria**:
- ✅ Auth error detected before analysis starts
- ✅ Graceful fallback to heuristics without blocking
- ✅ Analysis completes successfully with heuristic scores
- ✅ Report clearly indicates "heuristic-only" method
- ✅ No crash or incomplete results

---

## Scenario 3: Cache Reuse on Repeated Analysis

### Setup

```bash
# Ensure valid API key and AI enabled
export SDEK_AI_OPENAI_KEY="sk-..."
sdek config set ai.provider openai
sdek config set ai.enabled true

# Run first analysis (cache miss)
sdek analyze
```

### Execute

```bash
# Run second analysis immediately (no event changes)
sdek analyze --verbose

# Expected output:
# ✓ Loaded 130 events from 5 sources
# ✓ Analyzing 124 controls across 3 frameworks
# ⚙ AI provider: openai (gpt-4-turbo-preview)
# ⚙ Cache hit: SOC2-CC1.1 (reusing previous result)
# ⚙ Cache hit: SOC2-CC1.2 (reusing previous result)
# ... (cache hits for all controls)
# ✓ Analysis complete: 124 controls, 18 findings
# ✓ AI cache hit rate: 100%
```

### Validate

```bash
# Check cache directory populated
ls -lh ~/.cache/sdek/ai-cache/
# Expected: 124 JSON files (one per control)

# Inspect cache entry
cat ~/.cache/sdek/ai-cache/{cache_key}.json | jq '.'
# Expected fields:
# {
#   "cache_key": "a3f5d9...",
#   "response": { ... },
#   "cached_at": "2025-10-11T10:00:00Z",
#   "event_ids": ["evt-1", "evt-2"],
#   "control_id": "SOC2-CC1.1",
#   "provider": "openai",
#   "model_version": "gpt-4-turbo-preview"
# }
```

**Success Criteria**:
- ✅ First run makes AI API calls (cache miss)
- ✅ Second run reuses cached results (100% hit rate)
- ✅ Analysis latency significantly reduced (no AI calls)
- ✅ Cache statistics logged at end
- ✅ Cache files persisted in correct directory

---

## Scenario 4: Cache Invalidation on Event Change

### Setup

```bash
# Run analysis with caching
sdek analyze

# Verify cache populated
ls ~/.cache/sdek/ai-cache/ | wc -l
# Expected: 124 files
```

### Execute

```bash
# Add new event (simulates event change)
sdek ingest --source git --events testdata/new_commit.json

# Run analysis again
sdek analyze --verbose

# Expected output:
# ✓ Loaded 131 events from 5 sources (1 new)
# ⚙ Cache invalidated: 12 controls affected by event changes
# ⚙ AI provider: openai (gpt-4-turbo-preview)
# ⚙ Analyzing control SOC2-CC1.1 with AI... (cache miss, re-analyze)
# ⚙ Cache hit: SOC2-CC2.1 (unaffected control)
# ... (mix of cache hits and misses)
# ✓ Analysis complete: 124 controls, 19 findings
# ✓ AI cache hit rate: 90% (12 invalidations)
```

### Validate

```bash
# Check cache statistics
# - 12 controls re-analyzed (new AI calls)
# - 112 controls cached (no changes)
```

**Success Criteria**:
- ✅ Event change detected by cache invalidation logic
- ✅ Only affected controls re-analyzed with AI
- ✅ Unaffected controls still served from cache
- ✅ Cache hit rate reflects partial invalidation
- ✅ New findings captured from new event

---

## Scenario 5: Switch AI Providers Mid-Stream

### Setup

```bash
# Start with OpenAI
export SDEK_AI_OPENAI_KEY="sk-..."
sdek config set ai.provider openai
sdek analyze
```

### Execute

```bash
# Switch to Anthropic
export SDEK_AI_ANTHROPIC_KEY="sk-ant-..."
sdek config set ai.provider anthropic
sdek config set ai.model claude-3-opus-20240229

# Run analysis again
sdek analyze --verbose

# Expected output:
# ✓ Loaded 130 events from 5 sources
# ⚙ AI provider: anthropic (claude-3-opus-20240229)
# ⚙ Cache miss: SOC2-CC1.1 (different provider, re-analyze)
# ... (all cache misses, new provider)
# ✓ Analysis complete: 124 controls, 18 findings
# ✓ AI cache hit rate: 0% (provider changed)
```

### Validate

```bash
# Check report shows Anthropic provider
cat report.json | jq '.frameworks[0].controls[0].evidence[0]'

# Expected: "provider": "anthropic"
```

**Success Criteria**:
- ✅ Provider switch detected in config
- ✅ OpenAI cache not reused (different provider)
- ✅ All controls re-analyzed with Anthropic
- ✅ New cache entries created for Anthropic
- ✅ No errors or mixed provider results

---

## Scenario 6: Disable AI for CI/CD

### Setup

```bash
# CI/CD environment (no API keys available)
unset SDEK_AI_OPENAI_KEY
unset SDEK_AI_ANTHROPIC_KEY
```

### Execute

```bash
# Run analysis with AI disabled via flag
sdek analyze --ai-provider=none --verbose

# Or via config
sdek config set ai.enabled false
sdek analyze --verbose

# Expected output:
# ✓ Loaded 130 events from 5 sources
# ⚙ AI analysis disabled (--ai-provider=none)
# ⚙ Using heuristic-only analysis
# ⚙ Analyzing control SOC2-CC1.1 with heuristics...
# ... (heuristic analysis for all controls)
# ✓ Analysis complete: 124 controls, 18 findings
```

### Validate

```bash
# Check report shows heuristic-only method
cat report.json | jq '.frameworks[0].controls[0].evidence[0].analysis_method'
# Expected: "heuristic-only"

# Verify no AI cache created
ls ~/.cache/sdek/ai-cache/ 2>/dev/null || echo "No AI cache directory"
# Expected: "No AI cache directory"
```

**Success Criteria**:
- ✅ AI disabled via flag or config
- ✅ Analysis runs successfully with heuristics
- ✅ No AI API calls attempted
- ✅ No cache directory created
- ✅ Results consistent with deterministic heuristics (reproducible in CI)

---

## Scenario 7: PII Redaction Before AI Transmission

### Setup

```bash
# Generate events with PII
cat > testdata/event_with_pii.json <<EOF
{
  "event_id": "evt-pii-1",
  "event_type": "message",
  "source": "slack",
  "description": "Support ticket discussion",
  "content": "User john.doe@example.com reported issue. Phone: 555-123-4567. API key: sk-1234567890abcdefghijklmnopqrstuv",
  "timestamp": "2025-10-11T10:00:00Z"
}
EOF

sdek ingest --source slack --events testdata/event_with_pii.json
```

### Execute

```bash
# Run analysis with verbose logging
sdek analyze --verbose

# Expected output:
# ⚙ Redacting PII from events before AI transmission...
# ⚙ Redacted 1 email, 1 phone number, 1 API key
# ⚙ Analyzing control SOC2-CC1.1 with AI...
# ... (analysis proceeds with redacted content)
```

### Validate

```bash
# Check logs for redaction statistics
# Expected: Redaction counts logged per pattern type

# Verify original events unchanged in state
sdek state show --event evt-pii-1 | grep "john.doe@example.com"
# Expected: Email still present in stored state (only redacted for AI)

# Check AI prompt did not contain PII (via audit logs if implemented)
# Expected: Redacted placeholders like "User <EMAIL_REDACTED> reported issue"
```

**Success Criteria**:
- ✅ PII detected before AI transmission
- ✅ Email, phone, API key redacted with placeholders
- ✅ Original events unchanged in state
- ✅ AI analysis proceeds with redacted content
- ✅ Redaction statistics logged

---

## Scenario 8: AI Timeout and Fallback

### Setup

```bash
# Simulate slow AI provider (mock or use real provider under load)
# For testing, can temporarily reduce timeout:
sdek config set ai.timeout 5  # 5 seconds instead of 60
```

### Execute

```bash
sdek analyze --verbose

# Expected output:
# ⚙ AI provider: openai (gpt-4-turbo-preview)
# ⚙ Analyzing control SOC2-CC1.1 with AI...
# ⚠ AI request timeout after 5s, falling back to heuristics
# ⚙ Analyzing control SOC2-CC1.2 with AI...
# ✓ AI confidence: 85%, heuristic: 70%, combined: 80%
# ... (mix of timeouts and successes)
# ✓ Analysis complete: 124 controls, 18 findings
# ⚠ 15 controls analyzed with heuristics due to AI timeouts
```

### Validate

```bash
# Check report shows mixed analysis methods
cat report.json | jq '[.frameworks[].controls[].evidence[].analysis_method] | unique'
# Expected: ["ai+heuristic", "heuristic-only"]
```

**Success Criteria**:
- ✅ Timeout detected after configured duration
- ✅ Graceful fallback to heuristics for timed-out controls
- ✅ Other controls continue with AI (no global abort)
- ✅ Report indicates which controls used fallback
- ✅ Analysis completes successfully despite timeouts

---

## Integration Test Automation

These scenarios should be codified as integration tests:

```bash
# Run all integration tests
go test ./tests/integration/... -v

# Specific AI integration tests
go test ./tests/integration/analyze_ai_test.go -v -run TestAnalyzeWithAI
go test ./tests/integration/analyze_ai_test.go -v -run TestAnalyzeFallback
go test ./tests/integration/analyze_ai_test.go -v -run TestAnalyzeCaching
```

**Test Fixtures**:
- `testdata/ai/fixtures/openai_response_*.json` - Golden file responses
- `testdata/ai/fixtures/anthropic_response_*.json` - Golden file responses
- `testdata/events_with_pii.json` - Events for redaction testing

---

## Manual Validation Checklist

After implementation, validate:

- [ ] AI analysis produces different (hopefully better) confidence scores than heuristics alone
- [ ] Justifications are contextually relevant and reference specific events
- [ ] Residual risk notes highlight genuine gaps or concerns
- [ ] Cache persists across CLI invocations
- [ ] Event changes correctly invalidate affected cache entries
- [ ] Provider switching works without errors or stale cache issues
- [ ] AI-disabled mode is identical to legacy heuristic behavior
- [ ] PII is never transmitted to AI providers (verify via audit logs or packet capture)
- [ ] Timeouts prevent indefinite hangs
- [ ] Fallback to heuristics is seamless (no user interruption)

---

## Rollback Plan

If AI feature causes issues in production:

1. **Immediate**: Disable AI via config
   ```bash
   sdek config set ai.enabled false
   ```

2. **Temporary**: Use environment variable override
   ```bash
   export SDEK_AI_ENABLED=false
   sdek analyze
   ```

3. **Permanent**: Remove AI provider keys to prevent accidental use
   ```bash
   sdek config unset ai.openai_key
   sdek config unset ai.anthropic_key
   ```

4. **Verify**: Confirm heuristic-only analysis works identically to pre-AI behavior
   ```bash
   # Compare reports before/after disabling AI
   diff report_ai_disabled.json report_pre_ai.json
   # Should be identical except timestamps
   ```
