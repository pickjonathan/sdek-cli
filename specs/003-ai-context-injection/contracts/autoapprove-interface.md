# Contract: Auto-Approve Policy Matcher

**Feature**: 003-ai-context-injection  
**Purpose**: Define interface for matching evidence plan items against auto-approve policies

---

## Interface Definition

```go
package ai

// AutoApproveMatcher determines if a plan item should be auto-approved.
type AutoApproveMatcher interface {
    // Matches checks if source+query matches auto-approve policy.
    //
    // Parameters:
    //   source: MCP connector name (e.g., "github", "aws", "jira")
    //   query: Search query or filter (e.g., "auth*", "INFOSEC-123")
    //
    // Returns:
    //   true if query matches any pattern for the source
    //   false if source not whitelisted OR no patterns match
    //
    // Behavior:
    //   - MUST return false if policy.Enabled = false
    //   - MUST be case-insensitive
    //   - MUST support glob wildcards: *, **, ?
    //   - MUST be fast (<1μs per match)
    Matches(source, query string) bool
}
```

---

## Example Patterns

**Config** (YAML):
```yaml
ai:
  autonomous:
    autoApprove:
      enabled: true
      rules:
        github: ["auth*", "*login*", "mfa*"]
        aws: ["iam*", "security*"]
        jira: ["INFOSEC-*"]
```

**Matching Examples**:
- `Matches("github", "authentication")` → true (matches "auth*")
- `Matches("github", "user-login-flow")` → true (matches "*login*")
- `Matches("github", "payment")` → false (no match)
- `Matches("slack", "security-channel")` → false (source not whitelisted)

---

## Testing

**Unit Tests**:
- Wildcard matching: `auth*` matches `authentication`, not `unauth`
- Case-insensitive: `AUTH*` matches `authentication`
- Mid-wildcard: `*login*` matches `user-login-flow`
- Source whitelist: Unlisted source returns false
- Disabled policy: Returns false for all
