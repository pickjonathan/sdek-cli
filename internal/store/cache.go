package store

import (
	"fmt"
	"sync"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Cache provides fast in-memory access to state data with indexes
type Cache struct {
	state *State
	mu    sync.RWMutex

	// Indexes for fast lookups
	sourceByID     map[string]*types.Source
	eventsBySource map[string][]types.Event
	controlByID    map[string]*types.Control
	frameworkByID  map[string]*types.Framework
	eventByID      map[string]*types.Event
}

// NewCache creates a new cache from the given state
func NewCache(state *State) *Cache {
	c := &Cache{
		state:          state,
		sourceByID:     make(map[string]*types.Source),
		eventsBySource: make(map[string][]types.Event),
		controlByID:    make(map[string]*types.Control),
		frameworkByID:  make(map[string]*types.Framework),
		eventByID:      make(map[string]*types.Event),
	}
	c.rebuild()
	return c
}

// rebuild rebuilds all indexes from the state
func (c *Cache) rebuild() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear existing indexes
	c.sourceByID = make(map[string]*types.Source)
	c.eventsBySource = make(map[string][]types.Event)
	c.controlByID = make(map[string]*types.Control)
	c.frameworkByID = make(map[string]*types.Framework)
	c.eventByID = make(map[string]*types.Event)

	// Rebuild source index
	for i := range c.state.Sources {
		c.sourceByID[c.state.Sources[i].ID] = &c.state.Sources[i]
	}

	// Rebuild events by source index
	for _, event := range c.state.Events {
		c.eventsBySource[event.SourceID] = append(c.eventsBySource[event.SourceID], event)
		c.eventByID[event.ID] = &event
	}

	// Rebuild control index
	for i := range c.state.Controls {
		c.controlByID[c.state.Controls[i].ID] = &c.state.Controls[i]
	}

	// Rebuild framework index
	for i := range c.state.Frameworks {
		c.frameworkByID[c.state.Frameworks[i].ID] = &c.state.Frameworks[i]
	}
}

// Invalidate marks the cache as invalid and rebuilds all indexes
func (c *Cache) Invalidate() {
	c.rebuild()
}

// GetSourceByID retrieves a source by ID from the cache
func (c *Cache) GetSourceByID(id string) (*types.Source, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	source, exists := c.sourceByID[id]
	if !exists {
		return nil, fmt.Errorf("source not found: %s", id)
	}
	return source, nil
}

// GetEventsBySource retrieves all events for a given source ID
func (c *Cache) GetEventsBySource(sourceID string) []types.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	events, exists := c.eventsBySource[sourceID]
	if !exists {
		return []types.Event{}
	}

	// Return a copy to prevent external modification
	result := make([]types.Event, len(events))
	copy(result, events)
	return result
}

// GetEventByID retrieves an event by ID from the cache
func (c *Cache) GetEventByID(id string) (*types.Event, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	event, exists := c.eventByID[id]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", id)
	}
	return event, nil
}

// GetControlByID retrieves a control by ID from the cache
func (c *Cache) GetControlByID(id string) (*types.Control, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	control, exists := c.controlByID[id]
	if !exists {
		return nil, fmt.Errorf("control not found: %s", id)
	}
	return control, nil
}

// GetFrameworkByID retrieves a framework by ID from the cache
func (c *Cache) GetFrameworkByID(id string) (*types.Framework, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	framework, exists := c.frameworkByID[id]
	if !exists {
		return nil, fmt.Errorf("framework not found: %s", id)
	}
	return framework, nil
}

// GetAllSources returns all sources from the cache
func (c *Cache) GetAllSources() []types.Source {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]types.Source, len(c.state.Sources))
	copy(result, c.state.Sources)
	return result
}

// GetAllEvents returns all events from the cache
func (c *Cache) GetAllEvents() []types.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]types.Event, len(c.state.Events))
	copy(result, c.state.Events)
	return result
}

// GetAllFrameworks returns all frameworks from the cache
func (c *Cache) GetAllFrameworks() []types.Framework {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]types.Framework, len(c.state.Frameworks))
	copy(result, c.state.Frameworks)
	return result
}

// GetAllControls returns all controls from the cache
func (c *Cache) GetAllControls() []types.Control {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]types.Control, len(c.state.Controls))
	copy(result, c.state.Controls)
	return result
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]int{
		"sources":    len(c.sourceByID),
		"events":     len(c.eventByID),
		"frameworks": len(c.frameworkByID),
		"controls":   len(c.controlByID),
	}
}
