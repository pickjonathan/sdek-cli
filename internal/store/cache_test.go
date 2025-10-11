package store

import (
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewCache(t *testing.T) {
	state := NewState()
	cache := NewCache(state)

	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	if cache.state != state {
		t.Error("Cache state pointer doesn't match")
	}

	stats := cache.GetStats()
	if stats["sources"] != 0 {
		t.Errorf("Expected 0 sources in cache, got %d", stats["sources"])
	}
}

func TestCacheSourceOperations(t *testing.T) {
	state := NewState()

	// Add sources to state
	source1 := types.Source{
		ID:         types.SourceTypeGit,
		Name:       "Git Repo 1",
		Type:       types.SourceTypeGit,
		Status:     "active",
		EventCount: 25,
		LastSync:   time.Now(),
		Enabled:    true,
	}
	source2 := types.Source{
		ID:         types.SourceTypeJira,
		Name:       "Jira Project",
		Type:       types.SourceTypeJira,
		Status:     "active",
		EventCount: 30,
		LastSync:   time.Now(),
		Enabled:    true,
	}

	if err := state.AddSource(source1); err != nil {
		t.Fatalf("Failed to add source1: %v", err)
	}
	if err := state.AddSource(source2); err != nil {
		t.Fatalf("Failed to add source2: %v", err)
	}

	// Create cache
	cache := NewCache(state)

	// Test GetSourceByID
	retrieved, err := cache.GetSourceByID(types.SourceTypeGit)
	if err != nil {
		t.Errorf("Failed to get source by ID: %v", err)
	}
	if retrieved.Name != "Git Repo 1" {
		t.Errorf("Expected name 'Git Repo 1', got '%s'", retrieved.Name)
	}

	// Test getting non-existent source
	_, err = cache.GetSourceByID("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent source")
	}

	// Test GetAllSources
	allSources := cache.GetAllSources()
	if len(allSources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(allSources))
	}

	// Test cache stats
	stats := cache.GetStats()
	if stats["sources"] != 2 {
		t.Errorf("Expected 2 sources in stats, got %d", stats["sources"])
	}
}

func TestCacheEventOperations(t *testing.T) {
	state := NewState()

	// Add events to state
	event1 := types.NewEvent("git-1", types.EventTypeCommit, "Commit 1", "author1")
	event2 := types.NewEvent("git-1", types.EventTypeCommit, "Commit 2", "author2")
	event3 := types.NewEvent("jira-1", types.EventTypeTicket, "Ticket 1", "author3")

	if err := state.AddEvent(*event1); err != nil {
		t.Fatalf("Failed to add event1: %v", err)
	}
	if err := state.AddEvent(*event2); err != nil {
		t.Fatalf("Failed to add event2: %v", err)
	}
	if err := state.AddEvent(*event3); err != nil {
		t.Fatalf("Failed to add event3: %v", err)
	}

	// Create cache
	cache := NewCache(state)

	// Test GetEventsBySource
	gitEvents := cache.GetEventsBySource("git-1")
	if len(gitEvents) != 2 {
		t.Errorf("Expected 2 events for git-1, got %d", len(gitEvents))
	}

	jiraEvents := cache.GetEventsBySource("jira-1")
	if len(jiraEvents) != 1 {
		t.Errorf("Expected 1 event for jira-1, got %d", len(jiraEvents))
	}

	// Test getting events for non-existent source
	noEvents := cache.GetEventsBySource("nonexistent")
	if len(noEvents) != 0 {
		t.Errorf("Expected 0 events for non-existent source, got %d", len(noEvents))
	}

	// Test GetEventByID
	retrieved, err := cache.GetEventByID(event1.ID)
	if err != nil {
		t.Errorf("Failed to get event by ID: %v", err)
	}
	if retrieved.Title != "Commit 1" {
		t.Errorf("Expected title 'Commit 1', got '%s'", retrieved.Title)
	}

	// Test GetAllEvents
	allEvents := cache.GetAllEvents()
	if len(allEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(allEvents))
	}

	// Test cache stats
	stats := cache.GetStats()
	if stats["events"] != 3 {
		t.Errorf("Expected 3 events in stats, got %d", stats["events"])
	}
}

func TestCacheFrameworkOperations(t *testing.T) {
	state := NewState()

	// Add frameworks to state
	framework1 := types.Framework{
		ID:           types.FrameworkSOC2,
		Name:         "SOC 2",
		Version:      "2017",
		Description:  "Service Organization Control 2",
		Category:     "security",
		ControlCount: 64,
	}
	framework2 := types.Framework{
		ID:           types.FrameworkISO27001,
		Name:         "ISO 27001",
		Version:      "2013",
		Description:  "Information Security Management",
		Category:     "security",
		ControlCount: 114,
	}

	if err := state.AddFramework(framework1); err != nil {
		t.Fatalf("Failed to add framework1: %v", err)
	}
	if err := state.AddFramework(framework2); err != nil {
		t.Fatalf("Failed to add framework2: %v", err)
	}

	// Create cache
	cache := NewCache(state)

	// Test GetFrameworkByID
	retrieved, err := cache.GetFrameworkByID(types.FrameworkSOC2)
	if err != nil {
		t.Errorf("Failed to get framework by ID: %v", err)
	}
	if retrieved.Name != "SOC 2" {
		t.Errorf("Expected name 'SOC 2', got '%s'", retrieved.Name)
	}

	// Test getting non-existent framework
	_, err = cache.GetFrameworkByID("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent framework")
	}

	// Test GetAllFrameworks
	allFrameworks := cache.GetAllFrameworks()
	if len(allFrameworks) != 2 {
		t.Errorf("Expected 2 frameworks, got %d", len(allFrameworks))
	}

	// Test cache stats
	stats := cache.GetStats()
	if stats["frameworks"] != 2 {
		t.Errorf("Expected 2 frameworks in stats, got %d", stats["frameworks"])
	}
}

func TestCacheControlOperations(t *testing.T) {
	state := NewState()

	// Add controls to state
	control1 := types.Control{
		ID:               "CC1.1",
		FrameworkID:      types.FrameworkSOC2,
		Title:            "Control Objective 1.1",
		Description:      "Test control 1",
		Category:         "access",
		RiskStatus:       types.RiskStatusGreen,
		RiskScore:        0.0,
		EvidenceCount:    5,
		ConfidenceLevel:  80.0,
		RequiredEvidence: 3,
	}
	control2 := types.Control{
		ID:               "CC1.2",
		FrameworkID:      types.FrameworkSOC2,
		Title:            "Control Objective 1.2",
		Description:      "Test control 2",
		Category:         "access",
		RiskStatus:       types.RiskStatusYellow,
		RiskScore:        0.5,
		EvidenceCount:    2,
		ConfidenceLevel:  50.0,
		RequiredEvidence: 3,
	}

	if err := state.AddControl(control1); err != nil {
		t.Fatalf("Failed to add control1: %v", err)
	}
	if err := state.AddControl(control2); err != nil {
		t.Fatalf("Failed to add control2: %v", err)
	}

	// Create cache
	cache := NewCache(state)

	// Test GetControlByID
	retrieved, err := cache.GetControlByID("CC1.1")
	if err != nil {
		t.Errorf("Failed to get control by ID: %v", err)
	}
	if retrieved.Title != "Control Objective 1.1" {
		t.Errorf("Expected title 'Control Objective 1.1', got '%s'", retrieved.Title)
	}

	// Test getting non-existent control
	_, err = cache.GetControlByID("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent control")
	}

	// Test GetAllControls
	allControls := cache.GetAllControls()
	if len(allControls) != 2 {
		t.Errorf("Expected 2 controls, got %d", len(allControls))
	}

	// Test cache stats
	stats := cache.GetStats()
	if stats["controls"] != 2 {
		t.Errorf("Expected 2 controls in stats, got %d", stats["controls"])
	}
}

func TestCacheInvalidate(t *testing.T) {
	state := NewState()

	// Add initial data
	source := types.Source{
		ID:         types.SourceTypeGit,
		Name:       "Git Repo 1",
		Type:       types.SourceTypeGit,
		Status:     "active",
		EventCount: 25,
		LastSync:   time.Now(),
		Enabled:    true,
	}
	if err := state.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Create cache
	cache := NewCache(state)

	// Verify initial state
	stats := cache.GetStats()
	if stats["sources"] != 1 {
		t.Errorf("Expected 1 source initially, got %d", stats["sources"])
	}

	// Add more data directly to state
	source2 := types.Source{
		ID:         types.SourceTypeJira,
		Name:       "Jira Project",
		Type:       types.SourceTypeJira,
		Status:     "active",
		EventCount: 30,
		LastSync:   time.Now(),
		Enabled:    true,
	}
	if err := state.AddSource(source2); err != nil {
		t.Fatalf("Failed to add source2: %v", err)
	}

	// Invalidate cache to pick up new data
	cache.Invalidate()

	// Verify updated state
	stats = cache.GetStats()
	if stats["sources"] != 2 {
		t.Errorf("Expected 2 sources after invalidation, got %d", stats["sources"])
	}

	// Verify we can retrieve the new source
	retrieved, err := cache.GetSourceByID(types.SourceTypeJira)
	if err != nil {
		t.Errorf("Failed to get new source after invalidation: %v", err)
	}
	if retrieved.Name != "Jira Project" {
		t.Errorf("Expected name 'Jira Project', got '%s'", retrieved.Name)
	}
}

func TestCacheConcurrency(t *testing.T) {
	state := NewState()
	cache := NewCache(state)

	// Test concurrent reads don't panic
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			cache.GetAllSources()
			cache.GetAllEvents()
			cache.GetAllFrameworks()
			cache.GetAllControls()
			cache.GetStats()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
