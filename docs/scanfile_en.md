# Scanfile — Concurrent File Scanning

## Overview

This module demonstrates concurrent directory scanning using Go's **goroutines**, **channels**, and a **thread-safe queue**. A producer goroutine reads directory entries and sends filenames through an unbuffered channel; the main goroutine consumes them while also managing a concurrent-safe queue.

## Use Case

When processing large directories (e.g., an image pipeline, a log archiver, or a file indexer), sequential scanning creates a bottleneck. Separating the **scanning** from the **processing** with a channel lets both happen concurrently, improving throughput. A concurrent-safe queue provides a secondary buffer for batch processing.

## Code Walkthrough

### Package Structure

```
scanfile/
├── main.go          — orchestration: scanning + channel receive
└── queue/queue.go   — thread-safe FIFO queue
```

### Thread-Safe Queue (`queue/queue.go`)

```go
type Queue struct {
    data *list.List
}

var lock sync.Mutex
```

The queue wraps `container/list.List` (a doubly-linked list) and uses a **package-level** `sync.Mutex` to serialize access. `PushFront` inserts at the head; `Pop` removes from the tail — this gives **FIFO** ordering.

```go
func (q *Queue) Push(v interface{}) {
    defer lock.Unlock()
    lock.Lock()
    q.data.PushFront(v)
}
```

Note the `defer lock.Unlock()` appears **before** `lock.Lock()` — this is intentional. `defer` registers the call at the time of the statement but executes it when the function returns, so the lock is acquired first and released on exit regardless of the code path.

### Producer Goroutine (`scanDir`)

```go
func scanDir(dir string, fn chan string) {
    files, err := ioutil.ReadDir(path)
    ...
    go func() {
        for _, f := range files {
            fn <- f.Name()   // send filename into channel
            q.Push(f.Name()) // also push into queue
        }
        close(fn) // signal: no more files
    }()
}
```

- `ioutil.ReadDir` returns the directory listing sorted by name.
- The goroutine sends each filename into the **unbuffered** channel `fn`, blocking until the receiver is ready — this naturally paces the producer.
- `close(fn)` tells the consumer that the stream is finished.

### Consumer Loop (`main`)

```go
for {
    v, ok := <-fn
    if ok == false {
        break
    }
    fmt.Println("Received", v, ok)

    if q.Size() > 0 {
        n := q.Pop()
        fmt.Println("File Name", n)
    }
}
```

The **ok-idiom** (`v, ok := <-fn`) distinguishes between a real value and a closed channel. When `ok` is `false`, the channel is closed and the loop exits cleanly.

## Key Concepts

- **Unbuffered channel** — sender blocks until receiver is ready, creating natural backpressure between producer and consumer
- **`close(ch)`** — signals end-of-stream; consumers detect it via the ok-idiom or `range`
- **ok-idiom (`v, ok := <-ch`)** — the safe way to receive from a channel that may be closed
- **`sync.Mutex`** — protects the shared queue from concurrent access; `defer Unlock()` ensures the lock is always released
- **`container/list`** — Go's built-in doubly-linked list, used here as the queue's underlying storage

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Producer is slower than consumer | Use unbuffered channel for natural flow control |
| Producer is faster than consumer | Use buffered channel to absorb bursts |
| Multiple goroutines sharing a data structure | Protect with `sync.Mutex` or use `sync/atomic` |
| Processing items as they arrive | Channel pipeline is cleaner than polling a shared queue |
