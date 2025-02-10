# XLORM 构建器方法文档

[中文版](Builder_methods_zh.md "访问中文版")

[English](Builder_methods_en.md "Access English Version")

## 概述

`builder.go` 提供了一组强大的查询构建工具，帮助开发者以灵活、直观的方式构建复杂的数据库查询。这些方法允许你通过链式调用和方法组合，动态地创建 SELECT、INSERT、UPDATE 和 DELETE 查询。

### 主要特点

- **链式调用**：支持连续调用多个方法，逐步构建复杂查询
- **动态条件**：可以根据运行时条件动态添加查询条件
- **类型安全**：提供类型安全的查询构建方法
- **高度可读**：使用直观的方法名和语法，提高代码可读性

### 使用场景

- 复杂的条件查询
- 动态生成查询语句
- 根据用户输入构建查询
- 减少重复的查询代码

### 性能考虑

- 构建器方法会尽量减少内存分配和性能开销
- 对于高性能场景，建议提前优化查询构建逻辑

## 创建构建器

### NewBuilder
- 创建查询构建器
- 签名：`NewBuilder(table string) *Builder`
- 示例：`builder := db.NewBuilder("users")`
- 实际使用：
```go
// 创建一个针对 users 表的查询构建器
userBuilder := db.NewBuilder("users")
// 生成的 SQL 语句：(无)
```

## 查询配置方法

### Fields
- 设置查询字段
- 签名：`Fields(fields ...string) *Builder`
- 示例：`builder.Fields("id", "name", "age")`
- 实际使用：
```go
// 只查询用户的 id、name 和 email 字段
query, args := db.NewBuilder("users").
    Fields("id", "name", "email").
    Build()
// 生成的 SQL 语句：SELECT id, name, email FROM users
```

### Where
- 添加查询条件
- 签名：`Where(condition string, args ...interface{}) *Builder`
- 示例：`builder.Where("age > ?", 18)`
- 实际使用：
```go
// 查询年龄大于 18 岁的用户
query, args := db.NewBuilder("users").
    Where("age > ?", 18).
    Build()
// 生成的 SQL 语句：SELECT * FROM users WHERE age > 18
```

### OrWhere
- 添加 OR 查询条件
- 签名：`OrWhere(condition string, args ...interface{}) *Builder`
- 示例：`builder.Where("age > ?", 18).OrWhere("status = ?", "active")`
- 实际使用：
```go
// 查询年龄大于 18 岁或状态为 active 的用户
query, args := db.NewBuilder("users").
    Where("age > ?", 18).
    OrWhere("status = ?", "active").
    Build()
// 生成的 SQL 语句：SELECT * FROM users WHERE age > 18 OR status = 'active'
```

### NotWhere
- 添加 NOT 查询条件
- 签名：`NotWhere(condition string, args ...interface{}) *Builder`
- 示例：`builder.NotWhere("status = ?", "deleted")`
- 实际使用：
```go
// 查询状态不是 deleted 的用户
query, args := db.NewBuilder("users").
    NotWhere("status = ?", "deleted").
    Build()
// 生成的 SQL 语句：SELECT * FROM users WHERE status != 'deleted'
```

### Join
- 添加连接
- 签名：`Join(join string) *Builder`
- 示例：`builder.Join("LEFT JOIN orders ON users.id = orders.user_id")`
- 实际使用：
```go
// 联表查询用户及其订单信息
query, args := db.NewBuilder("users").
    Fields("users.id", "users.name", "orders.order_id").
    Join("LEFT JOIN orders ON users.id = orders.user_id").
    Build()
// 生成的 SQL 语句：SELECT users.id, users.name, orders.order_id FROM users LEFT JOIN orders ON users.id = orders.user_id
```

### GroupBy
- 添加分组
- 签名：`GroupBy(groupBy string) *Builder`
- 示例：`builder.GroupBy("category")`
- 实际使用：
```go
// 按城市分组统计用户数量
query, args := db.NewBuilder("users").
    Fields("city", "COUNT(*) as user_count").
    GroupBy("city").
    Build()
// 生成的 SQL 语句：SELECT city, COUNT(*) as user_count FROM users GROUP BY city
```

### Having
- 添加分组条件
- 签名：`Having(having string) *Builder`
- 示例：`builder.Having("count(*) > 10")`
- 实际使用：
```go
// 查询用户数量大于 10 的城市
query, args := db.NewBuilder("users").
    Fields("city", "COUNT(*) as user_count").
    GroupBy("city").
    Having("user_count > 10").
    Build()
// 生成的 SQL 语句：SELECT city, COUNT(*) as user_count FROM users GROUP BY city HAVING user_count > 10
```

### OrderBy
- 添加排序
- 签名：`OrderBy(orderBy string) *Builder`
- 示例：`builder.OrderBy("created_at desc")`
- 实际使用：
```go
// 按创建时间降序排序用户
query, args := db.NewBuilder("users").
    OrderBy("created_at desc").
    Limit(10).
    Build()
// 生成的 SQL 语句：SELECT * FROM users ORDER BY created_at desc LIMIT 10
```

### Limit
- 添加记录数限制
- 签名：`Limit(limit int64) *Builder`
- 示例：`builder.Limit(10)`
- 实际使用：
```go
// 查询前 10 条用户记录
query, args := db.NewBuilder("users").
    Limit(10).
    Build()
// 生成的 SQL 语句：SELECT * FROM users LIMIT 10
```

### Page
- 设置分页
- 签名：`Page(page, pageSize int64) *Builder`
- 示例：`builder.Page(1, 20)` // 第1页，每页20条记录
- 实际使用：
```go
// 查询第 2 页用户，每页 15 条记录
query, args := db.NewBuilder("users").
    Page(2, 15).
    Build()
// 生成的 SQL 语句：SELECT * FROM users LIMIT 15 OFFSET 15
```

### Offset
- 添加偏移
- 签名：`Offset(offset int64) *Builder`
- 示例：`builder.Offset(10)` // 跳过前10条记录
- 实际使用：
```go
// 跳过前 10 条记录，获取接下来的用户
query, args := db.NewBuilder("users").
    Offset(10).
    Limit(5).
    Build()
// 生成的 SQL 语句：SELECT * FROM users LIMIT 5 OFFSET 10
```

### ForUpdate
- 添加行锁
- 签名：`ForUpdate() *Builder`
- 示例：`builder.ForUpdate()`
- 实际使用：
```go
// 查询并锁定特定用户记录，防止并发修改
query, args := db.NewBuilder("users").
    Where("id = ?", 123).
    ForUpdate().
    Build()
// 生成的 SQL 语句：SELECT * FROM users WHERE id = 123 FOR UPDATE
```

## 构建方法

### Build
- 构建SQL语句
- 签名：`Build() (string, []interface{})`
- 示例：
```go
query, args := builder.Build()
// query 为生成的 SQL 语句
// args 为对应的参数
```
- 实际使用：
```go
// 复杂查询示例：组合多个查询条件
query, args := db.NewBuilder("users").
    Fields("id", "name", "email").
    Where("age > ?", 18).
    OrWhere("status = ?", "active").
    OrderBy("created_at desc").
    Limit(10).
    Build()
// 生成的 SQL 语句：SELECT id, name, email FROM users WHERE age > 18 OR status = 'active' ORDER BY created_at desc LIMIT 10
```

### ReleaseBuilder
- 释放Builder对象到池中
- 签名：`ReleaseBuilder()`
- 示例：`builder.ReleaseBuilder()`
- 注意：通常不需要手动调用此方法，因为 `Build()` 方法已经内置了自动释放 Builder 对象到对象池的功能
- 实际使用：
```go
// 大多数情况下，你不需要手动调用 ReleaseBuilder
query, args := db.NewBuilder("users").
    Where("status = ?", "active").
    Build()
// Build 方法已经自动处理了 Builder 对象的释放
```

## 使用示例

```go
query, args := db.NewBuilder("users").
    Fields("id", "name").
    Where("age > ?", 18).
    OrderBy("created_at desc").
    Limit(10).
    Build()
// 生成的 SQL 语句：SELECT id, name FROM users WHERE age > 18 ORDER BY created_at DESC LIMIT 10
```

## 注意事项
- 支持链式调用
- 可以灵活配置查询条件
- 使用对象池管理 Builder 对象，提高性能
- 自动处理字段、条件、分组、排序等查询参数
