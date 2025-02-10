# XLORM 表操作方法文档

[中文版](Table_methods_zh.md "访问中文版")

[English](Table_methods_en.md "Access English Version")

## 概述

`table.go` 提供了一组全面的表操作方法，简化了数据库表的增删改查操作。这些方法封装了常见的数据库交互模式，提供了一种直观、高效的数据库操作方式。

### 主要特点

- **简化数据库操作**：提供直接、简洁的方法进行数据库交互
- **类型安全**：支持结构体和映射的数据操作
- **灵活性**：支持多种查询、插入、更新和删除场景
- **事务支持**：无缝集成事务管理
- **性能优化**：内部实现了高效的数据库操作逻辑

### 使用场景

- 单表增删改查操作
- 批量数据处理
- 复杂条件查询
- 数据持久化
- 快速原型开发

### 性能考虑

- 方法针对常见操作进行了性能优化
- 支持批量操作，减少数据库交互次数
- 对于大规模数据处理，建议使用批量操作和事务

### 安全性

- 支持参数化查询，防止 SQL 注入
- 提供数据验证和类型转换机制
- 错误处理和异常捕获

## 查询条件方法

### Where
- 添加查询条件
- 签名：`Where(condition string, args ...interface{}) *table`
- 示例：`table.Where("id = ?", 1)`

### OrderBy
- 添加排序条件
- 签名：`OrderBy(order string) *table`
- 示例：`table.OrderBy("created_at desc")`

### Limit
- 添加记录数限制
- 签名：`Limit(limit int64) *table`
- 示例：`table.Limit(10)`

### Page
- 设置分页
- 签名：`Page(page, pageSize int64) *table`
- 示例：`table.Page(1, 20)` // 第1页，每页20条记录

### Offset
- 添加偏移量
- 签名：`Offset(offset int64) *table`
- 示例：`table.Offset(10)` // 跳过前10条记录

### Fields
- 设置查询字段
- 签名：`Fields(fields string) *table`
- 示例：`table.Fields("id, name, age")`

### Join
- 添加表连接
- 签名：`Join(join string) *table`
- 示例：`table.Join("LEFT JOIN users ON users.id = orders.user_id")`

### GroupBy
- 添加分组条件
- 签名：`GroupBy(groupBy string) *table`
- 示例：`table.GroupBy("category")`

### Having
- 添加分组过滤条件
- 签名：`Having(having string) *table`
- 示例：`table.Having("count(*) > 10")`

## 查询方法

### Count
获取符合条件的记录数量。

```go
// 示例1：基本计数
total, err := db.M("users").
    Where("age >= ?", 18).
    Count()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("成年用户数量: %d\n", total)

// 示例2：复杂条件计数
total, err = db.M("orders").
    Where("status = ?", "completed").
    Where("total > ?", 1000).
    Count()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("高价订单数量: %d\n", total)

// 示例3：联表计数
total, err = db.M("products").
    Join("categories").
    Where("categories.status = ?", "active").
    Count()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("活跃分类下的商品数量: %d\n", total)
```

### Find
查询单条记录，返回 `map[string]interface{}` 类型。

```go
// 示例1：根据主键查找
user, err := db.M("users").
    Where("id = ?", 1).
    Find()
if err != nil {
    if err == sql.ErrNoRows {
        fmt.Println("未找到用户")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("用户信息: %+v\n", user)

// 示例2：复杂条件查找
activeUser, err := db.M("users").
    Where("status = ?", "active").
    Where("age >= ?", 18).
    OrderBy("created_at DESC").
    Find()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("活跃用户信息: %+v\n", activeUser)

// 示例3：联表查询
userOrder, err := db.M("users").
    Join("orders").
    Where("users.id = ?", 1).
    Find()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("用户订单信息: %+v\n", userOrder)
```

### FindAll
查询多条记录，返回 `[]map[string]interface{}` 类型。

```go
// 示例1：查询所有成年用户
users, err := db.M("users").
    Where("age >= ?", 18).
    OrderBy("age DESC").
    Limit(10).
    FindAll()
if err != nil {
    log.Fatal(err)
}
for _, user := range users {
    fmt.Printf("用户: %+v\n", user)
}

// 示例2：复杂条件查询
activeUsers, err := db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35).
    FindAll()
if err != nil {
    log.Fatal(err)
}
for _, user := range activeUsers {
    fmt.Printf("活跃用户: %+v\n", user)
}

// 示例3：联表查询并获取总数
db.M("users").HasTotal(true)
usersWithOrders, err := db.M("users").
    Join("orders").
    GroupBy("users.id").
    FindAll()
if err != nil {
    log.Fatal(err)
}
total := db.M("users").GetTotal()
fmt.Printf("用户订单信息: %+v\n总数: %d\n", usersWithOrders, total)
```

### FindAllWithCursor
使用游标逐行读取数据，减少内存占用。

```go
// 示例：使用游标处理大量数据
err := db.M("orders").
    Where("status = ?", "completed").
    FindAllWithCursor(context.Background(), func(order map[string]interface{}) error {
        // 处理每一条订单
        fmt.Printf("处理订单: %+v\n", order)
        
        // 可以根据业务逻辑决定是否继续处理
        // 返回错误将中止游标遍历
        return nil
    })
if err != nil {
    log.Fatal(err)
}
```

### FindAllWithContext
带上下文的多记录查询，支持超时和取消。

```go
// 示例：带超时的查询
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

users, err := db.M("users").
    Where("status = ?", "active").
    FindAllWithContext(ctx)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("查询超时")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("活跃用户: %+v\n", users)
```

### Insert
插入单条记录。

```go
// 示例1：插入用户
userData := map[string]interface{}{
    "name":     "张三",
    "email":    "zhangsan@example.com",
    "age":      25,
    "status":   "active",
}

id, err := db.M("users").Insert(userData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("插入用户，ID: %d\n", id)

// 示例2：插入订单
orderData := map[string]interface{}{
    "user_id":     1,
    "total":       199.99,
    "status":      "pending",
    "created_at":  time.Now(),
}

id, err = db.M("orders").Insert(orderData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("插入订单，ID: %d\n", id)
```

### InsertWithContext
带上下文的插入操作。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

userData := map[string]interface{}{
    "name":     "李四",
    "email":    "lisi@example.com",
    "age":      30,
    "status":   "active",
}

id, err := db.M("users").InsertWithContext(ctx, userData)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("插入超时")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("插入用户，ID: %d\n", id)
```

### Update
更新记录。

```go
// 示例1：更新用户
userData := map[string]interface{}{
    "id":     1,
    "name":   "张三（已更新）",
    "status": "active",
}

rowsAffected, err := db.M("users").Update(userData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("更新用户，影响行数: %d\n", rowsAffected)

// 示例2：条件更新
rowsAffected, err = db.M("users").
    Where("status = ?", "active").
    Where("age < ?", 25).
    Update(map[string]interface{}{
        "status": "potential",
    })
if err != nil {
    log.Fatal(err)
}
fmt.Printf("批量更新用户状态，影响行数: %d\n", rowsAffected)
```

### UpdateWithContext
带上下文的更新操作。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

userData := map[string]interface{}{
    "id":     1,
    "status": "inactive",
}

rowsAffected, err := db.M("users").UpdateWithContext(ctx, userData)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("更新超时")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("更新用户，影响行数: %d\n", rowsAffected)
```

### Delete
删除记录。

```go
// 示例1：根据主键删除
userData := map[string]interface{}{
    "id": 1,
}

rowsAffected, err := db.M("users").Delete(userData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("删除用户，影响行数: %d\n", rowsAffected)

// 示例2：条件删除
rowsAffected, err = db.M("orders").
    Where("status = ?", "canceled").
    Where("created_at < ?", time.Now().AddDate(0, -3, 0)).
    Delete()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("删除过期订单，影响行数: %d\n", rowsAffected)
```

### DeleteWithContext
带上下文的删除操作。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rowsAffected, err := db.M("logs").
    Where("created_at < ?", time.Now().AddDate(0, -1, 0)).
    DeleteWithContext(ctx)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("删除超时")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("删除过期日志，影响行数: %d\n", rowsAffected)
```

## 上下文方法

### WithContext
- 设置上下文
- 签名：`WithContext(ctx context.Context) *table`
- 示例：`table.WithContext(ctx)`

### FindAllWithContext
- 带上下文的多记录查询
- 签名：`FindAllWithContext(ctx context.Context) ([]map[string]interface{}, error)`
- 示例：`records, err := table.FindAllWithContext(ctx)`

## 总数控制方法

### HasTotal
- 设置是否需要获取总数
- 签名：`HasTotal(need bool) *table`
- 示例：`table.HasTotal(true)`

### GetTotal
- 获取记录集总数
- 签名：`GetTotal() int64`
- 示例：`total := table.GetTotal()`

## 数据操作方法

### Insert
- 插入记录
- 签名：`Insert(data interface{}) (lastInsertId int64, err error)`
- 示例：`id, err := table.Insert(data)`

### InsertWithContext
- 带上下文的插入记录
- 签名：`InsertWithContext(ctx context.Context, data interface{}) (lastInsertId int64, err error)`
- 示例：`id, err := table.InsertWithContext(ctx, data)`

### Update
- 更新记录
- 签名：`Update(data interface{}) (rowsAffected int64, err error)`
- 示例：`affected, err := table.Update(data)`

### UpdateWithContext
- 带上下文的更新记录
- 签名：`UpdateWithContext(ctx context.Context, data interface{}) (rowsAffected int64, err error)`
- 示例：`affected, err := table.UpdateWithContext(ctx, data)`

### Delete
- 删除记录
- 签名：`Delete() (rowsAffected int64, err error)`
- 示例：`affected, err := table.Delete()`

### DeleteWithContext
- 带上下文的删除记录
- 签名：`DeleteWithContext(ctx context.Context) (rowsAffected int64, err error)`
- 示例：`affected, err := table.DeleteWithContext(ctx)`

## 批量操作方法

### BatchInsert
- 批量插入记录
- 签名：`BatchInsert(data []map[string]interface{}, batchSize int) (totalAffecteds int64, err error)`
- 示例：
```go
users := []map[string]interface{}{
    {"name": "Alice", "age": 25},
    {"name": "Bob", "age": 30},
}
affected, err := table.BatchInsert(users, 100)
```

### BatchUpdate
- 批量更新记录
- 签名：`BatchUpdate(records []map[string]interface{}, keyField string, batchSize int) (totalAffecteds int64, err error)`
- 示例：
```go
users := []map[string]interface{}{
    {"id": 1, "name": "Alice Updated", "age": 26},
    {"id": 2, "name": "Bob Updated", "age": 31},
}
affected, err := table.BatchUpdate(users, "id", 100)
```

## 批量操作注意事项
- 批量操作支持大规模数据处理
- 可以自定义批次大小
- 支持灵活的数据处理

## 事务方法

### Commit
- 提交事务
- 签名：`Commit() error`
- 实际使用：
```go
// 提交当前事务
err := transaction.Commit()
// 成功提交事务
```

### Rollback
- 回滚事务
- 签名：`Rollback() error`
- 实际使用：
```go
// 回滚当前事务
err := transaction.Rollback()
// 成功回滚事务
```

### DB
- 获取数据库实例
- 签名：`DB() *DB`
- 实际使用：
```go
// 获取与事务关联的数据库实例
db := transaction.DB()
```

## 注意事项
- 所有方法都提供了详细的调试和日志功能
- 批量操作使用事务确保数据一致性
- 支持灵活的 SQL 生成和调试
- 提供安全的参数化查询

### DeleteWithContext
带上下文的记录删除，支持超时和取消操作。

#### 基本使用示例

```go
// 示例1：基本上下文删除
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user := User{
    ID: 1,
}

// 带上下文的删除操作
rowsAffected, err := db.M("users").DeleteWithContext(ctx, &user)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("删除操作超时")
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("删除用户，影响行数: %d\n", rowsAffected)

// 示例2：可取消的删除
ctx, cancel = context.WithCancel(context.Background())
go func() {
    // 模拟在某些条件下取消删除
    time.Sleep(2 * time.Second)
    cancel()
}()

rowsAffected, err = db.M("logs").
    Where("created_at < ?", time.Now().AddDate(0, -1, 0)).
    DeleteWithContext(ctx)
// 删除一个月前的日志
if err != nil {
    if err == context.Canceled {
        fmt.Println("删除操作被取消")
    } else {
        log.Fatal(err)
    }
}

// 示例3：批量条件删除
ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

rowsAffected, err = db.M("temp_files").
    Where("status = ?", "processed").
    Where("created_at < ?", time.Now().AddDate(0, 0, -7)).
    DeleteWithContext(ctx)
// 删除7天前已处理的临时文件
if err != nil {
    log.Fatal(err)
}
fmt.Printf("删除过期临时文件，影响行数: %d\n", rowsAffected)
```

## 事务处理
使用事务确保多个数据库操作的原子性和一致性。

#### 基本使用示例

```go
// 示例1：基本事务处理
err := db.WithTransaction(context.Background(), func(tx *DB) error {
    // 扣减用户余额
    _, err := tx.M("users").
        Where("id = ?", 1).
        UpdateColumns(map[string]interface{}{
            "balance": gorm.Expr("balance - ?", 100),
        })
    if err != nil {
        return err
    }

    // 创建支付记录
    _, err = tx.M("payment_logs").Insert(&PaymentLog{
        UserID:   1,
        Amount:   100,
        Status:   "completed",
        CreateAt: time.Now(),
    })
    return err
})

if err != nil {
    log.Fatal("事务处理失败:", err)
}
fmt.Println("支付事务处理成功")

// 示例2：带上下文的事务处理
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err = db.WithTransaction(ctx, func(tx *DB) error {
    // 创建订单
    order := Order{
        UserID:     1,
        Total:      199.99,
        Status:     "pending",
        CreateTime: time.Now(),
    }
    orderID, err := tx.M("orders").InsertWithContext(ctx, &order)
    if err != nil {
        return err
    }

    // 扣减库存
    _, err = tx.M("products").
        Where("id = ?", 100).
        UpdateColumns(map[string]interface{}{
            "stock": gorm.Expr("stock - 1"),
        })
    if err != nil {
        return err
    }

    // 创建订单明细
    _, err = tx.M("order_items").InsertWithContext(ctx, &OrderItem{
        OrderID:   orderID,
        ProductID: 100,
        Quantity:  1,
        Price:     199.99,
    })
    return err
})

if err != nil {
    log.Fatal("订单创建事务处理失败:", err)
}
fmt.Println("订单创建事务处理成功")

// 示例3：嵌套事务和错误处理
err = db.WithTransaction(context.Background(), func(tx *DB) error {
    // 第一个操作：用户注册
    userID, err := tx.M("users").Insert(&User{
        Name:     "新用户",
        Email:    "newuser@example.com",
        Status:   "active",
    })
    if err != nil {
        return err
    }

    // 嵌套事务：创建用户配置和初始积分
    return tx.WithTransaction(func(subTx *DB) error {
        // 创建用户配置
        _, err := subTx.M("user_configs").Insert(&UserConfig{
            UserID:    userID,
            Theme:     "default",
            Language:  "zh_CN",
        })
        if err != nil {
            return err
        }

        // 添加初始积分
        _, err = subTx.M("user_credits").Insert(&UserCredit{
            UserID: userID,
            Credits: 100,
            Reason:  "新用户注册奖励",
        })
        return err
    })
})

if err != nil {
    log.Fatal("用户注册事务处理失败:", err)
}
fmt.Println("用户注册事务处理成功")
```

#### 注意事项
- `WithTransaction()` 确保事务内的所有操作要么全部成功，要么全部回滚
- 支持嵌套事务
- 可以与上下文结合使用，支持超时和取消
- 返回 `error` 将触发事务回滚
- 适用于需要保证数据一致性的复杂操作
- 建议将相关的数据库操作放在同一个事务中
- 避免在事务中执行耗时的非数据库操作

## 事务基本用法

事务是数据库操作中确保数据一致性的重要机制。在 xlorm 中，提供了两种事务处理方式：`Begin()` 和 `ExecTx()`。

### Begin() 方法
`Begin()` 方法用于手动开启一个事务，返回一个 `*Transaction` 对象，需要手动管理事务的提交和回滚。

```go
// 示例1：手动管理事务
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

defer func() {
    if err != nil {
        // 发生错误时回滚
        if rbErr := tx.Rollback(); rbErr != nil {
            log.Printf("回滚失败: %v", rbErr)
        }
    } else {
        // 无错误时提交
        if err = tx.Commit(); err != nil {
            log.Printf("提交事务失败: %v", err)
        }
    }
}()

// 在事务中执行操作
_, err = tx.M("users").Insert(map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
})
if err != nil {
    return
}

// 可以执行多个操作
_, err = tx.M("orders").Insert(map[string]interface{}{
    "user_id": 1,
    "total":   100.00,
})
if err != nil {
    return
}
```

### ExecTx() 方法
`ExecTx()` 提供了更简洁的事务处理方式，自动管理事务的提交和回滚。

```go
// 示例2：使用 ExecTx() 简化事务处理
err := db.ExecTx(func(tx *Transaction) error {
    // 插入用户
    _, err := tx.M("users").Insert(map[string]interface{}{
        "name":  "李四",
        "email": "lisi@example.com",
    })
    if err != nil {
        return err
    }

    // 插入订单
    _, err = tx.M("orders").Insert(map[string]interface{}{
        "user_id": 1,
        "total":   200.00,
    })
    if err != nil {
        return err
    }

    // 返回 nil 表示事务成功
    return nil
})

if err != nil {
    log.Printf("事务执行失败: %v", err)
}
```

### 批量操作事务

对于大量数据的插入或更新，xlorm 提供了批量操作的事务支持。

```go
// 示例3：批量插入
users := []map[string]interface{}{
    {"name": "王五", "email": "wangwu@example.com"},
    {"name": "赵六", "email": "zhaoliu@example.com"},
}

// 批量插入，自动使用事务
affectedRows, err := db.M("users").BatchInsert(users, 1000)
if err != nil {
    log.Printf("批量插入失败: %v", err)
}
fmt.Printf("插入行数: %d\n", affectedRows)
```

### 事务处理注意事项

1. `Begin()` 方法需要手动管理 `Commit()` 和 `Rollback()`
2. `ExecTx()` 方法自动处理事务的提交和回滚
3. 事务中的所有操作要么全部成功，要么全部回滚
4. 使用 `defer` 可以确保事务正确处理
5. 批量操作默认使用事务，提高性能和数据一致性
6. 对于复杂的业务逻辑，推荐使用 `ExecTx()` 方法

### 错误处理示例

```go
// 示例4：复杂错误处理
err := db.ExecTx(func(tx *Transaction) error {
    // 模拟一个可能失败的操作
    _, err := tx.M("users").Insert(map[string]interface{}{
        "name":  "小明",
        "email": "xiaoming@example.com",
    })
    if err != nil {
        return err  // 返回错误将自动回滚事务
    }

    // 模拟业务逻辑验证
    if !validateUser(tx) {
        return errors.New("用户验证失败")
    }

    return nil
})

if err != nil {
    // 处理事务执行错误
    log.Printf("事务执行失败: %v", err)
}

// 用户验证函数示例
func validateUser(tx *Transaction) bool {
    // 执行复杂的用户验证逻辑
    return true
}
```

### 性能和调试

xlorm 的事务处理支持调试模式，可以记录事务的跟踪信息：

```go
// 开启调试模式
db.SetDebug(true)

// 执行事务，将输出详细的调试信息
err := db.ExecTx(func(tx *Transaction) error {
    // 事务操作
    return nil
})
