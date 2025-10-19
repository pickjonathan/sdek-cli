package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReportCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "help flag",
			args:        []string{"report", "--help"},
			expectError: false,
		},
		{
			name:        "report without data shows error",
			args:        []string{"report"},
			expectError: true, // No data in state
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

func TestReportCommandGeneratesJSON(t *testing.T) {
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
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("seed command failed: %v", err)
	}

	// Set output path in temp directory
	outputPath := filepath.Join(tmpDir, "report.json")

	// Run report command
	rootCmd.SetArgs([]string{"report", "--output", outputPath})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("report command failed: %v", err)
	}

	// Verify report file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("expected report file to be created at %s", outputPath)
	}

	// Verify report file is not empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat report file: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("expected report file to have content, got 0 bytes")
	}
}

func TestReportCommandRoleFilter(t *testing.T) {
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

	roles := []string{"manager", "engineer"}

	for _, role := range roles {
		outputPath := filepath.Join(tmpDir, "report-"+role+".json")
		
		// Run report command with role filter
		rootCmd.SetArgs([]string{"report", "--output", outputPath, "--role", role})
		buf = new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("report command failed for role %s: %v", role, err)
		}

		// Verify report file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("expected report file to be created for role %s at %s", role, outputPath)
		}
	}
}

func TestReportCommandOutputFormat(t *testing.T) {
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

	outputPath := filepath.Join(tmpDir, "report.json")

	// Run report command
	rootCmd.SetArgs([]string{"report", "--output", outputPath})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("report command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected summary information
	expectedStrings := []string{
		"report",
		outputPath,
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, but it didn't", expected)
		}
	}
}
