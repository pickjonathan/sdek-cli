package integrationpackage integrationpackage integrationpackage integration



import (

	"context"

	"testing"import (



	"github.com/pickjonathan/sdek-cli/internal/ai"	"context"

)

	"testing"import (import (

// TestProvider_AnalyzeWithContext tests the provider interface used in autonomous mode

func TestProvider_AnalyzeWithContext(t *testing.T) {

	// Create mock provider

	mockProvider := ai.NewMockProvider()	"github.com/pickjonathan/sdek-cli/internal/ai"	"context"	"context"

	expectedResponse := "Analysis complete: Access controls properly implemented"

	mockProvider.SetResponse(expectedResponse))



	ctx := context.Background()	"testing"	"testing"

	prompt := "Analyze the following evidence for SOC2 CC6.1 compliance"

// TestProvider_AnalyzeWithContext tests the provider interface used in autonomous mode

	// Test AnalyzeWithContext

	response, err := mockProvider.AnalyzeWithContext(ctx, prompt)func TestProvider_AnalyzeWithContext(t *testing.T) {	"time"

	if err != nil {

		t.Fatalf("AnalyzeWithContext failed: %v", err)	// Create mock provider

	}

	mockProvider := ai.NewMockProvider()	"github.com/pickjonathan/sdek-cli/internal/ai"

	if response != expectedResponse {

		t.Errorf("Expected response %q, got %q", expectedResponse, response)	expectedResponse := "Analysis complete: Access controls properly implemented"

	}

	mockProvider.SetResponse(expectedResponse))	"github.com/pickjonathan/sdek-cli/internal/ai"

	// Verify call tracking

	if mockProvider.GetCallCount() != 1 {

		t.Errorf("Expected call count 1, got %d", mockProvider.GetCallCount())

	}	ctx := context.Background()	"github.com/pickjonathan/sdek-cli/pkg/types"



	if mockProvider.GetLastPrompt() != prompt {	prompt := "Analyze the following evidence for SOC2 CC6.1 compliance"

		t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())

	}// TestProvider_AnalyzeWithContext tests the provider interface used in autonomous mode)

}

	// Test AnalyzeWithContext

// TestProvider_MultipleCallsTracking tests call count increments correctly

func TestProvider_MultipleCallsTracking(t *testing.T) {	response, err := mockProvider.AnalyzeWithContext(ctx, prompt)func TestProvider_AnalyzeWithContext(t *testing.T) {

	mockProvider := ai.NewMockProvider()

	mockProvider.SetResponse("test response")	if err != nil {



	ctx := context.Background()		t.Fatalf("AnalyzeWithContext failed: %v", err)	// Create mock provider// TestProvider_AnalyzeWithContext tests the provider interface used in autonomous mode



	// Make multiple calls	}

	for i := 1; i <= 5; i++ {

		prompt := "Test prompt"	mockProvider := ai.NewMockProvider()func TestProvider_AnalyzeWithContext(t *testing.T) {

		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)

		if err != nil {	if response != expectedResponse {

			t.Fatalf("Call %d failed: %v", i, err)

		}		t.Errorf("Expected response %q, got %q", expectedResponse, response)	expectedResponse := "Analysis complete: Access controls properly implemented"	// Create mock provider



		if mockProvider.GetCallCount() != i {	}

			t.Errorf("After call %d, expected count %d, got %d", i, i, mockProvider.GetCallCount())

		}	mockProvider.SetResponse(expectedResponse)	mockProvider := ai.NewMockProvider()

	}

}	// Verify call tracking



// TestProvider_ErrorHandling tests error propagation	if mockProvider.GetCallCount() != 1 {	expectedResponse := "Analysis complete: Access controls properly implemented"

func TestProvider_ErrorHandling(t *testing.T) {

	mockProvider := ai.NewMockProvider()		t.Errorf("Expected call count 1, got %d", mockProvider.GetCallCount())

	expectedErr := ai.ErrProviderAuth

	mockProvider.SetError(expectedErr)	}	ctx := context.Background()	mockProvider.SetResponse(expectedResponse)



	ctx := context.Background()



	_, err := mockProvider.AnalyzeWithContext(ctx, "test prompt")	if mockProvider.GetLastPrompt() != prompt {	prompt := "Analyze the following evidence for SOC2 CC6.1 compliance"

	if err == nil {

		t.Fatal("Expected error, got nil")		t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())

	}

	}	ctx := context.Background()

	if err != expectedErr {

		t.Errorf("Expected error %v, got %v", expectedErr, err)}

	}

}	// Test AnalyzeWithContext	prompt := "Analyze the following evidence for SOC2 CC6.1 compliance"



// TestProvider_EmptyPrompt tests empty prompt validation// TestProvider_MultipleCallsTracking tests call count increments correctly

func TestProvider_EmptyPrompt(t *testing.T) {

	mockProvider := ai.NewMockProvider()func TestProvider_MultipleCallsTracking(t *testing.T) {	response, err := mockProvider.AnalyzeWithContext(ctx, prompt)

	ctx := context.Background()

	mockProvider := ai.NewMockProvider()

	_, err := mockProvider.AnalyzeWithContext(ctx, "")

	if err == nil {	mockProvider.SetResponse("test response")	if err != nil {	// Test AnalyzeWithContext

		t.Error("Expected error for empty prompt, got nil")

	}

}

	ctx := context.Background()		t.Fatalf("AnalyzeWithContext failed: %v", err)	response, err := mockProvider.AnalyzeWithContext(ctx, prompt)

// TestProvider_ContextCancellation tests context cancellation handling

func TestProvider_ContextCancellation(t *testing.T) {

	mockProvider := ai.NewMockProvider()

	mockProvider.SetResponse("response")	// Make multiple calls	}	if err != nil {



	// Create cancelled context	for i := 1; i <= 5; i++ {

	ctx, cancel := context.WithCancel(context.Background())

	cancel()		prompt := "Test prompt"		t.Fatalf("AnalyzeWithContext failed: %v", err)



	_, err := mockProvider.AnalyzeWithContext(ctx, "test")		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)

	// Should handle cancellation gracefully

	t.Logf("Context cancellation result: %v", err)		if err != nil {	if response != expectedResponse {	}

}

			t.Fatalf("Call %d failed: %v", i, err)

// TestProvider_LastPromptTracking tests that last prompt is tracked correctly

func TestProvider_LastPromptTracking(t *testing.T) {		}		t.Errorf("Expected response %q, got %q", expectedResponse, response)

	mockProvider := ai.NewMockProvider()

	mockProvider.SetResponse("test")



	ctx := context.Background()		if mockProvider.GetCallCount() != i {	}	if response != expectedResponse {



	prompts := []string{			t.Errorf("After call %d, expected count %d, got %d", i, i, mockProvider.GetCallCount())

		"First prompt",

		"Second prompt",		}		t.Errorf("Expected response %q, got %q", expectedResponse, response)

		"Third prompt",

	}	}



	for _, prompt := range prompts {}	// Verify call tracking	}

		_, _ = mockProvider.AnalyzeWithContext(ctx, prompt)

		

		if mockProvider.GetLastPrompt() != prompt {

			t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())// TestProvider_ErrorHandling tests error propagation	if mockProvider.GetCallCount() != 1 {

		}

	}func TestProvider_ErrorHandling(t *testing.T) {

}

	mockProvider := ai.NewMockProvider()		t.Errorf("Expected call count 1, got %d", mockProvider.GetCallCount())	// Verify call tracking

	expectedErr := ai.ErrProviderAuth

	mockProvider.SetError(expectedErr)	}	if mockProvider.GetCallCount() != 1 {



	ctx := context.Background()		t.Errorf("Expected call count 1, got %d", mockProvider.GetCallCount())



	_, err := mockProvider.AnalyzeWithContext(ctx, "test prompt")	if mockProvider.GetLastPrompt() != prompt {	}

	if err == nil {

		t.Fatal("Expected error, got nil")		t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())

	}

	}	if mockProvider.GetLastPrompt() != prompt {

	if err != expectedErr {

		t.Errorf("Expected error %v, got %v", expectedErr, err)}		t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())

	}

}	}



// TestProvider_EmptyPrompt tests empty prompt validation// TestProvider_MultipleCallsTracking tests call count increments correctly}

func TestProvider_EmptyPrompt(t *testing.T) {

	mockProvider := ai.NewMockProvider()func TestProvider_MultipleCallsTracking(t *testing.T) {

	ctx := context.Background()

	mockProvider := ai.NewMockProvider()// TestProvider_MultipleCallsTracking tests call count increments correctly

	_, err := mockProvider.AnalyzeWithContext(ctx, "")

	if err == nil {	mockProvider.SetResponse("test response")func TestProvider_MultipleCallsTracking(t *testing.T) {

		t.Error("Expected error for empty prompt, got nil")

	}	mockProvider := ai.NewMockProvider()

}

	ctx := context.Background()	mockProvider.SetResponse("test response")

// TestProvider_ContextCancellation tests context cancellation handling

func TestProvider_ContextCancellation(t *testing.T) {

	mockProvider := ai.NewMockProvider()

	mockProvider.SetResponse("response")	// Make multiple calls	ctx := context.Background()



	// Create cancelled context	for i := 1; i <= 5; i++ {

	ctx, cancel := context.WithCancel(context.Background())

	cancel()		prompt := "Test prompt"	// Make multiple calls



	_, err := mockProvider.AnalyzeWithContext(ctx, "test")		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)	for i := 1; i <= 5; i++ {

	// Should handle cancellation gracefully

	t.Logf("Context cancellation result: %v", err)		if err != nil {		prompt := "Test prompt"

}

			t.Fatalf("Call %d failed: %v", i, err)		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)

// TestProvider_LastPromptTracking tests that last prompt is tracked correctly

func TestProvider_LastPromptTracking(t *testing.T) {		}		if err != nil {

	mockProvider := ai.NewMockProvider()

	mockProvider.SetResponse("test")			t.Fatalf("Call %d failed: %v", i, err)



	ctx := context.Background()		if mockProvider.GetCallCount() != i {		}



	prompts := []string{			t.Errorf("After call %d, expected count %d, got %d", i, i, mockProvider.GetCallCount())

		"First prompt",

		"Second prompt",		}		if mockProvider.GetCallCount() != i {

		"Third prompt",

	}	}			t.Errorf("After call %d, expected count %d, got %d", i, i, mockProvider.GetCallCount())



	for _, prompt := range prompts {}		}

		_, _ = mockProvider.AnalyzeWithContext(ctx, prompt)

			}

		if mockProvider.GetLastPrompt() != prompt {

			t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())// TestProvider_ErrorHandling tests error propagation}

		}

	}func TestProvider_ErrorHandling(t *testing.T) {

}

	mockProvider := ai.NewMockProvider()// TestProvider_ErrorHandling tests error propagation

	expectedErr := ai.ErrProviderAuthfunc TestProvider_ErrorHandling(t *testing.T) {

	mockProvider.SetError(expectedErr)	mockProvider := ai.NewMockProvider()

	expectedErr := ai.ErrProviderAuth

	ctx := context.Background()	mockProvider.SetError(expectedErr)



	_, err := mockProvider.AnalyzeWithContext(ctx, "test prompt")	ctx := context.Background()

	if err == nil {

		t.Fatal("Expected error, got nil")	_, err := mockProvider.AnalyzeWithContext(ctx, "test prompt")

	}	if err == nil {

		t.Fatal("Expected error, got nil")

	if err != expectedErr {	}

		t.Errorf("Expected error %v, got %v", expectedErr, err)

	}	if err != expectedErr {

}		t.Errorf("Expected error %v, got %v", expectedErr, err)

	}

// TestProvider_EmptyPrompt tests empty prompt validation}

func TestProvider_EmptyPrompt(t *testing.T) {

	mockProvider := ai.NewMockProvider()// TestProvider_EmptyPrompt tests empty prompt validation

	ctx := context.Background()func TestProvider_EmptyPrompt(t *testing.T) {

	mockProvider := ai.NewMockProvider()

	_, err := mockProvider.AnalyzeWithContext(ctx, "")	ctx := context.Background()

	if err == nil {

		t.Error("Expected error for empty prompt, got nil")	_, err := mockProvider.AnalyzeWithContext(ctx, "")

	}	if err == nil {

}		t.Error("Expected error for empty prompt, got nil")

	}

// TestProvider_ContextCancellation tests context cancellation handling}

func TestProvider_ContextCancellation(t *testing.T) {

	mockProvider := ai.NewMockProvider()// TestProvider_ContextCancellation tests context cancellation handling

	mockProvider.SetResponse("response")func TestProvider_ContextCancellation(t *testing.T) {

	mockProvider := ai.NewMockProvider()

	// Create cancelled context	mockProvider.SetResponse("response")

	ctx, cancel := context.WithCancel(context.Background())

	cancel()	// Create cancelled context

	ctx, cancel := context.WithCancel(context.Background())

	_, err := mockProvider.AnalyzeWithContext(ctx, "test")	cancel()

	// Should handle cancellation gracefully

	t.Logf("Context cancellation result: %v", err)	_, err := mockProvider.AnalyzeWithContext(ctx, "test")

}	// Should handle cancellation gracefully

	t.Logf("Context cancellation result: %v", err)

// TestProvider_LastPromptTracking tests that last prompt is tracked correctly}

func TestProvider_LastPromptTracking(t *testing.T) {

	mockProvider := ai.NewMockProvider()	// Test 1: Analyze with context injection

	mockProvider.SetResponse("test")	t.Run("AnalyzeWithContextInjection", func(t *testing.T) {

		ctx := context.Background()

	ctx := context.Background()

		req := &ai.AnalysisRequest{

	prompts := []string{			Events: []types.EvidenceEvent{

		"First prompt",				mockConnector.GetEvents("github")[0],

		"Second prompt",				mockConnector.GetEvents("github")[1],

		"Third prompt",			},

	}			Controls: []types.ControlObjective{

				{

	for _, prompt := range prompts {					ID:          "CC6.1",

		_, _ = mockProvider.AnalyzeWithContext(ctx, prompt)					Name:        "Logical and Physical Access Controls",

							Description: "The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives.",

		if mockProvider.GetLastPrompt() != prompt {				},

			t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())			},

		}		}

	}

}		resp, err := engine.Analyze(ctx, req)

		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected response, got nil")
		}

		if len(resp.Findings) == 0 {
			t.Error("Expected findings in response")
		}

		// Verify mock provider was called
		if mockProvider.GetCallCount() == 0 {
			t.Error("Expected provider to be called")
		}
	})

	// Test 2: ProposePlan for autonomous collection
	t.Run("ProposePlanAutonomous", func(t *testing.T) {
		ctx := context.Background()

		// Set mock plan items
		mockProvider.SetPlanItems([]types.PlanItem{
			{
				Source:         "github",
				Query:          "commits related to authentication",
				SignalStrength: 0.9,
				Rationale:      "Authentication commits likely contain access control evidence",
			},
			{
				Source:         "github",
				Query:          "commits related to authorization",
				SignalStrength: 0.85,
				Rationale:      "Authorization commits support CC6.1 control objectives",
			},
		})

		plan, err := engine.ProposePlan(ctx, &ai.PlanRequest{
			Controls: []types.ControlObjective{
				{
					ID:          "CC6.1",
					Name:        "Logical and Physical Access Controls",
					Description: "Implement logical access security controls",
				},
			},
			Context: map[string]interface{}{
				"available_sources": []string{"github", "jira"},
			},
		})

		if err != nil {
			t.Fatalf("ProposePlan failed: %v", err)
		}

		if plan == nil {
			t.Fatal("Expected plan, got nil")
		}

		if len(plan.Items) == 0 {
			t.Error("Expected plan items")
		}

		// Verify plan items have required fields
		for i, item := range plan.Items {
			if item.Source == "" {
				t.Errorf("Plan item %d missing source", i)
			}
			if item.Query == "" {
				t.Errorf("Plan item %d missing query", i)
			}
			if item.SignalStrength <= 0 {
				t.Errorf("Plan item %d has invalid signal strength: %f", i, item.SignalStrength)
			}
		}
	})

	// Test 3: ExecutePlan autonomous collection
	t.Run("ExecutePlanAutonomous", func(t *testing.T) {
		ctx := context.Background()

		plan := &types.Plan{
			Items: []types.PlanItem{
				{
					Source:         "github",
					Query:          "commits related to authentication",
					SignalStrength: 0.9,
				},
			},
		}

		results, err := engine.ExecutePlan(ctx, plan)
		if err != nil {
			t.Fatalf("ExecutePlan failed: %v", err)
		}

		if results == nil {
			t.Fatal("Expected results, got nil")
		}

		if len(results.Events) == 0 {
			t.Error("Expected collected events")
		}

		// Verify events have correct source
		for _, event := range results.Events {
			if event.Source != "GitHub" {
				t.Errorf("Expected source GitHub, got %s", event.Source)
			}
		}
	})

	// Test 4: Call count tracking
	t.Run("CallCountTracking", func(t *testing.T) {
		initialCount := mockProvider.GetCallCount()

		ctx := context.Background()
		prompt := "Test prompt for tracking"

		_, err := mockProvider.AnalyzeWithContext(ctx, prompt)
		if err != nil {
			t.Fatalf("AnalyzeWithContext failed: %v", err)
		}

		newCount := mockProvider.GetCallCount()
		if newCount != initialCount+1 {
			t.Errorf("Expected call count %d, got %d", initialCount+1, newCount)
		}

		if mockProvider.GetLastPrompt() != prompt {
			t.Errorf("Expected last prompt %q, got %q", prompt, mockProvider.GetLastPrompt())
		}
	})
}

// TestAutonomousFlow_ProviderErrorHandling tests error handling in autonomous mode
func TestAutonomousFlow_ProviderErrorHandling(t *testing.T) {
	// Create mock provider with error
	mockProvider := ai.NewMockProvider()
	mockProvider.SetError(ai.ErrProviderAuth)

	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Enabled:  true,
		},
	}

	engine := ai.NewEngineWithProvider(config, mockProvider, nil)

	ctx := context.Background()

	// Test that errors propagate correctly
	t.Run("ProviderErrorPropagates", func(t *testing.T) {
		req := &ai.AnalysisRequest{
			Events: []types.EvidenceEvent{
				{
					EventID:     "evt-1",
					Source:      "Test",
					Description: "Test event",
				},
			},
			Controls: []types.ControlObjective{
				{
					ID:   "CC6.1",
					Name: "Test Control",
				},
			},
		}

		_, err := engine.Analyze(ctx, req)
		if err == nil {
			t.Error("Expected error from provider, got nil")
		}
	})
}

// TestAutonomousFlow_ConfidenceScoring tests confidence score handling
func TestAutonomousFlow_ConfidenceScoring(t *testing.T) {
	tests := []struct {
		name            string
		confidenceScore float64
		expectHighConf  bool
	}{
		{
			name:            "high confidence",
			confidenceScore: 0.95,
			expectHighConf:  true,
		},
		{
			name:            "medium confidence",
			confidenceScore: 0.75,
			expectHighConf:  false,
		},
		{
			name:            "low confidence",
			confidenceScore: 0.45,
			expectHighConf:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := ai.NewMockProvider()
			mockProvider.SetConfidenceScore(tt.confidenceScore)

			config := &types.Config{
				AI: types.AIConfig{
					Provider: "mock",
					Enabled:  true,
				},
			}

			engine := ai.NewEngineWithProvider(config, mockProvider, nil)

			ctx := context.Background()
			req := &ai.AnalysisRequest{
				Events: []types.EvidenceEvent{
					{
						EventID:     "evt-1",
						Source:      "Test",
						Description: "Test event",
					},
				},
				Controls: []types.ControlObjective{
					{
						ID:   "CC6.1",
						Name: "Test Control",
					},
				},
			}

			resp, err := engine.Analyze(ctx, req)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			if len(resp.Findings) == 0 {
				t.Fatal("Expected findings")
			}

			// Check if confidence score is properly reflected
			// High confidence should result in findings with high confidence
			hasHighConfidence := false
			for _, finding := range resp.Findings {
				if finding.ConfidenceScore >= 0.9 {
					hasHighConfidence = true
					break
				}
			}

			if tt.expectHighConf && !hasHighConfidence {
				t.Error("Expected high confidence finding")
			}
		})
	}
}

// TestAutonomousFlow_ContextCancellation tests context cancellation handling
func TestAutonomousFlow_ContextCancellation(t *testing.T) {
	mockProvider := ai.NewMockProvider()

	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Enabled:  true,
		},
	}

	engine := ai.NewEngineWithProvider(config, mockProvider, nil)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &ai.AnalysisRequest{
		Events: []types.EvidenceEvent{
			{
				EventID:     "evt-1",
				Source:      "Test",
				Description: "Test event",
			},
		},
		Controls: []types.ControlObjective{
			{
				ID:   "CC6.1",
				Name: "Test Control",
			},
		},
	}

	_, err := engine.Analyze(ctx, req)
	// Should handle cancellation gracefully (may or may not return error depending on timing)
	// At minimum, should not crash
	t.Logf("Context cancellation result: %v", err)
}

// TestAutonomousFlow_EmptyEvents tests handling of empty event list
func TestAutonomousFlow_EmptyEvents(t *testing.T) {
	mockProvider := ai.NewMockProvider()

	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Enabled:  true,
		},
	}

	engine := ai.NewEngineWithProvider(config, mockProvider, nil)

	ctx := context.Background()
	req := &ai.AnalysisRequest{
		Events: []types.EvidenceEvent{}, // Empty events
		Controls: []types.ControlObjective{
			{
				ID:   "CC6.1",
				Name: "Test Control",
			},
		},
	}

	resp, err := engine.Analyze(ctx, req)
	if err != nil {
		t.Fatalf("Analyze failed with empty events: %v", err)
	}

	// Should handle gracefully, possibly with low confidence or no findings
	t.Logf("Response with empty events: %+v", resp)
}
