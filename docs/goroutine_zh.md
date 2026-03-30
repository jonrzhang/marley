# Goroutine — 并行图片下载

## 概述

本模块演示 Go 的 **goroutine**——轻量级并发执行单元。5 张图片的下载若顺序执行需要 500ms–1500ms，而并发运行时只需约等于最慢那次下载的时间。`sync.WaitGroup` 协调所有 goroutine 完成后，主函数再汇总结果。

## 使用场景

任何包含相互独立工作项的批量操作都适合使用 goroutine：图片/视频处理、并行 API 调用、数据库批量写入、文件格式转换等。关键前提是每个任务的工作**不依赖其他任务的结果**。

## 代码解析

### 结果类型

```go
type DownloadResult struct {
    URL      string
    Size     int
    Duration time.Duration
    Err      error
}
```

每个 goroutine 在预分配的 `results` 切片中写入**固定索引**位置。由于没有两个 goroutine 写同一个索引，无需加锁——这是设计上的安全保证。

### 启动 Goroutine

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

**三个关键细节：**

1. `wg.Add(1)` 在 `go func(...)` **之前**调用——如果放在 goroutine 内部，主 goroutine 可能在任何 goroutine 调用 `Add` 之前就调用了 `wg.Wait()`，导致立即返回。

2. goroutine 内的 `defer wg.Done()` 确保即使 `download` 发生 panic，`Done` 也会被调用。

3. `(i, url)` 以**函数参数**传入，而非通过闭包捕获。如果直接捕获循环变量，所有 goroutine 执行时可能看到的是循环结束时的最终值。

### download 函数

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

`time.Sleep` 模拟网络延迟。10% 的失败率演示了 goroutine 优雅处理局部失败的方式——失败的 goroutine 返回一个 `Err` 值，而不是让程序崩溃。

### 汇总结果

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

`wg.Wait()` 返回后，所有 goroutine 都已完成对 `results` 的写入。主 goroutine 无需任何同步即可安全遍历切片。

## 核心知识点

- **`go func()`** — `go` 关键字将任何函数调用变为 goroutine；goroutine 非常轻量（初始栈约 2KB，而 OS 线程约 1MB）
- **`sync.WaitGroup`** — 启动前 `Add(n)`，每个 goroutine 内 `Done()`，主函数 `Wait()` 阻塞直到全部完成
- **循环变量捕获陷阱** — 始终将循环变量作为参数传入 goroutine 函数，而非通过闭包捕获
- **基于索引的结果收集** — 预分配切片并写入固定索引；索引不重叠时无需加锁
- **`defer wg.Done()`** — 保证即使发生 panic，WaitGroup 计数器也能正确递减

## 适用场景

| 场景 | 建议 |
|------|------|
| N 个独立任务，需要所有结果 | Goroutine + WaitGroup + 结果切片 |
| 只需要第一个成功结果 | Goroutine + `select` 监听结果 channel |
| 需要限制并发数量 | 使用信号量 channel：`sem := make(chan struct{}, N)` |
| 任务有依赖关系（A 先于 B）| 使用 channel 或 `errgroup` 表达顺序 |
