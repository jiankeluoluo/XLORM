# config.go 文档

[中文版](Config_zh.md "访问中文版")

[English](Config_en.md "Access English Version")

## 概述
`config.go` 定义了 xlorm 数据库连接和配置的核心结构体，提供了灵活且安全的数据库配置方案。

## 主要结构体

### Config 数据库配置

#### 字段详解

##### 连接配置
- `DBName`: 数据库别名，用于区分不同数据库
- `Driver`: 数据库驱动（如 mysql）
- `Host`: 数据库主机地址
- `Username`: 数据库用户名
- `Password`: 数据库密码
- `Database`: 数据库名称
- `Port`: 数据库端口号

##### 连接参数
- `Charset`: 字符集（默认：utf8mb4）
- `TablePrefix`: 表名前缀
- `ConnMaxLifetime`: 连接最大生命周期
- `ConnMaxIdleTime`: 连接最大空闲时间
- `ConnTimeout`: 连接超时时间
- `ReadTimeout`: 读取超时时间
- `WriteTimeout`: 写入超时时间
- `SlowQueryTime`: 慢查询阈值

##### 连接池配置
- `MaxOpenConns`: 最大打开连接数（默认：0）
- `MaxIdleConns`: 最大空闲连接数（默认：0）
- `EnablePoolStats`: 是否启用连接池性能指标（默认：false）
- `PoolStatsInterval`: 连接池统计频率

##### 日志配置
- `LogDir`: 日志目录
- `LogLevel`: 日志级别（debug|info|warn|error）
- `LogBufferSize`: 日志缓冲区大小（默认：5000）
- `LogRotationEnabled`: 是否启用日志轮转
- `LogRotationMaxAge`: 日志保留天数（默认：30）

##### 调试配置
- `Debug`: 是否开启调试模式（默认：false）
- `DBMetricsBufferSize`: 异步指标缓冲区大小（默认：1000）

## 主要方法

### Validate() error
验证配置的有效性

#### 验证规则
1. 配置不能为空
2. 主机地址必填
3. 端口号有效性检查（1-65535）
4. 用户名必填
5. 数据库名称必填
6. 日志级别有效性检查

#### 示例
```go
config := &xlorm.Config{
    Host:     "localhost",
    Port:     3306,
    Username: "root",
    Database: "mydb",
}
err := config.Validate()
```

## 使用建议

1. 使用环境变量或配置文件管理敏感信息
2. 合理设置连接池参数
3. 启用调试模式进行性能分析
4. 配置适当的日志级别和轮转策略

## 最佳实践

### 配置示例

```go
config := &xlorm.Config{
    Driver:            "mysql",
    Host:             "localhost",
    Port:             3306,
    Username:         "username",
    Password:         "password",
    Database:         "mydb",
    Charset:          "utf8mb4",
    TablePrefix:      "app_",
    MaxOpenConns:     50,
    MaxIdleConns:     10,
    ConnMaxLifetime:  time.Hour,
    SlowQueryTime:    time.Second * 2,
    LogLevel:         "info",
    Debug:            true,
}
```

## 注意事项

- 不要在代码中硬编码敏感信息
- 定期轮换数据库密码
- 根据实际负载调整连接池参数
- 监控数据库连接状态

## XLORM 配置文档

### 概述

`Config` 结构体是 XLORM 框架的核心配置对象，用于定义数据库连接和系统行为的详细参数。通过精细配置，可以优化数据库性能、日志管理和连接行为。

### 配置字段详解

#### 基本连接配置

| 字段名 | 类型 | 描述 | 默认值 |
|--------|------|------|--------|
| `DBName` | `string` | 数据库别名，用于区分不同数据库 | 无 |
| `Driver` | `string` | 数据库驱动类型 | 无 |
| `Host` | `string` | 数据库主机地址 | 必填 |
| `Username` | `string` | 数据库用户名 | 必填 |
| `Password` | `string` | 数据库密码 | 必填 |
| `Database` | `string` | 数据库名称 | 必填 |
| `Port` | `int` | 数据库端口号 | 必填 |

#### 连接增强配置

| 字段名 | 类型 | 描述 | 默认值 |
|--------|------|------|--------|
| `Charset` | `string` | 数据库字符集 | `"utf8mb4"` |
| `TablePrefix` | `string` | 表名前缀 | 无 |
| `ConnMaxLifetime` | `time.Duration` | 连接最大生命周期 | 无 |
| `ConnMaxIdleTime` | `time.Duration` | 连接最大空闲时间 | 无 |
| `ConnTimeout` | `time.Duration` | 连接超时时间 | 无 |
| `ReadTimeout` | `time.Duration` | 读取超时时间 | 无 |
| `WriteTimeout` | `time.Duration` | 写入超时时间 | 无 |

#### 日志配置

| 字段名 | 类型 | 描述 | 默认值 |
|--------|------|------|--------|
| `LogDir` | `string` | 日志存储目录 | 无 |
| `LogLevel` | `string` | 日志级别（debug/info/warn/error） | `"debug"` |
| `LogBufferSize` | `int` | 日志缓冲区大小 | `5000` |
| `LogRotationEnabled` | `bool` | 是否启用日志轮转 | `false` |
| `LogRotationMaxAge` | `int` | 日志保留天数 | `30` |

#### 性能和调试配置

| 字段名 | 类型 | 描述 | 默认值 |
|--------|------|------|--------|
| `MaxOpenConns` | `int` | 最大打开连接数 | `0`（不限制） |
| `MaxIdleConns` | `int` | 最大空闲连接数 | `0`（不限制） |
| `SlowQueryTime` | `time.Duration` | 慢查询阈值 | 无 |
| `PoolStatsInterval` | `time.Duration` | 连接池统计频率 | 无 |
| `DBMetricsBufferSize` | `int` | 异步指标缓冲区大小 | `1000` |
| `EnablePoolStats` | `bool` | 是否启用性能指标 | `false` |
| `Debug` | `bool` | 是否开启调试模式 | `false` |

### 配置示例

```go
config := &xlorm.Config{
    DBName:            "mydb",
    Driver:            "mysql",
    Host:              "localhost",
    Username:          "root",
    Password:          "password",
    Database:          "myapp",
    Port:              3306,
    Charset:           "utf8mb4",
    LogLevel:          "info",
    MaxOpenConns:      100,
    MaxIdleConns:      50,
    ConnMaxLifetime:   30 * time.Minute,
    LogRotationEnabled: true,
    Debug:             true,
}
```

### 配置验证

`Config` 结构体提供了 `Validate()` 方法，用于验证配置的有效性。验证规则包括：

1. 配置不能为空
2. 主机地址不能为空
3. 端口号必须在有效范围内（1-65535）
4. 用户名不能为空
5. 数据库名称不能为空
6. 日志级别必须有效

### 最佳实践

1. 始终使用 `Validate()` 方法检查配置
2. 根据业务需求合理设置连接池参数
3. 在生产环境中禁用调试模式
4. 设置适当的超时和日志轮转策略
5. 使用环境变量或配置文件管理敏感信息

### 注意事项

- 敏感信息（如密码）不要硬编码
- 根据实际负载调整连接池参数
- 日志级别会影响系统性能，生产环境建议使用 `info` 或 `warn`
