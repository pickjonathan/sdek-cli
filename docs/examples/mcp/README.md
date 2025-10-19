# MCP Configuration Examples

This directory contains example MCP (Model Context Protocol) configuration files for popular integrations.

## Quick Start

**New to MCP testing?** Start here: [How to Verify AWS MCP Usage](./HOW_TO_VERIFY_AWS_MCP_USAGE.md)

This guide shows you exactly how to verify that the AI is using AWS MCP tools for evidence collection, with:
- Quick 5-indicator checklist
- Automated test script
- Debug log patterns to watch for
- Troubleshooting tips

## Available Examples

### GitHub (`github.json`)
Integrates with GitHub API for accessing:
- Commits and commit history
- Pull requests
- Issues
- Branches and tags
- Code search

**Setup**:
```bash
# Copy to your MCP config directory
cp docs/examples/mcp/github.json ~/.sdek/mcp/

# Set required environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_OWNER="your-org-or-username"
export GITHUB_REPO="your-repo"

# Test the configuration
sdek mcp validate ~/.sdek/mcp/github.json
sdek mcp test github
```

### Slack (`slack.json`)
Integrates with Slack API for accessing:
- Messages and conversations
- Channel information and history
- User data
- Search functionality

**Setup**:
```bash
# Copy to your MCP config directory
cp docs/examples/mcp/slack.json ~/.sdek/mcp/

# Set required environment variables
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_TEAM_ID="T0123456789"

# Test the configuration
sdek mcp validate ~/.sdek/mcp/slack.json
sdek mcp test slack
```

### Jira (`jira.json`)
Integrates with Jira API for accessing:
- Issues and workflows
- Projects and boards
- Sprints
- User search

**Setup**:
```bash
# Copy to your MCP config directory
cp docs/examples/mcp/jira.json ~/.sdek/mcp/

# Set required environment variables
export JIRA_HOST="your-domain.atlassian.net"
export JIRA_EMAIL="your@email.com"
export JIRA_API_TOKEN="your-api-token"

# Test the configuration
sdek mcp validate ~/.sdek/mcp/jira.json
sdek mcp test jira
```

## Configuration File Structure

All MCP configuration files follow this schema:

```json
{
  "schemaVersion": "1.0.0",          // Required: Schema version
  "name": "tool-name",                // Required: Unique identifier
  "description": "Tool description",  // Optional: Human-readable description
  "command": "executable",            // Required: Command to run (npx, uvx, python, etc.)
  "args": ["arg1", "arg2"],          // Required: Command arguments
  "transport": "stdio",               // Required: "stdio" or "http"
  "env": {                            // Optional: Environment variables
    "KEY": "${ENV_VAR}"               // Supports ${} expansion
  },
  "capabilities": [                   // Required: List of supported operations
    "tool.resource.action"
  ],
  "timeout": "30s",                   // Optional: Request timeout (default: 30s)
  "retryPolicy": {                    // Optional: Retry configuration
    "maxAttempts": 3,
    "backoff": "exponential"
  },
  "metadata": {                       // Optional: Additional metadata
    "category": "integration-type",
    "documentation": "https://...",
    "setup": ["instruction1", "..."]
  }
}
```

## Environment Variable Expansion

Configuration files support environment variable expansion using `${VAR_NAME}` syntax:

```json
{
  "env": {
    "API_TOKEN": "${MY_SERVICE_TOKEN}",
    "BASE_URL": "${SERVICE_URL}"
  }
}
```

## Configuration Discovery

SDEK discovers MCP configuration files from multiple locations (in order of precedence):

1. **Custom path**: `$SDEK_MCP_PATH` environment variable
2. **Project-specific**: `./.sdek/mcp/` (current directory)
3. **User global**: `~/.sdek/mcp/` (home directory)

Files with the same name in higher-precedence locations override lower-precedence ones.

## Validation

Validate your configuration files before using them:

```bash
# Validate a single file
sdek mcp validate path/to/config.json

# Validate all discovered configs
sdek mcp list
```

## Testing Connections

Test that your MCP tool can connect successfully:

```bash
# Test a specific tool
sdek mcp test tool-name

# List all tools with status
sdek mcp list
```

## Creating Custom Integrations

To create your own MCP integration:

1. **Copy a template**:
   ```bash
   cp docs/examples/mcp/github.json ~/.sdek/mcp/my-service.json
   ```

2. **Edit the configuration**:
   - Change `name` to a unique identifier
   - Update `command` and `args` for your MCP server
   - Define `capabilities` for your service
   - Add required environment variables in `env`

3. **Validate and test**:
   ```bash
   sdek mcp validate ~/.sdek/mcp/my-service.json
   sdek mcp test my-service
   ```

4. **Use in AI workflows**:
   ```bash
   sdek ai analyze --agent evidence-collector
   # The agent will automatically discover and use configured MCP tools
   ```

## Common Issues

### Tool shows as "degraded" or "offline"
- Verify environment variables are set correctly
- Check that the MCP server executable is available (`npx`, `uvx`, etc.)
- Test manually: `npx -y @modelcontextprotocol/server-github`
- Check logs: `sdek mcp test tool-name --verbose`

### "Permission denied" errors
- Ensure your API tokens have the required scopes/permissions
- Check token expiration
- Verify network connectivity to the service

### Environment variables not expanding
- Use `${VAR_NAME}` syntax, not `$VAR_NAME`
- Ensure variables are exported: `export VAR_NAME=value`
- Check variable spelling and case

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [MCP Server Implementations](https://github.com/modelcontextprotocol/servers)
- [SDEK MCP Documentation](../../commands.md#mcp-commands)
