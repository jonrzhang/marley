# Context — HTTP API Client

## Overview

This module demonstrates Go's **`context`** package through an HTTP API client that makes multiple sequential calls. Context provides a tree-shaped cancellation signal and key-value store that flows through an entire call chain — from the entry point down to the deepest library call — without needing explicit parameters at every level.

## Use Case

When a user cancels a request, every in-flight operation triggered by that request should stop promptly. Without context, each layer would need to poll a cancellation flag or use bespoke signaling. Context standardizes this: a single cancel/timeout at the top propagates to every function that respects `ctx.Done()`. The `requestID` carried in the context enables distributed tracing without threading it through every function signature.

## Code Walkthrough

### Custom Context Key

```go
type contextKey string
const requestIDKey contextKey = "requestID"
```

Using a **custom unexported type** for the key prevents collisions with keys from other packages. Two packages that both use `"requestID"` as a plain string key would collide. Two packages using their own `contextKey` type never will, because the key is matched by both value and type.

### httpGet — Respecting Cancellation

```go
func httpGet(ctx context.Context, url string) (*APIResponse, error) {
    reqID, _ := ctx.Value(requestIDKey).(string)
    fmt.Printf("[%s] GET %s\n", reqID, url)

    delay := time.Duration(rand.Intn(300)+100) * time.Millisecond

    select {
    case <-time.After(delay):      // simulated response
        return &APIResponse{StatusCode: 200, Body: `{"status":"ok"}`}, nil
    case <-ctx.Done():             // context cancelled or timed out
        return nil, fmt.Errorf("[%s] request to %s: %w", reqID, url, ctx.Err())
    }
}
```

`ctx.Done()` returns a channel that is closed when the context is cancelled or times out. The `select` races between a successful response and a cancellation signal — whichever happens first wins. `ctx.Err()` returns the reason: `context.DeadlineExceeded` or `context.Canceled`.

### fetchUserOrders — Propagating Context

```go
func fetchUserOrders(ctx context.Context, userID int) error {
    user, err := httpGet(ctx, fmt.Sprintf("/api/users/%d", userID))
    if err != nil {
        return fmt.Errorf("fetch user: %w", err)
    }
    orders, err := httpGet(ctx, fmt.Sprintf("/api/users/%d/orders", userID))
    if err != nil {
        return fmt.Errorf("fetch orders: %w", err)
    }
    return nil
}
```

The same `ctx` is passed to every sub-call. If the context is cancelled after the first call succeeds, the second call will immediately return an error without waiting for the simulated delay — the cancellation propagates automatically.

### The Three Context Patterns

**1. WithTimeout — automatic deadline:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()   // always call cancel to release resources
```
The context automatically cancels after 1 second. `defer cancel()` is essential — even if the operation finishes before the timeout, resources allocated for the timer must be freed.

**2. WithCancel — manual cancellation:**
```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(80 * time.Millisecond)
    cancel()   // trigger cancellation from another goroutine
}()
```
`cancel()` can be called from any goroutine. This models user-initiated cancellation (e.g., closing a browser tab).

**3. WithValue — request-scoped data:**
```go
ctx := context.WithValue(context.Background(), requestIDKey, "req-abc-001")
```
`WithValue` creates a derived context with an additional key-value pair. Values should be request-scoped metadata (trace IDs, auth tokens) — not business logic parameters (use function arguments for those).

## Key Concepts

- **`context.Background()`** — the root context; always non-nil, never cancelled; the starting point
- **`ctx.Done()`** — a channel closed when the context is cancelled; use in `select` to respond to cancellation
- **`ctx.Err()`** — returns `nil` if not cancelled, `DeadlineExceeded` or `Canceled` otherwise
- **`defer cancel()`** — must always be called after `WithTimeout`/`WithCancel` to prevent resource leaks
- **Custom key types** — prevent collisions between packages storing values in the same context
- **Context propagation** — pass `ctx` as the first parameter to every function that does I/O or could block

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| HTTP/RPC call with a deadline | `context.WithTimeout` |
| User-initiated cancellation | `context.WithCancel`, call `cancel()` on user action |
| Attaching trace ID / auth info | `context.WithValue` with a typed key |
| Checking if work should stop | `select { case <-ctx.Done(): return ctx.Err() }` |
| Library function signature | Accept `ctx context.Context` as the first parameter |
