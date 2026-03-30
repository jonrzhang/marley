# Goroutine — Parallel Image Downloading

## Overview

This module demonstrates Go's **goroutines** — lightweight concurrent execution units. Five image downloads that would take 500ms–1500ms sequentially complete in roughly the time of the slowest single download, because all five run concurrently. `sync.WaitGroup` coordinates the completion of all goroutines before the main function aggregates results.

## Use Case

Any batch operation with independent work items is a candidate for goroutines: image/video processing, parallel API calls, database bulk inserts, or file format conversions. The key requirement is that each item's work does **not depend on the results of other items**.

## Code Walkthrough

### Result Type

```go
type DownloadResult struct {
    URL      string
    Size     int
    Duration time.Duration
    Err      error
}
```

Each goroutine writes into a pre-allocated `results` slice at a **fixed index**. Because no two goroutines write to the same index, no mutex is needed — this is safe by design.

### Launching Goroutines

```go
results := make([]DownloadResult, len(urls))
var wg sync.WaitGroup

for i, url := range urls {
    wg.Add(1)
    go func(idx int, u string) {
        defer wg.Done()
        results[idx] = download(u)
    }(i, url)
}
wg.Wait()
```

**Three critical details:**

1. `wg.Add(1)` is called **before** `go func(...)` — if it were inside the goroutine, the main goroutine might call `wg.Wait()` before any goroutine calls `Add`, causing an immediate return.

2. `defer wg.Done()` inside the goroutine ensures `Done` is called even if `download` panics.

3. `(i, url)` are passed as **function parameters**, not captured by closure. If you captured `i` and `url` from the loop directly, all goroutines would see the loop's final values by the time they execute.

### The download Function

```go
func download(url string) DownloadResult {
    start := time.Now()
    delay := time.Duration(rand.Intn(300)+100) * time.Millisecond
    time.Sleep(delay)

    if rand.Intn(10) == 0 {
        return DownloadResult{URL: url, Err: fmt.Errorf("connection timeout")}
    }
    return DownloadResult{URL: url, Size: rand.Intn(500) + 100, Duration: time.Since(start)}
}
```

`time.Sleep` simulates network latency. The 10% failure rate demonstrates that goroutines handle partial failures gracefully — failed goroutines return an `Err` value rather than crashing the program.

### Result Aggregation

```go
for _, r := range results {
    if r.Err != nil {
        fmt.Printf("[FAIL] %s — %v\n", r.URL, r.Err)
        failed++
    } else {
        fmt.Printf("[OK]   %s — %dKB in %v\n", r.URL, r.Size, r.Duration.Round(time.Millisecond))
        totalSize += r.Size
        success++
    }
}
```

After `wg.Wait()` returns, all goroutines have finished writing to `results`. The main goroutine can safely iterate the slice without any synchronization.

## Key Concepts

- **`go func()`** — the `go` keyword turns any function call into a goroutine; goroutines are cheap (2KB initial stack vs ~1MB for OS threads)
- **`sync.WaitGroup`** — `Add(n)` before launch, `Done()` per goroutine, `Wait()` to block until all finish
- **Loop variable capture** — always pass loop variables as parameters to goroutine functions, not via closure
- **Index-based result collection** — pre-allocate a slice and write at a fixed index; no mutex needed when indices don't overlap
- **`defer wg.Done()`** — guarantees the WaitGroup is decremented even on panic

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| N independent tasks, want results from all | Goroutines + WaitGroup + result slice |
| Want first successful result only | Goroutines + `select` on a result channel |
| Need to limit concurrency | Use a semaphore channel: `sem := make(chan struct{}, N)` |
| Tasks have dependencies (A before B) | Use channels or `errgroup` to express ordering |
