# Scanfile — 并发文件扫描

## 概述

本模块演示如何使用 Go 的 **goroutine**、**channel** 和**线程安全队列**实现并发目录扫描。生产者 goroutine 读取目录条目并通过无缓冲 channel 发送文件名；主 goroutine 消费这些文件名，同时维护一个并发安全的队列。

## 使用场景

处理大型目录（如图片处理流水线、日志归档、文件索引）时，顺序扫描会成为瓶颈。用 channel 将**扫描**和**处理**分离，让两者并发进行，从而提升吞吐量。并发安全队列则提供了一个二级缓冲，用于批量处理。

## 代码解析

### 包结构

```
scanfile/
├── main.go          — 编排：扫描 + channel 接收
└── queue/queue.go   — 线程安全的 FIFO 队列
```

### 线程安全队列（`queue/queue.go`）

```go
type Queue struct {
    data *list.List
}

var lock sync.Mutex
```

队列封装了 `container/list.List`（双向链表），使用**包级别**的 `sync.Mutex` 串行化访问。`PushFront` 从头部插入，`Pop` 从尾部移除，实现 **FIFO** 顺序。

```go
func (q *Queue) Push(v interface{}) {
    defer lock.Unlock()
    lock.Lock()
    q.data.PushFront(v)
}
```

注意 `defer lock.Unlock()` 写在 `lock.Lock()` **之前**——这是有意为之。`defer` 在语句执行时注册调用，但在函数返回时才执行，因此锁先被获取，无论函数如何退出都会被释放。

### 生产者 Goroutine（`scanDir`）

```go
func scanDir(dir string, fn chan string) {
    files, err := ioutil.ReadDir(path)
    ...
    go func() {
        for _, f := range files {
            fn <- f.Name()   // 将文件名发送到 channel
            q.Push(f.Name()) // 同时推入队列
        }
        close(fn) // 通知消费者：没有更多文件了
    }()
}
```

- `ioutil.ReadDir` 返回按名称排序的目录列表。
- goroutine 将每个文件名发送到**无缓冲** channel `fn`，在接收方准备好之前会阻塞——这天然地控制了生产者速度。
- `close(fn)` 通知消费者数据流已结束。

### 消费者循环（`main`）

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

**ok 惯用法**（`v, ok := <-fn`）区分了真实值和已关闭的 channel。当 `ok` 为 `false` 时，channel 已关闭，循环干净地退出。

## 核心知识点

- **无缓冲 channel** — 发送方阻塞直到接收方就绪，在生产者和消费者之间形成天然的背压机制
- **`close(ch)`** — 发送流结束信号；消费者通过 ok 惯用法或 `range` 检测到关闭
- **ok 惯用法（`v, ok := <-ch`）** — 从可能关闭的 channel 安全接收的方式
- **`sync.Mutex`** — 保护共享队列免受并发访问；`defer Unlock()` 确保锁总被释放
- **`container/list`** — Go 内置双向链表，此处用作队列的底层存储

## 适用场景

| 场景 | 建议 |
|------|------|
| 生产者比消费者慢 | 使用无缓冲 channel 实现自然的流量控制 |
| 生产者比消费者快 | 使用有缓冲 channel 吸收突发流量 |
| 多个 goroutine 共享数据结构 | 用 `sync.Mutex` 保护，或使用 `sync/atomic` |
| 边到达边处理 | Channel 流水线比轮询共享队列更简洁 |
