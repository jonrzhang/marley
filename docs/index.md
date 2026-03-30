---
layout: default
title: Marley — Go Learning Examples
---

# Marley — Go Learning Examples

A personal Go learning project demonstrating core language features through realistic application scenarios.
Each module is a standalone runnable example with bilingual documentation (English / 中文).

---

## Basic Modules

| Module | English | 中文 |
|--------|---------|------|
| GORM — ORM with SQLite | [EN](gorm_en) | [中文](gorm_zh) |
| Scanfile — Concurrent File Scanning | [EN](scanfile_en) | [中文](scanfile_zh) |
| MQTT — Publish / Subscribe | [EN](mqtt_en) | [中文](mqtt_zh) |

---

## Go Feature Examples

Each example demonstrates a core Go feature through a realistic scenario.

| Module | Feature | Scenario | English | 中文 |
|--------|---------|----------|---------|------|
| `goroutine/` | Goroutines | Parallel image downloading | [EN](goroutine_en) | [中文](goroutine_zh) |
| `channel/` | Channels | Log processing pipeline | [EN](channel_en) | [中文](channel_zh) |
| `select/` | Select | Multi-service health checker | [EN](select_en) | [中文](select_zh) |
| `defer/` | Defer | Database transaction management | [EN](defer_en) | [中文](defer_zh) |
| `interface/` | Interfaces | Notification system (Email/SMS/Push) | [EN](interface_en) | [中文](interface_zh) |
| `errors/` | Error handling | Config file parser | [EN](errors_en) | [中文](errors_zh) |
| `closure/` | Closures | HTTP middleware chain | [EN](closure_en) | [中文](closure_zh) |
| `embedding/` | Struct embedding | User permission system | [EN](embedding_en) | [中文](embedding_zh) |
| `context/` | Context | HTTP API client with timeout | [EN](context_en) | [中文](context_zh) |
| `generics/` | Generics | Type-safe in-memory cache | [EN](generics_en) | [中文](generics_zh) |
| `cgo/` | Cgo | System resource monitor | [EN](cgo_en) | [中文](cgo_zh) |

---

## Running the Examples

Each module is a standalone Go program:

```bash
cd <module> && go run main.go
```

Source code: [github.com/jonrzhang/marley](https://github.com/jonrzhang/marley)
