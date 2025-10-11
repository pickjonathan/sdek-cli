package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewState(t *testing.T) {
	state := NewState()

	if state == nil {
		t.Fatal("NewState returned nil")
	}

	if state.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", state.Version)
	}

	if len(state.Sources) != 0 {
		t.Errorf("Expected empty sources, got %d", len(state.Sources))
	}

	if len(state.Users) == 0 {
		t.Error("Expected default users to be populated")
	}

	if state.Config == nil {
		t.Error("Expected config to be initialized")
	}
}

func TestStateSaveLoad(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()

	// Override the state file path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a new state with some data
	state := NewState()
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

	// Save the state
	if err := state.Save(); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify the file was created
	statePath, err := GetStateFilePath()
	if err != nil {
		t.Fatalf("Failed to get state file path: %v", err)
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	// Load the state
	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify the loaded state matches
	if len(loadedState.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(loadedState.Sources))
	}

	if loadedState.Sources[0].Type != types.SourceTypeGit {
		t.Errorf("Expected source type %s, got %s", types.SourceTypeGit, loadedState.Sources[0].Type)
	}
}

func TestStateLoadNonExistent(t *testing.T) {
	// Create a temporary directory for test state
	tmpDir := t.TempDir()

	// Override the state file path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load state from non-existent file should return empty state
	state, err := Load()
	if err != nil {
		t.Fatalf("Load should not error on non-existent file: %v", err)
	}

	if state == nil {
		t.Fatal("Load returned nil state")
	}

	if len(state.Sources) != 0 {
		t.Errorf("Expected empty sources, got %d", len(state.Sources))
	}
}

func TestStateClear(t *testing.T) {
	state := NewState()

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

	event := types.NewEvent("git-1", types.EventTypeCommit, "Test Event", "test-author")
	if err := state.AddEvent(*event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Clear the state
	state.Clear()

	// Verify everything is empty
	if len(state.Sources) != 0 {
		t.Errorf("Expected empty sources after clear, got %d", len(state.Sources))
	}

	if len(state.Events) != 0 {
		t.Errorf("Expected empty events after clear, got %d", len(state.Events))
	}

	if len(state.Users) == 0 {
		t.Error("Expected default users to be restored after clear")
	}
}

func TestStateAddSource(t *testing.T) {
	state := NewState()

	validSource := types.Source{
		ID:         types.SourceTypeGit,
		Name:       "Test Git Repo",
		Type:       types.SourceTypeGit,
		Status:     "active",
		EventCount: 25,
		LastSync:   time.Now(),
		Enabled:    true,
	}

	if err := state.AddSource(validSource); err != nil {
		t.Errorf("Failed to add valid source: %v", err)
	}

	if len(state.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(state.Sources))
	}

	// Test adding invalid source
	invalidSource := types.Source{
		ID:         "invalid",
		Name:       "Invalid Source",
		EventCount: 5, // Too few events
		LastSync:   time.Now(),
	}

	if err := state.AddSource(invalidSource); err == nil {
		t.Error("Expected error when adding invalid source")
	}
}

func TestStateAddEvent(t *testing.T) {
	state := NewState()

	validEvent := types.NewEvent("git-1", types.EventTypeCommit, "Test Event", "test-author")

	if err := state.AddEvent(*validEvent); err != nil {
		t.Errorf("Failed to add valid event: %v", err)
	}

	if len(state.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(state.Events))
	}

	// Test adding invalid event
	invalidEvent := types.Event{
		ID:        "invalid-id", // Not a valid UUID
		SourceID:  "git-1",
		EventType: types.EventTypeCommit,
		Timestamp: time.Now(),
		Title:     "",
		Content:   "Test content",
		Author:    "test-author",
	}

	if err := state.AddEvent(invalidEvent); err == nil {
		t.Error("Expected error when adding invalid event")
	}
}

func TestStateAddFramework(t *testing.T) {
	state := NewState()

	validFramework := types.Framework{
		ID:           types.FrameworkSOC2,
		Name:         "SOC 2",
		Version:      "2017",
		Description:  "Service Organization Control 2",
		Category:     "security",
		ControlCount: 2,
	}

	if err := state.AddFramework(validFramework); err != nil {
		t.Errorf("Failed to add valid framework: %v", err)
	}

	if len(state.Frameworks) != 1 {
		t.Errorf("Expected 1 framework, got %d", len(state.Frameworks))
	}

	// Test adding invalid framework
	invalidFramework := types.Framework{
		ID:           "invalid",
		Name:         "Invalid Framework",
		Version:      "1.0",
		Description:  "Invalid",
		Category:     "test",
		ControlCount: 0,
	}

	if err := state.AddFramework(invalidFramework); err == nil {
		t.Error("Expected error when adding invalid framework")
	}
}

func TestGetStateFilePath(t *testing.T) {
	path, err := GetStateFilePath()
	if err != nil {
		t.Fatalf("Failed to get state file path: %v", err)
	}

	if path == "" {
		t.Error("State file path is empty")
	}

	// Verify it ends with .sdek/state.json
	if filepath.Base(path) != "state.json" {
		t.Errorf("Expected state.json, got %s", filepath.Base(path))
	}

	if filepath.Base(filepath.Dir(path)) != ".sdek" {
		t.Errorf("Expected .sdek directory, got %s", filepath.Base(filepath.Dir(path)))
	}
}
