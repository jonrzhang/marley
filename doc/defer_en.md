# Defer — Database Transaction Management

## Overview

This module demonstrates Go's **`defer` statement** through two realistic resource-management scenarios: database transaction commit/rollback and file handle cleanup. `defer` guarantees that cleanup code runs when a function exits, regardless of whether the exit is normal, an early return, or a panic.

## Use Case

Resource cleanup is error-prone when written manually: every early return path must remember to release the resource. `defer` solves this by registering cleanup at the point of acquisition, so future maintainers cannot accidentally add a return path that skips cleanup.

The database transaction pattern is the canonical example: you want automatic rollback on any error path, but if everything succeeds, the explicit `Commit` should suppress the rollback.

## Code Walkthrough

### The Idempotent Rollback Pattern

```go
func createOrder(userID int, fail bool) error {
    tx := beginTX()
    defer tx.Rollback()   // registered immediately after Open

    // ... execute operations ...

    return tx.Commit()    // sets tx.closed = true
}
```

```go
func (tx *TX) Rollback() {
    if !tx.closed {      // idempotent: safe to call even after Commit
        fmt.Printf("[DB] Rolled back (%d ops undone)\n", len(tx.ops))
        tx.closed = true
    }
}

func (tx *TX) Commit() error {
    fmt.Printf("[DB] Committed (%d ops)\n", len(tx.ops))
    tx.closed = true     // prevents Rollback from doing anything
    return nil
}
```

The `closed` flag makes `Rollback` **idempotent** — safe to call multiple times. The flow is:
- **Success path:** `Commit()` sets `closed = true` → deferred `Rollback()` sees `closed=true`, does nothing
- **Error path:** function returns an error before `Commit()` → deferred `Rollback()` fires, undoes all ops

This pattern means the caller never needs to explicitly call `Rollback` — it's always handled.

### LIFO Execution Order

```go
// If you had multiple defers:
defer fmt.Println("third")
defer fmt.Println("second")
defer fmt.Println("first")
// Output: first, second, third
```

Deferred calls execute in **Last-In, First-Out** order. This matches resource teardown semantics: if you open A then B, you want to close B before A.

### defer Modifies Named Return Values

```go
func deferWithReturn() (result int) {
    defer func() {
        result *= 2    // modifies the named return variable
    }()
    return 5           // sets result=5, then deferred func runs: result=10
}
```

When a function has **named return values**, deferred functions can read and modify them. `return 5` sets `result = 5`, then the deferred function runs and multiplies it by 2, so the caller receives 10. This is rarely needed but useful for things like wrapping errors at function exit.

### File Cleanup

```go
func processReport(filename string) {
    f := openFile(filename)
    defer f.Close()   // registered right after open

    // ... process file ...
}   // f.Close() guaranteed to run here
```

The file `Close` is registered immediately after `openFile`, making the pairing explicit and impossible to forget. Any code added between `defer` and the end of the function automatically has the cleanup guaranteed.

## Key Concepts

- **`defer` execution timing** — runs when the enclosing function returns, not when the `defer` statement is reached
- **LIFO order** — multiple defers execute in reverse registration order
- **Named return values** — deferred functions can modify named return variables
- **Idempotent cleanup** — use a `closed` or `done` flag to make cleanup safe to call multiple times
- **Arguments evaluated immediately** — `defer f(x)` evaluates `x` at the `defer` statement, not at execution time

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Unlock a mutex after locking | `lock.Lock(); defer lock.Unlock()` |
| Close a file/connection after opening | `defer f.Close()` immediately after open |
| Database transaction | `defer tx.Rollback()` + explicit `tx.Commit()` |
| Recover from panic | `defer func() { if r := recover(); r != nil { ... } }()` |
| Timing a function | `start := time.Now(); defer func() { log(time.Since(start)) }()` |
