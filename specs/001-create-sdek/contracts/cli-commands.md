# CLI Command Contracts

## Command: sdek

Root command with global flags.

**Usage**: `sdek [command] [flags]`

**Global Flags**:
- `--config string`: Config file path (default: `$HOME/.sdek/config.yaml`)
- `--data-dir string`: Data directory path (default: `$HOME/.sdek`)
- `--log-level string`: Log level (debug|info|warn|error) (default: "info")
- `--verbose`: Enable verbose output (equivalent to --log-level=debug)
- `--version`: Display version information

**Subcommands**:
- `tui`: Launch terminal UI
- `ingest`: Simulate data ingestion
- `analyze`: Run evidence mapping
- `report`: Export compliance report
- `seed`: Generate demo data
- `config`: Manage configuration

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Invalid arguments
- `3`: Configuration error
- `4`: State file error

---

## Command: sdek tui

Launch interactive terminal UI.

**Usage**: `sdek tui [flags]`

**Flags**:
- Inherits all global flags
- `--role string`: User role (compliance_manager|engineer) (default from config)

**Behavior**:
1. Load state from `$HOME/.sdek/state.json`
2. Check terminal dimensions (minimum 80×24)
3. Initialize Bubble Tea application
4. Display home screen with Sources/Frameworks/Findings sections
5. Enable keyboard navigation
6. Auto-save state on exit or 'e' key press

**Keyboard Shortcuts**:
- `Tab`: Switch between sections
- `↑/↓`: Navigate lists
- `Enter`: Select/open item
- `q`: Quit
- `r`: Refresh data
- `e`: Export report
- `/`: Search
- `?`: Show help

**Exit Codes**:
- `0`: Normal exit
- `1`: Terminal too small
- `2`: State file error
- `130`: SIGINT (Ctrl+C)

**Contract**:
```
Input: None
Output: Interactive TUI session
State: Read from state.json, write on exit
```

---

## Command: sdek ingest

Simulate data ingestion from sources.

**Usage**: `sdek ingest [flags]`

**Flags**:
- Inherits all global flags
- `--source strings`: Specific sources to ingest (default: all enabled)
- `--events int`: Events per source (default: random 10-50)
- `--seed int`: Random seed for reproducibility (default: timestamp)

**Behavior**:
1. Load current state
2. For each specified source:
   - Generate 10-50 simulated events
   - Assign timestamps within last 90 days
   - Add to state
3. Save updated state
4. Print summary (sources updated, events added)

**Output Format** (stdout):
```
Ingesting data from sources...
  ✓ Git: 25 events
  ✓ Jira: 30 events
  ✓ Slack: 18 events
  ✓ CI/CD: 42 events
  ✓ Docs: 15 events

Total: 130 events ingested
State saved to $HOME/.sdek/state.json
```

**Exit Codes**:
- `0`: Success
- `1`: Ingestion failed
- `4`: State file error

**Contract**:
```
Input: --source, --events, --seed flags
Output: Event count summary
State: Read state.json, append events, write state.json
Side Effects: Updates state file
```

---

## Command: sdek analyze

Run evidence mapping analysis.

**Usage**: `sdek analyze [flags]`

**Flags**:
- Inherits all global flags
- `--framework strings`: Specific frameworks to analyze (default: all enabled)
- `--remapping`: Clear existing evidence and remap from scratch (default: false)

**Behavior**:
1. Load current state (events, controls)
2. For each event:
   - Extract keywords from title/content
   - Match against control keywords
   - Calculate confidence score
   - Assign confidence level (Low/Medium/High)
   - Create Evidence record if score > 30
3. For each control:
   - Aggregate evidence
   - Calculate risk score
   - Determine risk status (complete/partial/missing)
   - Assign risk color (green/yellow/red)
4. Generate findings for gaps
5. Save updated state
6. Print summary (mappings created, controls updated, findings generated)

**Output Format** (stdout):
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

State saved to $HOME/.sdek/state.json
```

**Exit Codes**:
- `0`: Success
- `1`: Analysis failed
- `4`: State file error

**Contract**:
```
Input: --framework, --remapping flags
Output: Analysis summary with compliance percentages
State: Read state.json, update evidence/controls/findings, write state.json
Side Effects: Updates state file
```

---

## Command: sdek report

Export compliance report.

**Usage**: `sdek report [flags]`

**Flags**:
- Inherits all global flags
- `--output string`: Output file path (default: `$HOME/sdek/reports/sdek-report-YYYY-MM-DD.json`)
- `--format string`: Output format (json) (default: "json")
- `--framework strings`: Include specific frameworks (default: all)
- `--include-events`: Include full event details (default: false)

**Behavior**:
1. Load current state
2. Filter data by specified frameworks
3. Serialize to JSON format
4. Create output directory if needed
5. Write report file
6. Print file location

**Output Format** (stdout):
```
Generating compliance report...
  Including frameworks: SOC2, ISO 27001, PCI DSS
  Total controls: 120
  Total evidence: 245
  Total findings: 18

Report exported to:
  $HOME/sdek/reports/sdek-report-2025-10-11.json
```

**Report JSON Structure**:
```json
{
  "metadata": {
    "version": "1.0.0",
    "generated_at": "2025-10-11T10:30:00Z",
    "sdek_version": "0.1.0",
    "user_role": "compliance_manager",
    "sources_included": ["git", "jira", "slack", "cicd", "docs"]
  },
  "summary": {
    "total_controls": 120,
    "compliant_controls": 88,
    "partial_controls": 22,
    "missing_controls": 10,
    "compliance_percentage": 73.3
  },
  "frameworks": [...],
  "findings": [...]
}
```

**Exit Codes**:
- `0`: Success
- `1`: Export failed
- `2`: Invalid output path
- `4`: State file error

**Contract**:
```
Input: --output, --format, --framework flags
Output: File path confirmation
State: Read state.json (no writes)
Side Effects: Creates report file
```

---

## Command: sdek seed

Generate demo data with predefined scenarios.

**Usage**: `sdek seed [flags]`

**Flags**:
- Inherits all global flags
- `--demo`: Use demo scenario (default: true)
- `--seed int`: Random seed (default: 42 for reproducibility)
- `--reset`: Clear existing state first (default: false)

**Behavior**:
1. Optionally clear existing state
2. Generate sources (5 sources)
3. Generate events (10-50 per source)
4. Generate frameworks (3 frameworks with controls)
5. Run initial analysis (map evidence)
6. Ensure each framework has green/yellow/red controls
7. Save complete state
8. Print summary

**Output Format** (stdout):
```
Seeding demo data...
  ✓ Created 5 sources
  ✓ Generated 130 events
  ✓ Loaded 3 frameworks (120 controls)
  ✓ Created 245 evidence mappings
  ✓ Generated 18 findings

Demo data ready! Run 'sdek tui' to explore.
State saved to $HOME/.sdek/state.json
```

**Exit Codes**:
- `0`: Success
- `1`: Seeding failed
- `4`: State file error

**Contract**:
```
Input: --demo, --seed, --reset flags
Output: Seeding summary
State: Create/overwrite state.json
Side Effects: Creates/resets state file
```

---

## Command: sdek config

Manage application configuration.

**Usage**: `sdek config [command] [flags]`

**Subcommands**:
- `init`: Create default configuration file
- `get <key>`: Get configuration value
- `set <key> <value>`: Set configuration value
- `list`: List all configuration values
- `validate`: Validate configuration file

**Flags**:
- Inherits all global flags

### sdek config init

**Behavior**:
1. Create `$HOME/.sdek` directory
2. Write default config.yaml
3. Print file location

**Output** (stdout):
```
Created default configuration at:
  $HOME/.sdek/config.yaml
```

### sdek config get

**Usage**: `sdek config get <key>`

**Output** (stdout):
```
log_level: info
```

### sdek config set

**Usage**: `sdek config set <key> <value>`

**Output** (stdout):
```
Updated configuration:
  log_level: debug

Configuration saved to $HOME/.sdek/config.yaml
```

### sdek config list

**Output** (stdout):
```
Configuration ($HOME/.sdek/config.yaml):
  data_dir: /Users/alice/.sdek
  log_level: info
  theme: dark
  user_role: compliance_manager
  ...
```

### sdek config validate

**Output** (stdout):
```
✓ Configuration file is valid
```

**Exit Codes**:
- `0`: Success
- `1`: Command failed
- `3`: Configuration error

**Contract**:
```
Input: subcommand, key, value
Output: Configuration status
State: Read/write config.yaml (not state.json)
Side Effects: Modifies config file
```

---

## Non-Interactive Mode

All commands support non-interactive (CI/CD) mode by default. No TTY required except for `sdek tui`.

**Detection**: Commands check `os.Stdout.Fd()` to detect pipe/redirect and disable color/formatting.

**Example** (piped):
```bash
sdek report --format json | jq '.summary.compliance_percentage'
```

---

## Error Handling Contract

**All commands MUST**:
1. Validate flags in PreRun hook
2. Return contextual errors with `fmt.Errorf("context: %w", err)`
3. Log errors to stderr (never stdout)
4. Use appropriate exit codes
5. Display help on invalid usage

**Error Message Format**:
```
Error: failed to load state: open /Users/alice/.sdek/state.json: no such file or directory

Run 'sdek seed --demo' to initialize demo data.
```

---

## Configuration Precedence

For all commands, configuration is resolved in this order:

1. CLI flags (highest priority)
2. Environment variables (`SDEK_*`)
3. Config file (`$HOME/.sdek/config.yaml`)
4. Built-in defaults (lowest priority)

**Environment Variable Mapping**:
- `SDEK_DATA_DIR` → `data_dir`
- `SDEK_LOG_LEVEL` → `log_level`
- `SDEK_USER_ROLE` → `user_role`
- `SDEK_CONFIG` → config file path
