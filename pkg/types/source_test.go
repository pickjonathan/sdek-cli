package types

import (
	"testing"
	"time"
)

func TestValidateSource(t *testing.T) {
	tests := []struct {
		name    string
		source  *Source
		wantErr bool
	}{
		{
			name:    "nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name: "valid git source",
			source: &Source{
				ID:         SourceTypeGit,
				Name:       "Git Commits",
				Type:       SourceTypeGit,
				Status:     "simulated",
				LastSync:   time.Now(),
				EventCount: 25,
				Enabled:    true,
			},
			wantErr: false,
		},
		{
			name: "invalid source ID",
			source: &Source{
				ID:         "invalid",
				Name:       "Invalid Source",
				Type:       "invalid",
				Status:     "simulated",
				LastSync:   time.Now(),
				EventCount: 25,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "event count too low",
			source: &Source{
				ID:         SourceTypeGit,
				Name:       "Git Commits",
				Type:       SourceTypeGit,
				Status:     "simulated",
				LastSync:   time.Now(),
				EventCount: 5,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "event count too high",
			source: &Source{
				ID:         SourceTypeGit,
				Name:       "Git Commits",
				Type:       SourceTypeGit,
				Status:     "simulated",
				LastSync:   time.Now(),
				EventCount: 60,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "last sync too old",
			source: &Source{
				ID:         SourceTypeGit,
				Name:       "Git Commits",
				Type:       SourceTypeGit,
				Status:     "simulated",
				LastSync:   time.Now().AddDate(0, 0, -100),
				EventCount: 25,
				Enabled:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSource(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSource(t *testing.T) {
	source := NewSource(SourceTypeGit, "Git Commits", SourceTypeGit)

	if source.ID != SourceTypeGit {
		t.Errorf("expected ID %s, got %s", SourceTypeGit, source.ID)
	}
	if source.Name != "Git Commits" {
		t.Errorf("expected Name 'Git Commits', got %s", source.Name)
	}
	if source.Status != "simulated" {
		t.Errorf("expected Status 'simulated', got %s", source.Status)
	}
	if !source.Enabled {
		t.Error("expected Enabled to be true")
	}
	if source.EventCount != 0 {
		t.Errorf("expected EventCount 0, got %d", source.EventCount)
	}
}

func TestValidSourceTypes(t *testing.T) {
	expected := []string{SourceTypeGit, SourceTypeJira, SourceTypeSlack, SourceTypeCICD, SourceTypeDocs}

	if len(ValidSourceTypes) != len(expected) {
		t.Errorf("expected %d source types, got %d", len(expected), len(ValidSourceTypes))
	}

	for i, st := range expected {
		if ValidSourceTypes[i] != st {
			t.Errorf("expected source type %s at index %d, got %s", st, i, ValidSourceTypes[i])
		}
	}
}
