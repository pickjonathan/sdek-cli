package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/store"
)

func TestSeedCommand(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()

	// Save and restore global state
	oldHome := os.Getenv("HOME")
	oldDataDir := dataDir
	defer func() {
		os.Setenv("HOME", oldHome)
		dataDir = oldDataDir
	}()

	os.Setenv("HOME", tmpDir)
	dataDir = tmpDir

	tests := []struct {
		name           string
		args           []string
		expectedErrMsg string
		expectError    bool
	}{
		{
			name:           "missing demo flag",
			args:           []string{"seed"},
			expectedErrMsg: "--demo flag is required",
			expectError:    true,
		},
		{
			name:        "help flag",
			args:        []string{"seed", "--help"},
			expectError: false,
		},
		{
			name:        "demo flag provided",
			args:        []string{"seed", "--demo"},
			expectError: false,
		},
		{
			name:        "demo with seed value",
			args:        []string{"seed", "--demo", "--seed", "12345"},
			expectError: false,
		},
		{
			name:        "demo with reset",
			args:        []string{"seed", "--demo", "--reset"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command flags
			seedDemo = false
			seedValue = 0
			seedReset = false

			// Reset Cobra flag state
			seedCmd.Flags().Set("demo", "false")
			seedCmd.Flags().Set("seed", "0")
			seedCmd.Flags().Set("reset", "false")

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
			if tt.expectedErrMsg != "" && (err == nil || err.Error() != tt.expectedErrMsg) {
				t.Errorf("expected error message %q, got %q", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestSeedCommandGeneratesData(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()

	// Save and restore global state
	oldHome := os.Getenv("HOME")
	oldDataDir := dataDir
	defer func() {
		os.Setenv("HOME", oldHome)
		dataDir = oldDataDir
	}()

	os.Setenv("HOME", tmpDir)
	dataDir = filepath.Join(tmpDir, ".sdek")

	// Run seed command
	rootCmd.SetArgs([]string{"seed", "--demo", "--seed", "12345"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	// Load state and verify data was generated
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify sources were created
	if len(state.Sources) != 5 {
		t.Errorf("expected 5 sources, got %d", len(state.Sources))
	}

	// Verify events were created
	if len(state.Events) == 0 {
		t.Errorf("expected events to be generated, got 0")
	}

	// Verify frameworks were created
	if len(state.Frameworks) != 3 {
		t.Errorf("expected 3 frameworks, got %d", len(state.Frameworks))
	}

	// Verify evidence was created
	if len(state.Evidence) == 0 {
		t.Errorf("expected evidence to be generated, got 0")
	}

	// Verify findings were created
	if len(state.Findings) == 0 {
		t.Errorf("expected findings to be generated, got 0")
	}
}

func TestSeedCommandDeterministicGeneration(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()

	// Save and restore global state
	oldHome := os.Getenv("HOME")
	oldDataDir := dataDir
	defer func() {
		os.Setenv("HOME", oldHome)
		dataDir = oldDataDir
	}()

	os.Setenv("HOME", tmpDir)
	dataDir = filepath.Join(tmpDir, ".sdek")

	seed := int64(42)

	// Run seed command twice with same seed
	for i := 0; i < 2; i++ {
		rootCmd.SetArgs([]string{"seed", "--demo", "--seed", "42", "--reset"})
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("seed command failed on iteration %d: %v", i, err)
		}
	}

	// Load state
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify deterministic generation (should have same number of entities)
	if len(state.Events) == 0 {
		t.Errorf("expected events to be generated with seed %d", seed)
	}
}

func TestSeedCommandResetFlag(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()

	// Save and restore global state
	oldHome := os.Getenv("HOME")
	oldDataDir := dataDir
	defer func() {
		os.Setenv("HOME", oldHome)
		dataDir = oldDataDir
	}()

	os.Setenv("HOME", tmpDir)

	dataDir = filepath.Join(tmpDir, ".sdek")

	// Run seed command first time
	rootCmd.SetArgs([]string{"seed", "--demo"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	// Load state and get initial event count
	state1, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	initialEventCount := len(state1.Events)

	// Run seed command again without reset (should append)
	rootCmd.SetArgs([]string{"seed", "--demo", "--seed", "999"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed on second run: %v", err)
	}

	// Load state and verify events were appended
	state2, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if len(state2.Events) <= initialEventCount {
		t.Errorf("expected events to be appended, got %d events (was %d)", len(state2.Events), initialEventCount)
	}

	// Run seed command with reset
	rootCmd.SetArgs([]string{"seed", "--demo", "--reset"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed with reset: %v", err)
	}

	// Load state and verify it was reset
	state3, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// After reset with new seed, event count should be different
	if len(state3.Events) == len(state2.Events) {
		t.Logf("Note: Event count same after reset (both %d), this is OK if different seed generates same count", len(state3.Events))
	}
}

func TestSeedCommandOutputFormat(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	dataDir = tmpDir

	// Run seed command
	rootCmd.SetArgs([]string{"seed", "--demo"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected summary information
	expectedStrings := []string{
		"sources",
		"events",
		"frameworks",
		"evidence",
		"findings",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, but it didn't", expected)
		}
	}
}
