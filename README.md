# marley

Marley is a cute puppy, and this project is for my own golang exercise.

## Modules

Each subdirectory is a standalone Go module. Run with `cd <module> && go run main.go`.

### Basics

| Module | Description |
|--------|-------------|
| `gorm/` | GORM ORM with SQLite — AutoMigrate, CRUD |
| `scanfile/` | Concurrent file scanning with goroutines and channels |
| `mqtt/` | MQTT pub/sub — subscribe, publish, WaitGroup |

### Go Feature Examples

Each module demonstrates a core Go feature through a realistic scenario:

| Module | Feature | Scenario |
|--------|---------|----------|
| `goroutine/` | Goroutines | Parallel image downloading |
| `channel/` | Channels | Log processing pipeline (read→parse→write) |
| `select/` | Select | Multi-service health checker with timeout |
| `defer/` | Defer | Database transaction auto commit/rollback |
| `interface/` | Interfaces | Notification system (Email/SMS/Push) |
| `errors/` | Error handling | Config file parser with errors.Is/As |
| `closure/` | Closures | HTTP middleware chain |
| `embedding/` | Struct embedding | User permission system |
| `context/` | Context | HTTP API client with timeout/cancellation |
| `generics/` | Generics | Type-safe in-memory cache with TTL |
| `cgo/` | Cgo | System resource monitor via C sysinfo() |
