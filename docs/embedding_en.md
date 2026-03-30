# Embedding — User Permission System

## Overview

This module demonstrates Go **struct embedding** as a composition mechanism. `AdminUser` embeds both `BaseUser` and `AuditLog` — their fields and methods are **promoted** to `AdminUser`, usable as if defined directly on it. `AdminUser` can also override promoted methods, achieving the equivalent of inheritance without the tight coupling.

## Use Case

A permission system has two user tiers: regular users (read-only access to public resources) and admins (broader access with a full audit trail). Traditional OOP would use inheritance, but Go's embedding composes these capabilities more flexibly — `AuditLog` can be reused in other types (e.g., `ServiceAccount`) without any inheritance hierarchy.

## Code Walkthrough

### BaseUser — The Foundation

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

`BaseUser` defines the shared state and default behavior. It provides a default `CanAccess` that only grants access to `public/` resources.

### AuditLog — Reusable Capability

```go
type AuditLog struct {
    logs []string
}

func (a *AuditLog) Log(action string) {
    entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), action)
    a.logs = append(a.logs, entry)
}
```

`AuditLog` is a self-contained capability. It can be embedded in any type that needs audit logging, without `BaseUser` needing to know about it — this is composition over inheritance.

### AdminUser — Embedding Both

```go
type AdminUser struct {
    BaseUser               // anonymous (embedded) field
    AuditLog               // anonymous (embedded) field
    Permissions []string
}
```

Embedding is declared with just the type name, no field name. This is an **anonymous field**. The promoted methods and fields can be accessed directly:

```go
admin.Name          // promoted from BaseUser.Name
admin.String()      // promoted from BaseUser.String()
admin.Log("...")    // promoted from AuditLog.Log()
admin.PrintLogs()   // promoted from AuditLog.PrintLogs()
```

### Method Override

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

`AdminUser` defines its own `CanAccess`, which **shadows** (overrides) the promoted one from `BaseUser`. The admin version checks the `Permissions` slice and records every access decision in the audit log. The original `BaseUser.CanAccess` is still accessible via `admin.BaseUser.CanAccess(resource)`.

### Interface Satisfaction via Promotion

```go
func checkAccess(user interface{ CanAccess(string) bool }, resource string) {
    if user.CanAccess(resource) { ... }
}

checkAccess(regular, resource)  // BaseUser.CanAccess
checkAccess(admin, resource)    // AdminUser.CanAccess (overrides)
```

Both `BaseUser` and `*AdminUser` satisfy the anonymous interface `{ CanAccess(string) bool }` because both have the method (with different implementations). This polymorphism comes from Go's implicit interface satisfaction, not from explicit inheritance.

### Accessing the Embedded Type Directly

```go
admin.BaseUser.CanAccess(resource)  // explicitly calls BaseUser's version
```

The embedded type is accessible by its type name, letting you bypass the override when needed.

## Key Concepts

- **Embedding** — declaring a type without a field name inside a struct; its fields and methods are promoted
- **Method promotion** — promoted methods appear directly on the outer type
- **Method override (shadowing)** — define a method with the same name on the outer type to shadow the promoted one
- **Multiple embedding** — embed as many types as needed; each contributes its own set of promoted methods
- **Explicit access** — `outer.EmbeddedType.Method()` bypasses shadowing to reach the embedded type's method

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Share fields/methods across types | Embed a common base type |
| Add capability to a type | Embed the capability type (e.g., `sync.Mutex`, `AuditLog`) |
| Override behavior | Redefine the method on the outer type |
| Multiple independent capabilities | Embed multiple types |
| Named field vs embedding | Use a named field when you don't want method promotion |
