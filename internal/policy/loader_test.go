package policy

import (
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	// Should have loaded default excerpts
	if len(loader.excerpts) == 0 {
		t.Error("Expected default excerpts to be loaded")
	}
}

func TestGetExcerpt_SOC2(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name      string
		controlID string
		wantError bool
	}{
		{
			name:      "valid SOC2 control",
			controlID: "SOC2-CC1.1",
			wantError: false,
		},
		{
			name:      "valid SOC2 control 2",
			controlID: "SOC2-CC6.1",
			wantError: false,
		},
		{
			name:      "invalid control",
			controlID: "SOC2-INVALID",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excerpt, err := loader.GetExcerpt(tt.controlID)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error for invalid control")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if excerpt == "" {
					t.Error("Expected non-empty excerpt")
				}
			}
		})
	}
}

func TestGetExcerpt_ISO27001(t *testing.T) {
	loader := NewLoader()

	excerpt, err := loader.GetExcerpt("ISO27001-A.5.1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if excerpt == "" {
		t.Error("Expected non-empty excerpt for ISO27001-A.5.1")
	}
}

func TestGetExcerpt_PCIDSS(t *testing.T) {
	loader := NewLoader()

	excerpt, err := loader.GetExcerpt("PCI-DSS-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if excerpt == "" {
		t.Error("Expected non-empty excerpt for PCI-DSS-1")
	}
}

func TestLoadExcerpts(t *testing.T) {
	loader := NewLoader()

	customExcerpts := map[string]string{
		"CUSTOM-1": "Custom control excerpt",
		"CUSTOM-2": "Another custom control",
	}

	loader.LoadExcerpts(customExcerpts)

	// Should be able to retrieve custom excerpts
	excerpt, err := loader.GetExcerpt("CUSTOM-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if excerpt != "Custom control excerpt" {
		t.Errorf("Expected 'Custom control excerpt', got %q", excerpt)
	}
}

func TestHasExcerpt(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name      string
		controlID string
		want      bool
	}{
		{
			name:      "existing SOC2 control",
			controlID: "SOC2-CC1.1",
			want:      true,
		},
		{
			name:      "existing ISO control",
			controlID: "ISO27001-A.5.1",
			want:      true,
		},
		{
			name:      "existing PCI control",
			controlID: "PCI-DSS-1",
			want:      true,
		},
		{
			name:      "non-existing control",
			controlID: "INVALID-1",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loader.HasExcerpt(tt.controlID)
			if got != tt.want {
				t.Errorf("HasExcerpt(%q) = %v, want %v", tt.controlID, got, tt.want)
			}
		})
	}
}

func TestGetAllControlIDs(t *testing.T) {
	loader := NewLoader()

	controlIDs := loader.GetAllControlIDs()

	if len(controlIDs) == 0 {
		t.Error("Expected non-empty control IDs")
	}

	// Should contain controls from all frameworks
	hasSOC2 := false
	hasISO := false
	hasPCI := false

	for _, id := range controlIDs {
		if len(id) > 4 {
			prefix := id[:4]
			if prefix == "SOC2" {
				hasSOC2 = true
			} else if prefix == "ISO2" {
				hasISO = true
			} else if prefix == "PCI-" {
				hasPCI = true
			}
		}
	}

	if !hasSOC2 {
		t.Error("Expected SOC2 controls in control IDs")
	}
	if !hasISO {
		t.Error("Expected ISO27001 controls in control IDs")
	}
	if !hasPCI {
		t.Error("Expected PCI-DSS controls in control IDs")
	}
}

func TestGetExcerpt_EmptyLoader(t *testing.T) {
	loader := &Loader{
		excerpts: make(map[string]string),
	}

	_, err := loader.GetExcerpt("ANY-CONTROL")
	if err == nil {
		t.Error("Expected error for empty loader")
	}
}
