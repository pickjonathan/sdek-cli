# MCP Concurrent Access Fix

## Problem

When executing AI plans with multiple evidence collection items, the MCP stdio transport was experiencing concurrent access corruption leading to errors:

```
invalid character '\x00' looking for beginning of value
```

The issue occurred because:
1. Multiple goroutines (evidence collection tasks) were calling the same MCP transport concurrently
2. The `bufio.Scanner` was shared across all concurrent calls
3. Even with mutex locking, Scanner's internal state was not thread-safe

## Root Causes

### Issue 1: Shared Scanner State
```go
// BEFORE - BROKEN
type StdioTransport struct {
    scanner *bufio.Scanner  // Shared across all concurrent calls
    mu      sync.Mutex
    // ...
}
```

`bufio.Scanner` maintains internal buffer state that gets corrupted when multiple goroutines call `Scan()` concurrently, even with a mutex protecting the calls.

### Issue 2: Deadlock from Nested Locking
```go
// BEFORE - DEADLOCK
func (t *StdioTransport) Invoke(...) {
    defer func() {
        latency := time.Since(start)
        t.recordLatency(latency)  // This also tries to acquire t.mu
    }()
    
    t.mu.Lock()
    defer t.mu.Unlock()
    // ... request/response cycle
}

func (t *StdioTransport) recordLatency(...) {
    t.mu.Lock()  // DEADLOCK - already locked by Invoke
    defer t.mu.Unlock()
    // ...
}
```

The deferred `recordLatency()` call tried to acquire the mutex that was already held by `Invoke()`, causing a deadlock.

## Solution

### Fix 1: Replace Scanner with Reader
```go
// AFTER - FIXED
type StdioTransport struct {
    reader *bufio.Reader  // Thread-safe when used with mutex
    mu     sync.Mutex
    // ...
}

func (t *StdioTransport) Invoke(...) {
    t.mu.Lock()
    
    // ... write request ...
    
    // Read response line (thread-safe with mutex)
    responseBytes, err := t.reader.ReadBytes('\n')
    if err != nil {
        t.mu.Unlock()
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // ... parse response ...
    
    t.mu.Unlock()
}
```

`bufio.Reader.ReadBytes()` is safe to use with a mutex because:
- Each read operation is atomic
- No internal state corruption when protected by mutex
- Reads exactly one line at a time (newline-delimited JSON-RPC)

### Fix 2: Record Latency After Unlock
```go
// AFTER - FIXED
func (t *StdioTransport) Invoke(...) {
    start := time.Now()
    
    t.mu.Lock()
    // ... entire request-response cycle ...
    t.mu.Unlock()
    
    // Record latency AFTER releasing lock
    latency := time.Since(start)
    t.recordLatency(latency)  // Can safely acquire its own lock now
    
    return response.Result, nil
}
```

## Testing Results

### Before Fix
```
6 plan items generated
1 successful invocation (evidence_id="884effb1-0e7f-4140-b014-941bdeaaaece")
5 failed invocations (invalid character '\x00' error)
```

### After Fix
```
6 plan items generated
6 successful invocations:
  - evidence_id="51c189dd-fd50-4760-9bec-48811a9feec7"
  - evidence_id="29200ffb-f97b-46f9-88fb-72de7dc11bf2"
  - evidence_id="46a9f367-de7d-41c4-aaea-c2bf615d284d"
  - evidence_id="4c9f5f04-d466-4a34-aa03-9c1de81a95dc"
  - evidence_id="8e93d0ec-5dfb-4c33-a7d1-2e338640883a"
  - evidence_id="e51e7065-e97d-4b9d-afb1-84674726e313"
Complete finding generated with 70% confidence
```

## Key Learnings

1. **`bufio.Scanner` is not thread-safe** even with mutex protection - it maintains internal buffer state
2. **Use `bufio.Reader`** for concurrent access scenarios - its methods are atomic operations
3. **Avoid nested locking** - deferred functions that acquire the same mutex cause deadlocks
4. **Latency recording should happen outside critical sections** to avoid lock contention

## Files Modified

- `internal/mcp/transport/stdio.go`:
  - Changed `scanner *bufio.Scanner` → `reader *bufio.Reader`
  - Updated `Invoke()` to use `ReadBytes('\n')` instead of `scanner.Scan()`
  - Moved latency recording after mutex unlock
  - Explicit unlock before each error return

## Impact

This fix enables **fully concurrent MCP tool invocations** for autonomous AI evidence collection, allowing multiple AWS API calls to execute in parallel without corruption or deadlocks.
