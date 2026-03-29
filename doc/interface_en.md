# Interface — Notification System

## Overview

This module demonstrates Go **interfaces** through a notification system that supports Email, SMS, and Push channels. Each channel type implements a common `Notifier` interface without any explicit declaration — Go's **implicit interface satisfaction** (duck typing) means any type with the required methods automatically satisfies the interface.

## Use Case

A notification system must evolve: today Email and SMS, tomorrow WhatsApp and Slack. If `NotificationService` depended on concrete types, adding a new channel would require modifying the service. By programming to the `Notifier` interface, the service is closed for modification but open for extension — new channels are added by implementing the interface, not by changing existing code.

## Code Walkthrough

### Interface Definition

```go
type Notifier interface {
    Send(msg Message) error
    Channel() string
}
```

The interface declares a **method set**: any type that has both `Send` and `Channel` methods with these exact signatures automatically satisfies `Notifier`. No `implements` keyword, no registration — it just works.

### Implementing the Interface

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

`EmailNotifier` satisfies `Notifier` because it has `Send` and `Channel`. The compiler verifies this when an `EmailNotifier` is used where a `Notifier` is expected. Each implementation carries its own configuration (`SMTPHost`, `APIKey`, `AppID`) — the interface doesn't need to know about these.

`SMSNotifier` adds channel-specific logic (truncating long messages to SMS limits), while `PushNotifier` might in the future add priority or badge count. These details are invisible to the service.

### Service Programming to Interface

```go
type NotificationService struct {
    notifiers []Notifier   // holds any type satisfying Notifier
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

`NotificationService` knows nothing about Email, SMS, or Push — it only knows `Notifier`. This makes it trivially testable: pass a mock `Notifier` in tests. It also makes the service future-proof: adding WeChat requires zero changes to `NotificationService`.

### Type Assertion

```go
var s Shape = Circle{Radius: 3}
if c, ok := s.(Circle); ok {
    fmt.Printf("It's a circle with radius %.1f\n", c.Radius)
}
```

When you need the concrete type back from an interface value, use a **type assertion** `v.(T)`. The two-return form `v, ok := i.(T)` is safe — `ok` is false if the assertion fails, instead of panicking. Use type assertions sparingly; if you find yourself asserting frequently, the interface may not be designed correctly.

## Key Concepts

- **Implicit satisfaction** — no `implements` keyword; any type with the required methods satisfies the interface
- **Method set** — an interface defines a contract; concrete types provide the implementation
- **Duck typing** — "if it has Send and Channel, it's a Notifier" — decouples interface from implementation package
- **Programming to interfaces** — `[]Notifier` instead of `[]EmailNotifier`; enables polymorphism and testability
- **Type assertion `v.(T)`** — recovers the concrete type; use the `v, ok` form to avoid panics
- **Empty interface `interface{}`** — satisfied by all types; use `any` (alias since Go 1.18) for generic values

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Multiple implementations of one behavior | Define an interface |
| Want to mock dependencies in tests | Program to interfaces, inject implementations |
| Plugin/extensible architecture | Accept an interface, let callers provide implementations |
| Interoperating with unknown types | `fmt.Stringer`, `io.Reader`, `io.Writer` are standard interfaces to implement |
| Type switching on variants | `switch v := i.(type)` for exhaustive type dispatch |
