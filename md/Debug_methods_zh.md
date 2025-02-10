# XLORM 调试工具文档

[中文版](Debug_methods_zh.md "访问中文版")

[English](Debug_methods_en.md "Access English Version")

## 概述

`debug.go` 提供了一组强大的 SQL 调试工具，帮助开发者更好地理解和分析数据库查询。这些方法可以帮助你在开发和测试阶段快速定位和解决问题。


## 主要功能

### 1. SQL 语句获取方法

#### GetQuerySQL(queryType string)
获取最终生成的查询 SQL 语句和参数

##### SELECT 查询示例

```go
// 示例：基本查询
users := db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10)

sql, args := users.GetQuerySQL("SELECT")
// sql: "SELECT * FROM users WHERE age > ? ORDER BY created_at DESC LIMIT 10"
// args: [18]

// 示例：指定查询字段
selectedUsers := db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC")

sql, args := selectedUsers.GetQuerySQL("SELECT")
// sql: "SELECT id, name, email FROM users WHERE status = ? ORDER BY created_at DESC"
// args: ["active"]

// 示例：复杂查询（JOIN）
complexQuery := db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5)

sql, args := complexQuery.GetQuerySQL("SELECT")
// sql: "SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > ? GROUP BY users.id HAVING COUNT(orders.id) > ?"
// args: [1000, 5]
```

##### COUNT 查询示例

```go
// 示例：基本计数
userCount := db.M("users").
    Where("age > ?", 18)

sql, args := userCount.GetQuerySQL("COUNT")
// sql: "SELECT COUNT(*) FROM users WHERE age > ?"
// args: [18]

// 示例：带条件的计数
activeUserCount := db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35)

sql, args := activeUserCount.GetQuerySQL("COUNT")
// sql: "SELECT COUNT(*) FROM users WHERE status = ? AND age BETWEEN ? AND ?"
// args: ["active", 18, 35]

// 示例：去重计数
distinctEmailCount := db.M("users").
    Where("registered = ?", true)

sql, args := distinctEmailCount.GetQuerySQL("COUNT DISTINCT")
// sql: "SELECT COUNT(DISTINCT email) FROM users WHERE registered = ?"
// args: [true]
```

##### DELETE 查询示例

```go
// 示例：基本删除
deleteInactiveUsers := db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0))

sql, args := deleteInactiveUsers.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE last_login < ?"
// args: [时间戳]

// 示例：带多个条件的删除
deleteSpecificUsers := db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18)

sql, args := deleteSpecificUsers.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE status = ? AND age < ?"
// args: ["inactive", 18]

// 示例：限制删除数量
limitedDelete := db.M("users").
    Where("is_test = ?", true).
    Limit(100)

sql, args := limitedDelete.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE is_test = ? LIMIT 100"
// args: [true]
```

#### GetInsertSQL(data interface{})
获取插入语句的 SQL

```go
// 示例：插入用户
user := User{
    Name: "张三",
    Age: 25,
    Email: "zhangsan@example.com",
}

sql, args, err := db.M("users").GetInsertSQL(user)
// sql: "INSERT INTO users (name, age, email) VALUES (?, ?, ?)"
// args: ["张三", 25, "zhangsan@example.com"]
```

#### GetBatchInsertSQL(data []interface{})
获取批量插入语句的 SQL

```go
// 示例：批量插入用户
users := []User{
    {Name: "张三", Age: 25},
    {Name: "李四", Age: 30},
    {Name: "王五", Age: 35},
}

sql, args, err := db.M("users").GetBatchInsertSQL(users)
// sql: "INSERT INTO users (name, age) VALUES (?, ?), (?, ?), (?, ?)"
// args: ["张三", 25, "李四", 30, "王五", 35]
```

#### GetUpdateSQL(data interface{})
获取更新语句的 SQL

```go
// 示例：更新用户
user := User{
    Name: "张三",
    Age: 26,
}

sql, args, err := db.M("users").
    Where("id = ?", 1).
    GetUpdateSQL(user)
// sql: "UPDATE users SET name = ?, age = ? WHERE id = ?"
// args: ["张三", 26, 1]
```

### 2. SQL 语句格式化方法

#### FormatQuerySQL(queryType string)
格式化查询 SQL 语句，将参数替换到查询中

##### SELECT 查询示例

```go
// 示例：基本查询
users := db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10)

formattedSQL := users.FormatQuerySQL("SELECT")
// 输出: SELECT * FROM users WHERE age > 18 ORDER BY created_at DESC LIMIT 10

// 示例：指定查询字段
selectedUsers := db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC")

formattedSQL := selectedUsers.FormatQuerySQL("SELECT")
// 输出: SELECT id, name, email FROM users WHERE status = 'active' ORDER BY created_at DESC

// 示例：复杂查询（JOIN）
complexQuery := db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5)

formattedSQL := complexQuery.FormatQuerySQL("SELECT")
// 输出: SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > 1000 GROUP BY users.id HAVING COUNT(orders.id) > 5
```

##### COUNT 查询示例

```go
// 示例：基本计数
userCount := db.M("users").
    Where("age > ?", 18)

formattedSQL := userCount.FormatQuerySQL("COUNT")
// 输出: SELECT COUNT(*) FROM users WHERE age > 18

// 示例：带条件的计数
activeUserCount := db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35)

formattedSQL := activeUserCount.FormatQuerySQL("COUNT")
// 输出: SELECT COUNT(*) FROM users WHERE status = 'active' AND age BETWEEN 18 AND 35

// 示例：去重计数
distinctEmailCount := db.M("users").
    Where("registered = ?", true)

formattedSQL := distinctEmailCount.FormatQuerySQL("COUNT DISTINCT")
// 输出: SELECT COUNT(DISTINCT email) FROM users WHERE registered = 1
```

##### DELETE 查询示例

```go
// 示例：基本删除
deleteInactiveUsers := db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0))

formattedSQL := deleteInactiveUsers.FormatQuerySQL("DELETE")
// 输出: DELETE FROM users WHERE last_login < '2024-08-06 15:18:10'

// 示例：带多个条件的删除
deleteSpecificUsers := db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18)

formattedSQL := deleteSpecificUsers.FormatQuerySQL("DELETE")
// 输出: DELETE FROM users WHERE status = 'inactive' AND age < 18

// 示例：限制删除数量
limitedDelete := db.M("users").
    Where("is_test = ?", true).
    Limit(100)

formattedSQL := limitedDelete.FormatQuerySQL("DELETE")
// 输出: DELETE FROM users WHERE is_test = 1 LIMIT 100
```

#### FormatInsertSQL(data interface{})
格式化插入 SQL 语句

```go
// 示例：插入用户
user := User{
    Name: "张三",
    Age: 25,
    Email: "zhangsan@example.com",
}

formattedSQL, err := db.M("users").FormatInsertSQL(user)
// 输出: INSERT INTO users (name, age, email) VALUES ('张三', 25, 'zhangsan@example.com')
```

#### FormatBatchInsertSQL(data []interface{})
格式化批量插入 SQL 语句

```go
// 示例：批量插入用户
users := []User{
    {Name: "张三", Age: 25},
    {Name: "李四", Age: 30},
    {Name: "王五", Age: 35},
}

formattedSQL, err := db.M("users").FormatBatchInsertSQL(users)
// 输出: INSERT INTO users (name, age) VALUES ('张三', 25), ('李四', 30), ('王五', 35)
```

#### FormatUpdateSQL(data interface{})
格式化更新 SQL 语句

```go
// 示例：更新用户
user := User{
    Name: "张三",
    Age: 26,
}

formattedSQL, err := db.M("users").
    Where("id = ?", 1).
    FormatUpdateSQL(user)
// 输出: UPDATE users SET name = '张三', age = 26 WHERE id = 1
```

### 3. SQL 语句日志打印方法

#### PrintQuerySQL(queryType string)
打印查询 SQL 语句到日志

##### SELECT 查询示例

```go
// 示例：基本查询
db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10).
    PrintQuerySQL("SELECT")
// 日志输出: 生成的SQL语句 sql=SELECT * FROM users WHERE age > 18 ORDER BY created_at DESC LIMIT 10

// 示例：指定查询字段
db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC").
    PrintQuerySQL("SELECT")
// 日志输出: 生成的SQL语句 sql=SELECT id, name, email FROM users WHERE status = 'active' ORDER BY created_at DESC

// 示例：复杂查询（JOIN）
db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5).
    PrintQuerySQL("SELECT")
// 日志输出: 生成的SQL语句 sql=SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > 1000 GROUP BY users.id HAVING COUNT(orders.id) > 5
```

##### COUNT 查询示例

```go
// 示例：基本计数
db.M("users").
    Where("age > ?", 18).
    PrintQuerySQL("COUNT")
// 日志输出: 生成的SQL语句 sql=SELECT COUNT(*) FROM users WHERE age > 18

// 示例：带条件的计数
db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35).
    PrintQuerySQL("COUNT")
// 日志输出: 生成的SQL语句 sql=SELECT COUNT(*) FROM users WHERE status = 'active' AND age BETWEEN 18 AND 35

// 示例：去重计数
db.M("users").
    Where("registered = ?", true).
    PrintQuerySQL("COUNT DISTINCT")
// 日志输出: 生成的SQL语句 sql=SELECT COUNT(DISTINCT email) FROM users WHERE registered = 1
```

##### DELETE 查询示例

```go
// 示例：基本删除
db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0)).
    PrintQuerySQL("DELETE")
// 日志输出: 生成的SQL语句 sql=DELETE FROM users WHERE last_login < '2024-08-06 15:18:10'

// 示例：带多个条件的删除
db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18).
    PrintQuerySQL("DELETE")
// 日志输出: 生成的SQL语句 sql=DELETE FROM users WHERE status = 'inactive' AND age < 18

// 示例：限制删除数量
db.M("users").
    Where("is_test = ?", true).
    Limit(100).
    PrintQuerySQL("DELETE")
// 日志输出: 生成的SQL语句 sql=DELETE FROM users WHERE is_test = 1 LIMIT 100
```

#### PrintInsertSQL(data interface{})
打印插入 SQL 语句到日志

```go
// 示例：插入用户
user := User{
    Name: "张三",
    Age: 25,
}

err := db.M("users").PrintInsertSQL(user)
// 日志输出: 生成的插入SQL语句 sql=INSERT INTO users (name, age) VALUES ('张三', 25)
```

#### PrintBatchInsertSQL(data []interface{})
打印批量插入 SQL 语句到日志

```go
// 示例：批量插入用户
users := []User{
    {Name: "张三", Age: 25},
    {Name: "李四", Age: 30},
}

err := db.M("users").PrintBatchInsertSQL(users)
// 日志输出: 生成的批量插入SQL语句 sql=INSERT INTO users (name, age) VALUES ('张三', 25), ('李四', 30)
```

#### PrintUpdateSQL(data interface{})
打印更新 SQL 语句到日志

```go
// 示例：更新用户
user := User{
    Name: "张三",
    Age: 26,
}

err := db.M("users").
    Where("id = ?", 1).
    PrintUpdateSQL(user)
// 日志输出: 生成的更新SQL语句 sql=UPDATE users SET name = '张三', age = 26 WHERE id = 1
```

## 注意事项

- 这些方法主要用于开发和调试阶段
- 格式化方法仅用于展示，不应用于实际数据库查询
- 方法会返回完整的 SQL 语句，包括参数

## 性能建议

- 在生产环境中谨慎使用这些方法
- 对于高性能场景，建议仅在必要时使用

## 许可

这些调试工具是 XLORM 框架的一部分，遵循框架的开源许可协议。
