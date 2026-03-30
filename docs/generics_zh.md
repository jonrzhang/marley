# Generics — 泛型内存缓存

## 概述

本模块通过一个带 TTL 过期的类型安全内存缓存演示 **泛型**（Go 1.18 引入）。同一套 `Cache` 实现可以用于 `User`、`Product`、`string` 或任何其他类型——无需代码重复，无需 `interface{}` 类型断言。

## 使用场景

泛型出现前，通用缓存要么需要为每种类型单独实现（代码重复），要么使用 `map[string]interface{}`（到处需要类型断言，丢失类型安全，运行时可能因类型错误而 panic）。有了泛型，一个 `Cache[K, V]` 实现服务于所有类型，编译器负责强制类型正确性。

## 代码解析

### 类型参数

```go
type Cache[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]cacheItem[V]
    ttl   time.Duration
}
```

`[K comparable, V any]` 声明了两个**类型参数**：
- `K comparable` — key 类型；必须支持 `==`（map key 的要求）。`comparable` 是内置约束。
- `V any` — value 类型；可以是任何类型。`any` 是 `interface{}` 的别名。

约束 `comparable` 在编译时强制执行——无法用不可比较的 key 类型（如 slice）实例化 `Cache`。

### 泛型条目类型

```go
type cacheItem[V any] struct {
    value     V
    expiresAt time.Time
}
```

`cacheItem` 也是泛型的，用同样的 `V` 参数化。泛型类型可以相互引用——`Cache[K,V]` 包含 `map[K]cacheItem[V]`。

### 带类型推断的构造函数

```go
func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
    c := &Cache[K, V]{
        items: make(map[K]cacheItem[V]),
        ttl:   ttl,
    }
    go c.cleanup()
    return c
}
```

用法：`NewCache[int, User](500 * time.Millisecond)` — 显式指定类型参数。Go 有时可以从函数参数推断类型，但没有类型化参数的构造函数需要显式指定。

### 泛型代码中的零值

```go
func (c *Cache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    if !ok || time.Now().After(item.expiresAt) {
        var zero V    // 类型 V 的零值
        return zero, false
    }
    return item.value, true
}
```

`var zero V` 创建类型参数 `V` 的**零值**。不能在这里用 `nil`，因为 `V` 可能是 `int` 或 `struct`。`var` 声明无论什么类型都能得到正确的零值。

### 泛型函数

```go
func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}
```

`Number` 是使用 `~` 运算符的自定义约束：
```go
type Number interface {
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}
```
`~int` 表示"int 以及任何底层类型为 int 的类型"（例如 `type MyInt int` 也满足 `~int`）。`|` 联合允许约束中有多种类型。

### 泛型函数的类型推断

```go
Sum([]int{1, 2, 3})        // T 推断为 int
Sum([]float64{1.1, 2.2})   // T 推断为 float64
```

编译器从参数类型推断 `T`——不需要显式写 `Sum[int](...)`。

## 核心知识点

- **类型参数** — `[T constraint]` 声明类型参数；在调用处实例化，或显式写 `[T]`
- **约束** — `comparable`、`any`、带 `|` 联合和 `~` 底层类型的自定义接口
- **`~T`** — 匹配 `T` 以及所有以 `T` 为底层类型的类型
- **零值** — `var zero V` 获取任意类型参数的零值
- **类型推断** — 编译器通常可以从函数参数推断类型参数
- **泛型结构体** — 结构体可以参数化；方法使用相同的 `[K, V]` 语法

## 适用场景

| 场景 | 建议 |
|------|------|
| 同一算法适用于多种类型 | 泛型函数 |
| 类型安全的容器（栈、队列、缓存）| 泛型结构体 |
| 当前使用 `interface{}` + 类型断言 | 用泛型替换 |
| 约束是"任意数字类型" | `~int \| ~float64 ...` 约束 |
| 需要运行时类型分发（type switch）| 接口，而非泛型 |
