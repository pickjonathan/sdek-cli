package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// AuditLogger logs all tool executions for compliance and debugging.
type AuditLogger struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	enabled  bool
}

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Event       string                 `json:"event"` // "started" | "completed" | "failed"
	ToolName    string                 `json:"tool_name"`
	Arguments   map[string]interface{} `json:"arguments,omitempty"`
	Context     map[string]string      `json:"context,omitempty"`
	Success     bool                   `json:"success,omitempty"`
	Error       string                 `json:"error,omitempty"`
	LatencyMs   int                    `json:"latency_ms,omitempty"`
	OutputSize  int                    `json:"output_size,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Approved    bool                   `json:"approved,omitempty"`
	RiskLevel   string                 `json:"risk_level,omitempty"`
}

// NewAuditLogger creates a new audit logger that writes to the specified file.
func NewAuditLogger(filePath string) (*AuditLogger, error) {
	if filePath == "" {
		// Disabled if no path provided
		return &AuditLogger{enabled: false}, nil
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filePath[:len(filePath)-len(filePath[len(filePath)-1:])], 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		file:     file,
		filePath: filePath,
		enabled:  true,
	}, nil
}

// LogStart logs the start of a tool execution.
func (l *AuditLogger) LogStart(call *types.ToolCall) {
	if !l.enabled {
		return
	}

	entry := AuditEntry{
		Timestamp: time.Now(),
		Event:     "started",
		ToolName:  call.ToolName,
		Arguments: call.Arguments,
		Context:   call.Context,
		UserID:    call.Context["user_id"],
		SessionID: call.Context["session_id"],
		Approved:  call.Context["approved"] == "true",
	}

	l.write(entry)
}

// LogComplete logs the completion of a tool execution.
func (l *AuditLogger) LogComplete(call *types.ToolCall, result *types.ToolExecutionResult, err error) {
	if !l.enabled {
		return
	}

	event := "completed"
	if err != nil || (result != nil && !result.Success) {
		event = "failed"
	}

	entry := AuditEntry{
		Timestamp:  time.Now(),
		Event:      event,
		ToolName:   call.ToolName,
		Context:    call.Context,
		UserID:     call.Context["user_id"],
		SessionID:  call.Context["session_id"],
	}

	if result != nil {
		entry.Success = result.Success
		entry.Error = result.Error
		entry.LatencyMs = int(result.LatencyMS)
		entry.OutputSize = l.estimateOutputSize(result.Output)
	}

	if err != nil && entry.Error == "" {
		entry.Error = err.Error()
	}

	l.write(entry)
}

// LogApproval logs a user approval decision.
func (l *AuditLogger) LogApproval(call *types.ToolCall, approved bool, riskLevel string) {
	if !l.enabled {
		return
	}

	event := "approved"
	if !approved {
		event = "denied"
	}

	entry := AuditEntry{
		Timestamp: time.Now(),
		Event:     event,
		ToolName:  call.ToolName,
		Arguments: call.Arguments,
		Context:   call.Context,
		UserID:    call.Context["user_id"],
		SessionID: call.Context["session_id"],
		Approved:  approved,
		RiskLevel: riskLevel,
	}

	l.write(entry)
}

// write writes an audit entry to the log file as JSON.
func (l *AuditLogger) write(entry AuditEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Log to stderr if marshaling fails
		fmt.Fprintf(os.Stderr, "audit: failed to marshal entry: %v\n", err)
		return
	}

	// Write JSON line
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		// Log to stderr if write fails
		fmt.Fprintf(os.Stderr, "audit: failed to write entry: %v\n", err)
		return
	}

	// Flush immediately for reliability
	l.file.Sync()
}

// estimateOutputSize estimates the size of output data in bytes.
func (l *AuditLogger) estimateOutputSize(output interface{}) int {
	if output == nil {
		return 0
	}

	// Serialize to JSON to estimate size
	data, err := json.Marshal(output)
	if err != nil {
		return 0
	}

	return len(data)
}

// Close closes the audit log file.
func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}

	return nil
}

// Rotate rotates the audit log file by renaming the current file
// and creating a new one.
func (l *AuditLogger) Rotate() error {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Close current file
	if l.file != nil {
		l.file.Close()
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := fmt.Sprintf("%s.%s", l.filePath, timestamp)
	if err := os.Rename(l.filePath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate audit log: %w", err)
	}

	// Open new file
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new audit log: %w", err)
	}

	l.file = file
	return nil
}
