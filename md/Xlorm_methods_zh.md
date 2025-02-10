# XLORM 核心方法文档

[中文版](Xlorm_methods_zh.md "访问中文版")

[English](Xlorm_methods_en.md "Access English Version")

## 概述

`xlorm.go` 是 XLORM 框架的核心文件，提供了数据库连接、事务管理、日志控制和性能监控等基础功能。这个文件定义了 `DB` 结构体，是整个框架的核心入口。

### 主要特点

- **数据库连接管理**：提供灵活的数据库连接配置和初始化
- **事务处理**：支持手动和自动事务管理
- **日志系统**：提供动态日志级别调整和调试功能
- **性能监控**：内置连接池统计和查询性能指标收集
- **上下文支持**：集成 Go 语言上下文机制

### 使用场景

- 创建数据库连接
- 管理数据库事务
- 配置日志和调试
- 监控数据库性能
- 管理数据库连接生命周期

## 数据库连接方法

### New
- 创建新的数据库连接
- 签名：`New(cfg *Config) (*DB, error)`
- 示例：
```go
// 创建 MySQL 数据库连接
config := &Config{
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: "password",
    DBName:   "mydb",
}
db, err := xlorm.New(config)
if err != nil {
    log.Fatal(err)
}
```

### M 和 Table
- 获取表操作对象
- 签名：`M(tableName string) *Table`
- 签名：`Table(tableName string) *Table`
- 示例：
```go
// 获取用户表操作对象
userTable := db.M("users")
// 或者
userTable := db.Table("users")
```

## 上下文管理方法

### WithContext
- 设置数据库操作上下文
- 签名：`WithContext(ctx context.Context) *DB`
- 示例：
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

db.WithContext(ctx)
```

### GetContext
- 获取当前数据库操作上下文
- 签名：`GetContext() context.Context`
- 示例：
```go
currentCtx := db.GetContext()
```

## 事务处理方法

### Begin
- 手动开启事务
- 签名：`Begin() (*Transaction, error)`
- 示例：
```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// 执行事务操作
err = tx.Commit()
```

### ExecTx
- 自动管理事务
- 签名：`ExecTx(fn func(*Transaction) error) error`
- 示例：
```go
err := db.ExecTx(func(tx *Transaction) error {
    // 执行事务操作
    _, err := tx.M("users").Insert(user)
    return err
})
```

## 缓存管理方法

### WithCache
- 使用缓存执行查询
- 签名：`WithCache(cache Cache, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error)`
- 示例：
```go
result, err := db.WithCache(redisCache, "user_key", 1*time.Hour, func() (interface{}, error) {
    // 查询逻辑
    return user, nil
})
```

### InvalidateCache
- 使缓存失效
- 签名：`InvalidateCache(cache Cache, keys ...string) error`
- 示例：
```go
err := db.InvalidateCache(redisCache, "user_key1", "user_key2")
```

## 查询和执行方法

### Query
- 执行查询并返回结果集
- 签名：`Query(query string, args ...interface{}) (*sql.Rows, error)`
- 示例：
```go
rows, err := db.Query("SELECT * FROM users WHERE age > ?", 18)
defer rows.Close()
```

### QueryWithContext
- 带上下文的查询方法
- 签名：`QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)`
- 示例：
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.QueryWithContext(ctx, "SELECT * FROM users WHERE status = ?", "active")
```

### Exec
- 执行更新操作
- 签名：`Exec(query string, args ...interface{}) (sql.Result, error)`
- 示例：
```go
result, err := db.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
```

## 日志和调试方法

### SetDebug
- 设置调试模式
- 签名：`SetDebug(bool) *DB`
- 示例：
```go
db.SetDebug(true)  // 开启调试模式
```

### SetLogLevel
- 动态调整日志级别
- 签名：`SetLogLevel(level string) error`
- 可用日志级别：
  - `"debug"`: 最详细的日志级别，记录所有调试信息
  - `"info"`: 常规信息，追踪系统运行状态
  - `"warn"`: 警告信息，记录潜在风险
  - `"error"`: 仅记录错误信息
- 示例：
```go
err := db.SetLogLevel("debug")  // 设置为调试模式
err = db.SetLogLevel("info")    // 设置为常规信息模式
err = db.SetLogLevel("warn")    // 设置为警告模式
err = db.SetLogLevel("error")   // 仅记录错误
```

## 性能监控方法

### DBMetrics
- 获取性能指标
- 签名：`DBMetrics() *dbMetrics`
- 示例：
```go
metrics := db.DBMetrics()
```

### GetPoolStats
- 获取连接池统计
- 签名：`GetPoolStats() *sql.DBStats`
- 示例：
```go
poolStats := db.GetPoolStats()
```

## 连接管理方法

### Ping
- 测试数据库连接
- 签名：`Ping(ctx context.Context) error`
- 示例：
```go
err := db.Ping(context.Background())
```

### Close
- 关闭数据库连接
- 签名：`Close() error`
- 示例：
```go
err := db.Close()
```

## 其他实用方法

### GetVersion
- 获取框架版本
- 签名：`GetVersion() string`
- 示例：
```go
version := db.GetVersion()
```

### GetDBName
- 获取数据库名称
- 签名：`GetDBName() string`
- 示例：
```go
dbName := db.GetDBName()
```

## 注意事项

- 大多数方法支持方法链
- 内置连接池管理和性能监控
- 支持动态日志级别调整
- 提供丰富的上下文和超时控制
- 性能和安全性是设计重点

## 性能建议

- 合理使用连接池
- 启用性能指标收集
- 设置适当的超时时间
- 在高并发场景下注意连接数控制
