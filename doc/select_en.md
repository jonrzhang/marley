# Select — Multi-Service Health Checker

## Overview

This module demonstrates Go's **`select` statement** for multiplexing multiple channel operations. It simultaneously probes several services, uses a **fan-in** pattern to merge all results into one channel, and uses `select` to either process each result or exit early on a timeout.

## Use Case

Health-check systems need to query many services in parallel and collect results within a time budget. Sequential checks are too slow; waiting for each service indefinitely causes hangs. `select` provides the exact tool: wait for whichever channel has data first, or take the timeout branch if nothing arrives in time.

## Code Walkthrough

### probe — Per-Service Goroutine

```go
func probe(service string, latency time.Duration) <-chan HealthResult {
    ch := make(chan HealthResult, 1)  // buffered: goroutine won't block after send
    go func() {
        time.Sleep(latency)          // simulate network round-trip
        ch <- HealthResult{
            Service: service,
            Latency: latency,
            Healthy: rand.Intn(5) != 0,
        }
    }()
    return ch
}
```

Each probe runs in its own goroutine. The channel is **buffered with size 1** so the goroutine can send its result and exit immediately, even if the main goroutine hasn't read it yet. Without the buffer, if the main goroutine times out and stops reading, the probe goroutine would leak (blocked forever on the send).

### fanIn — Merging Multiple Channels

```go
func fanIn(channels ...<-chan HealthResult) <-chan HealthResult {
    merged := make(chan HealthResult, len(channels))
    for _, ch := range channels {
        go func(c <-chan HealthResult) {
            merged <- <-c       // receive one value, forward to merged
        }(ch)
    }
    return merged
}
```

The fan-in pattern merges N channels into 1. Each channel gets a dedicated goroutine that blocks on `<-c` and forwards the result to `merged`. The `merged` channel is buffered with `len(channels)` so all forwarder goroutines can complete without blocking, even if the consumer stops early.

### select — Result Collection with Timeout

```go
timeout := time.After(500 * time.Millisecond)
received := 0

for received < len(services) {
    select {
    case r := <-results:
        // process result
        received++
    case <-timeout:
        fmt.Printf("[TIMEOUT] %d service(s) did not respond within 500ms\n", len(services)-received)
        return
    }
}
```

`time.After(d)` returns a channel that receives a value after duration `d`. In each loop iteration, `select` blocks until either:
- A result arrives on `results` → process it
- The timeout fires → report how many services didn't respond and exit

**Important:** `timeout` is created **once before the loop**, not inside it. If it were inside the loop, each iteration would create a fresh 500ms timer, effectively disabling the overall timeout.

`select` with multiple ready cases picks one **at random** — this prevents starvation when multiple results arrive simultaneously.

## Key Concepts

- **`select`** — like a `switch` for channel operations; blocks until one case is ready, picks randomly among multiple ready cases
- **`time.After(d)`** — returns `<-chan time.Time` that fires after duration; the idiomatic timeout pattern
- **Fan-in** — the pattern of merging N channels into 1 using forwarding goroutines
- **Buffered channels in probes** — prevent goroutine leaks when the consumer exits early due to timeout
- **Single timeout per operation** — create `time.After` once outside a loop, not inside

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Wait for first of N responses | `select` with N channel cases |
| Overall deadline for multiple operations | `time.After` outside the loop |
| Merge N concurrent results into one stream | Fan-in pattern |
| Non-blocking channel check | `select` with a `default` case |
| Periodic polling | `time.Tick` or `time.NewTicker` |
