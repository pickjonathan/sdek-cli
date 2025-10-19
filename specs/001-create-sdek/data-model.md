# Data Model: Create sdek

**Feature**: 001-create-sdek  
**Date**: 2025-10-11

## Entity Definitions

### Source
Represents a data integration point (Git, Jira, Slack, CI/CD, Docs).

**Fields**:
- `ID` (string): Unique identifier (e.g., "git", "jira", "slack", "cicd", "docs")
- `Name` (string): Display name (e.g., "Git Commits", "Jira Tickets")
- `Type` (string): Source type identifier
- `Status` (string): Connection status (always "simulated" in Phase 1)
- `LastSync` (time.Time): Last synchronization timestamp
- `EventCount` (int): Number of events from this source (10-50)
- `Enabled` (bool): Whether source is active

**Validation Rules**:
- ID must be one of: "git", "jira", "slack", "cicd", "docs"
- EventCount must be between 10 and 50
- LastSync must be within 90 days of current date (simulated)

**State Transitions**:
```
Initialized → Seeded → Ready
```

---

### Event
A discrete signal from a source (commit, ticket, message, pipeline run, document).

**Fields**:
- `ID` (string): Unique identifier (UUID format)
- `SourceID` (string): Reference to parent Source
- `Timestamp` (time.Time): When event occurred
- `EventType` (string): Type of event (commit, ticket, message, build, document_change)
- `Title` (string): Event summary (commit message, ticket title, etc.)
- `Content` (string): Full event content/description
- `Author` (string): Creator/committer name
- `Metadata` (map[string]interface{}): Source-specific data
  - Git: commit_sha, branch, files_changed
  - Jira: ticket_id, status, priority
  - Slack: channel, thread_id, reactions
  - CI/CD: pipeline_id, status, duration
  - Docs: file_path, change_type, reviewer

**Validation Rules**:
- ID must be valid UUID
- SourceID must reference existing Source
- Timestamp must be within 90 days of current date
- EventType must match source type
- Title required, max 200 characters
- Content optional, max 10,000 characters

**Relationships**:
- Many-to-one with Source (Event.SourceID → Source.ID)
- Many-to-many with Evidence (Event can map to multiple Evidence items)

---

### Framework
A compliance standard (SOC2, ISO 27001, PCI DSS).

**Fields**:
- `ID` (string): Unique identifier ("soc2", "iso27001", "pci_dss")
- `Name` (string): Full framework name
- `Version` (string): Framework version
- `ControlCount` (int): Total number of controls
- `CompliancePercentage` (float64): Overall compliance score (0-100)
- `Description` (string): Framework description
- `Category` (string): Framework category (security, privacy, financial)

**Validation Rules**:
- ID must be one of: "soc2", "iso27001", "pci_dss"
- ControlCount must be positive integer
- CompliancePercentage must be between 0 and 100
- Must have at least one control in each risk state (green, yellow, red)

**Relationships**:
- One-to-many with Control (Framework has multiple Controls)

**Compliance Calculation**:
```
CompliancePercentage = (GreenControls / TotalControls) * 100
```

---

### Control
A specific compliance requirement within a framework.

**Fields**:
- `ID` (string): Control identifier (e.g., "CC6.1", "A.12.6.1", "6.3.2")
- `FrameworkID` (string): Reference to parent Framework
- `Title` (string): Control title/name
- `Description` (string): Control requirement description
- `Category` (string): Control category (access_control, change_management, etc.)
- `SeverityLevel` (string): Control criticality (critical, high, medium, low)
- `EvidenceIDs` ([]string): List of mapped Evidence IDs
- `ConfidenceScore` (float64): Aggregate confidence (0-100)
- `RiskStatus` (string): Compliance state (complete, partial, missing)
- `RiskColor` (string): Visual indicator (green, yellow, red)
- `IssueCount` (map[string]int): Count by severity (critical, high, medium, low)
- `CriticalEquivalent` (float64): Weighted severity score

**Validation Rules**:
- ID format must match framework convention
- FrameworkID must reference existing Framework
- SeverityLevel must be one of: critical, high, medium, low
- RiskStatus must be one of: complete, partial, missing
- RiskColor must be one of: green, yellow, red
- ConfidenceScore must be between 0 and 100

**Relationships**:
- Many-to-one with Framework (Control.FrameworkID → Framework.ID)
- One-to-many with Evidence (Control has multiple Evidence items)

**Risk Calculation**:
```go
CriticalEquivalent = 
    IssueCount["critical"] * 1.0 +
    IssueCount["high"] * 0.333 +     // 3 high = 1 critical
    IssueCount["medium"] * 0.167 +   // 6 medium = 1 critical
    IssueCount["low"] * 0.083        // 12 low = 1 critical

if CriticalEquivalent > 0:
    RiskStatus = "missing"
    RiskColor = "red"
elif ConfidenceScore >= 70 && len(EvidenceIDs) >= 2:
    RiskStatus = "complete"
    RiskColor = "green"
else:
    RiskStatus = "partial"
    RiskColor = "yellow"
```

---

### Evidence
A mapping between an Event and a Control.

**Fields**:
- `ID` (string): Unique identifier (UUID format)
- `EventID` (string): Reference to source Event
- `ControlID` (string): Reference to target Control
- `ConfidenceLevel` (string): Mapping strength (Low, Medium, High)
- `ConfidenceScore` (float64): Numeric confidence (0-100)
- `RelevanceScore` (float64): How well event matches control (0-100)
- `Timestamp` (time.Time): When mapping was created
- `MappingRationale` (string): Why event maps to control (keywords matched)
- `Keywords` ([]string): Extracted keywords that triggered mapping

**Validation Rules**:
- ID must be valid UUID
- EventID must reference existing Event
- ControlID must reference existing Control
- ConfidenceLevel must be one of: Low, Medium, High
- ConfidenceScore must be between 0 and 100
- RelevanceScore must be between 0 and 100
- ConfidenceLevel derived from ConfidenceScore:
  - Low: 0-30
  - Medium: 31-70
  - High: 71-100

**Relationships**:
- Many-to-one with Event (Evidence.EventID → Event.ID)
- Many-to-one with Control (Evidence.ControlID → Control.ID)

**Confidence Calculation**:
```go
ConfidenceScore = (KeywordMatchScore * 0.6) + (ContextScore * 0.4)

where:
  KeywordMatchScore = (MatchedKeywords / TotalControlKeywords) * 100
  ContextScore = based on event metadata and control category alignment
```

---

### Finding
An identified compliance gap or risk.

**Fields**:
- `ID` (string): Unique identifier (UUID format)
- `ControlID` (string): Reference to affected Control
- `SeverityLevel` (string): Issue severity (critical, high, medium, low)
- `Title` (string): Finding summary
- `Description` (string): Detailed gap description
- `RecommendedActions` ([]string): List of remediation steps
- `EvidenceDeficit` (int): Number of missing evidence items
- `CriticalEquivalentScore` (float64): Weighted severity contribution
- `DetectedAt` (time.Time): When finding was identified
- `Status` (string): Finding state (open, acknowledged, resolved)

**Validation Rules**:
- ID must be valid UUID
- ControlID must reference existing Control
- SeverityLevel must be one of: critical, high, medium, low
- EvidenceDeficit must be non-negative integer
- Status must be one of: open, acknowledged, resolved

**Relationships**:
- Many-to-one with Control (Finding.ControlID → Control.ID)

**Severity Contribution**:
```go
CriticalEquivalentScore = 
    1.0     if SeverityLevel == "critical"
    0.333   if SeverityLevel == "high"
    0.167   if SeverityLevel == "medium"
    0.083   if SeverityLevel == "low"
```

---

### User
Simulated user with role-based view preferences.

**Fields**:
- `ID` (string): Unique identifier (UUID format)
- `Name` (string): Display name
- `Email` (string): Email address
- `Role` (string): User role (compliance_manager, engineer)
- `VisibilityPreferences` (map[string]bool): What data user can see
  - `show_technical_details` (bool)
  - `show_summary_only` (bool)
  - `show_evidence_ids` (bool)
  - `show_confidence_scores` (bool)
- `Active` (bool): Whether user is selected

**Validation Rules**:
- ID must be valid UUID
- Role must be one of: compliance_manager, engineer
- Email must be valid format
- Exactly one user must have Active = true at any time

**Predefined Users**:
```go
Users = [
    {
        ID: "user-001",
        Name: "Alice Chen",
        Email: "alice.chen@example.com",
        Role: "compliance_manager",
        VisibilityPreferences: {
            "show_technical_details": false,
            "show_summary_only": true,
            "show_evidence_ids": false,
            "show_confidence_scores": true,
        },
        Active: true,
    },
    {
        ID: "user-002",
        Name: "Bob Martinez",
        Email: "bob.martinez@example.com",
        Role: "engineer",
        VisibilityPreferences: {
            "show_technical_details": true,
            "show_summary_only": false,
            "show_evidence_ids": true,
            "show_confidence_scores": true,
        },
        Active: false,
    },
    {
        ID: "user-003",
        Name: "Carol Zhang",
        Email: "carol.zhang@example.com",
        Role: "engineer",
        VisibilityPreferences: {
            "show_technical_details": true,
            "show_summary_only": false,
            "show_evidence_ids": true,
            "show_confidence_scores": true,
        },
        Active: false,
    },
]
```

---

### Config
Application configuration.

**Fields**:
- `DataDir` (string): Data directory path (default: `$HOME/.sdek`)
- `ConfigFile` (string): Config file path (default: `$HOME/.sdek/config.yaml`)
- `StateFile` (string): State persistence file (default: `$HOME/.sdek/state.json`)
- `ReportsDir` (string): Export directory (default: `$HOME/sdek/reports`)
- `LogLevel` (string): Logging verbosity (debug, info, warn, error)
- `ColorTheme` (string): UI theme (dark, light, custom)
- `SelectedUserRole` (string): Active user role (compliance_manager, engineer)
- `FrameworksEnabled` ([]string): Active frameworks
- `SourcesEnabled` ([]string): Active sources

**Validation Rules**:
- DataDir must be valid directory path
- LogLevel must be one of: debug, info, warn, error
- ColorTheme must be one of: dark, light, custom
- SelectedUserRole must be valid role
- FrameworksEnabled items must be valid framework IDs
- SourcesEnabled items must be valid source IDs

**Default Values**:
```go
DefaultConfig = Config{
    DataDir: "$HOME/.sdek",
    ConfigFile: "$HOME/.sdek/config.yaml",
    StateFile: "$HOME/.sdek/state.json",
    ReportsDir: "$HOME/sdek/reports",
    LogLevel: "info",
    ColorTheme: "dark",
    SelectedUserRole: "compliance_manager",
    FrameworksEnabled: []string{"soc2", "iso27001", "pci_dss"},
    SourcesEnabled: []string{"git", "jira", "slack", "cicd", "docs"},
}
```

---

## State Schema

**JSON State File Structure** (`$HOME/.sdek/state.json`):

```json
{
  "version": "1.0.0",
  "updated_at": "2025-10-11T10:30:00Z",
  "sources": [
    {
      "id": "git",
      "name": "Git Commits",
      "type": "git",
      "status": "simulated",
      "last_sync": "2025-10-11T10:00:00Z",
      "event_count": 25,
      "enabled": true
    }
  ],
  "events": [
    {
      "id": "evt-001",
      "source_id": "git",
      "timestamp": "2025-10-10T15:30:00Z",
      "event_type": "commit",
      "title": "Add authentication middleware",
      "content": "Implemented JWT-based authentication...",
      "author": "Bob Martinez",
      "metadata": {
        "commit_sha": "a1b2c3d4",
        "branch": "main",
        "files_changed": 5
      }
    }
  ],
  "frameworks": [
    {
      "id": "soc2",
      "name": "SOC 2 Type II",
      "version": "2017",
      "control_count": 45,
      "compliance_percentage": 73.3,
      "description": "Trust Services Criteria",
      "category": "security"
    }
  ],
  "controls": [
    {
      "id": "CC6.1",
      "framework_id": "soc2",
      "title": "Logical and Physical Access Controls",
      "description": "...",
      "category": "access_control",
      "severity_level": "high",
      "evidence_ids": ["evd-001", "evd-002"],
      "confidence_score": 85.0,
      "risk_status": "complete",
      "risk_color": "green",
      "issue_count": {"critical": 0, "high": 0, "medium": 1, "low": 2},
      "critical_equivalent": 0.333
    }
  ],
  "evidence": [
    {
      "id": "evd-001",
      "event_id": "evt-001",
      "control_id": "CC6.1",
      "confidence_level": "High",
      "confidence_score": 85.0,
      "relevance_score": 90.0,
      "timestamp": "2025-10-11T10:30:00Z",
      "mapping_rationale": "Keywords: authentication, access, JWT",
      "keywords": ["authentication", "access", "security", "JWT"]
    }
  ],
  "findings": [
    {
      "id": "find-001",
      "control_id": "CC6.7",
      "severity_level": "medium",
      "title": "Insufficient documentation for password policy",
      "description": "Control requires documented password policy...",
      "recommended_actions": [
        "Document password complexity requirements",
        "Add password policy to security handbook"
      ],
      "evidence_deficit": 1,
      "critical_equivalent_score": 0.167,
      "detected_at": "2025-10-11T10:30:00Z",
      "status": "open"
    }
  ],
  "users": [
    {
      "id": "user-001",
      "name": "Alice Chen",
      "email": "alice.chen@example.com",
      "role": "compliance_manager",
      "visibility_preferences": {
        "show_technical_details": false,
        "show_summary_only": true
      },
      "active": true
    }
  ]
}
```

---

## Relationships Diagram

```
Source (1) ──< (N) Event
                    │
                    │ (N)
                    ↓
Framework (1) ──< (N) Control ──< (N) Evidence
                         │
                         │ (N)
                         ↓
                      Finding

User (independent selection)
Config (application-level)
```

---

## Indexes and Performance

**In-Memory Indexes** (for fast lookups):
- `EventsBySource`: map[string][]Event
- `ControlsByFramework`: map[string][]Control
- `EvidenceByControl`: map[string][]Evidence
- `EvidenceByEvent`: map[string][]Evidence
- `FindingsByControl`: map[string][]Finding

**Lazy Loading Strategy**:
- Load full state on startup
- Build indexes in background
- Cache computed values (compliance percentages, risk scores)
- Invalidate cache only on data changes

---

## Migration Strategy

**Future Phases**:
- Add `external_system_id` field to Event for real integrations
- Add `last_analyzed` timestamp to Control for incremental analysis
- Add `audit_trail` array to track evidence changes
- Add `custom_frameworks` support with user-defined controls
