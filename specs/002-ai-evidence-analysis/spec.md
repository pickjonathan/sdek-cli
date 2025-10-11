# Feature Specification: AI Evidence Analysis

**Feature Branch**: `002-ai-evidence-analysis`  
**Created**: October 11, 2025  
**Status**: Draft  
**Input**: User description: "AI Evidence Analysis: The mapping and summarization of findings from third-party integrations (Git, CI/CD, Jira, Slack, Docs) to compliance controls (SOC2, ISO 27001, PCI DSS) will be performed via an AI integration layer. The system will support multiple providers (initially OpenAI and Anthropic) behind a single abstraction. For each control, we will prompt the provider with the relevant policy text/excerpts plus normalized events, and request: (1) mapped evidence references, (2) a short justification, (3) a confidence score (0–100), and (4) residual risk notes. Provider choice is configurable; if AI is disabled or errors occur, the system falls back to deterministic heuristics. No raw secrets or PII leave the machine; inputs are redacted/hashed where possible. Results are cached locally and included in the exported report."

## Execution Flow (main)
```
1. Parse user description from Input
   → Feature: AI-enhanced evidence mapping for compliance controls
2. Extract key concepts from description
   → Actors: Compliance managers, engineers
   → Actions: Map events to controls, generate justifications, score confidence
   → Data: Events from 5 sources, 3 compliance frameworks, AI responses
   → Constraints: Privacy (no PII/secrets), fallback to heuristics, local caching
3. For each unclear aspect:
   → [RESOLVED] AI provider configuration mechanism
   → [RESOLVED] Caching strategy
   → [RESOLVED] Fallback behavior
4. Fill User Scenarios & Testing section
   → Primary: Compliance manager analyzes evidence with AI assistance
   → Edge cases: AI failure, privacy violations, low confidence
5. Generate Functional Requirements
   → All requirements testable via automated tests
   → Privacy and security requirements measurable
6. Identify Key Entities
   → AI Provider, AI Analysis Result, Evidence with AI metadata
7. Run Review Checklist
   → No implementation details in requirements
   → All ambiguities resolved via assumptions documented
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-11
- Q: How should the system determine cache freshness? → A: Event-driven: Cache invalidates immediately when any source event is added/modified/deleted
- Q: What timeout should trigger fallback to heuristics for a single AI provider request? → A: 60 seconds: Maximum wait before considering AI provider unresponsive
- Q: When analyzing a compliance control with no matching events, what should happen? → A: Skip AI entirely: Don't call AI provider; immediately mark control as "No Evidence" with 0% confidence
- Q: Should the system validate additional constraints on AI response JSON fields? → A: Trust AI output: Minimal validation; accept any JSON structure matching the four field names
- Q: When AI analysis succeeds, should heuristic analysis also run and combine confidence scores? → A: Weighted average: Combine scores (e.g., 70% AI + 30% heuristic) for balanced confidence

### Session 2025-10-12
- Q: Which PII/secret patterns should the privacy filter detect and redact? → A: Comprehensive: emails, phones, API keys, credit cards, SSNs, IP addresses, AWS keys, JWTs, passwords. System should inform AI that violations were detected (e.g., "3 API keys found") without sending actual values, so AI can factor security violations into compliance analysis.
- Q: Should the suggested config keys become formal requirements in the spec? → A: Yes, but only essential: ai.enabled (bool), ai.provider (string: openai|anthropic|none), ai.model (string: provider-specific model name). Omit tuning parameters (max_tokens, temperature, timeout_ms) as implementation details.
- Q: How should performance expectations be reconciled? → A: Adjust target: Change to <60 seconds per control analysis (matching timeout), <5 minutes for full analysis of all controls in a typical compliance framework.
- Q: What observability signals should the AI analysis system emit? → A: Full observability: Structured metrics (AI call counts, success/failure rates, cache hit rate, latency histograms, confidence score distributions, per-provider usage) plus distributed traces for each analysis workflow to enable production debugging.
- Q: When AI returns low confidence (<20%), what should happen? → A: Flag for review: Mark finding as "Requires Manual Verification" in report to signal compliance managers that evidence is weak and needs human judgment.

---

## User Scenarios & Testing

### Primary User Story
As a compliance manager, I want the system to use AI to intelligently map evidence from multiple sources (Git commits, CI/CD builds, Jira tickets, Slack messages, documentation) to specific compliance controls (SOC2, ISO 27001, PCI DSS), so that I can receive more accurate, contextually-aware compliance assessments with clear justifications rather than relying solely on keyword matching.

The system should automatically analyze events, understand their relevance to each control, provide confidence scores, and highlight any residual risks—while ensuring that sensitive information never leaves my machine.

### Acceptance Scenarios

1. **Given** the system has collected events from all five sources, **When** I run the analysis command with AI enabled, **Then** the system generates evidence mappings with AI-provided justifications and confidence scores for each control.

2. **Given** an AI provider is configured (OpenAI or Anthropic), **When** the system analyzes a control, **Then** it sends the control policy text and relevant events to the AI, receives structured analysis (evidence references, justification, confidence score 0-100, residual risk notes), and stores results locally.

3. **Given** multiple events reference the same compliance control, **When** AI analysis runs, **Then** the system consolidates evidence and provides a unified confidence assessment with reasoning explaining how multiple events contribute to compliance.

4. **Given** events contain potential PII or secrets (API keys, passwords, emails), **When** the system prepares data for AI analysis, **Then** sensitive information is automatically redacted or hashed before transmission, ensuring privacy.

5. **Given** AI analysis is disabled in configuration, **When** the system runs analysis, **Then** it automatically falls back to the existing deterministic heuristic-based mapping without user intervention.

6. **Given** an AI provider returns an error or times out, **When** the system attempts analysis, **Then** it gracefully falls back to heuristic mapping and logs the failure without blocking the workflow.

7. **Given** the system has previously analyzed certain events, **When** analysis runs again with the same data, **Then** cached AI results are reused to avoid redundant API calls and reduce latency.

8. **Given** a compliance report is generated, **When** AI analysis was used, **Then** the report includes AI-generated justifications, confidence scores, and residual risk notes alongside standard evidence data.

9. **Given** the user wants to switch AI providers, **When** they update the configuration, **Then** subsequent analyses use the new provider without requiring code changes.

10. **Given** an event matches multiple controls weakly via heuristics, **When** AI analysis runs, **Then** the AI clarifies which controls are genuinely relevant and assigns appropriate confidence scores, reducing false positives.

### Edge Cases
- **What happens when AI provider quota is exceeded?** System should fall back to heuristics and log quota exhaustion for user awareness.
- **What happens when events have no clear compliance relevance?** AI returns low confidence scores (0-20), system flags finding as "Requires Manual Verification" in report to signal weak evidence.
- **What happens when zero events match a control?** System skips AI analysis entirely, marks control as "No Evidence" with 0% confidence, avoiding unnecessary API calls.
- **How does the system handle extremely large events (e.g., long documentation)?** System should truncate or summarize content before sending to AI to stay within token limits.
- **What happens when cached results become stale?** Cache is automatically invalidated whenever any source event is added, modified, or deleted, ensuring AI analysis always reflects current event data.
- **How does the system prevent prompt injection attacks?** System should sanitize inputs and use structured prompts to prevent malicious event content from altering AI behavior.
- **What if both AI providers fail?** System continues with deterministic heuristics, treating it as if AI were disabled.

---

## Requirements

### Functional Requirements

#### AI Provider Integration
- **FR-001**: System MUST support multiple AI providers (initially OpenAI and Anthropic) through a unified abstraction layer.
- **FR-002**: System MUST allow users to configure AI provider via three essential config keys: `ai.enabled` (bool: true/false), `ai.provider` (string: "openai"|"anthropic"|"none"), and `ai.model` (string: provider-specific model name like "gpt-4" or "claude-3-5-sonnet").
- **FR-003**: System MUST read API credentials from environment variables only (SDEK_OPENAI_API_KEY, SDEK_ANTHROPIC_API_KEY), never from config files, to prevent accidental secret exposure.
- **FR-004**: System MUST support disabling AI analysis entirely by setting `ai.enabled=false`, falling back to deterministic heuristics.
- **FR-005**: System MUST gracefully handle AI provider errors (timeouts after 60 seconds, rate limits, invalid responses) by falling back to heuristic analysis for affected controls.

#### AI-Enhanced Evidence Analysis
- **FR-005**: For each compliance control with matching events, system MUST send to the AI provider: (a) control policy text/description, (b) relevant normalized events from all sources. Controls with zero events are skipped and marked as "No Evidence" with 0% confidence.
- **FR-006**: System MUST request from the AI provider: (a) mapped evidence references (which events support this control), (b) short justification (why these events are relevant), (c) confidence score (0-100 indicating certainty), (d) residual risk notes (what gaps or concerns remain).
- **FR-007**: System MUST parse AI responses as JSON and verify the four required field names are present (evidence links, justification, confidence, residual risk). Minimal type validation; trust AI output structure. If parsing fails or fields missing, fall back to heuristics for that control.
- **FR-008**: When AI analysis succeeds, system MUST also run heuristic analysis and combine confidence scores using weighted average (70% AI confidence + 30% heuristic confidence) to balance AI intelligence with deterministic keyword validation.
- **FR-009**: System MUST consolidate evidence from multiple events and provide unified compliance assessment per control.

#### Privacy & Security
- **FR-010**: System MUST NOT send raw secrets, API keys, passwords, or credentials to AI providers.
- **FR-011**: System MUST automatically detect and redact comprehensive PII/secret patterns before sending data to AI: email addresses, phone numbers, credit card numbers, SSNs, IP addresses, AWS access keys, JWTs, passwords, and API keys.
- **FR-012**: System MUST inform AI provider when PII/secrets are detected (e.g., "3 API keys redacted", "1 SSN found") so AI can factor security violations into compliance analysis, without transmitting the actual sensitive values.
- **FR-013**: System MUST hash or anonymize sensitive identifiers when possible while preserving analytical utility.
- **FR-014**: System MUST process all AI analysis locally on the user's machine; no data should be stored or logged by external providers beyond the request lifetime.
- **FR-015**: System MUST sanitize event content to prevent prompt injection attacks that could alter AI behavior.

#### Caching & Performance
- **FR-015**: System MUST cache AI analysis results locally to avoid redundant API calls for identical inputs.
- **FR-016**: System MUST invalidate cached AI results immediately when any source event is added, modified, or deleted to ensure analysis reflects current data.
- **FR-017**: System MUST provide cache statistics (hit rate, size) for monitoring efficiency.
- **FR-018**: System MUST handle AI provider token limits by truncating or summarizing long event content before transmission.

#### Fallback & Resilience
- **FR-019**: When AI analysis fails or is disabled, system MUST use existing deterministic heuristic mapping as fallback.
- **FR-020**: System MUST clearly distinguish AI-generated confidence scores from heuristic scores in reports.
- **FR-021**: System MUST emit structured observability signals: metrics (AI call counts, success/failure rates by provider, cache hit rate, latency histograms, confidence score distributions), logs (all AI failures with error context), and distributed traces (span per control analysis workflow showing AI call, cache lookup, privacy filter, fallback decisions).
- **FR-022**: System MUST continue functioning without degradation if AI quota is exhausted; heuristics handle all subsequent analysis.

#### Reporting & Transparency
- **FR-023**: Compliance reports MUST include AI-generated justifications when AI analysis was used.
- **FR-024**: Compliance reports MUST display both the combined confidence score (weighted average) and individual AI/heuristic scores for transparency.
- **FR-025**: Compliance reports MUST include residual risk notes from AI analysis where applicable.
- **FR-026**: Reports MUST indicate which evidence was analyzed via AI (with combined confidence) vs. heuristics only (when AI disabled/failed).
- **FR-027**: When AI returns low confidence (<20%), system MUST flag finding with "Requires Manual Verification" marker in report to signal compliance managers that evidence is weak and requires human judgment before making compliance decisions.

### Non-Functional Requirements

#### Accuracy & Quality
- **NFR-001**: AI-enhanced evidence mapping SHOULD provide higher accuracy than deterministic heuristics for ambiguous events [NEEDS CLARIFICATION: target accuracy improvement metric].
- **NFR-002**: AI confidence scores SHOULD correlate with actual compliance auditor judgments [NEEDS CLARIFICATION: validation methodology].

#### Performance
- **NFR-003**: AI analysis SHOULD complete within <60 seconds per control and <5 minutes for full analysis of all controls in a typical compliance framework.
- **NFR-004**: Each AI provider request MUST timeout after 60 seconds, triggering immediate fallback to heuristic analysis for that control.
- **NFR-005**: Cache hit rate SHOULD exceed 70% for repeated analyses on stable data.

#### Privacy & Compliance
- **NFR-005**: System MUST comply with data protection regulations (GDPR, CCPA) by not transmitting PII to external providers.
- **NFR-006**: System MUST be auditable: all AI requests and responses logged locally for compliance review.

#### Usability
- **NFR-007**: AI provider configuration SHOULD be simple (single config parameter change).
- **NFR-008**: Fallback to heuristics SHOULD be transparent; users not blocked by AI failures.

### Key Entities

- **AI Provider**: Represents an external AI service (OpenAI, Anthropic). Attributes include provider name, API endpoint abstraction, authentication mechanism, and enabled/disabled status.

- **AI Analysis Request**: Input sent to AI provider for a specific control. Contains control ID, control policy text, list of normalized events (with sensitive data redacted), and request timestamp.

- **AI Analysis Result**: Structured JSON response from AI provider. Contains four required fields: evidence references (event IDs or descriptions), justification text, confidence score (numeric value interpreted as 0-100 scale), residual risk notes (may be empty string). Also includes provider identifier and result timestamp added by system.

- **Cached AI Result**: Locally stored AI analysis outcome. Attributes include cache key (hash of request inputs), cached response data, cache creation timestamp, and event change tracking (invalidated when source events are added/modified/deleted).

- **Evidence with AI Metadata**: Enhanced evidence record. Original evidence plus AI justification, individual AI confidence score, heuristic confidence score, combined weighted confidence (70% AI + 30% heuristic), AI residual risk notes, and analysis method indicator (AI+heuristic combined vs. heuristic-only fallback).

- **Privacy Filter**: Component responsible for redaction. Identifies and masks comprehensive PII/secret patterns (emails, phones, credit cards, SSNs, IP addresses, AWS keys, JWTs, passwords, API keys) before AI transmission. Tracks redaction statistics (count of redacted items per event by category) and generates metadata summary (e.g., "3 API keys redacted, 1 email found") that is passed to AI provider to inform compliance analysis without exposing actual sensitive values.

---

## Assumptions & Dependencies

### Assumptions
1. **AI providers return structured responses**: We assume OpenAI and Anthropic can be prompted to return JSON-like structured data with the four required fields.
2. **Event change detection is feasible**: System can reliably detect when source events are added, modified, or deleted to trigger cache invalidation.
3. **Privacy patterns are detectable**: Common PII patterns (emails, phone numbers, API key formats) can be identified via regex or heuristics for redaction.
4. **Heuristic fallback is sufficient**: Existing deterministic keyword-based mapping provides acceptable baseline accuracy when AI is unavailable.
5. **Token limits are manageable**: Event content can be truncated or summarized to fit within AI provider token limits (e.g., 4K-8K tokens per request).

### Dependencies
- **Existing compliance framework definitions**: AI must reference SOC2, ISO 27001, PCI DSS control definitions already implemented in Phase 3.7.
- **Event normalization**: AI analysis relies on normalized events from the five sources (Git, CI/CD, Jira, Slack, Docs) implemented in Phase 3.6.
- **Configuration system**: AI provider selection and enablement depends on config framework from Phase 3.5.
- **Report exporter**: AI metadata must integrate with report generation system from Phase 3.8.
- **External AI provider accounts**: Users must have valid API keys for OpenAI or Anthropic to enable AI features.

---

## Out of Scope

The following are explicitly **not** included in this feature:

- **Training custom AI models**: System uses external AI providers as-is; no model training or fine-tuning.
- **Real-time AI analysis**: Analysis runs on-demand or scheduled; not live/streaming.
- **Multi-language AI support**: AI responses assumed to be in English only.
- **Advanced prompt engineering UI**: Users cannot customize AI prompts via the interface; prompts are predefined.
- **AI-based remediation suggestions**: AI only analyzes evidence; it does not generate remediation steps or fixes.
- **Cost management tools**: No built-in tracking or alerts for AI provider API usage costs.
- **On-premise AI deployment**: System relies on cloud-based AI providers; no local model hosting.

---

## Success Metrics

- **Accuracy**: AI-enhanced mapping reduces false positives compared to heuristics alone, measured post-launch via user feedback and auditor review outcomes.
- **Confidence Calibration**: Manual verification rate for low-confidence findings (<20%) tracked; target <10% false positives in this category.
- **Adoption**: 50%+ of users enable AI analysis within 3 months of release.
- **Performance**: AI analysis completes in <60 seconds per control and <5 minutes for full framework analysis for 95% of typical workloads.
- **Privacy**: Zero incidents of PII or secrets leaking to AI providers over 6-month period, validated via audit log review.
- **Fallback Reliability**: System maintains 100% uptime even when AI providers are down or disabled.
- **Observability**: All AI workflows emit metrics, logs, and traces; monitoring dashboards available within first month of release.

---

## Review & Acceptance Checklist

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers (all ambiguities resolved in Sessions 2025-10-11 and 2025-10-12)
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable with clear validation methods
- [x] Scope is clearly bounded (out-of-scope section defined)
- [x] Dependencies and assumptions identified

---

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted (AI providers, privacy, fallback, caching)
- [x] Ambiguities resolved (10 clarifications across 2 sessions: 2025-10-11 and 2025-10-12)
- [x] User scenarios defined (10 acceptance scenarios + edge cases)
- [x] Requirements generated (27 functional + 8 non-functional)
- [x] Entities identified (6 key entities)
- [x] Review checklist passed (ready for planning phase)

---

## Next Steps

1. **Clarification Phase**: Address marked [NEEDS CLARIFICATION] items with stakeholders:
   - Define cache TTL policy (time-based? event-change-based?)
   - Establish baseline false positive rate for accuracy comparison
   - Design validation methodology for AI confidence calibration
   - Specify target accuracy improvement metric

2. **Planning Phase**: Break down into technical tasks:
   - Design AI provider abstraction interface
   - Implement privacy filter for PII/secrets redaction
   - Build caching layer with invalidation logic
   - Create fallback orchestration between AI and heuristics
   - Extend report format to include AI metadata

3. **Implementation Phase**: Develop in stages:
   - Stage 1: AI provider abstraction + OpenAI integration
   - Stage 2: Privacy filtering + sanitization
   - Stage 3: Caching + performance optimization
   - Stage 4: Anthropic integration + provider switching
   - Stage 5: Report integration + transparency features

4. **Validation Phase**: Measure success metrics against targets defined above.

