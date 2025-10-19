package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// AuditLogger writes MCP invocation logs to a JSONL file.
type AuditLogger struct {
	logPath string
	mu      sync.Mutex
}

// NewAuditLogger creates a new audit logger with default path.
func NewAuditLogger() *AuditLogger {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".sdek", "logs")
	logPath := filepath.Join(logDir, "mcp-invocations.jsonl")

	os.MkdirAll(logDir, 0755)

	return &AuditLogger{
		logPath: logPath,
	}
}

// Write appends an invocation log entry to the audit file.
func (a *AuditLogger) Write(ctx context.Context, log *types.MCPInvocationLog) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	f, err := os.OpenFile(a.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// Rotate removes log entries older than 7 days.
func (a *AuditLogger) Rotate() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	f, err := os.Open(a.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open audit log for rotation: %w", err)
	}
	defer f.Close()

	var logs []types.MCPInvocationLog
	decoder := json.NewDecoder(f)

	cutoff := time.Now().AddDate(0, 0, -7)

	for decoder.More() {
		var log types.MCPInvocationLog
		if err := decoder.Decode(&log); err != nil {
			continue
		}

		if log.Timestamp.After(cutoff) {
			logs = append(logs, log)
		}
	}

	f.Close()

	tmpPath := a.logPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	for _, log := range logs {
		data, _ := json.Marshal(log)
		tmpFile.Write(append(data, '\n'))
	}

	tmpFile.Close()

	if err := os.Rename(tmpPath, a.logPath); err != nil {
		return fmt.Errorf("failed to rotate log: %w", err)
	}

	return nil
}
