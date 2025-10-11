package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewAutoSave(t *testing.T) {
	state := NewState()
	autoSave := NewAutoSave(state)

	if autoSave == nil {
		t.Fatal("NewAutoSave returned nil")
	}

	if autoSave.state != state {
		t.Error("AutoSave state pointer doesn't match")
	}

	if autoSave.debounceTime != 2*time.Second {
		t.Errorf("Expected debounce time 2s, got %v", autoSave.debounceTime)
	}

	if autoSave.IsRunning() {
		t.Error("AutoSave should not be running initially")
	}
}

func TestAutoSaveStartStop(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	state := NewState()
	autoSave := NewAutoSave(state)

	ctx := context.Background()
	if err := autoSave.Start(ctx); err != nil {
		t.Fatalf("Failed to start auto-save: %v", err)
	}

	if !autoSave.IsRunning() {
		t.Error("AutoSave should be running after Start()")
	}

	// Test starting again should fail
	if err := autoSave.Start(ctx); err == nil {
		t.Error("Expected error when starting already running auto-save")
	}

	if err := autoSave.Stop(); err != nil {
		t.Errorf("Failed to stop auto-save: %v", err)
	}

	if autoSave.IsRunning() {
		t.Error("AutoSave should not be running after Stop()")
	}
}

func TestAutoSaveDebounce(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	state := NewState()
	autoSave := NewAutoSave(state)

	ctx := context.Background()
	if err := autoSave.Start(ctx); err != nil {
		t.Fatalf("Failed to start auto-save: %v", err)
	}
	defer autoSave.Stop()

	// Add some data
	source := types.Source{
		ID:         types.SourceTypeGit,
		Name:       "Test Git Repo",
		Type:       types.SourceTypeGit,
		Status:     "active",
		EventCount: 25,
		LastSync:   time.Now(),
		Enabled:    true,
	}
	if err := state.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Mark dirty and wait for debounce
	autoSave.MarkDirty()

	// Wait slightly longer than debounce time
	time.Sleep(2500 * time.Millisecond)

	// Load the state and verify it was saved
	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if len(loadedState.Sources) != 1 {
		t.Errorf("Expected 1 source after auto-save, got %d", len(loadedState.Sources))
	}
}

func TestAutoSaveMultipleMarks(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	state := NewState()
	autoSave := NewAutoSave(state)

	ctx := context.Background()
	if err := autoSave.Start(ctx); err != nil {
		t.Fatalf("Failed to start auto-save: %v", err)
	}
	defer autoSave.Stop()

	// Mark dirty multiple times in quick succession
	for i := 0; i < 10; i++ {
		autoSave.MarkDirty()
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for debounce to complete
	time.Sleep(2500 * time.Millisecond)

	// Should have saved only once after the debounce period
	// We can't easily verify save count, but we can verify the state was saved
	_, err := Load()
	if err != nil {
		t.Fatalf("Failed to load state after multiple marks: %v", err)
	}
}

func TestAutoSaveContextCancellation(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	state := NewState()
	autoSave := NewAutoSave(state)

	ctx, cancel := context.WithCancel(context.Background())
	if err := autoSave.Start(ctx); err != nil {
		t.Fatalf("Failed to start auto-save: %v", err)
	}

	// Cancel the context
	cancel()

	// Give it some time to stop
	time.Sleep(100 * time.Millisecond)

	// AutoSave should have stopped
	if autoSave.IsRunning() {
		t.Error("AutoSave should have stopped after context cancellation")
	}
}

func TestAutoSaveFinalSave(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	state := NewState()
	autoSave := NewAutoSave(state)

	ctx := context.Background()
	if err := autoSave.Start(ctx); err != nil {
		t.Fatalf("Failed to start auto-save: %v", err)
	}

	// Add some data
	source := types.Source{
		ID:         types.SourceTypeGit,
		Name:       "Test Git Repo",
		Type:       types.SourceTypeGit,
		Status:     "active",
		EventCount: 25,
		LastSync:   time.Now(),
		Enabled:    true,
	}
	if err := state.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Stop immediately without waiting for debounce
	// This should trigger a final save
	if err := autoSave.Stop(); err != nil {
		t.Fatalf("Failed to stop auto-save: %v", err)
	}

	// Load and verify the data was saved
	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if len(loadedState.Sources) != 1 {
		t.Errorf("Expected 1 source after final save, got %d", len(loadedState.Sources))
	}
}
