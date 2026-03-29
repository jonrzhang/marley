# Closure — HTTP Middleware Chain

## Overview

This module demonstrates Go **closures** through an HTTP middleware chain. Each middleware (logging, auth, rate limiting) is a function that returns a function — the returned function **closes over** its configuration, maintaining private state across calls. Middleware instances are composed into a chain using the `chain` helper.

## Use Case

HTTP servers need cross-cutting concerns (authentication, logging, rate limiting) applied consistently across handlers. Hardcoding these in every handler leads to duplication. The middleware pattern solves this: each concern is a wrapper that intercepts requests, does its work, and calls the next handler. Closures make each middleware instance carry its own private state (e.g., the rate limit counter is per-instance, not global).

## Code Walkthrough

### Types

```go
type HandlerFunc func(path, token string) (int, string)
type Middleware  func(HandlerFunc) HandlerFunc
```

`HandlerFunc` models a simplified HTTP handler (in real life, `func(http.ResponseWriter, *http.Request)`). `Middleware` is a higher-order function: it takes a handler and returns a new handler that wraps it.

### withLogging — Simple Closure

```go
func withLogging(next HandlerFunc) HandlerFunc {
    return func(path, token string) (int, string) {
        start := time.Now()
        status, body := next(path, token)       // call the wrapped handler
        fmt.Printf("[LOG] %s → %d (%v)\n", path, status, time.Since(start).Round(time.Microsecond))
        return status, body
    }
}
```

The returned function closes over `next` — a reference to the handler being wrapped. When called, it records `start`, delegates to `next`, then logs the result. This is the **decorator pattern** implemented as a closure.

### withAuth — Closure Over Configuration

```go
func withAuth(validToken string) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(path, token string) (int, string) {
            if token != validToken {    // validToken captured from outer scope
                return 401, "Unauthorized"
            }
            return next(path, token)
        }
    }
}
```

`withAuth` is a **factory function** — it takes a configuration value (`validToken`) and returns a `Middleware`. The innermost function closes over both `next` and `validToken`. Each call to `withAuth("different-token")` creates a completely independent closure with its own `validToken`.

### withRateLimit — Mutable Captured State

```go
func withRateLimit(maxReqs int) Middleware {
    count := 0    // mutable state, private to this closure instance
    return func(next HandlerFunc) HandlerFunc {
        return func(path, token string) (int, string) {
            count++           // modifies the captured variable
            if count > maxReqs {
                return 429, fmt.Sprintf("Rate limit exceeded (%d/%d)", count, maxReqs)
            }
            return next(path, token)
        }
    }
}
```

`count` is defined in `withRateLimit`'s scope and **captured by reference** by the closure. Every call to the returned handler increments the same `count`. Two separate calls to `withRateLimit(3)` create two independent closures, each with its own `count` — they don't share state.

### chain — Composing Middleware

```go
func chain(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

`chain` applies middlewares in reverse order so the first middleware in the list is the outermost wrapper (first to execute). Given `chain(h, A, B, C)`, the execution order is `A → B → C → h → C → B → A`.

```go
handler := chain(apiHandler,
    withLogging,           // executes first (outermost)
    withAuth("secret"),    // executes second
    withRateLimit(3),      // executes third (innermost before handler)
)
```

## Key Concepts

- **Closure** — a function that captures variables from its enclosing scope; the variables outlive the scope they were defined in
- **Capture by reference** — closures capture variables, not values; modifications to the variable inside the closure affect the original
- **Factory pattern** — a function that returns a closure pre-configured with specific values
- **Independent instances** — each call to a factory produces a closure with its own private state
- **Middleware / decorator pattern** — wrap a function to add behavior before/after it without modifying it

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Function needs per-call configuration | Factory function returning a closure |
| Need private, persistent state per instance | Variable captured in closure |
| Wrapping functions to add behavior | Middleware / decorator pattern with closures |
| Callback with context | Close over the needed context variables |
| Loop variable in goroutine (trap!) | Pass as parameter, don't capture from loop |
