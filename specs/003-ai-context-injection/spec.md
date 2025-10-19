# Feature Specification: AI Context Injection & Autonomous Evidence Collection

**Feature Branch**: `003-ai-context-injection`  
**Created**: 2025-10-17  
**Status**: Draft  
**Input**: User description: "AI Context Injection & Autonomous Evidence Collection - Phase 1: Deterministic context injection (framework + section excerpt) into all AI analyses. Phase 2: AI proposes an evidence plan, gets approval, collects via MCPs, then analyzes."

## Execution Flow (main)
```
1. Parse user description from Input
   → Feature clearly defined: two-phase enhancement to AI analysis
2. Extract key concepts from description
   → Actors: Compliance Managers, Security Engineers
   → Actions: inject context, propose plans, approve, collect evidence, analyze
   → Data: framework metadata, control excerpts, evidence bundles, findings
   → Constraints: 95% context injection, 30% time reduction, 0.8 confidence, zero PII leakage
3. For each unclear aspect:
   → Marked where clarification needed (approval workflows, auto-approve policies)
4. Fill User Scenarios & Testing section
   → Three primary flows: context mode, autonomous mode, redaction validation
5. Generate Functional Requirements
   → 31 functional + 6 non-functional testable requirements covering both phases
6. Identify Key Entities
   → ContextPreamble, EvidencePlan, Finding, RedactionMap
7. Run Review Checklist
   → All sections complete, testable requirements, 3 clarifications marked
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-17
- Q: What default budget limits should be used for autonomous mode (max sources, max API calls, max tokens)? → A: Enterprise defaults (50 max sources, 500 max API calls, 250K max tokens) with configuration support via AI config file
- Q: How should the auto-approve policy be structured in the configuration? → A: Source + query patterns mapping (e.g., `github: ["auth*", "login*"]`)
- Q: What roles and permissions should be supported for role-based access control? → A: Three-tier model - Admin (full access), Analyst (read + analyze), Viewer (read-only)
- Q: What should happen when confidence score falls below a threshold? → A: Flag findings with confidence <0.6 with review-required badge, display all findings
- Q: What is the maximum number of concurrent analyses? → A: Configurable limit with default of 25 concurrent analyses

---

## Problem Statement

Current AI-powered compliance mappings lack explicit grounding in the specific framework controls being analyzed. Compliance managers must manually prepare context (framework names, section excerpts, control definitions) before each analysis, which is time-consuming and error-prone. Additionally, the system cannot autonomously determine what evidence is needed for a given control and collect it automatically.

This results in:
- **Context gaps**: AI responses may not reflect the precise control requirements
- **Manual prep overhead**: Analysts spend significant time gathering and formatting evidence
- **Inconsistent analysis**: Without standardized context injection, results vary across runs
- **Missed evidence**: Analysts may not identify all relevant data sources for a control

## Business Value & Outcomes

### Success Metrics
- **95% context injection rate**: All AI analyses include explicit framework metadata and section excerpts in prompts
- **30% reduction in prep time**: Time from control selection to first finding decreases by 30%
- **≥0.8 confidence score**: Seeded demo controls consistently achieve high confidence ratings across runs
- **Zero PII leakage**: No sensitive data reaches AI providers (validated via redaction logs)

### Business Benefits
- **Faster audits**: Automated evidence collection reduces manual work
- **Higher accuracy**: Explicit context grounding improves finding quality
- **Better compliance**: Mandatory redaction ensures regulatory compliance
- **Scalability**: Autonomous mode enables analysis of hundreds of controls

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story 1: Context Mode Analysis
**As a** Compliance Manager analyzing SOC 2 controls  
**I want to** analyze CC6.1 (Logical Access Security) with the official control excerpt automatically injected  
**So that** the AI's findings are grounded in the precise control language and I can trust the confidence scores

**Flow:**
1. User selects SOC 2 framework and CC6.1 control
2. User provides evidence files (or points to existing evidence)
3. System injects framework name, version, and CC6.1 excerpt into AI prompt
4. System sends normalized evidence + context to AI provider
5. AI returns finding with confidence score, residual risk, justification, and citations
6. User reviews finding with full context visibility

### Primary User Story 2: Autonomous Evidence Collection
**As a** Compliance Manager preparing for an audit  
**I want to** run autonomous mode on PCI-DSS requirements  
**So that** the system proposes what evidence to collect, I approve it, and findings are generated without manual prep

**Flow:**
1. User selects PCI-DSS framework and requirement 8.2.3
2. User enables autonomous mode (no evidence provided upfront)
3. System generates Evidence Plan: sources (GitHub, Jira, AWS), queries, estimated signal strength
4. User reviews plan in TUI and approves specific items
5. System executes plan via MCP connectors, normalizing collected data
6. System runs context mode analysis with collected evidence
7. User receives findings with provenance showing evidence sources

### Primary User Story 3: Redaction & Privacy Validation
**As a** Security Engineer  
**I want to** enforce redaction of PII and secrets before AI provider calls  
**So that** we comply with GDPR/CCPA and never leak sensitive data externally

**Flow:**
1. Security Engineer configures redaction policies (regex patterns, denylist terms)
2. During any AI analysis, system scans evidence and excerpts for secrets
3. System redacts matches, stores redaction map locally only
4. System sends sanitized prompt to AI provider
5. System logs redaction events (what was redacted, not the values)
6. Security Engineer audits redaction logs to verify zero leakage

### Acceptance Scenarios

#### Phase 1: Context Injection
1. **Given** SOC 2 framework is selected and CC6.1 excerpt is available  
   **When** I run context mode analysis with evidence files  
   **Then** the AI prompt includes "Framework: SOC2, Version: 2017, Section: CC6.1" and the full excerpt text  
   **And** the finding includes confidence score (0-1), residual risk (low/medium/high), justification, and citations

2. **Given** evidence contains PII (emails, IP addresses) and secrets (API keys)  
   **When** the prompt is constructed  
   **Then** all PII and secrets are redacted with placeholders  
   **And** the redaction map is stored locally (not sent to provider)  
   **And** a redaction event is logged

3. **Given** identical framework, section, and evidence (by digest)  
   **When** I run the analysis a second time without `--no-cache`  
   **Then** cached results are returned within 100ms  
   **And** no AI provider call is made

4. **Given** AI provider is unavailable or returns an error  
   **When** the analysis runs  
   **Then** the system falls back to heuristics mode  
   **And** the finding includes `mode: heuristics` flag  
   **And** basic keyword/regex mappings are used

#### Phase 2: Autonomous Collection
5. **Given** autonomous mode is enabled without evidence  
   **When** I analyze ISO 27001 A.9.4.2 (Secure log-on procedures)  
   **Then** the system proposes an Evidence Plan listing:
   - Sources: GitHub (authentication code), AWS (IAM policies), CI/CD (security scans)
   - Queries/Filters: "auth", "login", "MFA", "password policy"
   - Estimated signal strength per source

6. **Given** an Evidence Plan is proposed  
   **When** I review it in the TUI  
   **Then** I can approve or deny each item individually  
   **And** the system only collects data for approved items

7. **Given** I approved GitHub and AWS sources in the plan  
   **When** the system executes the plan  
   **Then** data is fetched via GitHub and AWS MCP connectors  
   **And** results are normalized into standard evidence format  
   **And** context mode analysis runs with collected evidence  
   **And** the finding shows provenance (which sources contributed)

8. **Given** auto-approve policy is configured for specific sources  
   **When** autonomous mode generates a plan  
   **Then** matching sources are auto-approved  
   **And** user only reviews non-auto-approved items

### Edge Cases
- **What happens when** a framework excerpt file is missing or empty?  
  → System errors with clear message: "Excerpt required for [framework] section [X]"

- **What happens when** autonomous mode proposes 20+ evidence sources?  
  → System applies configurable budget limits (max sources, max API calls) and prioritizes by estimated signal strength

- **What happens when** redaction removes all meaningful content from evidence?  
  → System warns user: "Evidence heavily redacted, confidence may be low" and proceeds with analysis

- **What happens when** an MCP connector fails during plan execution?  
  → System logs failure, marks that item as incomplete, proceeds with available evidence, and flags partial results

- **What happens when** a user denies all items in an Evidence Plan?  
  → System errors: "No evidence sources approved, cannot proceed"

- **What happens when** cache key collision occurs (unlikely but possible)?  
  → System includes evidence digest in cache key to prevent false hits

---

## Requirements *(mandatory)*

### Functional Requirements - Phase 1: Context Injection

- **FR-001**: System MUST inject framework metadata (name, version, identifier) into all AI analysis prompts
- **FR-002**: System MUST inject the exact control/section excerpt text into all AI analysis prompts
- **FR-003**: System MUST include normalized evidence in the prompt payload
- **FR-004**: System MUST include analysis rubrics (confidence criteria, risk levels) in the prompt
- **FR-005**: AI responses MUST return: finding_summary, mapped_controls, confidence_score (0-1), residual_risk (low/medium/high), justification, citations
- **FR-005a**: System MUST flag findings with confidence score <0.6 with a "review-required" badge in TUI and exports
- **FR-006**: System MUST redact PII (emails, phone numbers, IP addresses, names) from all evidence before sending to AI providers
- **FR-007**: System MUST redact secrets (API keys, passwords, tokens) from all evidence before sending to AI providers
- **FR-008**: System MUST store redaction maps locally only (never sent to providers)
- **FR-009**: System MUST log redaction events (count, types, timestamps) without logging actual redacted values
- **FR-010**: System MUST cache AI prompts and responses using keys: framework, section hash, evidence digest
- **FR-011**: System MUST support `--no-cache` flag to bypass cache
- **FR-012**: System MUST fall back to heuristics mode if AI provider fails or is disabled
- **FR-013**: Heuristics mode MUST mark findings with `mode: heuristics` flag
- **FR-014**: System MUST export findings to JSON and HTML formats with all metadata

### Functional Requirements - Phase 2: Autonomous Collection

- **FR-015**: System MUST generate Evidence Plans given framework + section excerpt (without evidence)
- **FR-016**: Evidence Plans MUST list: sources (GitHub, Jira, Slack, AWS, CI/CD, docs), queries/filters, estimated signal strength per source
- **FR-017**: System MUST require user approval for evidence collection (interactive or dry-run mode)
- **FR-018**: System MUST support per-item approval/denial in Evidence Plans
- **FR-019**: System MUST support policy-driven auto-approval using source-to-query-patterns mapping in configuration (e.g., `github: ["auth*", "login*"]` auto-approves GitHub queries matching those patterns)
- **FR-020**: System MUST execute approved plan items via existing MCP connectors
- **FR-021**: System MUST normalize collected data into standard evidence format (EvidenceEvent schema)
- **FR-022**: System MUST run context mode analysis with collected evidence after plan execution
- **FR-023**: Findings from autonomous mode MUST include provenance (which sources contributed which evidence)
- **FR-024**: System MUST apply configurable budget limits (max sources, max API calls, max tokens) via AI configuration file with enterprise defaults: 50 max sources, 500 max API calls, 250K max tokens
- **FR-025**: System MUST prioritize evidence sources by estimated signal strength when budget limits apply

### Functional Requirements - Cross-Cutting

- **FR-026**: System MUST enforce three-tier role-based access: Admin (full access including config changes), Analyst (read + run analyses + approve plans), Viewer (read-only access to findings and reports)
- **FR-027**: System MUST generate audit events for: plan proposals, approvals, evidence collection, redactions, findings
- **FR-028**: CLI MUST support commands: `sdek ai analyze`, `sdek ai plan`
- **FR-029**: TUI MUST display Context Preview panel showing injected framework + excerpt
- **FR-030**: TUI MUST display Autonomous Plan panel with approve/deny controls per item
- **FR-031**: TUI MUST show status pills (green/yellow/red), mode badges (AI/heuristics), and review-required badges for low-confidence findings

### Non-Functional Requirements

- **NFR-001**: Context mode analysis MUST complete within 30 seconds for typical evidence bundles (< 100 events)
- **NFR-002**: Autonomous mode MUST complete within 5 minutes for plans with ≤10 sources
- **NFR-003**: Cache hits MUST return results within 100ms
- **NFR-004**: Redaction MUST not remove more than 40% of evidence content (warn if exceeded)
- **NFR-005**: System MUST support configurable concurrent analysis with default maximum of 25 concurrent controls
- **NFR-006**: Evidence Plans MUST be deterministic for the same framework + section + policy configuration

### Key Entities *(include if feature involves data)*

- **ContextPreamble**: Represents framework metadata, section identifier, excerpt text, and analysis rubrics. Used to ground AI prompts.
  
- **EvidencePlan**: Represents a proposed evidence collection plan. Includes:
  - Sources (GitHub, Jira, AWS, etc.)
  - Queries/Filters per source
  - Estimated signal strength per source
  - Approval status per item (approved/denied/pending)
  - Auto-approval matching (whether query matched configured patterns)

- **Finding**: Represents the output of AI analysis. Includes:
  - finding_summary (string)
  - mapped_controls (array)
  - confidence_score (float 0-1)
  - residual_risk (enum: low/medium/high)
  - justification (string)
  - citations (array of evidence references)
  - mode (enum: ai/heuristics)
  - provenance (array of sources, for autonomous mode)
  - review_required (boolean: true if confidence <0.6)

- **RedactionMap**: Represents the mapping of redacted content. Stored locally only. Includes:
  - Original position/hash
  - Placeholder token
  - Redaction type (PII/secret)
  - Timestamp

- **EvidenceBundle**: Collection of normalized evidence events, redaction status, and metadata. Used as input to AI analysis.

---

## Scope & Boundaries

### In Scope
- Phase 1: Deterministic context injection for all AI analyses
- Phase 2: AI-generated evidence plans, approval workflows, autonomous collection
- Redaction of PII and secrets with local-only storage
- Prompt/response caching with digest-based keys
- Fallback to heuristics on AI failures
- CLI commands: `sdek ai analyze`, `sdek ai plan`
- TUI panels: Context Preview, Autonomous Plan
- JSON/HTML exports with full metadata
- Audit event generation

### Out of Scope
- New MCP connectors (reuse existing: GitHub, Jira, Slack, AWS, CI/CD, docs)
- Changes to severity-weighted risk scoring engine
- Real-time evidence streaming
- Multi-user collaboration features
- Custom AI model training
- Evidence versioning/history tracking
- Automated plan scheduling
- Integration with third-party compliance platforms

### Dependencies
- Existing MCP connectors (GitHub, Jira, Slack, AWS, CI/CD, docs) must be functional
- AI provider APIs (Anthropic, OpenAI) must be accessible
- Framework excerpt files (SOC2, ISO27001, PCI-DSS, etc.) must be available in policy loader
- Redaction patterns (regex, denylist) must be configured

### Assumptions
- Users have necessary credentials for MCP connectors (GitHub tokens, AWS keys, etc.)
- Framework excerpts are curated and accurate
- Evidence normalization already handles all connector formats
- TUI framework (Bubble Tea) supports required interactive components

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] Requirements are testable and unambiguous (3 clarifications marked)
- [x] Success criteria are measurable (95% injection, 30% time reduction, 0.8 confidence, zero leakage)
- [x] Scope is clearly bounded (in/out of scope defined)
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (3 clarifications)
- [x] User scenarios defined (3 primary stories, 8 acceptance scenarios, 6 edge cases)
- [x] Requirements generated (31 functional, 6 non-functional)
- [x] Entities identified (5 key entities)
- [x] Review checklist passed

---

## Next Steps

1. **Planning Phase**: Create detailed plan with tasks, milestones, and estimates
2. **Clarifications**: Resolve outstanding questions with stakeholders:
   - Define auto-approve policy schema
   - Set default budget limits
   - Specify role-based permission model
3. **Data Model**: Document ContextPreamble, EvidencePlan, Finding schemas
4. **Contracts**: Define internal API interfaces (Analyze, ProposePlan, ExecutePlan)
5. **Implementation**: Begin with Phase 1 (context injection), then Phase 2 (autonomous collection)
