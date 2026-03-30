# Interface — 通知系统

## 概述

本模块通过一个支持 Email、SMS 和 Push 渠道的通知系统演示 Go **接口**。每种渠道类型实现一个公共的 `Notifier` 接口，无需任何显式声明——Go 的**隐式接口满足**（鸭子类型）意味着任何拥有所需方法的类型都自动满足接口。

## 使用场景

通知系统需要不断演进：今天是 Email 和 SMS，明天是微信和 Slack。如果 `NotificationService` 依赖具体类型，添加新渠道就需要修改服务代码。通过面向 `Notifier` 接口编程，服务对修改封闭、对扩展开放——新渠道通过实现接口来添加，无需改动现有代码。

## 代码解析

### 接口定义

```go
type Notifier interface {
    Send(msg Message) error
    Channel() string
}
```

接口声明一个**方法集**：任何同时拥有 `Send` 和 `Channel` 这两个方法（签名完全匹配）的类型，都自动满足 `Notifier`。无需 `implements` 关键字，无需注册——直接生效。

### 实现接口

```go
type EmailNotifier struct {
    SMTPHost string
}

func (e EmailNotifier) Send(msg Message) error {
    fmt.Printf("[Email] To: %s | Subject: %s\n", msg.To, msg.Subject)
    return nil
}
func (e EmailNotifier) Channel() string { return "email" }
```

`EmailNotifier` 满足 `Notifier`，因为它有 `Send` 和 `Channel`。当 `EmailNotifier` 被用在需要 `Notifier` 的地方时，编译器会验证这一点。每种实现携带自己的配置（`SMTPHost`、`APIKey`、`AppID`）——接口无需了解这些细节。

`SMSNotifier` 添加了渠道特定的逻辑（截断超过 SMS 长度限制的消息），`PushNotifier` 将来可能添加优先级或角标数量。这些细节对服务层完全透明。

### 面向接口编程的服务

```go
type NotificationService struct {
    notifiers []Notifier   // 持有任何满足 Notifier 的类型
}

func (ns *NotificationService) Register(n Notifier) {
    ns.notifiers = append(ns.notifiers, n)
}

func (ns *NotificationService) Broadcast(msg Message) {
    var failed []string
    for _, n := range ns.notifiers {
        if err := n.Send(msg); err != nil {
            failed = append(failed, n.Channel())
        }
    }
    ...
}
```

`NotificationService` 对 Email、SMS、Push 一无所知——它只认识 `Notifier`。这使得测试极为简单：在测试中传入一个模拟的 `Notifier` 即可。同时也使服务面向未来：添加微信渠道不需要对 `NotificationService` 做任何修改。

### 类型断言

```go
var s Shape = Circle{Radius: 3}
if c, ok := s.(Circle); ok {
    fmt.Printf("It's a circle with radius %.1f\n", c.Radius)
}
```

当需要从接口值中取回具体类型时，使用**类型断言** `v.(T)`。两返回值形式 `v, ok := i.(T)` 是安全的——如果断言失败，`ok` 为 false，而不是 panic。尽量少用类型断言；如果频繁断言，说明接口设计可能有问题。

## 核心知识点

- **隐式满足** — 无需 `implements` 关键字；任何拥有所需方法的类型都满足接口
- **方法集** — 接口定义契约；具体类型提供实现
- **鸭子类型** — "如果它有 Send 和 Channel，它就是 Notifier"——将接口与实现包解耦
- **面向接口编程** — 用 `[]Notifier` 而非 `[]EmailNotifier`；实现多态性和可测试性
- **类型断言 `v.(T)`** — 获取具体类型；使用 `v, ok` 形式避免 panic
- **空接口 `interface{}`** — 被所有类型满足；Go 1.18 起可用 `any` 作为别名

## 适用场景

| 场景 | 建议 |
|------|------|
| 一种行为有多种实现 | 定义接口 |
| 测试时需要模拟依赖 | 面向接口编程，注入实现 |
| 插件/可扩展架构 | 接受接口参数，让调用方提供实现 |
| 与未知类型互操作 | 实现 `fmt.Stringer`、`io.Reader`、`io.Writer` 等标准接口 |
| 对变体进行类型分发 | 使用 `switch v := i.(type)` 进行穷举类型判断 |
