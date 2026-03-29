# GORM — 使用 ORM 操作 SQLite

## 概述

本模块演示如何使用 [GORM](https://gorm.io)（Go 最流行的 ORM 库）连接数据库、通过结构体标签定义数据模型，并自动完成表结构迁移。示例使用**内存 SQLite** 数据库，无需任何外部依赖即可运行。

## 使用场景

任何需要将数据持久化到关系型数据库的应用，都需要一个 ORM 层来避免手写 SQL。GORM 提供：
- 根据 Go 结构体自动生成并迁移表结构
- 流畅的查询、关联、事务 API
- 支持 SQLite、PostgreSQL、MySQL 等多种数据库

内存 SQLite 非常适合单元测试或原型开发——数据库仅在进程生命周期内存在。

## 代码解析

### 模型定义

```go
type Face struct {
    ID string `json:"id" gorm:"primary_key"`
}
```

`Face` 结构体映射到数据库中的一张表。结构体标签同时控制 JSON 序列化（`json:"id"`）和 GORM 行为（`gorm:"primary_key"`）。GORM 遵循约定优于配置原则，结构体名 `Face` 自动映射到表名 `faces`。

### 打开连接

```go
db, err := gorm.Open("sqlite3", ":memory:")
if err != nil {
    panic("Database open failed")
}
defer db.Close()
```

- `"sqlite3"` 是驱动名称。
- `":memory:"` 是特殊的 SQLite DSN，表示数据库仅存在于内存中。
- `defer db.Close()` 确保无论函数如何退出，数据库连接都会被释放。

### AutoMigrate 自动迁移

```go
db.AutoMigrate(&Face{})
```

`AutoMigrate` 检查结构体定义，自动创建或修改对应的表结构。它是**非破坏性**的——只会新增缺失的列或索引，不会删除已有数据。

## 核心知识点

- **结构体标签** — `gorm:"..."` 注解控制列名、主键、索引和约束，无需编写 DDL 语句
- **`defer db.Close()`** — 标准的资源清理模式，确保无论函数如何退出，连接都被释放
- **AutoMigrate** — 代码即 Schema：Go 结构体就是数据库表结构的唯一来源
- **内存数据库** — `":memory:"` 创建快速的临时 SQLite 实例，非常适合测试

## 适用场景

| 场景 | 建议 |
|------|------|
| 新 Go 项目需要数据库层 | 使用 GORM + `AutoMigrate` 快速原型开发 |
| 编写单元/集成测试 | 使用 `":memory:"` SQLite，避免依赖外部数据库 |
| 生产环境大数据量 | 对性能敏感的查询考虑使用原生 SQL 或 `sqlx` |
| 需要支持多种数据库 | GORM 的方言系统允许最小改动切换驱动 |
