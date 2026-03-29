# Channel — 日志处理 Pipeline

## 概述

本模块演示 Go **channel** 作为并发 pipeline 的连接纽带。三个独立阶段——读取原始日志行、解析为结构化条目、过滤并写入重要条目——各自在独立的 goroutine 中运行，通过有类型的 channel 连接。数据在 pipeline 中流动，无需任何共享内存或锁。

## 使用场景

日志处理系统需要高效处理高吞吐量的数据流。使用 pipeline 的好处：
- 各阶段并发运行——阶段 1 仍在读取时，阶段 2 已开始解析
- 阶段解耦——可以替换解析器或新增过滤阶段，无需修改其他阶段
- 背压自动实现——较慢的阶段通过 channel 阻塞自然减缓上游速度

实际类比：ETL 流水线、流处理器（如 Kafka 消费者）、图像处理链。

## 代码解析

### 有方向性的 Channel 类型

```go
func readLines(lines []string) <-chan string       // 返回只读 channel
func parseEntries(in <-chan string) <-chan LogEntry // 接收只读，返回只读
func filterAndWrite(in <-chan LogEntry) <-chan int  // 接收只读，返回只读
```

在函数签名中使用**有方向性的 channel**（`<-chan T` 只读，`chan<- T` 只写），在编译时强制约束数据流方向。只负责读取的阶段不会意外发送数据，反之亦然。

### 阶段 1：生产者

```go
func readLines(lines []string) <-chan string {
    out := make(chan string)  // 无缓冲
    go func() {
        defer close(out)     // 完成后始终关闭
        for _, line := range lines {
            out <- line      // 阻塞直到阶段 2 就绪
        }
    }()
    return out
}
```

模式是：创建 channel，启动一个 goroutine 向其发送数据并在完成后关闭，返回 channel。goroutine 并发运行；调用方立即获得一个可读取的 channel。

`defer close(out)` 至关重要——它通知下游阶段不会再有数据，让它们的 `range` 循环能够退出。

### 阶段 2：转换器

```go
func parseEntries(in <-chan string) <-chan LogEntry {
    out := make(chan LogEntry)
    go func() {
        defer close(out)
        for line := range in {       // 上游 channel 关闭时 range 自动退出
            parts := strings.SplitN(line, " ", 3)
            if len(parts) < 3 {
                continue             // 静默跳过格式错误的行
            }
            out <- LogEntry{Level: parts[0], Time: time.Now(), Message: parts[2]}
        }
    }()
    return out
}
```

`for line := range in` 阻塞等待下一个值，处理后，当上游 channel 关闭时自动退出。这是 pipeline 阶段中消费 channel 的惯用方式。

### 阶段 3：消费者

```go
func filterAndWrite(in <-chan LogEntry) <-chan int {
    count := make(chan int, 1)        // 缓冲大小 1：写入不阻塞
    go func() {
        n := 0
        for entry := range in {
            if entry.Level == "ERROR" || entry.Level == "WARN" {
                fmt.Printf("[STORE] %s | %s\n", entry.Level, entry.Message)
                n++
            }
        }
        count <- n                   // 完成后发送最终计数
    }()
    return count
}
```

返回**缓冲大小为 1 的 channel** 用于传递计数。goroutine 发送一次计数后退出；主 goroutine 通过 `<-count` 读取最终结果。缓冲防止 goroutine 在 main 还未读取时因发送而阻塞。

### Pipeline 组装

```go
lines   := readLines(rawLogs)
entries := parseEntries(lines)
count   := filterAndWrite(entries)

fmt.Printf("\nStored %d entries\n", <-count)
```

三行代码连接起三阶段并发 pipeline。每次函数调用立即返回；实际工作在后台 goroutine 中并发进行。

## 核心知识点

- **Pipeline 模式** — 链式调用各返回 channel 的函数；各阶段通过 channel 连接并发运行
- **`close(ch)` + `range`** — 发送流结束信号的标准方式；对 channel 的 `range` 在关闭时退出
- **有方向性的 channel** — `<-chan T`（只读）和 `chan<- T`（只写）在编译时强制约束数据流
- **无缓冲 channel** — 发送方阻塞直到接收方就绪，形成天然的背压机制
- **有缓冲 channel** — `make(chan T, N)` 允许最多 N 次无阻塞发送；用于结果 channel

## 适用场景

| 场景 | 建议 |
|------|------|
| 独立的转换阶段 | Channel Pipeline |
| 单个生产者，单个消费者 | 无缓冲 channel |
| 突发生产者，较慢消费者 | 有缓冲 channel 吸收峰值 |
| 扇出（一个源，多个消费者）| 多个 goroutine 从同一 channel 读取 |
| 扇入（多个源，一个消费者）| 用扇入函数合并多个 channel |
