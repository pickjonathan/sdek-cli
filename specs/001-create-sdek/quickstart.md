# Quickstart Guide: Create sdek

**Feature**: 001-create-sdek  
**Purpose**: Validate the sdek CLI and TUI implementation through end-to-end user scenarios

## Prerequisites

- Go 1.23+ installed
- Terminal with minimum 80×24 dimensions
- Unix-like OS (Linux, macOS) or Windows with WSL/PowerShell

## Installation

```bash
# Clone repository
git clone https://github.com/yourorg/sdek-cli.git
cd sdek-cli

# Build the CLI
go build -o sdek main.go

# Move to PATH (optional)
sudo mv sdek /usr/local/bin/

# Verify installation
sdek --version
```

**Expected Output**:
```
sdek version 0.1.0 (build: abc123, date: 2025-10-11)
```

---

## Scenario 1: First-Time Setup with Demo Data

**User Story**: As a new user, I want to initialize sdek with demo data so I can explore the tool immediately.

### Steps

1. **Seed demo data**:
```bash
sdek seed --demo
```

**Expected Output**:
```
Seeding demo data...
  ✓ Created 5 sources
  ✓ Generated 130 events
  ✓ Loaded 3 frameworks (120 controls)
  ✓ Created 245 evidence mappings
  ✓ Generated 18 findings

Demo data ready! Run 'sdek tui' to explore.
State saved to /Users/alice/.sdek/state.json
```

2. **Verify state file created**:
```bash
ls -lh ~/.sdek/state.json
```

**Expected**: File exists, size ~50-200KB

3. **Check configuration**:
```bash
sdek config list
```

**Expected Output**:
```
Configuration (/Users/alice/.sdek/config.yaml):
  data_dir: /Users/alice/.sdek
  log_level: info
  theme: dark
  user_role: compliance_manager
  export.default_path: /Users/alice/sdek/reports
  export.format: json
  frameworks.enabled: [soc2, iso27001, pci_dss]
  sources.enabled: [git, jira, slack, cicd, docs]
```

**Validation**:
- ✅ State file created successfully
- ✅ Config file created with defaults
- ✅ No errors displayed

---

## Scenario 2: Interactive TUI Navigation

**User Story**: As a compliance manager, I want to navigate the TUI to view compliance status across frameworks.

### Steps

1. **Launch TUI**:
```bash
sdek tui
```

**Expected**: TUI launches with home screen showing three sections:
```
┌─ sdek ─────────────────────────────────────────────────────────┐
│                                                                 │
│  [Sources]     Frameworks     Findings                         │
│                                                                 │
│  > Git Commits           25 events    Last sync: 2 hours ago   │
│    Jira Tickets          30 events    Last sync: 1 hour ago    │
│    Slack Messages        18 events    Last sync: 30 mins ago   │
│    CI/CD Pipelines       42 events    Last sync: 15 mins ago   │
│    Documentation         15 events    Last sync: 5 mins ago    │
│                                                                 │
│  Tab: Switch sections  ↑/↓: Navigate  Enter: Open  q: Quit     │
└─────────────────────────────────────────────────────────────────┘
```

2. **Switch to Frameworks section**:
   - Press `Tab`

**Expected**: Frameworks section becomes active:
```
┌─ sdek ─────────────────────────────────────────────────────────┐
│                                                                 │
│  Sources     [Frameworks]     Findings                         │
│                                                                 │
│  > SOC2 Type II          75% compliant    34/45 controls       │
│    ISO 27001            70% compliant    42/60 controls       │
│    PCI DSS              80% compliant    12/15 controls       │
│                                                                 │
│                                                                 │
│                                                                 │
│  Tab: Switch sections  ↑/↓: Navigate  Enter: Open  q: Quit     │
└─────────────────────────────────────────────────────────────────┘
```

3. **Open SOC2 framework**:
   - Press `Enter`

**Expected**: Control list displays with color-coded status:
```
┌─ SOC2 Type II Controls (75% compliant) ────────────────────────┐
│                                                                 │
│  > CC6.1  Logical and Physical Access Controls     ✓ Green     │
│    CC6.2  Prior to Issuing System Credentials     ⚠ Yellow    │
│    CC6.3  Removal of Access                        ✓ Green     │
│    CC6.7  Password Policies                        ✗ Red       │
│    ...                                                         │
│                                                                 │
│  ← Back  Enter: View Evidence  r: Refresh  e: Export           │
└─────────────────────────────────────────────────────────────────┘
```

4. **View control evidence**:
   - Navigate to `CC6.1` (use `↓`)
   - Press `Enter`

**Expected**: Evidence list with details:
```
┌─ CC6.1: Logical and Physical Access Controls ──────────────────┐
│  Status: Complete ✓  Confidence: 85%  Evidence: 3 items        │
│                                                                 │
│  > Git Commit: Add authentication middleware                   │
│    Source: git  Confidence: High (90%)  2 hours ago            │
│    Keywords: authentication, access, JWT                       │
│                                                                 │
│    Jira Ticket: Implement MFA for admin users                  │
│    Source: jira  Confidence: High (85%)  1 day ago             │
│    Keywords: MFA, authentication, admin, security              │
│                                                                 │
│    CI/CD Pipeline: Security scan passed                        │
│    Source: cicd  Confidence: Medium (65%)  3 hours ago         │
│    Keywords: security, scan, access                            │
│                                                                 │
│  ← Back  r: Refresh  e: Export                                 │
└─────────────────────────────────────────────────────────────────┘
```

5. **Navigate to Findings section**:
   - Press `←` to go back twice (to home)
   - Press `Tab` twice to reach Findings

**Expected**: Findings list with severity indicators:
```
┌─ sdek ─────────────────────────────────────────────────────────┐
│                                                                 │
│  Sources     Frameworks     [Findings]                         │
│                                                                 │
│  > [Medium] CC6.7 - Insufficient password policy docs          │
│    [High]   A.12.6.1 - Missing vulnerability scan evidence     │
│    [Low]    6.3.2 - Code review documentation incomplete       │
│    ...                                                         │
│                                                                 │
│                                                                 │
│  Tab: Switch sections  ↑/↓: Navigate  Enter: View  q: Quit     │
└─────────────────────────────────────────────────────────────────┘
```

6. **Quit TUI**:
   - Press `q`

**Expected**: Clean exit to shell, no errors

**Validation**:
- ✅ TUI displays home screen with three sections
- ✅ Tab navigation works between sections
- ✅ Arrow keys navigate lists
- ✅ Enter key opens details
- ✅ Color coding visible (green/yellow/red)
- ✅ Keyboard shortcuts displayed in status bar
- ✅ q key exits cleanly

---

## Scenario 3: CLI Workflow (Ingest → Analyze → Report)

**User Story**: As an engineer, I want to run the full compliance workflow via CLI commands for automation.

### Steps

1. **Clear existing demo data** (optional):
```bash
sdek seed --demo --reset
```

2. **Ingest simulated data**:
```bash
sdek ingest --verbose
```

**Expected Output**:
```
Ingesting data from sources...
  ✓ Git: 25 events
  ✓ Jira: 30 events
  ✓ Slack: 18 events
  ✓ CI/CD: 42 events
  ✓ Docs: 15 events

Total: 130 events ingested
State saved to /Users/alice/.sdek/state.json
```

3. **Verify events in state file**:
```bash
jq '.events | length' ~/.sdek/state.json
```

**Expected Output**: `130`

4. **Run evidence analysis**:
```bash
sdek analyze --verbose
```

**Expected Output**:
```
Analyzing evidence mappings...
  Processing 130 events across 5 sources
  Mapping to 120 controls across 3 frameworks

Results:
  Evidence mappings: 245 created
  Controls updated: 120
  Findings generated: 18

Risk Status:
  SOC2: 75% compliant (34/45 green, 8/45 yellow, 3/45 red)
  ISO 27001: 70% compliant (42/60 green, 12/60 yellow, 6/60 red)
  PCI DSS: 80% compliant (12/15 green, 2/15 yellow, 1/15 red)

State saved to /Users/alice/.sdek/state.json
```

5. **Export compliance report**:
```bash
sdek report --output ~/compliance-report.json
```

**Expected Output**:
```
Generating compliance report...
  Including frameworks: SOC2, ISO 27001, PCI DSS
  Total controls: 120
  Total evidence: 245
  Total findings: 18

Report exported to:
  /Users/alice/compliance-report.json
```

6. **Verify report structure**:
```bash
jq 'keys' ~/compliance-report.json
```

**Expected Output**:
```json
[
  "frameworks",
  "findings",
  "metadata",
  "summary"
]
```

7. **Check compliance summary**:
```bash
jq '.summary' ~/compliance-report.json
```

**Expected Output**:
```json
{
  "total_controls": 120,
  "compliant_controls": 88,
  "partial_controls": 22,
  "missing_controls": 10,
  "compliance_percentage": 73.3
}
```

**Validation**:
- ✅ Ingest command adds events to state
- ✅ Analyze command creates evidence mappings
- ✅ Analyze calculates risk scores correctly
- ✅ Report exports valid JSON
- ✅ Report contains all required sections
- ✅ Compliance percentages match analysis output

---

## Scenario 4: Configuration Management

**User Story**: As a user, I want to customize sdek configuration to match my preferences.

### Steps

1. **View current configuration**:
```bash
sdek config list
```

2. **Change log level**:
```bash
sdek config set log_level debug
```

**Expected Output**:
```
Updated configuration:
  log_level: debug

Configuration saved to /Users/alice/.sdek/config.yaml
```

3. **Change user role**:
```bash
sdek config set user_role engineer
```

4. **Verify configuration**:
```bash
sdek config get user_role
```

**Expected Output**: `engineer`

5. **Test configuration precedence** (CLI flag overrides config):
```bash
sdek tui --role compliance_manager
```

**Expected**: TUI launches with compliance_manager view (shows summaries, not technical details)

6. **Test environment variable override**:
```bash
SDEK_LOG_LEVEL=error sdek analyze
```

**Expected**: Only error-level logs displayed (overrides config file setting)

**Validation**:
- ✅ config set updates config file
- ✅ config get retrieves correct values
- ✅ CLI flags override config file
- ✅ Environment variables override config file
- ✅ Changes persist across commands

---

## Scenario 5: Role-Based Visibility

**User Story**: As a user switching roles, I want to see different information based on my role.

### Steps

1. **Launch TUI as compliance manager**:
```bash
sdek tui --role compliance_manager
```

2. **Open a control and observe details**:
   - Navigate to Frameworks → SOC2 → CC6.1

**Expected**: Summary view without technical details:
```
┌─ CC6.1: Logical and Physical Access Controls ──────────────────┐
│  Status: Complete ✓  Confidence: 85%                           │
│                                                                 │
│  Evidence Summary:                                             │
│    3 evidence items mapped                                     │
│    High confidence level                                       │
│    Last updated: 2 hours ago                                   │
│                                                                 │
│  Risk Assessment:                                              │
│    No critical gaps identified                                 │
│    Control fully satisfied                                     │
│                                                                 │
│  ← Back  e: Export                                             │
└─────────────────────────────────────────────────────────────────┘
```

3. **Quit and relaunch as engineer**:
```bash
sdek tui --role engineer
```

4. **Open the same control**:
   - Navigate to Frameworks → SOC2 → CC6.1

**Expected**: Technical details visible:
```
┌─ CC6.1: Logical and Physical Access Controls ──────────────────┐
│  Status: Complete ✓  Confidence: 85%  Control ID: CC6.1        │
│                                                                 │
│  Evidence Details:                                             │
│    evt-001  Git Commit: Add authentication middleware          │
│            SHA: a1b2c3d4  Confidence: 90%                      │
│            Author: Bob Martinez                                │
│                                                                 │
│    evt-042  Jira Ticket: PROJ-123 Implement MFA                │
│            Status: Done  Confidence: 85%                       │
│            Assignee: Carol Zhang                               │
│                                                                 │
│    evt-089  CI/CD: Security scan #245                          │
│            Pipeline: main  Confidence: 65%                     │
│            Duration: 3m 42s                                    │
│                                                                 │
│  ← Back  e: Export                                             │
└─────────────────────────────────────────────────────────────────┘
```

**Validation**:
- ✅ Compliance manager sees high-level summaries
- ✅ Engineer sees technical details (IDs, SHAs, metadata)
- ✅ Role switching changes visible information
- ✅ Both roles can navigate and export

---

## Scenario 6: Error Handling

**User Story**: As a user, I want clear error messages when something goes wrong.

### Steps

1. **Test invalid command**:
```bash
sdek invalid-command
```

**Expected Output**:
```
Error: unknown command "invalid-command" for "sdek"
Run 'sdek --help' for usage.
```

**Exit Code**: 2

2. **Test missing state file**:
```bash
rm ~/.sdek/state.json
sdek report
```

**Expected Output**:
```
Error: failed to load state: open /Users/alice/.sdek/state.json: no such file or directory

Run 'sdek seed --demo' to initialize demo data.
```

**Exit Code**: 4

3. **Test invalid config value**:
```bash
sdek config set log_level invalid
```

**Expected Output**:
```
Error: invalid log level "invalid"
Valid values: debug, info, warn, error
```

**Exit Code**: 3

4. **Test terminal too small** (resize terminal to 70×20):
```bash
sdek tui
```

**Expected Output**:
```
Error: terminal size too small (70×20)
Minimum required: 80×24

Please resize your terminal and try again.
```

**Exit Code**: 1

**Validation**:
- ✅ Clear error messages with context
- ✅ Appropriate exit codes
- ✅ Helpful recovery suggestions
- ✅ No stack traces or panics

---

## Scenario 7: Performance Validation

**User Story**: As a user, I want sdek to start quickly and render smoothly.

### Steps

1. **Measure cold start time**:
```bash
time sdek --version
```

**Expected**: < 100ms total execution time

2. **Measure TUI launch time**:
```bash
time (sdek tui &); sleep 0.5; killall sdek
```

**Expected**: TUI appears within 100ms

3. **Test large dataset** (generate maximum events):
```bash
sdek seed --demo --reset
sdek ingest --events 50  # Maximum per source
sdek analyze
```

**Expected**:
- Ingest completes in < 1 second
- Analyze completes in < 3 seconds
- State file size < 5MB

4. **Test TUI responsiveness** with large dataset:
```bash
sdek tui
```

- Navigate lists rapidly (hold `↓`)

**Expected**:
- Smooth scrolling (no lag)
- Frame rate stays at 60fps
- No visible stuttering

**Validation**:
- ✅ Cold start under 100ms
- ✅ Commands complete quickly
- ✅ TUI rendering smooth and responsive
- ✅ No performance degradation with maximum data

---

## Success Criteria

### Must Pass
- ✅ All scenarios complete without errors
- ✅ TUI displays correctly in 80×24 terminal
- ✅ Keyboard navigation works in TUI
- ✅ CLI commands execute successfully
- ✅ State persistence works correctly
- ✅ Configuration management functions properly
- ✅ Role-based visibility works as expected
- ✅ Error handling provides clear messages
- ✅ Performance meets targets (<100ms startup, 60fps TUI)

### Nice to Have
- Terminal color support detection
- Export in multiple formats
- Custom seed data scenarios
- Help screens in TUI
- Search functionality in TUI

---

## Cleanup

```bash
# Remove state and config
rm -rf ~/.sdek

# Remove exported reports
rm ~/compliance-report.json
rm -rf ~/sdek/reports
```

---

## Troubleshooting

### TUI doesn't render correctly
- Ensure terminal supports ANSI colors
- Try `TERM=xterm-256color sdek tui`
- Update terminal emulator

### Commands fail with permission errors
- Check directory permissions: `ls -ld ~/.sdek`
- Fix: `chmod 755 ~/.sdek`

### State file corruption
- Backup: `cp ~/.sdek/state.json ~/.sdek/state.json.bak`
- Reset: `sdek seed --demo --reset`

### Slow performance
- Check state file size: `du -h ~/.sdek/state.json`
- If > 10MB, consider resetting: `sdek seed --demo --reset`
