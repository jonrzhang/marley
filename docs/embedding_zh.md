# Embedding — 用户权限系统

## 概述

本模块通过用户权限系统演示 Go **结构体嵌入**作为组合机制。`AdminUser` 嵌入了 `BaseUser` 和 `AuditLog`——它们的字段和方法被**提升（promoted）**到 `AdminUser`，可以直接使用，仿佛是在 `AdminUser` 上直接定义的。`AdminUser` 还可以覆盖提升的方法，实现继承的等效效果，但没有紧耦合。

## 使用场景

权限系统有两个用户层级：普通用户（只读访问公开资源）和管理员（更广泛的访问权限加完整审计追踪）。传统 OOP 会使用继承，但 Go 的嵌入以更灵活的方式组合这些能力——`AuditLog` 可以在其他类型（如 `ServiceAccount`）中复用，无需任何继承层次结构。

## 代码解析

### BaseUser — 基础结构

```go
type BaseUser struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}

func (u BaseUser) String() string {
    return fmt.Sprintf("User{id=%d, name=%s, email=%s}", u.ID, u.Name, u.Email)
}

func (u BaseUser) CanAccess(resource string) bool {
    return strings.HasPrefix(resource, "public/")
}
```

`BaseUser` 定义了共享状态和默认行为。它提供的默认 `CanAccess` 只允许访问 `public/` 资源。

### AuditLog — 可复用的能力

```go
type AuditLog struct {
    logs []string
}

func (a *AuditLog) Log(action string) {
    entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), action)
    a.logs = append(a.logs, entry)
}
```

`AuditLog` 是一个自包含的能力。它可以嵌入任何需要审计日志的类型，而 `BaseUser` 无需了解它——这就是组合优于继承。

### AdminUser — 嵌入两者

```go
type AdminUser struct {
    BaseUser               // 匿名（嵌入）字段
    AuditLog               // 匿名（嵌入）字段
    Permissions []string
}
```

嵌入只需写类型名，不写字段名。这就是**匿名字段**。提升的方法和字段可以直接访问：

```go
admin.Name          // 提升自 BaseUser.Name
admin.String()      // 提升自 BaseUser.String()
admin.Log("...")    // 提升自 AuditLog.Log()
admin.PrintLogs()   // 提升自 AuditLog.PrintLogs()
```

### 方法覆盖

```go
func (a *AdminUser) CanAccess(resource string) bool {
    for _, perm := range a.Permissions {
        if strings.HasPrefix(resource, perm) {
            a.Log(fmt.Sprintf("access granted: %s", resource))
            return true
        }
    }
    a.Log(fmt.Sprintf("access denied: %s", resource))
    return false
}
```

`AdminUser` 定义了自己的 `CanAccess`，**遮蔽（shadows）**了从 `BaseUser` 提升的同名方法。管理员版本检查 `Permissions` 切片，并在审计日志中记录每次访问决定。通过 `admin.BaseUser.CanAccess(resource)` 仍然可以访问原来的 `BaseUser.CanAccess`。

### 通过提升满足接口

```go
func checkAccess(user interface{ CanAccess(string) bool }, resource string) {
    if user.CanAccess(resource) { ... }
}

checkAccess(regular, resource)  // BaseUser.CanAccess
checkAccess(admin, resource)    // AdminUser.CanAccess（覆盖）
```

`BaseUser` 和 `*AdminUser` 都满足匿名接口 `{ CanAccess(string) bool }`，因为两者都有这个方法（实现不同）。这种多态性来自 Go 的隐式接口满足，而非显式继承。

### 直接访问嵌入类型

```go
admin.BaseUser.CanAccess(resource)  // 显式调用 BaseUser 的版本
```

嵌入类型可以通过其类型名访问，让你在需要时绕过方法覆盖，直接调用被嵌入类型的方法。

## 核心知识点

- **嵌入** — 在结构体中声明一个没有字段名的类型；其字段和方法被提升到外层类型
- **方法提升** — 提升的方法直接出现在外层类型上
- **方法覆盖（遮蔽）** — 在外层类型上定义同名方法，遮蔽提升的方法
- **多重嵌入** — 可以嵌入任意多个类型；每个类型贡献自己的提升方法集
- **显式访问** — `outer.EmbeddedType.Method()` 绕过遮蔽，直接调用嵌入类型的方法

## 适用场景

| 场景 | 建议 |
|------|------|
| 跨类型共享字段/方法 | 嵌入公共基础类型 |
| 为类型添加能力 | 嵌入能力类型（如 `sync.Mutex`、`AuditLog`）|
| 覆盖行为 | 在外层类型上重新定义方法 |
| 多个独立能力 | 嵌入多个类型 |
| 命名字段 vs 嵌入 | 不需要方法提升时使用命名字段 |
