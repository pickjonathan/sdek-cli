package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/store"
)

func TestIngestCommand(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	tests := []struct {
		name           string
		args           []string
		expectedErrMsg string
		expectError    bool
	}{
		{
			name:        "help flag",
			args:        []string{"ingest", "--help"},
			expectError: false,
		},
		{
			name:           "missing source flag",
			args:           []string{"ingest"},
			expectedErrMsg: "--source flag is required",
			expectError:    true,
		},
		{
			name:           "invalid source type",
			args:           []string{"ingest", "--source", "invalid"},
			expectedErrMsg: "invalid source type",
			expectError:    true,
		},
		{
			name:        "valid git source",
			args:        []string{"ingest", "--source", "git"},
			expectError: false,
		},
		{
			name:        "valid jira source",
			args:        []string{"ingest", "--source", "jira"},
			expectError: false,
		},
		{
			name:        "valid slack source",
			args:        []string{"ingest", "--source", "slack"},
			expectError: false,
		},
		{
			name:        "valid cicd source",
			args:        []string{"ingest", "--source", "cicd"},
			expectError: false,
		},
		{
			name:        "valid docs source",
			args:        []string{"ingest", "--source", "docs"},
			expectError: false,
		},
		{
			name:        "with events count",
			args:        []string{"ingest", "--source", "git", "--events", "25"},
			expectError: false,
		},
		{
			name:        "with seed",
			args:        []string{"ingest", "--source", "git", "--seed", "42"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command flags
			ingestSource = ""
			ingestEvents = 0
			ingestSeed = 0

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
				t.Errorf("expected error message to contain %q, got %q", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestIngestCommandGeneratesEvents(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	// Run ingest command
	rootCmd.SetArgs([]string{"ingest", "--source", "git", "--events", "30", "--seed", "42"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ingest command failed: %v", err)
	}

	// Load state and verify events were generated
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify events were created
	if len(state.Events) != 30 {
		t.Errorf("expected 30 events, got %d", len(state.Events))
	}

	// Verify source was created
	if len(state.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(state.Sources))
	}

	// Verify source type is correct
	if len(state.Sources) > 0 && state.Sources[0].Type != "git" {
		t.Errorf("expected source type to be 'git', got '%s'", state.Sources[0].Type)
	}
}

func TestIngestCommandMultipleSources(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	sources := []string{"git", "jira", "slack"}

	for _, source := range sources {
		rootCmd.SetArgs([]string{"ingest", "--source", source, "--events", "20"})
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("ingest command failed for source %s: %v", source, err)
		}
	}

	// Load state and verify all sources were created
	state, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if len(state.Sources) != 3 {
		t.Errorf("expected 3 sources, got %d", len(state.Sources))
	}

	// Verify total events (20 per source * 3 sources = 60)
	expectedEvents := 60
	if len(state.Events) != expectedEvents {
		t.Errorf("expected %d events, got %d", expectedEvents, len(state.Events))
	}
}

func TestIngestCommandOutputFormat(t *testing.T) {
	// Create temporary directory for test state
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	dataDir = filepath.Join(tmpDir, ".sdek")

	// Run ingest command
	rootCmd.SetArgs([]string{"ingest", "--source", "git", "--events", "15"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ingest command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected summary information
	expectedStrings := []string{
		"events",
		"git",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, but it didn't", expected)
		}
	}
}
