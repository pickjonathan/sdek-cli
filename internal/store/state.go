package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// State represents the complete application state that gets persisted
type State struct {
	Sources    []types.Source    `json:"sources"`
	Events     []types.Event     `json:"events"`
	Frameworks []types.Framework `json:"frameworks"`
	Controls   []types.Control   `json:"controls"`
	Evidence   []types.Evidence  `json:"evidence"`
	Findings   []types.Finding   `json:"findings"`
	Users      []*types.User     `json:"users"`
	Config     *types.Config     `json:"config"`
	Version    string            `json:"version"`
}

// NewState creates a new empty state with default configuration
func NewState() *State {
	return &State{
		Sources:    make([]types.Source, 0),
		Events:     make([]types.Event, 0),
		Frameworks: make([]types.Framework, 0),
		Controls:   make([]types.Control, 0),
		Evidence:   make([]types.Evidence, 0),
		Findings:   make([]types.Finding, 0),
		Users:      types.AllUsers(),
		Config:     types.DefaultConfig(),
		Version:    "1.0",
	}
}

// GetStateFilePath returns the path to the state file
func GetStateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	sdekDir := filepath.Join(homeDir, ".sdek")
	if err := os.MkdirAll(sdekDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .sdek directory: %w", err)
	}

	return filepath.Join(sdekDir, "state.json"), nil
}

// Load reads the state from the JSON file
func Load() (*State, error) {
	filePath, err := GetStateFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get state file path: %w", err)
	}

	// If file doesn't exist, return a new empty state
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return NewState(), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// Save writes the state to the JSON file
func (s *State) Save() error {
	filePath, err := GetStateFilePath()
	if err != nil {
		return fmt.Errorf("failed to get state file path: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to a temporary file first, then rename to ensure atomicity
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file if rename fails
		return fmt.Errorf("failed to rename temporary state file: %w", err)
	}

	return nil
}

// Clear removes all data from the state
func (s *State) Clear() {
	s.Sources = make([]types.Source, 0)
	s.Events = make([]types.Event, 0)
	s.Frameworks = make([]types.Framework, 0)
	s.Controls = make([]types.Control, 0)
	s.Evidence = make([]types.Evidence, 0)
	s.Findings = make([]types.Finding, 0)
	s.Users = types.AllUsers()
	s.Config = types.DefaultConfig()
}

// AddSource adds a source to the state
func (s *State) AddSource(source types.Source) error {
	if err := types.ValidateSource(&source); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}
	s.Sources = append(s.Sources, source)
	return nil
}

// AddEvent adds an event to the state
func (s *State) AddEvent(event types.Event) error {
	if err := types.ValidateEvent(&event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	s.Events = append(s.Events, event)
	return nil
}

// AddFramework adds a framework to the state
func (s *State) AddFramework(framework types.Framework) error {
	if err := types.ValidateFramework(&framework); err != nil {
		return fmt.Errorf("invalid framework: %w", err)
	}
	s.Frameworks = append(s.Frameworks, framework)
	return nil
}

// AddControl adds a control to the state
func (s *State) AddControl(control types.Control) error {
	s.Controls = append(s.Controls, control)
	return nil
}

// AddEvidence adds evidence to the state
func (s *State) AddEvidence(evidence types.Evidence) error {
	s.Evidence = append(s.Evidence, evidence)
	return nil
}

// AddFinding adds a finding to the state
func (s *State) AddFinding(finding types.Finding) error {
	s.Findings = append(s.Findings, finding)
	return nil
}
