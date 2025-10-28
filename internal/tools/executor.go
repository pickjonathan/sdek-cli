package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Executor handles parallel execution of tool calls with concurrency limits,
// timeout management, and result aggregation.
type Executor struct {
	registry       *ToolRegistry
	maxConcurrency int
	defaultTimeout time.Duration
	auditor        *AuditLogger
}

// NewExecutor creates a new tool executor with the given configuration.
func NewExecutor(registry *ToolRegistry, maxConcurrency int, defaultTimeout time.Duration, auditor *AuditLogger) *Executor {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // Default
	}
	if defaultTimeout <= 0 {
		defaultTimeout = 60 * time.Second // Default
	}

	return &Executor{
		registry:       registry,
		maxConcurrency: maxConcurrency,
		defaultTimeout: defaultTimeout,
		auditor:        auditor,
	}
}

// ExecuteParallel executes multiple tool calls in parallel with concurrency limits.
// Returns aggregated results and any errors encountered.
func (e *Executor) ExecuteParallel(ctx context.Context, calls []*types.ToolCall) ([]*types.ToolExecutionResult, error) {
	if len(calls) == 0 {
		return []*types.ToolExecutionResult{}, nil
	}

	// Create semaphore channel for concurrency control
	semaphore := make(chan struct{}, e.maxConcurrency)

	// Results and errors channels
	results := make(chan *types.ToolExecutionResult, len(calls))
	errors := make(chan error, len(calls))

	var wg sync.WaitGroup

	// Launch goroutines for each tool call
	for _, call := range calls {
		wg.Add(1)
		go func(c *types.ToolCall) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute with timeout
			result, err := e.executeWithTimeout(ctx, c)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(call)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)
	close(errors)

	// Aggregate results
	var allResults []*types.ToolExecutionResult
	for result := range results {
		allResults = append(allResults, result)
	}

	// Aggregate errors
	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	// If we have partial results, return them with error information
	if len(allErrors) > 0 {
		return allResults, fmt.Errorf("partial execution completed: %d succeeded, %d failed", len(allResults), len(allErrors))
	}

	return allResults, nil
}

// Execute executes a single tool call with timeout and audit logging.
func (e *Executor) Execute(ctx context.Context, call *types.ToolCall) (*types.ToolExecutionResult, error) {
	return e.executeWithTimeout(ctx, call)
}

// executeWithTimeout executes a tool call with timeout management.
func (e *Executor) executeWithTimeout(ctx context.Context, call *types.ToolCall) (*types.ToolExecutionResult, error) {
	// Determine timeout
	timeout := e.defaultTimeout
	if timeoutStr, ok := call.Context["timeout_seconds"]; ok && timeoutStr != "" {
		// Parse timeout from string (e.g., "30" means 30 seconds)
		var timeoutSec int
		if _, err := fmt.Sscanf(timeoutStr, "%d", &timeoutSec); err == nil && timeoutSec > 0 {
			timeout = time.Duration(timeoutSec) * time.Second
		}
	}

	// Create timeout context
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Log start
	startTime := time.Now()
	if e.auditor != nil {
		e.auditor.LogStart(call)
	}

	// Execute the tool call
	result, err := e.registry.Execute(execCtx, call)

	// Calculate latency
	latency := time.Since(startTime)

	// Handle timeout
	if execCtx.Err() == context.DeadlineExceeded {
		result = &types.ToolExecutionResult{
			ToolName:   call.ToolName,
			Success:    false,
			Output:     nil,
			Error:      fmt.Sprintf("execution timeout after %v", timeout),
			LatencyMS:  int64(latency.Milliseconds()),
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		err = fmt.Errorf("tool %s timed out after %v", call.ToolName, timeout)
	}

	// Handle cancellation
	if execCtx.Err() == context.Canceled {
		result = &types.ToolExecutionResult{
			ToolName:   call.ToolName,
			Success:    false,
			Output:     nil,
			Error:      "execution cancelled",
			LatencyMS:  int64(latency.Milliseconds()),
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		err = fmt.Errorf("tool %s execution cancelled", call.ToolName)
	}

	// If result is nil, create error result
	if result == nil && err != nil {
		result = &types.ToolExecutionResult{
			ToolName:   call.ToolName,
			Success:    false,
			Output:     nil,
			Error:      err.Error(),
			LatencyMS:  int64(latency.Milliseconds()),
			Timestamp:  time.Now().Format(time.RFC3339),
		}
	}

	// Ensure latency is set
	if result != nil {
		result.LatencyMS = int64(latency.Milliseconds())
		result.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Log completion
	if e.auditor != nil {
		e.auditor.LogComplete(call, result, err)
	}

	return result, err
}

// ExecutionProgress tracks progress of parallel execution.
type ExecutionProgress struct {
	Total      int
	Completed  int
	Failed     int
	InProgress int
}

// GetProgress returns current execution progress.
// This is a placeholder for future progress tracking implementation.
func (e *Executor) GetProgress() *ExecutionProgress {
	// TODO: Implement real-time progress tracking
	return &ExecutionProgress{}
}
