# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Marley is a personal Go learning/exercise project. It contains independent modules exploring different Go concepts.

## Running Code

Each subdirectory is a standalone Go program:

```bash
cd gorm && go run main.go
cd scanfile && go run main.go
```

Note: There is no `go.mod` file. If dependencies are missing, initialize modules first:

```bash
cd gorm && go mod init github.com/jonrzhang/marley/gorm && go mod tidy
cd scanfile && go mod init github.com/jonrzhang/marley/scanfile && go mod tidy
```

## Architecture

Two independent learning modules:

### `gorm/`
Demonstrates GORM ORM with SQLite. Uses `github.com/jinzhu/gorm` and an in-memory SQLite3 database with AutoMigrate.

### `scanfile/`
Demonstrates concurrent file scanning with goroutines and channels. Has two packages:
- `main.go` — orchestrates directory scanning, sends filenames through Go channels
- `queue/queue.go` — thread-safe queue wrapping `container/list.List` with `sync.Mutex`

The scanfile module has a hardcoded macOS path (`/Users/zhangrong/Downloads/Images/Cropped`) that needs to be changed for local use.

### `mqtt/`
Demonstrates MQTT pub/sub with `github.com/eclipse/paho.mqtt.golang`. Connects to a public broker (`broker.emqx.io:1883`), subscribes to a topic, publishes 5 messages, and uses `sync.WaitGroup` to wait for all messages to be received before disconnecting.

## Go Feature Examples

Each of the following modules demonstrates a core Go feature through a realistic application scenario:

| Module | Feature | Scenario |
|--------|---------|---------|
| `goroutine/` | Goroutines + WaitGroup | Parallel image downloading |
| `channel/` | Channels + Pipeline | Log processing pipeline (read→parse→write) |
| `select/` | Select statement | Multi-service health checker with timeout |
| `defer/` | Defer | Database transaction auto commit/rollback |
| `interface/` | Interfaces | Notification system (Email/SMS/Push) |
| `errors/` | Error wrapping | Config file parser with errors.Is/As |
| `closure/` | Closures | HTTP middleware chain (auth, logging, rate limit) |
| `embedding/` | Struct embedding | User permission system (BaseUser → AdminUser) |
| `context/` | Context | HTTP API client with timeout and cancellation |
| `generics/` | Generics | Type-safe in-memory cache with TTL |
| `cgo/` | Cgo | System resource monitor via C sysinfo() |

Run any example: `cd <module> && go run main.go`

## Documentation

The `doc/` directory contains detailed explanations for every module, in both English and Chinese:

```
doc/<module>_en.md   — English
doc/<module>_zh.md   — Chinese
```

Covers: implementation logic, code walkthrough, key concepts, and when-to-use guidance.
