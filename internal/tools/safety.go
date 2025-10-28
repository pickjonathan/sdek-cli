package tools

import (
	"fmt"
	"strings"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// SafetyValidator implements three-tier safety validation for tool calls.
// It classifies tool calls into safe, interactive, or resource-modifying categories
// and determines if user approval is required.
type SafetyValidator struct {
	// Interactive commands that require terminal interaction
	interactiveCommands []string

	// Dangerous verbs that indicate resource modification
	dangerousVerbs []string

	// Custom user-defined rules
	denyList  []string // Patterns that always require approval
	allowList []string // Patterns that skip validation
}

// NewSafetyValidator creates a new safety validator with default rules.
func NewSafetyValidator() *SafetyValidator {
	return &SafetyValidator{
		interactiveCommands: defaultInteractiveCommands(),
		dangerousVerbs:      defaultDangerousVerbs(),
		denyList:            []string{},
		allowList:           []string{},
	}
}

// Analyze performs three-tier safety analysis on a tool call.
// Returns a ToolCallAnalysis with safety assessment.
func (v *SafetyValidator) Analyze(call *types.ToolCall) *types.ToolCallAnalysis {
	analysis := &types.ToolCallAnalysis{
		IsInteractive:     false,
		ModifiesResource:  false,
		RequiresApproval:  false,
		RiskLevel:         types.RiskLevelLow,
		Rationale:         "Tool call is read-only and safe",
	}

	// Extract command string from arguments
	command := v.extractCommand(call)

	// Check allow list first (skip validation if matched)
	if v.isInAllowList(command) {
		analysis.Rationale = "Tool call matched allow list"
		return analysis
	}

	// Check deny list (always require approval)
	if v.isInDenyList(command) {
		analysis.RequiresApproval = true
		analysis.RiskLevel = types.RiskLevelHigh
		analysis.Rationale = "Tool call matched deny list"
		return analysis
	}

	// Tier 1: Interactive command detection
	if v.isInteractive(command) {
		analysis.IsInteractive = true
		analysis.RequiresApproval = true
		analysis.RiskLevel = types.RiskLevelHigh
		analysis.Rationale = "Tool requires interactive terminal (vim, bash, python REPL, etc.)"
		return analysis
	}

	// Tier 2: Resource modification detection
	if v.modifiesResources(command) {
		analysis.ModifiesResource = true
		analysis.RequiresApproval = true
		analysis.RiskLevel = types.RiskLevelMedium
		analysis.Rationale = v.getModificationReason(command)
		return analysis
	}

	// Tier 3: Safe operations (no approval needed)
	return analysis
}

// isInteractive checks if a command requires interactive terminal.
func (v *SafetyValidator) isInteractive(command string) bool {
	cmd := v.getFirstWord(command)
	for _, interactive := range v.interactiveCommands {
		if cmd == interactive {
			return true
		}
	}
	return false
}

// modifiesResources checks if a command may modify system resources.
func (v *SafetyValidator) modifiesResources(command string) bool {
	commandLower := strings.ToLower(command)
	for _, verb := range v.dangerousVerbs {
		if strings.Contains(commandLower, verb) {
			return true
		}
	}
	return false
}

// isInAllowList checks if command matches allow list patterns.
func (v *SafetyValidator) isInAllowList(command string) bool {
	commandLower := strings.ToLower(command)
	for _, pattern := range v.allowList {
		patternLower := strings.ToLower(pattern)
		if strings.Contains(commandLower, patternLower) {
			return true
		}
	}
	return false
}

// isInDenyList checks if command matches deny list patterns.
func (v *SafetyValidator) isInDenyList(command string) bool {
	commandLower := strings.ToLower(command)
	for _, pattern := range v.denyList {
		patternLower := strings.ToLower(pattern)
		if strings.Contains(commandLower, patternLower) {
			return true
		}
	}
	return false
}

// extractCommand extracts the command string from tool call arguments.
func (v *SafetyValidator) extractCommand(call *types.ToolCall) string {
	// Try common argument field names
	if cmd, ok := call.Arguments["command"].(string); ok {
		return cmd
	}
	if cmd, ok := call.Arguments["cmd"].(string); ok {
		return cmd
	}
	if query, ok := call.Arguments["query"].(string); ok {
		return query
	}

	// If no command field, return tool name as fallback
	return call.ToolName
}

// getFirstWord extracts the first word from a command string.
func (v *SafetyValidator) getFirstWord(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// getModificationReason returns a human-readable reason for resource modification flag.
func (v *SafetyValidator) getModificationReason(command string) string {
	commandLower := strings.ToLower(command)

	// Check which dangerous verb was matched
	for _, verb := range v.dangerousVerbs {
		if strings.Contains(commandLower, verb) {
			return fmt.Sprintf("Command contains potentially destructive verb: '%s'", verb)
		}
	}

	return "Command may modify system resources"
}

// SetDenyList sets custom deny list patterns.
func (v *SafetyValidator) SetDenyList(patterns []string) {
	v.denyList = patterns
}

// SetAllowList sets custom allow list patterns.
func (v *SafetyValidator) SetAllowList(patterns []string) {
	v.allowList = patterns
}

// AddDenyPattern adds a single pattern to the deny list.
func (v *SafetyValidator) AddDenyPattern(pattern string) {
	v.denyList = append(v.denyList, pattern)
}

// AddAllowPattern adds a single pattern to the allow list.
func (v *SafetyValidator) AddAllowPattern(pattern string) {
	v.allowList = append(v.allowList, pattern)
}

// defaultInteractiveCommands returns the default list of interactive commands.
func defaultInteractiveCommands() []string {
	return []string{
		// Editors
		"vim", "vi", "nano", "emacs", "pico", "ed",

		// Interactive shells
		"bash", "sh", "zsh", "fish", "ksh", "csh", "tcsh",

		// REPLs
		"python", "python3", "ipython",
		"node", "nodejs",
		"irb", "ruby",
		"php",
		"psql", "mysql", "mongo", "redis-cli",

		// Interactive tools
		"less", "more",
		"top", "htop",
		"watch",
	}
}

// defaultDangerousVerbs returns the default list of dangerous operation verbs.
func defaultDangerousVerbs() []string {
	return []string{
		// Destructive operations
		"delete", "rm", "remove",
		"destroy", "terminate", "kill",
		"drop", "truncate", "purge",

		// Modification operations
		"create", "add", "insert",
		"update", "modify", "edit", "patch",
		"set", "put", "post",
		"apply", "deploy",

		// Execution operations
		"exec", "execute", "run",
		"ssh", "sudo",

		// AWS destructive operations
		"terminate-instances",
		"delete-stack",
		"delete-bucket",
		"revoke-security-group",

		// Kubernetes destructive operations
		"delete",
		"scale",
		"drain",
		"cordon",

		// Git destructive operations
		"push --force",
		"reset --hard",
		"rebase",

		// Database operations
		"drop table",
		"truncate",
		"delete from",
	}
}
