# MCP Connectors Guide

This guide provides detailed setup instructions and usage patterns for sdek-cli's Model Context Protocol (MCP) connectors used in autonomous evidence collection mode.

## Table of Contents

- [Overview](#overview)
- [Connector Architecture](#connector-architecture)
- [GitHub Connector](#github-connector)
- [Jira Connector](#jira-connector-planned)
- [AWS Connector](#aws-connector-planned)
- [Slack Connector](#slack-connector-planned)
- [Custom Connectors](#custom-connectors)
- [Troubleshooting](#troubleshooting)

## Overview

MCP connectors enable autonomous evidence collection by:

1. **Standardized Interface**: All connectors implement the `MCPConnector` interface
2. **Query Execution**: Convert AI-generated queries into source-specific API calls
3. **Result Normalization**: Transform diverse data formats into standard Events
4. **Error Handling**: Graceful degradation with detailed error messages
5. **Rate Limiting**: Respect API quotas and implement backoff strategies

### Connector Interface

All connectors implement this core interface:

```go
type MCPConnector interface {
    // Execute query and return normalized events
    Query(ctx context.Context, query string, filters map[string]interface{}) ([]types.Event, error)
    
    // Get connector name
    Name() string
    
    // Validate connection and credentials
    Validate(ctx context.Context) error
    
    // Get supported query capabilities
    Capabilities() ConnectorCapabilities
}
```

## Connector Architecture

### Configuration Schema

Each connector uses a common configuration structure:

```go
type ConnectorConfig struct {
    Enabled    bool                   `yaml:"enabled"`     // Master on/off switch
    APIKey     string                 `yaml:"api_key"`     // Authentication token
    Endpoint   string                 `yaml:"endpoint"`    // API base URL
    RateLimit  int                    `yaml:"rate_limit"`  // Max requests per hour
    Timeout    int                    `yaml:"timeout"`     // Request timeout (seconds)
    Extra      map[string]interface{} `yaml:"extra"`       // Connector-specific settings
}
```

### Query Filters

Connectors accept standardized filters:

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `time_range` | string | Time window for results | `"last_90_days"`, `"2025-01-01:2025-03-31"` |
| `repositories` | []string | Repository/project filter | `["backend", "frontend"]` |
| `status` | []string | Status filter | `["open", "closed"]`, `["done"]` |
| `labels` | []string | Label/tag filter | `["security", "compliance"]` |
| `author` | string | Author/assignee filter | `"john.doe"` |
| `limit` | int | Maximum results | `100` |

## GitHub Connector

**Status**: ✅ Fully Implemented

The GitHub connector fetches evidence from GitHub repositories via the GitHub REST API.

### Setup

**1. Generate Personal Access Token**

Navigate to: https://github.com/settings/tokens/new

Required scopes:
- `repo` - Full control of private repositories
- `read:org` - Read org and team membership
- `read:user` - Read user profile data

**2. Configure Connector**

Add to `~/.sdek/config.yaml`:

```yaml
ai:
  connectors:
    github:
      enabled: true
      api_key: ${GITHUB_TOKEN}  # Use environment variable
      endpoint: https://api.github.com
      rate_limit: 5000  # GitHub's default rate limit
      timeout: 30
      extra:
        owner: your-org-name      # GitHub organization/user
        default_repos:            # Default repositories to search
          - auth-service
          - user-management
          - api-gateway
        include_forks: false      # Exclude forks from searches
```

**3. Set Environment Variable**

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

**4. Validate Connection**

```bash
./sdek config validate
```

### Supported Queries

#### 1. Commit Search

**Query Format:**
```
commits: "search terms" [in:repos] [author:name] [date:range]
```

**Examples:**
```bash
# Find commits related to authentication
commits: "authentication OR MFA OR 2FA"

# Find commits in specific repos
commits: "password policy" in:auth-service,user-mgmt

# Find commits by author
commits: "security" author:john.doe

# Find commits in date range
commits: "access control" date:2025-01-01..2025-03-31
```

**AI-Generated Example:**
```
Query: "authentication OR password OR MFA OR access control"
Filters:
  - time_range: "last_90_days"
  - repositories: ["auth-service", "user-management"]
  - limit: 50
```

#### 2. Pull Request Search

**Query Format:**
```
pulls: "search terms" [state:open|closed|all] [label:name]
```

**Examples:**
```bash
# Find security-related PRs
pulls: "security OR vulnerability" state:closed

# Find PRs with specific label
pulls: "compliance" label:security-review

# Find open PRs
pulls: "authentication" state:open
```

#### 3. Issue Search

**Query Format:**
```
issues: "search terms" [state:open|closed] [label:name] [assignee:name]
```

**Examples:**
```bash
# Find closed security issues
issues: "security vulnerability" state:closed

# Find issues by label
issues: "access control" label:compliance

# Find assigned issues
issues: "MFA" assignee:john.doe
```

#### 4. Release Search

**Query Format:**
```
releases: [limit:n] [prerelease:true|false]
```

**Examples:**
```bash
# Get last 10 releases
releases: limit:10

# Get only stable releases
releases: prerelease:false
```

### Query Response

The GitHub connector returns events in this format:

```json
{
  "id": "github-commit-abc123",
  "source_id": "github",
  "event_type": "commit",
  "timestamp": "2025-10-15T10:30:00Z",
  "title": "Add MFA support with TOTP",
  "content": "Implemented multi-factor authentication using time-based one-time passwords...",
  "author": "john.doe",
  "metadata": {
    "repository": "auth-service",
    "sha": "abc123def456",
    "url": "https://github.com/org/auth-service/commit/abc123",
    "files_changed": 5,
    "additions": 234,
    "deletions": 12
  }
}
```

### Rate Limiting

- **GitHub Free**: 5,000 requests/hour (authenticated)
- **GitHub Pro**: 5,000 requests/hour
- **GitHub Enterprise**: Higher limits (configurable)

The connector implements:
- ✅ Exponential backoff on rate limit errors
- ✅ Respect `X-RateLimit-*` headers
- ✅ Pre-flight rate limit checks
- ✅ Graceful degradation on quota exhaustion

### Error Handling

| Error | Description | Resolution |
|-------|-------------|------------|
| `401 Unauthorized` | Invalid or expired token | Regenerate GitHub token |
| `403 Forbidden` | Insufficient token permissions | Add required scopes |
| `404 Not Found` | Repository/organization not found | Check owner/repo names |
| `422 Unprocessable` | Invalid query syntax | Simplify query terms |
| `Rate Limit Exceeded` | API quota exhausted | Wait for reset or increase limit |

### Advanced Configuration

```yaml
ai:
  connectors:
    github:
      enabled: true
      api_key: ${GITHUB_TOKEN}
      endpoint: https://api.github.com
      rate_limit: 5000
      timeout: 30
      extra:
        owner: acme-corp
        default_repos: ["*"]  # All repositories
        
        # Search options
        max_results_per_query: 100
        search_code: true     # Enable code content search
        search_commits: true
        search_issues: true
        search_prs: true
        
        # Filtering
        include_forks: false
        include_archived: false
        min_stars: 0
        
        # Performance
        parallel_requests: 5   # Concurrent API requests
        cache_ttl: 300        # Cache results for 5 minutes
```

## Jira Connector (Planned)

**Status**: 🔨 Planned Implementation

The Jira connector will fetch tickets, comments, and transitions from Jira Cloud or Server.

### Planned Setup

```yaml
ai:
  connectors:
    jira:
      enabled: true
      api_key: ${JIRA_API_TOKEN}
      endpoint: https://your-domain.atlassian.net
      rate_limit: 100  # Jira Cloud default
      timeout: 30
      extra:
        email: your-email@company.com  # For Jira Cloud
        project_key: PROJ                # Default project
        default_jql: 'labels in (security, compliance)'
```

### Planned Queries

```
# Find tickets by JQL
tickets: "project=SEC AND status=Done"

# Find security-related tickets
tickets: "labels in (security, compliance) AND updated >= -90d"

# Find tickets by assignee
tickets: "assignee=john.doe AND status!=Closed"
```

### Planned Features

- ✅ JQL query support
- ✅ Custom field extraction
- ✅ Comment and attachment retrieval
- ✅ Transition history tracking
- ✅ Worklog analysis
- ✅ Sprint/epic filtering

## AWS Connector (Planned)

**Status**: 🔨 Planned Implementation

The AWS connector will fetch CloudTrail logs, IAM events, and AWS Config changes.

### Planned Setup

```yaml
ai:
  connectors:
    aws:
      enabled: true
      # Uses standard AWS SDK credential chain:
      # 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
      # 2. Shared credentials file (~/.aws/credentials)
      # 3. IAM instance profile (for EC2/ECS)
      extra:
        regions:
          - us-east-1
          - us-west-2
        services:
          - cloudtrail
          - iam
          - config
        cloudtrail_bucket: my-cloudtrail-logs
```

### Planned Queries

```
# Find IAM policy changes
cloudtrail: "eventName:CreatePolicy OR eventName:AttachUserPolicy"

# Find authentication events
cloudtrail: "eventName:ConsoleLogin OR eventName:AssumeRole"

# Find config changes
config: "resourceType:AWS::IAM::Role AND configurationItemStatus:ResourceModified"
```

### Planned Features

- ✅ CloudTrail log aggregation across regions
- ✅ IAM event filtering
- ✅ AWS Config compliance tracking
- ✅ Multi-account support
- ✅ S3 bucket log parsing
- ✅ GuardDuty findings integration

## Slack Connector (Planned)

**Status**: 🔨 Planned Implementation

The Slack connector will fetch messages and threads from channels.

### Planned Setup

```yaml
ai:
  connectors:
    slack:
      enabled: true
      api_key: ${SLACK_BOT_TOKEN}
      endpoint: https://slack.com/api
      rate_limit: 50  # Tier 1 methods
      timeout: 30
      extra:
        workspace_id: T1234567890
        default_channels:
          - security
          - compliance
          - incidents
        include_private: false
```

### Planned Queries

```
# Find security discussions
messages: "security OR vulnerability" channel:security

# Find incident reports
messages: "incident" channel:incidents date:last_30_days

# Find mentions
messages: "MFA OR 2FA" mentions:@security-team
```

### Planned Features

- ✅ Channel message search
- ✅ Thread retrieval
- ✅ File attachment metadata
- ✅ User mention tracking
- ✅ Reaction/emoji analysis
- ✅ Workspace-wide search

## Custom Connectors

You can implement custom connectors for proprietary systems by implementing the `MCPConnector` interface.

### Example: Custom SIEM Connector

```go
package connectors

import (
    "context"
    "github.com/pickjonathan/sdek-cli/pkg/types"
)

type SIEMConnector struct {
    config *types.ConnectorConfig
    client *SIEMClient
}

func NewSIEMConnector(cfg *types.ConnectorConfig) (*SIEMConnector, error) {
    client, err := NewSIEMClient(cfg.Endpoint, cfg.APIKey)
    if err != nil {
        return nil, err
    }
    
    return &SIEMConnector{
        config: cfg,
        client: client,
    }, nil
}

func (c *SIEMConnector) Query(ctx context.Context, query string, filters map[string]interface{}) ([]types.Event, error) {
    // Convert query to SIEM-specific format
    siemQuery := convertToSIEMQuery(query, filters)
    
    // Execute query
    results, err := c.client.Search(ctx, siemQuery)
    if err != nil {
        return nil, err
    }
    
    // Normalize results to Event format
    events := make([]types.Event, len(results))
    for i, result := range results {
        events[i] = types.Event{
            ID:        result.ID,
            SourceID:  "siem",
            EventType: result.Type,
            Timestamp: result.Timestamp,
            Title:     result.Title,
            Content:   result.Description,
            Author:    result.User,
            Metadata:  result.Metadata,
        }
    }
    
    return events, nil
}

func (c *SIEMConnector) Name() string {
    return "siem"
}

func (c *SIEMConnector) Validate(ctx context.Context) error {
    return c.client.Ping(ctx)
}

func (c *SIEMConnector) Capabilities() ConnectorCapabilities {
    return ConnectorCapabilities{
        SupportsSearch:    true,
        SupportsFiltering: true,
        SupportsTimeRange: true,
        MaxResultsPerQuery: 1000,
    }
}
```

### Register Custom Connector

```go
// In main.go or connector initialization
registry := connectors.NewRegistry()
siemConnector, err := connectors.NewSIEMConnector(cfg.AI.Connectors["siem"])
if err == nil {
    registry.Register("siem", siemConnector)
}
```

### Custom Connector Configuration

```yaml
ai:
  connectors:
    siem:
      enabled: true
      api_key: ${SIEM_API_KEY}
      endpoint: https://siem.company.com/api
      rate_limit: 100
      timeout: 60
      extra:
        tenant_id: company-tenant
        default_index: security-events
        max_results: 1000
```

## Troubleshooting

### Connection Issues

**Problem**: `Failed to connect to connector`

**Solutions**:
1. Check network connectivity: `curl -I <endpoint>`
2. Verify API endpoint is correct
3. Check firewall/proxy settings
4. Validate DNS resolution

### Authentication Failures

**Problem**: `401 Unauthorized` or `403 Forbidden`

**Solutions**:
1. Verify API key/token is correct and not expired
2. Check token has required permissions/scopes
3. Confirm token is properly set in environment variables
4. Test authentication with API provider's CLI tool

### Rate Limiting

**Problem**: `429 Too Many Requests` or `Rate limit exceeded`

**Solutions**:
1. Reduce `rate_limit` in connector config
2. Increase `timeout` to allow for backoff
3. Enable result caching with `cache_ttl`
4. Reduce `max_results_per_query`
5. Contact API provider to increase quota

### Query Syntax Errors

**Problem**: `422 Unprocessable Entity` or `Invalid query`

**Solutions**:
1. Simplify query terms (remove special characters)
2. Check connector-specific query syntax
3. Test query directly with API provider's search
4. Review connector documentation for supported operators

### Empty Results

**Problem**: Query returns 0 results

**Solutions**:
1. Expand time range filter
2. Remove or loosen filters
3. Check query terms for typos
4. Verify repository/project names are correct
5. Confirm data exists in source system

### Validation Errors

**Problem**: `Connector validation failed`

**Solutions**:

```bash
# Check configuration
./sdek config validate

# Test connector manually
./sdek config get ai.connectors.github

# Verify environment variables
echo $GITHUB_TOKEN
echo $JIRA_API_TOKEN

# Run with debug logging
./sdek config set log.level debug
./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file policies.json
```

### Performance Issues

**Problem**: Queries are slow or timeout

**Solutions**:
1. Increase `timeout` value
2. Reduce `max_results_per_query`
3. Enable connector caching
4. Use more specific filters
5. Increase `parallel_requests` (if supported)

### Debug Mode

Enable verbose logging to troubleshoot connector issues:

```yaml
log:
  level: debug  # Set to debug for verbose output
```

```bash
# Run with debug output
./sdek config set log.level debug
./sdek ai plan --framework SOC2 --section CC6.1 --excerpts-file policies.json 2>&1 | tee debug.log
```

Debug output includes:
- Connector initialization
- Query construction
- API request/response details
- Rate limit status
- Error stack traces

## Best Practices

### 1. Security

- ✅ Use environment variables for API keys (never commit to version control)
- ✅ Rotate API tokens regularly
- ✅ Use minimal required permissions/scopes
- ✅ Store credentials in secure vaults (1Password, HashiCorp Vault)
- ✅ Enable audit logging for connector usage

### 2. Performance

- ✅ Set appropriate rate limits to avoid quota exhaustion
- ✅ Use caching for repeated queries
- ✅ Implement parallel requests where supported
- ✅ Use specific filters to reduce result sets
- ✅ Monitor API usage and costs

### 3. Reliability

- ✅ Implement exponential backoff on failures
- ✅ Handle partial failures gracefully
- ✅ Validate queries before execution
- ✅ Log all connector operations
- ✅ Test connectors before autonomous mode

### 4. Cost Management

- ✅ Track API usage per connector
- ✅ Set budget alerts with API providers
- ✅ Use auto-approve carefully in production
- ✅ Review AI-generated queries before execution
- ✅ Implement query result caching

## Next Steps

- Review [README.md](../README.md#autonomous-evidence-collection-experimental) for autonomous mode overview
- See [AI_WORKFLOW_ARCHITECTURE.md](./AI_WORKFLOW_ARCHITECTURE.md) for technical architecture
- Check [examples/](../examples/) for connector usage patterns
- Run `./sdek ai plan --help` for command options

## Support

For connector-specific issues:
- **GitHub**: https://docs.github.com/en/rest
- **Jira**: https://developer.atlassian.com/cloud/jira/platform/rest/v3/
- **AWS**: https://docs.aws.amazon.com/cloudtrail/
- **Slack**: https://api.slack.com/docs

For sdek-cli issues:
- Open an issue: https://github.com/pickjonathan/sdek-cli/issues
- Email: jonathan@example.com
