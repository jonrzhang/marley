# Closure — HTTP 中间件链

## 概述

本模块通过 HTTP 中间件链演示 Go **闭包**。每个中间件（日志、认证、限流）都是一个返回函数的函数——返回的函数**封闭（close over）**了其配置，在调用间保持私有状态。中间件实例通过 `chain` 工具函数组合成链。

## 使用场景

HTTP 服务器需要将横切关注点（认证、日志、限流）一致地应用于所有处理器。在每个处理器中硬编码这些逻辑会导致重复。中间件模式解决了这个问题：每个关注点是一个包装器，拦截请求、完成工作后调用下一个处理器。闭包让每个中间件实例携带自己的私有状态（例如限流计数器是实例级别的，不是全局的）。

## 代码解析

### 类型定义

```go
type HandlerFunc func(path, token string) (int, string)
type Middleware  func(HandlerFunc) HandlerFunc
```

`HandlerFunc` 建模了一个简化的 HTTP 处理器（实际中是 `func(http.ResponseWriter, *http.Request)`）。`Middleware` 是一个高阶函数：接受一个处理器，返回包装了它的新处理器。

### withLogging — 简单闭包

```go
func withLogging(next HandlerFunc) HandlerFunc {
    return func(path, token string) (int, string) {
        start := time.Now()
        status, body := next(path, token)       // 调用被包装的处理器
        fmt.Printf("[LOG] %s → %d (%v)\n", path, status, time.Since(start).Round(time.Microsecond))
        return status, body
    }
}
```

返回的函数封闭了 `next`——被包装处理器的引用。调用时，记录 `start`，委托给 `next`，然后记录结果。这是用闭包实现的**装饰器模式**。

### withAuth — 封闭配置的闭包

```go
func withAuth(validToken string) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(path, token string) (int, string) {
            if token != validToken {    // validToken 从外层作用域捕获
                return 401, "Unauthorized"
            }
            return next(path, token)
        }
    }
}
```

`withAuth` 是一个**工厂函数**——接受配置值（`validToken`）并返回 `Middleware`。最内层函数同时封闭了 `next` 和 `validToken`。每次调用 `withAuth("different-token")` 都会创建一个完全独立的闭包，拥有自己的 `validToken`。

### withRateLimit — 可变捕获状态

```go
func withRateLimit(maxReqs int) Middleware {
    count := 0    // 可变状态，对这个闭包实例私有
    return func(next HandlerFunc) HandlerFunc {
        return func(path, token string) (int, string) {
            count++           // 修改捕获的变量
            if count > maxReqs {
                return 429, fmt.Sprintf("Rate limit exceeded (%d/%d)", count, maxReqs)
            }
            return next(path, token)
        }
    }
}
```

`count` 定义在 `withRateLimit` 的作用域中，被闭包**按引用捕获**。每次调用返回的处理器都会递增同一个 `count`。两次独立调用 `withRateLimit(3)` 创建两个独立的闭包，各有自己的 `count`——它们不共享状态。

### chain — 组合中间件

```go
func chain(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

`chain` 以逆序应用中间件，使列表中第一个中间件成为最外层包装器（最先执行）。对于 `chain(h, A, B, C)`，执行顺序为 `A → B → C → h → C → B → A`。

```go
handler := chain(apiHandler,
    withLogging,           // 最先执行（最外层）
    withAuth("secret"),    // 第二执行
    withRateLimit(3),      // 第三执行（最内层，紧邻处理器）
)
```

## 核心知识点

- **闭包** — 捕获外层作用域变量的函数；这些变量的生命周期超出了定义它们的作用域
- **按引用捕获** — 闭包捕获的是变量，而非值；闭包内对变量的修改会影响原始变量
- **工厂模式** — 返回预先配置了特定值的闭包的函数
- **独立实例** — 每次调用工厂函数都会产生拥有独立私有状态的闭包
- **中间件/装饰器模式** — 包装函数以在不修改它的情况下添加前/后行为

## 适用场景

| 场景 | 建议 |
|------|------|
| 函数需要按调用配置 | 返回闭包的工厂函数 |
| 每个实例需要私有持久状态 | 在闭包中捕获变量 |
| 包装函数以添加行为 | 使用闭包的中间件/装饰器模式 |
| 带上下文的回调 | 封闭所需的上下文变量 |
| 循环中的 goroutine（陷阱！）| 以参数传入，不要从循环中捕获 |
