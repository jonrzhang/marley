# Defer — 数据库事务管理

## 概述

本模块通过两个实际的资源管理场景演示 Go 的 **`defer` 语句**：数据库事务的提交/回滚和文件句柄的清理。`defer` 保证清理代码在函数退出时执行，无论是正常返回、提前 return 还是 panic。

## 使用场景

手动编写资源清理代码容易出错：每个提前返回的路径都必须记住释放资源。`defer` 在资源获取时就注册清理操作，解决了这个问题——未来的维护者无法意外地添加一个跳过清理的 return 路径。

数据库事务模式是经典示例：任何错误路径都要自动回滚，但如果一切成功，显式的 `Commit` 应该阻止回滚。

## 代码解析

### 幂等回滚模式

```go
func createOrder(userID int, fail bool) error {
    tx := beginTX()
    defer tx.Rollback()   // 紧跟在打开事务后注册

    // ... 执行操作 ...

    return tx.Commit()    // 设置 tx.closed = true
}
```

```go
func (tx *TX) Rollback() {
    if !tx.closed {      // 幂等：即使在 Commit 后调用也安全
        fmt.Printf("[DB] Rolled back (%d ops undone)\n", len(tx.ops))
        tx.closed = true
    }
}

func (tx *TX) Commit() error {
    fmt.Printf("[DB] Committed (%d ops)\n", len(tx.ops))
    tx.closed = true     // 防止 Rollback 执行任何操作
    return nil
}
```

`closed` 标志使 `Rollback` 具有**幂等性**——多次调用安全。执行流程为：
- **成功路径：** `Commit()` 设置 `closed = true` → 延迟的 `Rollback()` 看到 `closed=true`，什么都不做
- **错误路径：** 函数在 `Commit()` 之前返回错误 → 延迟的 `Rollback()` 触发，撤销所有操作

这个模式意味着调用方永远不需要显式调用 `Rollback`——它总是被自动处理。

### LIFO 执行顺序

```go
// 如果有多个 defer：
defer fmt.Println("third")
defer fmt.Println("second")
defer fmt.Println("first")
// 输出：first, second, third
```

延迟调用按**后进先出（LIFO）**顺序执行。这符合资源释放的语义：如果先打开 A 再打开 B，应该先关闭 B 再关闭 A。

### defer 可修改命名返回值

```go
func deferWithReturn() (result int) {
    defer func() {
        result *= 2    // 修改命名返回变量
    }()
    return 5           // 设置 result=5，然后延迟函数执行：result=10
}
```

当函数有**命名返回值**时，延迟函数可以读取并修改它们。`return 5` 将 `result` 设为 5，然后延迟函数将其乘以 2，调用方最终收到 10。这种用法不常见，但在函数退出时包装错误等场景中很有用。

### 文件清理

```go
func processReport(filename string) {
    f := openFile(filename)
    defer f.Close()   // 紧跟在打开文件后注册

    // ... 处理文件 ...
}   // f.Close() 保证在此处执行
```

`defer f.Close()` 紧跟在 `openFile` 之后注册，使打开-关闭的配对关系清晰且不可遗漏。在 `defer` 之后函数内添加的任何代码，都自动享有清理保障。

## 核心知识点

- **`defer` 执行时机** — 在外围函数返回时执行，而非到达 `defer` 语句时
- **LIFO 顺序** — 多个 defer 按注册的逆序执行
- **命名返回值** — 延迟函数可以修改命名返回变量
- **幂等清理** — 使用 `closed` 或 `done` 标志，使清理函数多次调用安全
- **参数立即求值** — `defer f(x)` 中的 `x` 在 `defer` 语句处立即求值，不是在执行时

## 适用场景

| 场景 | 建议 |
|------|------|
| 加锁后解锁 | `lock.Lock(); defer lock.Unlock()` |
| 打开文件/连接后关闭 | 打开后立即 `defer f.Close()` |
| 数据库事务 | `defer tx.Rollback()` + 显式 `tx.Commit()` |
| 从 panic 中恢复 | `defer func() { if r := recover(); r != nil { ... } }()` |
| 函数计时 | `start := time.Now(); defer func() { log(time.Since(start)) }()` |
