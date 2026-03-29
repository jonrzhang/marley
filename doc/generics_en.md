# Generics — Type-Safe In-Memory Cache

## Overview

This module demonstrates **generics** (introduced in Go 1.18) through a type-safe in-memory cache with TTL expiration. The same `Cache` implementation works for `User`, `Product`, `string`, or any other type — without code duplication or `interface{}` type assertions.

## Use Case

Before generics, a generic cache required either a separate implementation per type (code duplication) or `map[string]interface{}` (requires type assertions everywhere, loses type safety, runtime panics on wrong types). With generics, one `Cache[K, V]` implementation serves all types while the compiler enforces type correctness.

## Code Walkthrough

### Type Parameters

```go
type Cache[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]cacheItem[V]
    ttl   time.Duration
}
```

`[K comparable, V any]` declares two **type parameters**:
- `K comparable` — the key type; must support `==` (needed for map keys). `comparable` is a built-in constraint.
- `V any` — the value type; can be anything. `any` is an alias for `interface{}`.

The constraint `comparable` is enforced at compile time — you cannot instantiate `Cache` with a non-comparable key type (e.g., a slice).

### Generic Item Type

```go
type cacheItem[V any] struct {
    value     V
    expiresAt time.Time
}
```

`cacheItem` is also generic, parameterized on the same `V`. Generic types can reference each other — `Cache[K,V]` contains `map[K]cacheItem[V]`.

### Constructor with Type Inference

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

Usage: `NewCache[int, User](500 * time.Millisecond)` — explicit type arguments. Go can sometimes infer types from function arguments, but constructors with no typed parameters require explicit specification.

### Zero Value in Generic Code

```go
func (c *Cache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    if !ok || time.Now().After(item.expiresAt) {
        var zero V    // zero value of type V
        return zero, false
    }
    return item.value, true
}
```

`var zero V` creates the **zero value** of the type parameter `V`. You cannot use `nil` here because `V` might be an `int` or `struct`. The `var` declaration always produces the correct zero value regardless of type.

### Generic Functions

```go
func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}
```

`Number` is a custom constraint using the `~` operator:
```go
type Number interface {
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}
```
`~int` means "int and any type whose underlying type is int" (e.g., `type MyInt int` also satisfies `~int`). The `|` union allows multiple types in a constraint.

### Type Inference in Generic Functions

```go
Sum([]int{1, 2, 3})        // T inferred as int
Sum([]float64{1.1, 2.2})   // T inferred as float64
```

The compiler infers `T` from the argument type — no explicit `Sum[int](...)` needed.

## Key Concepts

- **Type parameters** — `[T constraint]` declares a type parameter; instantiated at call site or with explicit `[T]`
- **Constraints** — `comparable`, `any`, custom interfaces with `|` unions and `~` underlying type
- **`~T`** — matches `T` and all types with `T` as underlying type
- **Zero value** — `var zero V` gets the zero value for any type parameter
- **Type inference** — the compiler often infers type parameters from function arguments
- **Generic structs** — structs can be parameterized; methods use the same `[K, V]` syntax

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| Same algorithm for multiple types | Generic function |
| Type-safe container (stack, queue, cache) | Generic struct |
| Currently using `interface{}` + type assertions | Replace with generics |
| Constraint is "any numeric type" | `~int \| ~float64 ...` constraint |
| Need runtime type dispatch (type switch) | Interfaces, not generics |
