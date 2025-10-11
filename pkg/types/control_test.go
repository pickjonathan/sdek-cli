package types

import "testing"

func TestValidateControl(t *testing.T) {
	tests := []struct {
		name    string
		control *Control
		wantErr bool
	}{
		{
			name:    "nil control",
			control: nil,
			wantErr: true,
		},
		{
			name: "valid control",
			control: &Control{
				ID:              "CC6.1",
				FrameworkID:     FrameworkSOC2,
				Title:           "Access Controls",
				RiskStatus:      RiskStatusGreen,
				ConfidenceLevel: 85.0,
			},
			wantErr: false,
		},
		{
			name: "empty control ID",
			control: &Control{
				ID:          "",
				FrameworkID: FrameworkSOC2,
			},
			wantErr: true,
		},
		{
			name: "empty framework ID",
			control: &Control{
				ID:          "CC6.1",
				FrameworkID: "",
			},
			wantErr: true,
		},
		{
			name: "invalid risk status",
			control: &Control{
				ID:          "CC6.1",
				FrameworkID: FrameworkSOC2,
				RiskStatus:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "confidence level too low",
			control: &Control{
				ID:              "CC6.1",
				FrameworkID:     FrameworkSOC2,
				ConfidenceLevel: -1.0,
			},
			wantErr: true,
		},
		{
			name: "confidence level too high",
			control: &Control{
				ID:              "CC6.1",
				FrameworkID:     FrameworkSOC2,
				ConfidenceLevel: 101.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateControl(tt.control)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateControl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name               string
		criticalSeverity   int
		highSeverity       int
		mediumSeverity     int
		lowSeverity        int
		evidenceCount      int
		requiredEvidence   int
		expectedRiskStatus string
	}{
		{
			name:               "no findings with sufficient evidence",
			criticalSeverity:   0,
			highSeverity:       0,
			mediumSeverity:     0,
			lowSeverity:        0,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusGreen,
		},
		{
			name:               "1 critical finding",
			criticalSeverity:   1,
			highSeverity:       0,
			mediumSeverity:     0,
			lowSeverity:        0,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusRed,
		},
		{
			name:               "3 high findings = 1 critical equivalent",
			criticalSeverity:   0,
			highSeverity:       3,
			mediumSeverity:     0,
			lowSeverity:        0,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusRed,
		},
		{
			name:               "6 medium findings = 1 critical equivalent",
			criticalSeverity:   0,
			highSeverity:       0,
			mediumSeverity:     6,
			lowSeverity:        0,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusRed,
		},
		{
			name:               "12 low findings = 1 critical equivalent",
			criticalSeverity:   0,
			highSeverity:       0,
			mediumSeverity:     0,
			lowSeverity:        12,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusRed,
		},
		{
			name:               "mixed findings below threshold",
			criticalSeverity:   0,
			highSeverity:       1,
			mediumSeverity:     2,
			lowSeverity:        3,
			evidenceCount:      3,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusYellow,
		},
		{
			name:               "insufficient evidence",
			criticalSeverity:   0,
			highSeverity:       0,
			mediumSeverity:     0,
			lowSeverity:        0,
			evidenceCount:      0,
			requiredEvidence:   1,
			expectedRiskStatus: RiskStatusYellow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := &Control{
				CriticalSeverity: tt.criticalSeverity,
				HighSeverity:     tt.highSeverity,
				MediumSeverity:   tt.mediumSeverity,
				LowSeverity:      tt.lowSeverity,
				EvidenceCount:    tt.evidenceCount,
				RequiredEvidence: tt.requiredEvidence,
			}

			control.CalculateRiskScore()

			if control.RiskStatus != tt.expectedRiskStatus {
				t.Errorf("expected risk status %s, got %s", tt.expectedRiskStatus, control.RiskStatus)
			}
		})
	}
}

func TestNewControl(t *testing.T) {
	keywords := []string{"access", "authentication", "authorization"}
	control := NewControl("CC6.1", FrameworkSOC2, "Access Controls", keywords)

	if control.ID != "CC6.1" {
		t.Errorf("expected ID 'CC6.1', got %s", control.ID)
	}

	if control.FrameworkID != FrameworkSOC2 {
		t.Errorf("expected FrameworkID %s, got %s", FrameworkSOC2, control.FrameworkID)
	}

	if control.Title != "Access Controls" {
		t.Errorf("expected Title 'Access Controls', got %s", control.Title)
	}

	if len(control.Keywords) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(control.Keywords))
	}

	if control.RequiredEvidence != 1 {
		t.Errorf("expected RequiredEvidence 1, got %d", control.RequiredEvidence)
	}
}
