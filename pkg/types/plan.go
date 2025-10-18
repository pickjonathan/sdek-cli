package types

import (
	"time"
)

// EvidencePlan represents a proposed evidence collection plan with approval workflow.
type EvidencePlan struct {
	// Plan metadata
	ID        string `json:"id"`        // Unique plan ID
	Framework string `json:"framework"` // From preamble
	Section   string `json:"section"`   // From preamble

	// Plan items
	Items []PlanItem `json:"items"` // Evidence sources to collect

	// Budget tracking
	EstimatedSources int `json:"estimated_sources"` // Total sources
	EstimatedCalls   int `json:"estimated_calls"`   // Total API calls
	EstimatedTokens  int `json:"estimated_tokens"`  // Total AI tokens

	// Status
	Status    PlanStatus `json:"status"`     // pending|approved|rejected|executing|complete
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PlanItem represents a single evidence source to query in the plan.
type PlanItem struct {
	// Source configuration
	Source  string   `json:"source"`  // "github", "jira", "aws", etc.
	Query   string   `json:"query"`   // Search query or filter
	Filters []string `json:"filters"` // Additional filters

	// Metadata
	SignalStrength float64 `json:"signal_strength"` // 0.0-1.0 estimated relevance
	Rationale      string  `json:"rationale"`       // Why this source/query

	// Approval
	ApprovalStatus ApprovalStatus `json:"approval_status"` // pending|approved|denied|auto_approved
	AutoApproved   bool           `json:"auto_approved"`   // Matched auto-approve policy

	// Execution
	ExecutionStatus ExecStatus `json:"execution_status,omitempty"` // pending|running|complete|failed
	EventsCollected int        `json:"events_collected,omitempty"` // Count after execution
	Error           string     `json:"error,omitempty"`            // Error if failed
}

// PlanStatus represents the overall status of an evidence plan.
type PlanStatus string

const (
	PlanPending   PlanStatus = "pending"
	PlanApproved  PlanStatus = "approved"
	PlanRejected  PlanStatus = "rejected"
	PlanExecuting PlanStatus = "executing"
	PlanComplete  PlanStatus = "complete"
)

// ApprovalStatus represents the approval state of a plan item.
type ApprovalStatus string

const (
	ApprovalPending      ApprovalStatus = "pending"
	ApprovalApproved     ApprovalStatus = "approved"
	ApprovalDenied       ApprovalStatus = "denied"
	ApprovalAutoApproved ApprovalStatus = "auto_approved"
)

// ExecStatus represents the execution state of a plan item.
type ExecStatus string

const (
	ExecPending  ExecStatus = "pending"
	ExecRunning  ExecStatus = "running"
	ExecComplete ExecStatus = "complete"
	ExecFailed   ExecStatus = "failed"
)
