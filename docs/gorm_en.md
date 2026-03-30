# GORM — ORM with SQLite

## Overview

This module demonstrates how to use [GORM](https://gorm.io), the most popular Go ORM library, to connect to a database, define models with struct tags, and automatically migrate the schema. It uses an **in-memory SQLite** database, so there are no external dependencies required to run it.

## Use Case

Any application that persists data to a relational database needs an ORM layer to avoid hand-writing SQL. GORM provides:
- Automatic schema migration from Go struct definitions
- A fluent API for queries, associations, and transactions
- Support for SQLite, PostgreSQL, MySQL, and others

In-memory SQLite is ideal for unit tests or prototyping — the database lives only for the duration of the process.

## Code Walkthrough

### Model Definition

```go
type Face struct {
    ID string `json:"id" gorm:"primary_key"`
}
```

The `Face` struct maps to a database table. Struct tags control both JSON serialization (`json:"id"`) and GORM behavior (`gorm:"primary_key"`). GORM uses convention over configuration — the struct name `Face` maps to a table named `faces`.

### Opening a Connection

```go
db, err := gorm.Open("sqlite3", ":memory:")
if err != nil {
    panic("Database open failed")
}
defer db.Close()
```

- `"sqlite3"` is the driver name.
- `":memory:"` is a special SQLite DSN meaning the database exists only in RAM.
- `defer db.Close()` ensures the connection is always released when `main` returns.

### AutoMigrate

```go
db.AutoMigrate(&Face{})
```

`AutoMigrate` inspects the struct definition and creates (or alters) the corresponding table to match. It is **non-destructive** — it adds missing columns and indexes but never drops existing ones.

## Key Concepts

- **Struct tags** — `gorm:"..."` annotations control column names, primary keys, indexes, and constraints without writing DDL
- **`defer db.Close()`** — the standard pattern to ensure resources are cleaned up regardless of how the function exits
- **AutoMigrate** — schema-as-code: the Go struct *is* the source of truth for the database schema
- **In-memory database** — `":memory:"` creates a fast, ephemeral SQLite instance perfect for tests

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| New Go project needing a database layer | Use GORM with `AutoMigrate` for rapid prototyping |
| Writing unit/integration tests | Use `":memory:"` SQLite to avoid external DB dependencies |
| Production with large datasets | Consider raw SQL or `sqlx` for performance-sensitive queries |
| Multiple database backends | GORM's dialect system lets you swap drivers with minimal code changes |
