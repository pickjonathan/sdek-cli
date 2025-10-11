package types

import "testing"

func TestValidateFramework(t *testing.T) {
	tests := []struct {
		name      string
		framework *Framework
		wantErr   bool
	}{
		{
			name:      "nil framework",
			framework: nil,
			wantErr:   true,
		},
		{
			name: "valid SOC2 framework",
			framework: &Framework{
				ID:                   FrameworkSOC2,
				Name:                 "SOC2 Type II",
				Version:              "2017",
				ControlCount:         45,
				CompliancePercentage: 75.0,
				Category:             "security",
			},
			wantErr: false,
		},
		{
			name: "invalid framework ID",
			framework: &Framework{
				ID:           "invalid",
				ControlCount: 10,
			},
			wantErr: true,
		},
		{
			name: "negative control count",
			framework: &Framework{
				ID:           FrameworkSOC2,
				ControlCount: -1,
			},
			wantErr: true,
		},
		{
			name: "compliance percentage out of range",
			framework: &Framework{
				ID:                   FrameworkSOC2,
				ControlCount:         10,
				CompliancePercentage: 150.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFramework(tt.framework)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFramework() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateCompliance(t *testing.T) {
	framework := NewFramework(FrameworkSOC2, "SOC2 Type II", "2017", 45)

	// Test with 34 green controls out of 45
	framework.CalculateCompliance(34, 45)

	expected := 75.55555555555556
	if framework.CompliancePercentage < expected-0.1 || framework.CompliancePercentage > expected+0.1 {
		t.Errorf("expected compliance ~75.56%%, got %.2f%%", framework.CompliancePercentage)
	}

	// Test with zero total controls
	framework.CalculateCompliance(0, 0)
	if framework.CompliancePercentage != 0 {
		t.Errorf("expected 0%% compliance for zero controls, got %.2f%%", framework.CompliancePercentage)
	}
}

func TestNewFramework(t *testing.T) {
	framework := NewFramework(FrameworkSOC2, "SOC2 Type II", "2017", 45)

	if framework.ID != FrameworkSOC2 {
		t.Errorf("expected ID %s, got %s", FrameworkSOC2, framework.ID)
	}

	if framework.CompliancePercentage != 0 {
		t.Errorf("expected initial compliance 0, got %.2f", framework.CompliancePercentage)
	}
}
