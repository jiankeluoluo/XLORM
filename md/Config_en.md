# XLORM Configuration Documentation

[中文版](Config_zh.md "访问中文版")

[English](Config_en.md "Access English Version")

## Overview

The `Config` struct is the core configuration object of the XLORM framework, used to define detailed parameters for database connections and system behavior. Through fine-grained configuration, you can optimize database performance, log management, and connection behavior.

## Configuration Field Explanation

### Basic Connection Configuration

| Field Name | Type | Description | Default Value |
|-----------|------|-------------|--------------|
| `DBName` | `string` | Database alias for distinguishing different databases | None |
| `Driver` | `string` | Database driver type | None |
| `Host` | `string` | Database host address | Required |
| `Username` | `string` | Database username | Required |
| `Password` | `string` | Database password | Required |
| `Database` | `string` | Database name | Required |
| `Port` | `int` | Database port number | Required |

### Connection Enhancement Configuration

| Field Name | Type | Description | Default Value |
|-----------|------|-------------|--------------|
| `Charset` | `string` | Database character set | `"utf8mb4"` |
| `TablePrefix` | `string` | Table name prefix | None |
| `ConnMaxLifetime` | `time.Duration` | Maximum connection lifetime | None |
| `ConnMaxIdleTime` | `time.Duration` | Maximum connection idle time | None |
| `ConnTimeout` | `time.Duration` | Connection timeout | None |
| `ReadTimeout` | `time.Duration` | Read timeout | None |
| `WriteTimeout` | `time.Duration` | Write timeout | None |

### Logging Configuration

| Field Name | Type | Description | Default Value |
|-----------|------|-------------|--------------|
| `LogDir` | `string` | Log storage directory | None |
| `LogLevel` | `string` | Log level (debug/info/warn/error) | `"debug"` |
| `LogBufferSize` | `int` | Log buffer size | `5000` |
| `LogRotationEnabled` | `bool` | Enable log rotation | `false` |
| `LogRotationMaxAge` | `int` | Log retention days | `30` |

### Performance and Debugging Configuration

| Field Name | Type | Description | Default Value |
|-----------|------|-------------|--------------|
| `MaxOpenConns` | `int` | Maximum open connections | `0` (unlimited) |
| `MaxIdleConns` | `int` | Maximum idle connections | `0` (unlimited) |
| `SlowQueryTime` | `time.Duration` | Slow query threshold | None |
| `PoolStatsInterval` | `time.Duration` | Connection pool statistics frequency | None |
| `DBMetricsBufferSize` | `int` | Async metrics buffer size | `1000` |
| `EnablePoolStats` | `bool` | Enable performance metrics | `false` |
| `Debug` | `bool` | Enable debug mode | `false` |

## Configuration Example

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

## Configuration Validation

The `Config` struct provides a `Validate()` method to verify the validity of the configuration. Validation rules include:

1. Configuration cannot be empty
2. Host address cannot be empty
3. Port number must be within a valid range (1-65535)
4. Username cannot be empty
5. Database name cannot be empty
6. Log level must be valid

## Best Practices

1. Always use the `Validate()` method to check configuration
2. Reasonably set connection pool parameters based on business requirements
3. Disable debug mode in production environments
4. Set appropriate timeout and log rotation policies
5. Use environment variables or configuration files to manage sensitive information

## Precautions

- Do not hardcode sensitive information (such as passwords)
- Adjust connection pool parameters according to actual load
- Log levels affect system performance; it is recommended to use `info` or `warn` in production environments
