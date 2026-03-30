# Errors — 配置文件解析器

## 概述

本模块演示 Go 的**错误处理**理念：错误是值，而非异常。它展示如何为配置解析器构建丰富的错误层次结构——用 `%w` 包装错误以保留上下文，定义哨兵错误表示特定的失败模式，并在调用处用 `errors.Is` 和 `errors.As` 检查错误。

## 使用场景

配置解析器的失败原因多种多样：缺少必填字段、无法解析的值、超出范围的数字。调用方需要区分这些情况，以提供有意义的反馈（"port 是必填项" vs "port 必须在 1-65535 之间"）。没有结构化的错误层次，调用方只能解析错误字符串——既脆弱又难以测试。Go 的错误包装提供了类型安全的、可遍历的错误链。

## 代码解析

### 哨兵错误

```go
var ErrMissingField = errors.New("missing required field")
var ErrOutOfRange   = errors.New("value out of range")
```

哨兵错误是包级别的错误值，用作**标识符**。调用方用 `errors.Is` 而非字符串匹配来比较它们。它们传达问题的类别，不携带上下文（上下文由包装层提供）。

### 自定义错误类型

```go
type ParseError struct {
    File  string
    Field string
    Value string
    Cause error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse %q in %s: invalid value %q", e.Field, e.File, e.Value)
}

func (e *ParseError) Unwrap() error { return e.Cause }
```

`ParseError` 携带丰富上下文：哪个文件、哪个字段、什么值被拒绝，以及底层原因。`Unwrap()` 方法是错误链遍历的关键——它告诉 `errors` 包如何向上追溯错误链。没有 `Unwrap`，`errors.Is` 和 `errors.As` 就无法穿透 `ParseError` 看到其 `Cause`。

### 用 %w 包装

```go
func parsePort(raw string, file string) (int, error) {
    if raw == "" {
        return 0, &ParseError{
            File:  file,
            Field: "port",
            Value: raw,
            Cause: ErrMissingField,   // 哨兵作为原因
        }
    }
    port, err := strconv.Atoi(raw)
    if err != nil {
        return 0, &ParseError{File: file, Field: "port", Value: raw, Cause: err}
    }
    ...
}

func loadConfig(file string, raw map[string]string) (map[string]interface{}, error) {
    port, err := parsePort(raw["port"], file)
    if err != nil {
        return nil, fmt.Errorf("loadConfig(%s): %w", file, err)  // %w 包装
    }
    ...
}
```

`fmt.Errorf("...: %w", err)` 创建一个包装了 `err` 的新错误。错误链现在是：
```
loadConfig error → ParseError → ErrMissingField
```

### 检查错误链

```go
switch {
case errors.Is(err, ErrMissingField):
    // 遍历错误链，在链中任意位置找到 ErrMissingField 则返回 true
    fmt.Printf("[MISS] %v\n", err)

case errors.Is(err, ErrOutOfRange):
    fmt.Printf("[RANGE] %v\n", err)

default:
    var pe *ParseError
    if errors.As(err, &pe) {
        // 从错误链中提取第一个 *ParseError
        fmt.Printf("[PARSE] field=%s file=%s\n", pe.Field, pe.File)
    }
}
```

- **`errors.Is(err, target)`** — 遍历错误链，在每一步调用 `Unwrap()`，如果链中任意错误 `==` target 则返回 true。适用于哨兵错误。
- **`errors.As(err, &target)`** — 遍历错误链，如果链中任意错误可赋值给目标类型，则返回 true 并设置 `target`。适用于自定义错误类型。

## 核心知识点

- **错误即值** — `error` 是内置接口；错误像普通值一样返回，不是抛出
- **`fmt.Errorf` 中的 `%w`** — 包装错误，为 `errors.Is`/`errors.As` 遍历保留错误链
- **`errors.Is`** — 检查错误链中任意位置的标识（哨兵错误）
- **`errors.As`** — 从错误链的任意位置提取特定类型的错误
- **`Unwrap() error`** — 使链遍历成为可能的方法；自定义错误类型必须实现它
- **哨兵错误** — 包级别的 `errors.New(...)` 值，用作错误类别标识

## 适用场景

| 场景 | 建议 |
|------|------|
| 调用方需要区分错误类别 | 定义哨兵错误 + `errors.Is` |
| 调用方需要结构化的错误数据 | 定义自定义错误类型 + `errors.As` |
| 为错误添加上下文 | `fmt.Errorf("doing X: %w", err)` |
| 来自标准库/第三方的错误 | 用上下文包装，`%w` 保留原始错误 |
| 不可恢复的内部错误 | `panic`（不是 error）——用于真正不可能发生的状态 |
