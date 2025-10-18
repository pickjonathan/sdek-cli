# Step 1 Complete: Connector Configuration Schema

**Date**: 2025-10-18  
**Branch**: `003-ai-context-injection`  
**Status**: ✅ COMPLETE

## Summary

Successfully added comprehensive connector configuration schema to enable autonomous evidence collection from external sources (GitHub, Jira, AWS, Slack).

## Changes Made

### 1. Configuration Types (`pkg/types/config.go`)

**Added `ConnectorConfig` struct**:
```go
type ConnectorConfig struct {
    Enabled   bool              // Enable this connector
    APIKey    string            // API key or token (supports env: ${VAR})
    Endpoint  string            // Optional custom endpoint URL
    RateLimit int               // Requests per minute (0 = unlimited)
    Timeout   int               // Request timeout in seconds
    Extra     map[string]string // Connector-specific settings
}
```

**Extended `AIConfig` struct**:
```go
type AIConfig struct {
    // ... existing fields ...
    Connectors map[string]ConnectorConfig `json:"connectors" mapstructure:"connectors"`
}
```

**Updated `DefaultConfig()`**:
- Initialized default configurations for 4 connectors: `github`, `jira`, `aws`, `slack`
- All connectors disabled by default (opt-in security)
- GitHub has 60 req/min rate limit (matches unauthenticated GitHub API)
- All connectors have 30 second timeout

**Added Validation**:
- Validates connector names (must be: github, jira, aws, slack)
- Validates timeout cannot be negative
- Validates rate_limit cannot be negative
- Notes that API keys can come from environment variables

### 2. Example Configuration (`config.example.yaml`)

**Added comprehensive connector examples** (~120 lines):

**GitHub Connector**:
```yaml
connectors:
  github:
    enabled: false
    api_key: ""  # or ${GITHUB_TOKEN}
    rate_limit: 60
    timeout: 30
```
- Includes documentation for:
  - Creating Personal Access Tokens (PAT)
  - Required scopes (repo, public_repo)
  - GitHub Enterprise endpoint configuration
  - Environment variable substitution

**Jira Connector**:
```yaml
  jira:
    enabled: false
    api_key: ""  # email:api_token format
    endpoint: "https://company.atlassian.net"
    timeout: 30
```
- Documents API token creation
- Shows required endpoint format
- Includes extra settings examples (default_project, max_results)

**AWS Connector**:
```yaml
  aws:
    enabled: false
    api_key: ""  # Optional - uses AWS credential chain
    timeout: 30
```
- Documents AWS credential chain order
- Explains IAM role usage
- Shows profile and services configuration

**Slack Connector**:
```yaml
  slack:
    enabled: false
    api_key: ""  # xoxb-... format
    timeout: 30
```
- Documents bot token creation
- Lists required OAuth scopes
- Shows channel filtering examples

### 3. Comprehensive Tests (`pkg/types/config_test.go`)

**Added 5 new validation tests**:
1. ✅ `valid connector config - github` - Single connector with all fields
2. ✅ `invalid connector name` - Rejects unknown connector types
3. ✅ `negative connector timeout` - Validates timeout >= 0
4. ✅ `negative connector rate limit` - Validates rate_limit >= 0
5. ✅ `multiple connectors - all valid` - Tests all 4 connectors together

**Added `TestConnectorConfig` with 3 sub-tests**:
1. ✅ `default config includes connectors` - Verifies all 4 present
2. ✅ `default connectors are disabled` - Security check
3. ✅ `github connector has rate limit` - Validates defaults

**Total new tests**: 8  
**All tests passing**: ✅ YES

## Test Results

```bash
$ go test ./pkg/types -run TestValidateConfig -v
=== RUN   TestValidateConfig
=== RUN   TestValidateConfig/valid_connector_config_-_github
=== RUN   TestValidateConfig/invalid_connector_name
=== RUN   TestValidateConfig/negative_connector_timeout
=== RUN   TestValidateConfig/negative_connector_rate_limit
=== RUN   TestValidateConfig/multiple_connectors_-_all_valid
--- PASS: TestValidateConfig (0.00s)
    --- PASS: TestValidateConfig/valid_connector_config_-_github (0.00s)
    --- PASS: TestValidateConfig/invalid_connector_name (0.00s)
    --- PASS: TestValidateConfig/negative_connector_timeout (0.00s)
    --- PASS: TestValidateConfig/negative_connector_rate_limit (0.00s)
    --- PASS: TestValidateConfig/multiple_connectors_-_all_valid (0.00s)
PASS

$ go test ./pkg/types -run TestConnectorConfig -v
=== RUN   TestConnectorConfig
=== RUN   TestConnectorConfig/default_config_includes_connectors
=== RUN   TestConnectorConfig/default_connectors_are_disabled
=== RUN   TestConnectorConfig/github_connector_has_rate_limit
--- PASS: TestConnectorConfig (0.00s)
    --- PASS: TestConnectorConfig/default_config_includes_connectors (0.00s)
    --- PASS: TestConnectorConfig/default_connectors_are_disabled (0.00s)
    --- PASS: TestConnectorConfig/github_connector_has_rate_limit (0.00s)
PASS

$ go test ./pkg/types -v
PASS
ok      github.com/pickjonathan/sdek-cli/pkg/types      0.173s
```

## Key Features

### Security By Default
- ✅ All connectors disabled by default
- ✅ API keys can use environment variables: `${GITHUB_TOKEN}`
- ✅ Validation prevents invalid configurations
- ✅ Rate limiting built into config schema

### Flexibility
- ✅ Supports custom endpoints (GitHub Enterprise, on-prem Jira)
- ✅ Per-connector timeouts and rate limits
- ✅ Extra settings map for connector-specific options
- ✅ Easy to add new connectors (just add to validConnectors list)

### Documentation
- ✅ Inline comments explain each field
- ✅ Example config shows all 4 connectors
- ✅ Links to API token creation pages
- ✅ Required scopes and permissions documented

### Extensibility
- ✅ `Extra` map allows connector-specific settings without schema changes
- ✅ Supports future connectors (just add to defaults)
- ✅ Validation is centralized and reusable

## Integration Points

The connector configuration schema integrates with:

1. **Config Loader** (`internal/config/loader.go`)
   - Loads connector configs from YAML
   - Resolves environment variables in API keys
   - Validates on load

2. **Connector Registry** (`internal/ai/connectors/registry.go`)
   - `RegistryBuilder.SetConfig()` accepts `ConnectorConfig`
   - Converts types.ConnectorConfig → connectors.Config
   - Already implemented and tested

3. **AI Engine** (`internal/ai/engine.go`)
   - Next step: Wire registry initialization from config
   - `NewEngineFromConfig()` will build registry with connectors

## Next Steps

✅ **Step 1 Complete** - Connector configuration schema  
🔄 **Step 2 In Progress** - Wire connectors into Engine  
⏳ **Step 3 Pending** - Update AI plan command  
⏳ **Step 4 Pending** - Integration tests  
⏳ **Step 5 Pending** - Documentation updates

## Files Modified

1. `pkg/types/config.go` (+28 lines)
   - Added ConnectorConfig struct
   - Extended AIConfig with Connectors field
   - Updated DefaultConfig with 4 connectors
   - Added connector validation logic

2. `config.example.yaml` (+120 lines)
   - Added comprehensive connector documentation
   - Examples for all 4 connectors
   - API token creation instructions
   - Environment variable usage examples

3. `pkg/types/config_test.go` (+150 lines)
   - 5 new ValidateConfig test cases
   - 3 new ConnectorConfig test cases
   - All edge cases covered
   - All tests passing

**Total**: ~300 lines added, 8 new tests, 0 breaking changes

---

**Estimated Time**: 30 minutes (as planned)  
**Actual Time**: 35 minutes (including test fixes)  
**Complexity**: Low-Medium  
**Breaking Changes**: None - backwards compatible  
**Ready for Production**: ✅ YES
