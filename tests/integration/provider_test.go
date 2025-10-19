package integration

import (
	"context"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/ai"
)

// TestProviderAnalyzeWithContext tests the provider interface used in autonomous mode
func TestProviderAnalyzeWithContext(t *testing.T) {
	// Create mock provider
	mockProvider := ai.NewMockProvider()

	ctx := context.Background()
	prompt := "Analyze the following evidence for SOC2 CC6.1 compliance"

	// Test AnalyzeWithContext
	response, err := mockProvider.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		t.Fatalf("AnalyzeWithContext failed: %v", err)
	}

	// Mock provider returns a JSON response with default confidence 0.85
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Verify response contains expected fields
	if !contains(response, "summary") {
		t.Error("Expected response to contain 'summary' field")
	}
	if !contains(response, "confidence_score") {
		t.Error("Expected response to contain 'confidence_score' field")
	}

	// Verify call tracking
	if mockProvider.GetCallCount() != 1 {
		t.Errorf("Expected call count 1, got %d", mockProvider.GetCallCount())
	}

	if mockProvider.GetLastPrompt() != prompt {
		t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestProviderMultipleCallsTracking tests call count increments correctly
func TestProviderMultipleCallsTracking(t *testing.T) {
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse("test response")

	ctx := context.Background()

	// Make multiple calls
	for i := 1; i <= 5; i++ {
		prompt := "Test prompt"
		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)
		if err != nil {
			t.Fatalf("Call %d failed: %v", i, err)
		}

		if mockProvider.GetCallCount() != i {
			t.Errorf("After call %d, expected count %d, got %d", i, i, mockProvider.GetCallCount())
		}
	}
}

// TestProviderErrorHandling tests error propagation
func TestProviderErrorHandling(t *testing.T) {
	mockProvider := ai.NewMockProvider()
	expectedErr := ai.ErrProviderAuth
	mockProvider.SetError(expectedErr)

	ctx := context.Background()

	_, err := mockProvider.AnalyzeWithContext(ctx, "test prompt")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestProviderEmptyPrompt tests empty prompt validation
func TestProviderEmptyPrompt(t *testing.T) {
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse("test") // Set a response so empty prompt can be tested
	ctx := context.Background()

	// Mock provider doesn't validate empty prompts, it just tracks them
	// This test documents current behavior - can be enhanced if validation is added
	response, err := mockProvider.AnalyzeWithContext(ctx, "")
	if err != nil {
		t.Logf("Empty prompt returned error (good): %v", err)
	} else {
		t.Logf("Empty prompt was accepted and returned: %s", response)
		// This is current behavior - mock doesn't validate
	}
}

// TestProviderContextCancellation tests context cancellation handling
func TestProviderContextCancellation(t *testing.T) {
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse("response")

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mockProvider.AnalyzeWithContext(ctx, "test")
	// Should handle cancellation gracefully
	t.Logf("Context cancellation result: %v", err)
}

// TestProviderLastPromptTracking tests that last prompt is tracked correctly
func TestProviderLastPromptTracking(t *testing.T) {
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse("test")

	ctx := context.Background()

	prompts := []string{
		"First prompt",
		"Second prompt",
		"Third prompt",
	}

	for _, prompt := range prompts {
		_, _ = mockProvider.AnalyzeWithContext(ctx, prompt)

		if mockProvider.GetLastPrompt() != prompt {
			t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())
		}
	}
}
