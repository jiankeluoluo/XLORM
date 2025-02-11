# XLORM Core Methods Documentation

[中文版](Xlorm_methods_zh.md "访问中文版")

[English](Xlorm_methods_en.md "Access English Version")

## Overview

`xlorm.go` is the core file of the XLORM framework, providing basic functionalities such as database connection, transaction management, log control, and performance monitoring. This file defines the `DB` struct, which is the main entry point of the entire framework.

### Key Features

- **Database Connection Management**: Provides flexible database connection configuration and initialization
- **Transaction Processing**: Supports manual and automatic transaction management
- **Logging System**: Offers dynamic log level adjustment and debugging capabilities
- **Performance Monitoring**: Built-in connection pool statistics and query performance metrics collection
- **Context Support**: Integrates Go language context mechanism

### Use Cases

- Creating database connections
- Managing database transactions
- Configuring logs and debugging
- Monitoring database performance
- Managing database connection lifecycle

## Database Connection Methods

### New
- Create a new database connection
- Signature: `New(cfg *Config) (*DB, error)`
- Example:
```go
// Create MySQL database connection
config := &xlorm.Config{
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

### M and Table
- Get table operation object
- Signature: `M(tableName string) *Table`
- Signature: `Table(tableName string) *Table`
- Example:
```go
// Get users table operation object
userTable := db.M("users")
// Or
userTable := db.Table("users")
```

## Context Management Methods

### WithContext
- Set database operation context
- Signature: `WithContext(ctx context.Context) *DB`
- Example:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

db.WithContext(ctx)
```

### GetContext
- Get current database operation context
- Signature: `GetContext() context.Context`
- Example:
```go
currentCtx := db.GetContext()
```

## Transaction Processing Methods

### Begin
- Manually start a transaction
- Signature: `Begin() (*Transaction, error)`
- Example:
```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// Perform transaction operations
err = tx.Commit()
```

### ExecTx
- Automatically manage transactions
- Signature: `ExecTx(fn func(*Transaction) error) error`
- Example:
```go
err := db.ExecTx(func(tx *Transaction) error {
    // Perform transaction operations
    _, err := tx.M("users").Insert(user)
    return err
})
```

## Cache Management Methods

### WithCache
- Execute query with cache
- Signature: `WithCache(cache Cache, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error)`
- Example:
```go
result, err := db.WithCache(redisCache, "user_key", 1*time.Hour, func() (interface{}, error) {
    // Query logic
    return user, nil
})
```

### InvalidateCache
- Invalidate cache
- Signature: `InvalidateCache(cache Cache, keys ...string) error`
- Example:
```go
err := db.InvalidateCache(redisCache, "user_key1", "user_key2")
```

## Query and Execution Methods

### Query
- Execute query and return result set
- Signature: `Query(query string, args ...interface{}) (*sql.Rows, error)`
- Example:
```go
rows, err := db.Query("SELECT * FROM users WHERE age > ?", 18)
defer rows.Close()
```

### QueryWithContext
- Query method with context
- Signature: `QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)`
- Example:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.QueryWithContext(ctx, "SELECT * FROM users WHERE status = ?", "active")
```

### Exec
- Execute update operation
- Signature: `Exec(query string, args ...interface{}) (sql.Result, error)`
- Example:
```go
result, err := db.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
```

## Logging and Debugging Methods

### SetDebug
- Set debug mode
- Signature: `SetDebug(bool) *DB`
- Example:
```go
db.SetDebug(true)  // Enable debug mode
```

### SetLogLevel
- Dynamically adjust log level
- Signature: `SetLogLevel(level string) error`
- Available log levels:
  - `"debug"`: Most detailed log level, records all debugging information
  - `"info"`: Regular information, tracking system runtime status
  - `"warn"`: Warning information, recording potential risks
  - `"error"`: Only records error information
- Example:
```go
err := db.SetLogLevel("debug")  // Set to debug mode
err = db.SetLogLevel("info")    // Set to regular information mode
err = db.SetLogLevel("warn")    // Set to warning mode
err = db.SetLogLevel("error")   // Only record errors
```

## Performance Monitoring Methods

### DBMetrics
- Get performance metrics
- Signature: `DBMetrics() *dbMetrics`
- Example:
```go
metrics := db.DBMetrics()
```

### GetPoolStats
- Get connection pool statistics
- Signature: `GetPoolStats() *sql.DBStats`
- Example:
```go
poolStats := db.GetPoolStats()
```

## Connection Management Methods

### Ping
- Test database connection
- Signature: `Ping(ctx context.Context) error`
- Example:
```go
err := db.Ping(context.Background())
```

### Close
- Close database connection
- Signature: `Close() error`
- Example:
```go
err := db.Close()
```

## Other Utility Methods

### GetVersion
- Get framework version
- Signature: `GetVersion() string`
- Example:
```go
version := db.GetVersion()
```

### GetDBName
- Get database name
- Signature: `GetDBName() string`
- Example:
```go
dbName := db.GetDBName()
```

## Precautions

- Most methods support method chaining
- Built-in connection pool management and performance monitoring
- Supports dynamic log level adjustment
- Provides rich context and timeout control
- Performance and security are design priorities

## Performance Recommendations

- Use connection pool reasonably
- Enable performance metrics collection
- Set appropriate timeout times
- Pay attention to connection count control in high concurrency scenarios
