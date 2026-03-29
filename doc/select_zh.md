# Select — 多服务健康检查

## 概述

本模块演示 Go 的 **`select` 语句**用于多路 channel 操作复用。程序同时探测多个服务，使用**扇入（fan-in）**模式将所有结果合并到一个 channel，再用 `select` 处理每条结果或在超时时提前退出。

## 使用场景

健康检查系统需要在时间预算内并行查询多个服务并收集结果。顺序检查太慢；无限期等待每个服务会导致挂起。`select` 正是解决这个问题的工具：等待哪个 channel 先有数据就处理哪个，若规定时间内没有数据则走超时分支。

## 代码解析

### probe — 单服务 Goroutine

```go
func probe(service string, latency time.Duration) <-chan HealthResult {
    ch := make(chan HealthResult, 1)  // 缓冲大小 1：goroutine 发送后不阻塞
    go func() {
        time.Sleep(latency)          // 模拟网络往返延迟
        ch <- HealthResult{
            Service: service,
            Latency: latency,
            Healthy: rand.Intn(5) != 0,
        }
    }()
    return ch
}
```

每个 probe 在独立 goroutine 中运行。channel **缓冲大小为 1**，goroutine 发送结果后可立即退出，无需等待主 goroutine 读取。如果没有缓冲，当主 goroutine 因超时停止读取时，probe goroutine 会泄漏（永久阻塞在发送）。

### fanIn — 合并多个 Channel

```go
func fanIn(channels ...<-chan HealthResult) <-chan HealthResult {
    merged := make(chan HealthResult, len(channels))
    for _, ch := range channels {
        go func(c <-chan HealthResult) {
            merged <- <-c       // 接收一个值，转发到 merged
        }(ch)
    }
    return merged
}
```

扇入模式将 N 个 channel 合并为 1 个。每个 channel 对应一个专用 goroutine，阻塞在 `<-c` 上并将结果转发到 `merged`。`merged` 的缓冲大小为 `len(channels)`，即使消费者提前退出，所有转发 goroutine 也能完成而不阻塞。

### select — 带超时的结果收集

```go
timeout := time.After(500 * time.Millisecond)
received := 0

for received < len(services) {
    select {
    case r := <-results:
        // 处理结果
        received++
    case <-timeout:
        fmt.Printf("[TIMEOUT] %d service(s) did not respond within 500ms\n", len(services)-received)
        return
    }
}
```

`time.After(d)` 返回一个在持续时间 `d` 后收到值的 channel。每次循环迭代，`select` 阻塞直到：
- `results` 上到达一条结果 → 处理它
- 超时触发 → 报告未响应的服务数量并退出

**重要：** `timeout` 在**循环外**创建一次，而非在循环内。如果在循环内创建，每次迭代都会重置 500ms 计时器，实际上禁用了整体超时。

多个 case 同时就绪时，`select` **随机**选择一个——防止某个 case 长期得不到处理。

## 核心知识点

- **`select`** — channel 操作版的 `switch`；阻塞直到某个 case 就绪，多个就绪时随机选择
- **`time.After(d)`** — 返回 `<-chan time.Time`，在指定时间后触发；惯用的超时模式
- **扇入（Fan-in）** — 用转发 goroutine 将 N 个 channel 合并为 1 个的模式
- **probe 中的缓冲 channel** — 防止消费者因超时提前退出时发生 goroutine 泄漏
- **单次超时** — 在循环外创建 `time.After`，而非循环内

## 适用场景

| 场景 | 建议 |
|------|------|
| 等待 N 个响应中的第一个 | `select` + N 个 channel case |
| 多个操作的整体超时 | 循环外的 `time.After` |
| 将 N 个并发结果合并为一个流 | 扇入模式 |
| 非阻塞 channel 检查 | `select` + `default` case |
| 定期轮询 | `time.Tick` 或 `time.NewTicker` |
