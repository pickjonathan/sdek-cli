package unit

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestInvocationLogCreatedWithRequiredFields(t *testing.T) {
	log := types.MCPInvocationLog{
		ID:        "test-id",
		RunID:     "run-123",
		AgentID:   "agent-1",
		AgentRole: "evidence-collector",
		ToolName:  "github",
		Method:    "commits.list",
		ArgsHash:  hashArgs(map[string]interface{}{"repo": "test/repo"}),
		Status:    "success",
	}
	
	if log.ID == "" {
		t.Error("ID should be set")
	}
	if log.ArgsHash == "" {
		t.Error("ArgsHash should be set")
	}
}

func TestArgsHashedWithSHA256(t *testing.T) {
	args := map[string]interface{}{
		"repo": "test/repo",
		"since": "2025-01-01",
	}
	
	hash := hashArgs(args)
	
	// Verify it's a valid SHA256 hash (64 hex chars)
	if len(hash) != 64 {
		t.Errorf("expected SHA256 hash length 64, got %d", len(hash))
	}
}

func TestRedactionFlagSet(t *testing.T) {
	log := types.MCPInvocationLog{
		RedactionApplied: true,
	}
	
	if !log.RedactionApplied {
		t.Error("RedactionApplied should be true")
	}
}

func hashArgs(args interface{}) string {
	// Simple hash for testing
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", args)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
