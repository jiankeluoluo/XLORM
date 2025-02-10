#### XLORM - Go 语言轻量级数据库 ORM 框架

[中文版](README_ZH.md "访问中文版")

[English](README.md "Access English Version")

## 简介

XLORM 是一个专为 Go 语言设计的高性能、易用性强的轻量级 ORM 框架，支持 MySQL 数据库。它提供了完整的 CRUD 操作、事务管理、批量处理、查询构建器、缓存支持等功能，并内置日志记录和性能监控模块，帮助开发者快速构建稳健的数据库应用。

**XLORM 由Windsurf AI功能命令式生成，并由DeepSeek R1进行代码审核、评估、提出性能优化建议。**

## 主要特性

- **完整的 CRUD 支持**：简化数据操作，支持结构体、Map 等多种数据类型
- **事务管理**：提供原子性操作和嵌套事务支持
- **批量操作优化**：高效处理大规模数据插入/更新
- **查询构建器**：链式调用生成安全 SQL，防止注入
- **缓存支持**：内置分片锁缓存，提升高频查询性能
- **日志与监控**：异步日志记录、慢查询报警、连接池统计
- **连接池管理**：支持连接探活、超时控制、连接复用

### 1. 高性能设计
- 异步性能指标收集
- 连接池管理
- 批量操作优化
- 缓存机制（结构体字段缓存、占位符缓存）

### 2. 灵活的数据库操作
- 支持多种数据库操作（增删改查）
- 事务处理
- 动态查询构建
- 字段映射和转换

### 3. 日志系统
- 结构化日志记录
- 日志轮转
- 可配置的日志级别
- 异步日志处理

### 4. 安全性
- SQL 注入防护
- 字段转义
- 安全的参数绑定
- 连接超时控制

### 5. 调试支持
- 详细的调试模式
- 性能指标追踪
- 慢查询监控
- 详细的系统运行状态报告

### 6. 动态配置
- 运行时修改部分配置
- 热更新连接池参数
- 灵活调整日志级别



## 与 GORM/XORM 的性能对比

### 1. GORM
#### 优势
- 功能丰富，支持模型关联、钩子、迁移等高级特性。
- 社区活跃，文档完善，插件生态丰富（如分页、软删除）。

#### 劣势
- 反射使用较多，性能相对较低，尤其在复杂查询或高频操作中。
- 日志和指标功能较弱，需依赖外部库（如 Prometheus）。

### 2. XORM
#### 优势
- 性能优于 GORM，支持读写分离和缓存优化。
- 提供更灵活的条件构建器（builder 模块）。

#### 劣势
- 异步处理和批量操作的优化不如 XLORM 深入。
- 日志和监控功能需手动集成。

### 3. XLORM
#### 性能亮点
- **批量操作**：通过事务分批提交，减少网络往返和锁竞争。
- **缓存设计**：分片 LRU 缓存提升高频查询效率。
- **对象池**：减少 GC 压力，适合高并发场景。

#### 适用场景
- 需要高性能批量写入（如日志入库、数据同步）。
- 对自定义日志和监控有较高要求的项目。

### 总结

| 维度     | GORM                | XORM                | XLORM               |
|----------|---------------------|---------------------|---------------------|
| 功能     | 丰富（关联、迁移）  | 中等（读写分离）    | 基础（CRUD/批量）   |
| 性能     | 较低                | 中等                | 较高（优化批量/缓存）|
| 安全性   | 中等                | 中等                | 高（注入防护）      |
| 易用性   | 高（文档完善）      | 中等                | 高（文档完善）        |
| 适用场景 | 复杂业务逻辑        | 常规 OLTP           | 高频批量写入        |

### 建议
- 若需要高级功能（如关联、迁移），选 **GORM**。
- 若需平衡性能与功能，选 **XORM**。
- 若追求极致性能且场景简单（如数据同步），可尝试 **XLORM**。

#### 基准测试对比报告、模块说明等请阅读md目录下的文件。

## 安装

```bash
go get github.com/go-sql-driver/mysql
go get github.com/jiankeluoluo/xlorm
```

## 快速开始

### 1. 数据库连接

```go
import "your_project/db"

// 配置数据库连接
config := &db.Config{
    DBName:            "master",               //数据库别名称、用于区分不同数据库
    Driver:            "MySQL",                 // 数据库驱动类型，目前仅支持 "MySQL"
    Host:              "localhost",             // 数据库服务器地址，支持 IP 或域名
    Port:              3306,                    // 数据库服务器端口号，MySQL 默认为 3306
    Username:          "your_username",         // 数据库登录用户名
    Password:          "your_password",         // 数据库登录密码
    Database:          "your_database",         // 要连接的具体数据库名称
    Charset:           "utf8mb4",               // 数据库字符集，推荐使用 utf8mb4 支持完整 Unicode
    TablePrefix:       "tb_",                   // 表名前缀，用于多项目或模块共享数据库时的命名空间
    LogDir:            "./logs",                // 日志文件存储目录，支持相对和绝对路径
    LogLevel:          "debug",                 // 日志级别，可选 debug/info/warn/error，推荐开发阶段使用 debug
    LogBufferSize:     5000,                   // 日志缓冲区大小，默认 5000
    LogRotationEnabled: true,  // 启用日志轮转
    LogRotationMaxAge:  30,    // 日志保留30天

    // 连接生命周期配置
    ConnMaxLifetime:   30 * time.Minute,        // 连接的最大生存时间，超过则重新创建
    ConnMaxIdleTime:   10 * time.Minute,        // 空闲连接的最大保持时间

    // 超时配置
    ConnTimeout:       5 * time.Second,         // 建立数据库连接的超时时间
    ReadTimeout:       3 * time.Second,         // 读取数据的超时时间
    WriteTimeout:      3 * time.Second,         // 写入数据的超时时间
    SlowQueryTime:     200 * time.Millisecond,  // 慢查询阈值，超过此时间的查询将被记录

    // 连接池配置
    PoolStatsInterval: 1 * time.Minute,         // 连接池统计信息收集间隔
    MaxOpenConns:      100,                     // 最大打开连接数，控制数据库的最大并发连接
    MaxIdleConns:      20,                      // 最大空闲连接数，减少频繁创建和销毁连接的开销

    // 调试和监控
    EnablePoolStats:     true,                    // 是否启用性能指标收集
    Debug:             true,                    // 是否开启调试模式，会输出更详细的日志信息
}

// 创建数据库连接
xdb, err := db.New(config)
if err != nil {
    log.Fatal(err)
}
defer xdb.Close()
```

## 配置选项详解

XLORM 提供了丰富且灵活的配置选项，每个配置都经过精心设计，以满足不同场景的需求。

#### 数据库连接配置

| 配置项 | 类型 | 默认值 | 描述 | 推荐设置 |
|--------|------|--------|------|----------|
| `DBName` | `string` | `""` | 数据库别名，用于区分不同数据库实例 | 建议设置有意义的名称 |
| `Driver` | `string` | `"MySQL"` | 数据库驱动类型 | 目前仅支持 MySQL |
| `Host` | `string` | `"localhost"` | 数据库服务器地址 | 生产环境使用实际服务器地址 |
| `Port` | `int` | `3306` | 数据库服务器端口 | MySQL 默认 3306 |
| `Username` | `string` | `""` | 数据库登录用户名 | 使用最小权限用户 |
| `Password` | `string` | `""` | 数据库登录密码 | 建议使用环境变量 |
| `Database` | `string` | `""` | 要连接的具体数据库名 | 必填 |

#### 连接池配置

| 配置项 | 类型 | 默认值 | 描述 | 性能建议 |
|--------|------|--------|------|----------|
| `MaxOpenConns` | `int` | `100` | 最大打开连接数 | 根据业务并发量调整，避免超载 |
| `MaxIdleConns` | `int` | `20` | 最大空闲连接数 | 平衡内存使用和连接复用 |
| `ConnMaxLifetime` | `time.Duration` | `30 * time.Minute` | 连接最大生存时间 | 防止长连接资源泄露 |
| `ConnMaxIdleTime` | `time.Duration` | `10 * time.Minute` | 空闲连接最大保持时间 | 控制连接回收频率 |

#### 超时与安全配置

| 配置项 | 类型 | 默认值 | 描述 | 安全建议 |
|--------|------|--------|------|----------|
| `ConnTimeout` | `time.Duration` | `5 * time.Second` | 建立连接超时时间 | 根据网络情况调整 |
| `ReadTimeout` | `time.Duration` | `3 * time.Second` | 读取数据超时 | 避免长时间阻塞 |
| `WriteTimeout` | `time.Duration` | `3 * time.Second` | 写入数据超时 | 防止写入操作挂起 |
| `SlowQueryTime` | `time.Duration` | `200 * time.Millisecond` | 慢查询阈值 | 合理设置，及时发现性能问题 |

#### 日志与调试配置

| 配置项 | 类型 | 默认值 | 描述 | 使用建议 |
|--------|------|--------|------|----------|
| `LogLevel` | `string` | `"warn"` | 日志级别 | 生产环境建议 `warn`/`error` |
| `Debug` | `bool` | `false` | 调试模式 | 开发阶段开启，生产关闭 |
| `LogDir` | `string` | `"./logs"` | 日志存储目录 | 确保目录可写 |
| `LogBufferSize` | `int` | `5000` | 日志缓冲区大小 | 根据日志量调整 |
| `EnablePoolStats` | `bool` | `false` | 是否启用连接池统计 | 性能监控时开启 |

#### 高级配置注意事项

1. **安全性**：
   - 避免在代码中硬编码敏感信息
   - 使用环境变量管理数据库凭证
   - 限制连接池大小防止资源耗尽

2. **性能优化**：
   - 根据实际业务场景调整 `MaxOpenConns` 和 `MaxIdleConns`
   - 合理设置超时时间，防止资源浪费
   - 使用 `SlowQueryTime` 追踪性能瓶颈

3. **日志管理**：
   - 生产环境建议使用 `warn` 或 `error` 级别
   - 定期清理和备份日志文件
   - 监控日志缓冲区使用情况

### 配置示例

```go
config := &db.Config{
    DBName:            "master",
    Driver:            "MySQL",
    Host:              os.Getenv("DB_HOST"),
    Port:              3306,
    Username:          os.Getenv("DB_USERNAME"),
    Password:          os.Getenv("DB_PASSWORD"),
    Database:          "myapp",
    
    // 连接池配置
    MaxOpenConns:      50,
    MaxIdleConns:      10,
    ConnMaxLifetime:   30 * time.Minute,
    
    // 超时配置
    ConnTimeout:       5 * time.Second,
    ReadTimeout:       3 * time.Second,
    WriteTimeout:      3 * time.Second,
    SlowQueryTime:     200 * time.Millisecond,
    
    // 日志配置
    LogLevel:          "warn",
    LogDir:            "./logs",
    Debug:             false,
    EnablePoolStats:   true,
}
```

**提示**：始终根据具体业务场景和性能需求调整配置参数。

## 基本查询操作

#### 2.1 查询单条记录

```go
// 查询单条记录
result, err := xdb.M("users").
    Where("id = ?", 1).
    Fields("id, name, age").
    Find()

users := db.M("users").
    NotWhere("status = ?", "banned").
    Find()
// SQL: SELECT * FROM users WHERE NOT (status = 'banned')

users := db.M("users").
    Where("age > ?", 18).
    NotWhere("status = ?", "banned").
    OrWhere("vip = ?", true).
    Find()
// SQL: SELECT * FROM users WHERE (age > 18 AND NOT (status = 'banned')) OR vip = true

products := db.M("products").
    Where("category = ?", "electronics").
    NotWhere("price < ?", 100).
    OrWhere("discount > ?", 0.5).
    Find()
// SQL: SELECT * FROM products WHERE (category = 'electronics' AND NOT (price < 100)) OR discount > 0.5
```

#### 2.2 查询多条记录

```go
// 查询多条记录，支持复杂查询条件
results, err := xdb.M("users").
    Where("age > ?", 18).
    OrderBy("age DESC").
    Limit(10).
    FindAll()

users := db.M("users").
    NotWhere("status = ?", "banned").
    FindAll()
// SQL: SELECT * FROM users WHERE NOT (status = 'banned')

users := db.M("users").
    Where("age > ?", 18).
    NotWhere("status = ?", "banned").
    OrWhere("vip = ?", true).
    FindAll()
// SQL: SELECT * FROM users WHERE (age > 18 AND NOT (status = 'banned')) OR vip = true

products := db.M("products").
    Where("category = ?", "electronics").
    NotWhere("price < ?", 100).
    OrWhere("discount > ?", 0.5).
    FindAll()
// SQL: SELECT * FROM products WHERE (category = 'electronics' AND NOT (price < 100)) OR discount > 0.5
```

#### 2.3 分页查询

```go
// 分页查询，获取第2页，每页10条记录
results, err := xdb.M("users").
    Where("status = ?", 1).
    Page(2, 10).
    FindAll()

// 获取总记录数
total, _ := xdb.M("users").
    Where("status = ?", 1).
    HasTotal(true). // 获取总记录数
    FindAll()
total_count := xdb.M("users").GetTotal()
```

#### 2.4 复杂查询

```go
// 复杂查询：联表、分组、过滤
results, err := xdb.M("orders").
    Fields("u.name, SUM(o.total_amount) as total_sales").
    Join("LEFT JOIN users u ON u.id = o.user_id").
    Where("o.status = ?", "completed").
    GroupBy("u.name").
    Having("total_sales > ?", 1000).
    FindAll()
```

### 3. 数据操作

#### 3.1 插入单条记录

```go
// 插入单条记录
user := map[string]interface{}{
    "name": "John Doe",
    "age":  30,
    "status": 1,
}
lastInsertId, err := xdb.M("users").Insert(user)
```

#### 3.2 批量插入

```go
// 批量插入多条记录
users := []map[string]interface{}{
    {"name": "Alice", "age": 25},
    {"name": "Bob", "age": 30},
    {"name": "Charlie", "age": 35},
}
rowsAffected, err := xdb.M("users").BatchInsert(users,500)
```

#### 3.3 更新记录

```go
// 更新单条记录
updateData := map[string]interface{}{
    "name": "John Smith",
    "age":  31,
}
rowsAffected, err := xdb.M("users").
    Where("id = ?", 1).
    Update(updateData)

// 批量更新
updateBatch := []map[string]interface{}{
    {"id": 1, "status": 2},
    {"id": 2, "status": 2},
}
rowsAffected, err := xdb.M("users").BatchUpdate(updateBatch, "id")
```

#### 3.4 删除记录

```go
// 删除单条记录
rowsAffected, err := xdb.M("users").
    Where("id = ?", 1).
    Delete()

// 批量删除
deleteIds := []int{1, 2, 3}
rowsAffected, err := xdb.M("users").
    Where("id IN (?)", deleteIds).
    Delete()
```

### 4. 事务处理

```go
err := xdb.ExecTx(func(tx *db.Transaction) error {
    // 转账示例：从账户A扣款，向账户B转账
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err
    }

    _, err = tx.Exec("UPDATE accounts SET balance = balance + 100 WHERE id = ?", 2)
    return err
})

if err != nil {
    // 事务失败，已自动回滚
    log.Fatal(err)
}
```

### 5. 缓存支持

```go
// 使用缓存执行查询
result, err := xdb.WithCache(redisCache, "user_key_1", 1*time.Hour, func() (interface{}, error) {
    return xdb.M("users").Where("id = ?", 1).Find()
})

// 手动使缓存失效
xdb.InvalidateCache(redisCache, "user_key_1")
```

### 6. 性能指标监控

```go
// 配置性能指标
config.EnablePoolStats = true
config.PoolStatsInterval = 30 * time.Second
```

## SQL 查询构建器

XLORM 提供了强大的 SQL 查询构建器，支持灵活且安全的查询构建。

### 基本使用

```go
// 使用 Builder 构建复杂查询
query := xdb.NewBuilder("users").
    Fields("id", "name", "age").
    Where("age > ?", 18).
    OrderBy("age DESC").
    Limit(10).
    Build()

// 执行查询
rows, err := xdb.Query(query)
```

### 高级查询构建

```go
// 复杂查询：联表、分组、过滤、排序
builder := xdb.NewBuilder("orders").
    Fields("u.name", "SUM(o.total) as total_sales").
    Join("LEFT JOIN users u ON u.id = o.user_id").
    Where("o.status = ?", "completed").
    GroupBy("u.name").
    Having("total_sales > ?", 1000).
    OrderBy("total_sales DESC").
    Limit(5)
query, args := builder.Build()
defer builder.ReleaseBuilder()
```

### 查询构建器特性

- 支持字段选择
- 动态 WHERE 条件
- 表连接（JOIN）
- 分组（GROUP BY）
- 分组过滤（HAVING）
- 排序（ORDER BY）
- 分页和限制
- 行锁支持

## 日志系统

XLORM 提供了高性能、异步的结构化日志系统，基于 Go 标准库 `log/slog`，支持灵活的日志处理。默认将日志写入 `./logs/db.log`。日志包含详细的操作信息和错误追踪。

### 日志特性

- 异步日志处理
- 结构化日志记录
- 高性能、低延迟
- 可配置的日志缓冲区
- 优雅的日志关闭机制
- 丢失日志追踪

### 日志轮转功能

XLORM 支持按天自动分割日志文件，并可配置日志保留时间：

- `LogRotationEnabled`: 是否启用日志轮转功能
- `LogRotationMaxAge`: 日志保留天数，默认为30天

特点：
- 自动按天创建日志文件
- 文件名格式：`db_2024-02-05.log`
- 可配置日志保留时间
- 自动清理过期日志文件

### 基本日志使用

```go

config := &db.Config{
    LogDir: "./logs",  // 自定义日志目录
    LogLevel: "debug",        // 设置日志级别
    LogBufferSize: 5000,      // 日志缓冲区数量（默认5000）
    LogRotationEnabled: true,  // 启用日志轮转
    LogRotationMaxAge:  30,    // 日志保留30天
    Debug:  true,             // 启用调试日志
}
// 获取日志实例
logger := xdb.Logger()

// 记录不同级别的日志
logger.Info("数据库操作",
    slog.String("operation", "query"),
    slog.Int("user_id", 123)
)

logger.Error("查询失败",
    slog.String("error", err.Error()),
    slog.String("query", sqlQuery)
)

// 使用LogAttrs提升性能
logger.LogAttrs(context.Background(), slog.LevelDebug,
    "SQL详情",
    slog.String("query", query),
    slog.Any("args", args),
)

//运行时动态调整
// 临时开启调试日志
xdb.SetLogLevel("debug")

// 恢复生产级别
xdb.SetLogLevel("warn")

```

### 异步日志处理

```go
// 获取日志指标
logMetrics := xdb.AsyncLogger().GetLogMetrics()
// 打印日志指标
fmt.Printf("日志指标:\n")
fmt.Printf("总日志数: %d\n", logMetrics["total_logs"])
fmt.Printf("丢弃的日志数: %d\n", logMetrics["dropped_logs"])
fmt.Printf("日志通道深度: %d\n", logMetrics["channel_depth"])
// metrics 包含：
// - total_logs: 总处理日志数
// - dropped_logs: 丢弃的日志数
// - channel_depth: 当前日志通道深度
```

### 日志性能监控

```go
asyncLogger := xdb.AsyncLogger()
// 获取总处理日志数量
totalLogsCount := asyncLogger.GetTotalLogsCount()
// 获取丢弃的日志数量(异步阻塞处理不及时被丢弃的日志)
droppedLogsCount := asyncLogger.GetDroppedLogsCount()

fmt.Printf("总日志数: %d\n", totalLogsCount)
fmt.Printf("丢弃的日志数: %d\n", droppedLogsCount)
```

### 日志系统高级特性

1. 非阻塞日志记录
2. 自动处理日志通道溢出
3. 可配置的日志缓冲区大小
4. 支持上下文和属性扩展
5. 内置错误收集机制

### 日志级别

XLORM 支持标准的日志级别：

- `Debug`: 调试信息
- `Info`: 普通信息
- `Warn`: 警告信息
- `Error`: 错误信息

### 性能建议

- 合理设置日志缓冲区大小
- 避免在高频调用中记录过多日志
- 使用结构化日志提高可读性
- 定期监控丢弃的日志数量

### 日志配置最佳实践

1. 生产环境建议：
   - 将 [LogLevel]设置为 `info` 或 `warn`
   - 适当配置 `LogRotationMaxAge`，平衡存储空间和日志保留需求
   - 确保日志目录有足够的写入权限

2. 调试场景：
   - 使用 `debug` 级别获取详细日志
   - 临时增大 `LogBufferSize`
   - 关注日志文件大小和数量

3. 性能考虑：
   - 日志缓冲区大小 `LogBufferSize` 影响日志记录性能
   - 过大可能增加内存消耗
   - 过小可能丢失日志

### 注意事项

- 日志轮转依赖系统时间，请确保系统时间正确
- 日志文件名包含日期，便于追踪和管理
- 超过保留天数的日志文件将被自动删除
- 建议定期备份重要日志

## 性能指标监控

XLORM 提供了详细的性能指标监控机制，帮助开发者了解数据库操作性能。

### 连接池统计
- `EnablePoolStats`: 开启连接池性能指标
- 可监控连接使用情况
- 实时跟踪连接数、空闲连接等

### 慢查询追踪
- `SlowQueryTime`: 设置慢查询阈值
- 自动记录超过阈值的查询
- 提供详细的查询耗时信息

### 性能指标统计

```go
// 获取性能指标统计
stats := xdb.DBMetrics()
metrics := stats.GetDBMetrics()

// 打印性能指标
// 打印基本指标
fmt.Printf("数据库性能指标:\n")
fmt.Printf("数据库名称: %s\n", metrics["db_name"])
fmt.Printf("总查询数: %d\n", metrics["total_queries"])
fmt.Printf("慢查询数: %d\n", metrics["slow_queries"])
fmt.Printf("查询错误数: %d\n", metrics["total_errors"])
fmt.Printf("影响的总行数: %d\n", metrics["total_affected_rows"])

// 打印查询耗时统计
queryStats := metrics["query_stats"].(map[string]interface{})
for queryType, stat := range queryStats {
    statMap := stat.(map[string]interface{})
    fmt.Printf("查询类型 %s 的耗时统计:\n", queryType)
    fmt.Printf("查询次数: %d\n", statMap["count"])
    fmt.Printf("总耗时: %v\n", statMap["total_time"])
    fmt.Printf("平均耗时: %v\n", statMap["average_time"])
}
```

### 重置性能指标

```go
// 重置所有性能指标
stats.ResetDBMetrics()
```

### 动态开启/关闭性能统计

```go
// 动态开启/关闭性能统计
stats.SetDBMetricsEnable(true)
stats.SetDBMetricsEnable(false)
```

### 异步性能指标

```go
// 异步性能指标，减少对主业务的性能影响
asyncMetrics := xdb.AsyncMetrics()

// 获取丢弃的指标数量（如果指标通道已满）
droppedMetricsCount := asyncMetrics.GetDroppedMetricsCount()
```

### 连接池指标统计

```go
// 获取连接池统计
stats := xdb.GetPoolStats()

// 打印连接池信息
fmt.Printf("连接池状态:\n")
fmt.Printf("最大连接数: %d\n", stats.MaxOpenConnections)
fmt.Printf("当前打开连接数: %d\n", stats.OpenConnections)
fmt.Printf("正在使用的连接数: %d\n", stats.InUse)
fmt.Printf("空闲连接数: %d\n", stats.Idle)
fmt.Printf("等待连接次数: %d\n", stats.WaitCount)
fmt.Printf("等待连接总时间: %v\n", stats.WaitDuration)
fmt.Printf("因超过最大空闲时间关闭的连接数: %d\n", stats.MaxIdleClosed)
fmt.Printf("因超过最大生命周期关闭的连接数: %d\n", stats.MaxLifetimeClosed)
```

## 安全性建议

### 配置安全
- 避免硬编码敏感信息
- 使用环境变量管理数据库凭证
- 限制连接池大小防止资源耗尽

### 数据脱敏
- 日志中自动脱敏敏感信息
- 支持自定义脱敏规则

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

[MIT License]
