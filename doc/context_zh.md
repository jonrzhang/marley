# Context — HTTP API 客户端

## 概述

本模块通过一个进行多次串行调用的 HTTP API 客户端演示 Go 的 **`context`** 包。Context 提供树状的取消信号和键值存储，从入口点一路流向最深层的库调用——无需在每个层级都显式传递参数。

## 使用场景

当用户取消一个请求时，该请求触发的所有进行中的操作都应立即停止。没有 context，每一层都需要轮询取消标志或使用特定的信号机制。Context 标准化了这一点：顶层的单个取消/超时会自动传播到所有遵循 `ctx.Done()` 的函数。context 中携带的 `requestID` 无需穿透每个函数签名，就能实现分布式追踪。

## 代码解析

### 自定义 Context Key

```go
type contextKey string
const requestIDKey contextKey = "requestID"
```

使用**自定义未导出类型**作为 key，防止与其他包的 key 冲突。两个包都用纯字符串 `"requestID"` 作为 key 会发生冲突。而使用各自的 `contextKey` 类型则永远不会冲突，因为 key 的匹配需要值和类型都相同。

### httpGet — 响应取消信号

```go
func httpGet(ctx context.Context, url string) (*APIResponse, error) {
    reqID, _ := ctx.Value(requestIDKey).(string)
    fmt.Printf("[%s] GET %s\n", reqID, url)

    delay := time.Duration(rand.Intn(300)+100) * time.Millisecond

    select {
    case <-time.After(delay):      // 模拟成功响应
        return &APIResponse{StatusCode: 200, Body: `{"status":"ok"}`}, nil
    case <-ctx.Done():             // context 被取消或超时
        return nil, fmt.Errorf("[%s] request to %s: %w", reqID, url, ctx.Err())
    }
}
```

`ctx.Done()` 返回一个在 context 被取消或超时时关闭的 channel。`select` 在成功响应和取消信号之间竞争——先发生的那个获胜。`ctx.Err()` 返回原因：`context.DeadlineExceeded` 或 `context.Canceled`。

### fetchUserOrders — 传播 Context

```go
func fetchUserOrders(ctx context.Context, userID int) error {
    user, err := httpGet(ctx, fmt.Sprintf("/api/users/%d", userID))
    if err != nil {
        return fmt.Errorf("fetch user: %w", err)
    }
    orders, err := httpGet(ctx, fmt.Sprintf("/api/users/%d/orders", userID))
    if err != nil {
        return fmt.Errorf("fetch orders: %w", err)
    }
    return nil
}
```

同一个 `ctx` 传递给每个子调用。如果 context 在第一次调用成功后被取消，第二次调用会立即返回错误，不会等待模拟延迟——取消信号自动传播。

### 三种 Context 使用模式

**1. WithTimeout — 自动截止时间：**
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()   // 始终调用 cancel 以释放资源
```
context 在 1 秒后自动取消。即使操作在超时前完成，`defer cancel()` 也是必须的——为计时器分配的资源需要被释放。

**2. WithCancel — 手动取消：**
```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(80 * time.Millisecond)
    cancel()   // 从另一个 goroutine 触发取消
}()
```
`cancel()` 可以从任何 goroutine 调用。这模拟了用户主动取消（如关闭浏览器标签页）。

**3. WithValue — 请求范围的数据：**
```go
ctx := context.WithValue(context.Background(), requestIDKey, "req-abc-001")
```
`WithValue` 创建一个附加了额外键值对的派生 context。值应该是请求范围的元数据（追踪 ID、认证令牌）——不应该是业务逻辑参数（业务参数应使用函数参数传递）。

## 核心知识点

- **`context.Background()`** — 根 context；始终非 nil，永不取消；是所有 context 的起点
- **`ctx.Done()`** — context 被取消时关闭的 channel；在 `select` 中用于响应取消
- **`ctx.Err()`** — 未取消时返回 `nil`，否则返回 `DeadlineExceeded` 或 `Canceled`
- **`defer cancel()`** — `WithTimeout`/`WithCancel` 后必须调用，防止资源泄漏
- **自定义 key 类型** — 防止不同包在同一 context 中存储值时发生冲突
- **Context 传播** — 将 `ctx` 作为第一个参数传递给每个进行 I/O 或可能阻塞的函数

## 适用场景

| 场景 | 建议 |
|------|------|
| 带截止时间的 HTTP/RPC 调用 | `context.WithTimeout` |
| 用户主动取消 | `context.WithCancel`，用户操作时调用 `cancel()` |
| 附加追踪 ID / 认证信息 | `context.WithValue` + 有类型的 key |
| 检查是否应停止工作 | `select { case <-ctx.Done(): return ctx.Err() }` |
| 库函数签名 | 将 `ctx context.Context` 作为第一个参数接受 |
