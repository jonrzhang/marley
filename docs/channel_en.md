# Channel — Log Processing Pipeline

## Overview

This module demonstrates Go **channels** as the connective tissue of a concurrent pipeline. Three independent stages — reading raw log lines, parsing them into structured entries, and filtering/writing important ones — each run in their own goroutine, connected by typed channels. Data flows through the pipeline without any shared memory or locks.

## Use Case

A log processing system must handle high-throughput streams efficiently. With a pipeline:
- Each stage runs concurrently, so stage 2 parses while stage 1 is still reading
- Stages are decoupled — you can swap out the parser or add a new filter stage without touching others
- Backpressure is automatic — a slow stage naturally slows the upstream stage via channel blocking

Real-world equivalents: ETL pipelines, stream processors (like Kafka consumers), image processing chains.

## Code Walkthrough

### Directional Channel Types

```go
func readLines(lines []string) <-chan string       // returns receive-only channel
func parseEntries(in <-chan string) <-chan LogEntry // takes receive-only, returns receive-only
func filterAndWrite(in <-chan LogEntry) <-chan int  // takes receive-only, returns receive-only
```

Using **directional channels** (`<-chan T` for receive-only, `chan<- T` for send-only) in function signatures enforces the data flow direction at compile time. A stage that is supposed to only read cannot accidentally send, and vice versa.

### Stage 1: Producer

```go
func readLines(lines []string) <-chan string {
    out := make(chan string)  // unbuffered
    go func() {
        defer close(out)     // always close when done
        for _, line := range lines {
            out <- line      // blocks until stage 2 is ready
        }
    }()
    return out
}
```

The pattern is: create a channel, launch a goroutine that sends into it and closes it when done, return the channel. The goroutine runs concurrently; the caller gets a channel it can read from immediately.

`defer close(out)` is critical — it signals downstream stages that no more data is coming, allowing them to exit their `range` loops.

### Stage 2: Transformer

```go
func parseEntries(in <-chan string) <-chan LogEntry {
    out := make(chan LogEntry)
    go func() {
        defer close(out)
        for line := range in {       // range exits when 'in' is closed
            parts := strings.SplitN(line, " ", 3)
            if len(parts) < 3 {
                continue             // silently skip malformed lines
            }
            out <- LogEntry{Level: parts[0], Time: time.Now(), Message: parts[2]}
        }
    }()
    return out
}
```

`for line := range in` blocks waiting for the next value, processes it, and exits automatically when the upstream channel is closed. This is the idiomatic way to drain a channel in a pipeline stage.

### Stage 3: Consumer

```go
func filterAndWrite(in <-chan LogEntry) <-chan int {
    count := make(chan int, 1)        // buffered: writer won't block
    go func() {
        n := 0
        for entry := range in {
            if entry.Level == "ERROR" || entry.Level == "WARN" {
                fmt.Printf("[STORE] %s | %s\n", entry.Level, entry.Message)
                n++
            }
        }
        count <- n                   // send final count when done
    }()
    return count
}
```

Returns a **buffered channel of size 1** for the count. The goroutine sends the count once and exits; the main goroutine reads it with `<-count` to get the final result. The buffer prevents the goroutine from blocking after sending the count, even if main hasn't read it yet.

### Pipeline Assembly

```go
lines   := readLines(rawLogs)
entries := parseEntries(lines)
count   := filterAndWrite(entries)

fmt.Printf("\nStored %d entries\n", <-count)
```

Three lines to wire up a three-stage concurrent pipeline. Each function call returns immediately; the work happens concurrently in the background goroutines.

## Key Concepts

- **Pipeline pattern** — chain functions that each return a channel; stages run concurrently connected by channels
- **`close(ch)` + `range`** — the standard way to signal end-of-stream; `range` on a channel exits when it's closed
- **Directional channels** — `<-chan T` (receive-only) and `chan<- T` (send-only) enforce data flow at compile time
- **Unbuffered channels** — sender blocks until receiver is ready, creating natural backpressure
- **Buffered channels** — `make(chan T, N)` allows up to N sends without blocking; used for result channels

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Independent transformation stages | Channel pipeline |
| Single goroutine producing, single consuming | Unbuffered channel |
| Bursty producer, slower consumer | Buffered channel to absorb spikes |
| Fan-out (one source, many consumers) | Multiple goroutines reading from the same channel |
| Fan-in (many sources, one consumer) | Merge channels with a fan-in function |
