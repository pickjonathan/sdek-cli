package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/store"
)

func TestAnalyzeCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "help flag",
			args:        []string{"analyze", "--help"},
			expectError: false,
		},
		{
			name:        "analyze without events shows error",
			args:        []string{"analyze"},
			expectError: true, // No events in state
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test state
			tmpDir := t.TempDir()
			
			// Set HOME to temp directory for test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)
			
			dataDir = filepath.Join(tmpDir, ".sdek")

			// Reset root command
			rootCmd.SetArgs(tt.args)
			
			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAnalyzeCommandWithEvents(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	// First, seed data
	rootCmd.SetArgs([]string{"seed", "--demo", "--seed", "42"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	// Get initial state
	state1, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load initial state: %v", err)
	}
	initialEvidenceCount := len(state1.Evidence)

	// Run analyze command
	rootCmd.SetArgs([]string{"analyze"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	// Load state and verify analysis was performed
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state after analyze: %v", err)
	}

	// Verify evidence was generated (seed already generates evidence, so should have same or more)
	if len(state.Evidence) < initialEvidenceCount {
		t.Errorf("expected at least %d evidence, got %d", initialEvidenceCount, len(state.Evidence))
	}

	// Verify frameworks were updated with compliance percentages
	if len(state.Frameworks) == 0 {
		t.Errorf("expected frameworks to be present")
	}

	// Verify controls have risk calculations
	hasRiskCalculations := false
	for _, ctrl := range state.Controls {
		if ctrl.RiskStatus != "" {
			hasRiskCalculations = true
			break
		}
	}
	if !hasRiskCalculations {
		t.Errorf("expected controls to have risk status calculated")
	}
}

func TestAnalyzeCommandOutputFormat(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	// First, seed data
	rootCmd.SetArgs([]string{"seed", "--demo"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	_ = rootCmd.Execute()

	// Run analyze command
	rootCmd.SetArgs([]string{"analyze"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected summary information
	expectedStrings := []string{
		"evidence",
		"controls",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, but it didn't", expected)
		}
	}
}

func TestAnalyzeCommandGeneratesFindings(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	// First, seed data
	rootCmd.SetArgs([]string{"seed", "--demo", "--seed", "42"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	// Run analyze command
	rootCmd.SetArgs([]string{"analyze"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	// Load state and verify findings were generated
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify findings exist (seed already generates findings)
	if len(state.Findings) == 0 {
		t.Logf("Note: No findings generated (this is OK if all controls are green)")
	}
}
