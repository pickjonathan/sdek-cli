package policy

import (
	"fmt"
)

// Loader handles loading policy excerpts for compliance frameworks
type Loader struct {
	excerpts map[string]string // control ID -> policy excerpt
}

// NewLoader creates a new policy loader with default excerpts
func NewLoader() *Loader {
	l := &Loader{
		excerpts: make(map[string]string),
	}
	l.loadDefaultExcerpts()
	return l
}

// GetExcerpt returns the policy excerpt for a given control ID
func (l *Loader) GetExcerpt(controlID string) (string, error) {
	excerpt, ok := l.excerpts[controlID]
	if !ok {
		return "", fmt.Errorf("no policy excerpt found for control %s", controlID)
	}
	return excerpt, nil
}

// LoadExcerpts loads policy excerpts from a map
func (l *Loader) LoadExcerpts(excerpts map[string]string) {
	for controlID, excerpt := range excerpts {
		l.excerpts[controlID] = excerpt
	}
}

// HasExcerpt checks if a policy excerpt exists for the given control ID
func (l *Loader) HasExcerpt(controlID string) bool {
	_, ok := l.excerpts[controlID]
	return ok
}

// GetAllControlIDs returns all control IDs that have policy excerpts
func (l *Loader) GetAllControlIDs() []string {
	controlIDs := make([]string, 0, len(l.excerpts))
	for controlID := range l.excerpts {
		controlIDs = append(controlIDs, controlID)
	}
	return controlIDs
}

// loadDefaultExcerpts loads the default policy excerpts from excerpts.go
func (l *Loader) loadDefaultExcerpts() {
	l.LoadExcerpts(SOC2Excerpts)
	l.LoadExcerpts(ISO27001Excerpts)
	l.LoadExcerpts(PCIDSSExcerpts)
}
