package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoSave manages automatic saving of state with debounce logic
type AutoSave struct {
	state        *State
	saveChan     chan struct{}
	stopChan     chan struct{}
	wg           sync.WaitGroup
	debounceTime time.Duration
	mu           sync.Mutex
	running      bool
}

// NewAutoSave creates a new AutoSave instance with 2-second debounce
func NewAutoSave(state *State) *AutoSave {
	return &AutoSave{
		state:        state,
		saveChan:     make(chan struct{}, 1),
		stopChan:     make(chan struct{}),
		debounceTime: 2 * time.Second,
		running:      false,
	}
}

// Start begins the auto-save loop
func (a *AutoSave) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("auto-save already running")
	}
	a.running = true
	a.mu.Unlock()

	a.wg.Add(1)
	go a.saveLoop(ctx)

	return nil
}

// Stop gracefully stops the auto-save loop and performs a final save
func (a *AutoSave) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	close(a.stopChan)
	a.wg.Wait()

	// Perform final save
	return a.state.Save()
}

// MarkDirty signals that the state has changed and should be saved
func (a *AutoSave) MarkDirty() {
	select {
	case a.saveChan <- struct{}{}:
		// Successfully marked dirty
	default:
		// Channel already has a pending save signal, no need to add another
	}
}

// saveLoop is the main loop that handles debounced saves
func (a *AutoSave) saveLoop(ctx context.Context) {
	defer a.wg.Done()
	defer func() {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
	}()

	var timer *time.Timer
	var timerRunning bool

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop the loop
			if timer != nil && timerRunning {
				timer.Stop()
			}
			return

		case <-a.stopChan:
			// Stop signal received
			if timer != nil && timerRunning {
				timer.Stop()
			}
			return

		case <-a.saveChan:
			// State marked dirty, start or reset the debounce timer
			if timer != nil && timerRunning {
				timer.Stop()
			}
			timer = time.NewTimer(a.debounceTime)
			timerRunning = true

		case <-func() <-chan time.Time {
			if timer != nil && timerRunning {
				return timer.C
			}
			// Return a nil channel that will never receive
			return nil
		}():
			// Debounce timer fired, perform save
			timerRunning = false
			if err := a.state.Save(); err != nil {
				// In a production app, we'd log this error
				// For now, we'll just ignore it as we can't return errors from this goroutine
				_ = err
			}
		}
	}
}

// IsRunning returns whether the auto-save loop is currently running
func (a *AutoSave) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}
