# Errors — Config File Parser

## Overview

This module demonstrates Go's **error handling** philosophy: errors are values, not exceptions. It shows how to build a rich error hierarchy for a config parser — wrapping errors with `%w` to preserve context, defining sentinel errors for specific failure modes, and using `errors.Is` and `errors.As` to inspect errors at the call site.

## Use Case

A config parser fails for many reasons: missing required fields, unparseable values, out-of-range numbers. The caller needs to distinguish between these to give meaningful feedback ("port is required" vs "port must be 1-65535"). Without a structured error hierarchy, callers must parse error strings — fragile and untestable. Go's error wrapping provides a type-safe, traversable error chain.

## Code Walkthrough

### Sentinel Errors

```go
var ErrMissingField = errors.New("missing required field")
var ErrOutOfRange   = errors.New("value out of range")
```

Sentinel errors are package-level error values used as **identifiers**. Callers compare against them with `errors.Is`, not string matching. They convey the category of problem without carrying context (the context comes from the wrapping layer).

### Custom Error Type

```go
type ParseError struct {
    File  string
    Field string
    Value string
    Cause error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse %q in %s: invalid value %q", e.Field, e.File, e.Value)
}

func (e *ParseError) Unwrap() error { return e.Cause }
```

`ParseError` carries rich context: which file, which field, what value was rejected, and the underlying cause. The `Unwrap()` method is the key to error chain traversal — it tells the `errors` package how to walk up the chain. Without `Unwrap`, `errors.Is` and `errors.As` cannot see past `ParseError` to its `Cause`.

### Wrapping with %w

```go
func parsePort(raw string, file string) (int, error) {
    if raw == "" {
        return 0, &ParseError{
            File:  file,
            Field: "port",
            Value: raw,
            Cause: ErrMissingField,   // sentinel as cause
        }
    }
    port, err := strconv.Atoi(raw)
    if err != nil {
        return 0, &ParseError{File: file, Field: "port", Value: raw, Cause: err}
    }
    ...
}

func loadConfig(file string, raw map[string]string) (map[string]interface{}, error) {
    port, err := parsePort(raw["port"], file)
    if err != nil {
        return nil, fmt.Errorf("loadConfig(%s): %w", file, err)  // %w wraps
    }
    ...
}
```

`fmt.Errorf("...: %w", err)` creates a new error that wraps `err`. The chain is now:
```
loadConfig error → ParseError → ErrMissingField
```

### Inspecting the Error Chain

```go
switch {
case errors.Is(err, ErrMissingField):
    // walks the chain: finds ErrMissingField anywhere in it
    fmt.Printf("[MISS] %v\n", err)

case errors.Is(err, ErrOutOfRange):
    fmt.Printf("[RANGE] %v\n", err)

default:
    var pe *ParseError
    if errors.As(err, &pe) {
        // extracts the first *ParseError from the chain
        fmt.Printf("[PARSE] field=%s file=%s\n", pe.Field, pe.File)
    }
}
```

- **`errors.Is(err, target)`** — walks the error chain calling `Unwrap()` at each step, returns true if any error in the chain `==` target. Works with sentinel errors.
- **`errors.As(err, &target)`** — walks the chain, returns true and sets `target` if any error in the chain is assignable to the target type. Works with custom error types.

## Key Concepts

- **Errors as values** — `error` is a built-in interface; errors are returned like any other value, not thrown
- **`%w` in fmt.Errorf** — wraps an error, preserving the chain for `errors.Is`/`errors.As` traversal
- **`errors.Is`** — checks identity (sentinel errors) anywhere in the chain
- **`errors.As`** — extracts a typed error from anywhere in the chain
- **`Unwrap() error`** — the method that enables chain traversal; must be implemented on custom error types
- **Sentinel errors** — package-level `errors.New(...)` values used as categories

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Callers need to distinguish error categories | Define sentinel errors + `errors.Is` |
| Callers need structured error data | Define custom error type + `errors.As` |
| Adding context to an error | `fmt.Errorf("doing X: %w", err)` |
| Error from stdlib/third-party code | Wrap with context, let `%w` preserve the original |
| Unrecoverable internal errors | `panic` (not errors) — for truly impossible states |
