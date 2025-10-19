package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors MCP config directories for changes.
type Watcher struct {
	watcher  *fsnotify.Watcher
	paths    []string
	debounce time.Duration
}

// NewWatcher creates a new file watcher for MCP configs.
func NewWatcher(paths []string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		watcher:  watcher,
		paths:    paths,
		debounce: 500 * time.Millisecond,
	}

	// Add all paths to watch
	for _, path := range paths {
		if err := watcher.Add(path); err != nil {
			// Path might not exist, that's ok
			continue
		}
	}

	return w, nil
}

// Reloader is an interface for triggering config reloads.
type Reloader interface {
	Reload(ctx context.Context) (int, error)
}

// Watch starts watching for file changes and triggers reloads.
func (w *Watcher) Watch(ctx context.Context, reloader Reloader, stopCh <-chan struct{}, wg *interface{}) {

	var debounceTimer *time.Timer
	var debounceCh <-chan time.Time

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only care about write/create/remove events on .json files
			if filepath.Ext(event.Name) != ".json" {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				// Debounce: reset timer on each event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.NewTimer(w.debounce)
				debounceCh = debounceTimer.C
			}

		case <-debounceCh:
			// Debounce period elapsed, trigger reload
			fmt.Println("MCP config change detected, reloading...")
			if _, err := reloader.Reload(ctx); err != nil {
				fmt.Printf("warning: reload failed: %v\n", err)
			}
			debounceTimer = nil
			debounceCh = nil

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("watcher error: %v\n", err)

		case <-stopCh:
			return

		case <-ctx.Done():
			return
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
