# XLORM Debug Methods Documentation

[中文版](Debug_methods_zh.md "访问中文版")

[English](Debug_methods_en.md "Access English Version")

## Overview

The `debug.go` provides a powerful set of SQL debugging tools to help developers better understand and analyze database queries. These methods can help you quickly locate and solve problems during development and testing.

## Main Features

### 1. SQL Statement Retrieval Methods

#### GetQuerySQL(queryType string)
Retrieve the final generated query SQL statement and parameters

##### SELECT Query Examples

```go
// Example: Basic query
users := db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10)

sql, args := users.GetQuerySQL("SELECT")
// sql: "SELECT * FROM users WHERE age > ? ORDER BY created_at DESC LIMIT 10"
// args: [18]

// Example: Specify query fields
selectedUsers := db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC")

sql, args := selectedUsers.GetQuerySQL("SELECT")
// sql: "SELECT id, name, email FROM users WHERE status = ? ORDER BY created_at DESC"
// args: ["active"]

// Example: Complex query (JOIN)
complexQuery := db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5)

sql, args := complexQuery.GetQuerySQL("SELECT")
// sql: "SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > ? GROUP BY users.id HAVING COUNT(orders.id) > ?"
// args: [1000, 5]
```

##### COUNT Query Examples

```go
// Example: Basic count
userCount := db.M("users").
    Where("age > ?", 18)

sql, args := userCount.GetQuerySQL("COUNT")
// sql: "SELECT COUNT(*) FROM users WHERE age > ?"
// args: [18]

// Example: Conditional count
activeUserCount := db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35)

sql, args := activeUserCount.GetQuerySQL("COUNT")
// sql: "SELECT COUNT(*) FROM users WHERE status = ? AND age BETWEEN ? AND ?"
// args: ["active", 18, 35]

// Example: Distinct count
distinctEmailCount := db.M("users").
    Where("registered = ?", true)

sql, args := distinctEmailCount.GetQuerySQL("COUNT DISTINCT")
// sql: "SELECT COUNT(DISTINCT email) FROM users WHERE registered = ?"
// args: [true]
```

##### DELETE Query Examples

```go
// Example: Basic delete
deleteInactiveUsers := db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0))

sql, args := deleteInactiveUsers.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE last_login < ?"
// args: [timestamp]

// Example: Delete with multiple conditions
deleteSpecificUsers := db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18)

sql, args := deleteSpecificUsers.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE status = ? AND age < ?"
// args: ["inactive", 18]

// Example: Limit delete quantity
limitedDelete := db.M("users").
    Where("is_test = ?", true).
    Limit(100)

sql, args := limitedDelete.GetQuerySQL("DELETE")
// sql: "DELETE FROM users WHERE is_test = ? LIMIT 100"
// args: [true]
```

#### GetInsertSQL(data interface{})
Get the SQL for insert statement

```go
// Example: Insert user
user := User{
    Name: "John Doe",
    Age: 25,
    Email: "johndoe@example.com",
}

sql, args, err := db.M("users").GetInsertSQL(user)
// sql: "INSERT INTO users (name, age, email) VALUES (?, ?, ?)"
// args: ["John Doe", 25, "johndoe@example.com"]
```

#### GetBatchInsertSQL(data []interface{})
Get the SQL for batch insert statement

```go
// Example: Batch insert users
users := []User{
    {Name: "John Doe", Age: 25},
    {Name: "Jane Smith", Age: 30},
    {Name: "Bob Johnson", Age: 35},
}

sql, args, err := db.M("users").GetBatchInsertSQL(users)
// sql: "INSERT INTO users (name, age) VALUES (?, ?), (?, ?), (?, ?)"
// args: ["John Doe", 25, "Jane Smith", 30, "Bob Johnson", 35]
```

#### GetUpdateSQL(data interface{})
Get the SQL for update statement

```go
// Example: Update user
user := User{
    Name: "John Doe",
    Age: 26,
}

sql, args, err := db.M("users").
    Where("id = ?", 1).
    GetUpdateSQL(user)
// sql: "UPDATE users SET name = ?, age = ? WHERE id = ?"
// args: ["John Doe", 26, 1]
```

### 2. SQL Statement Formatting Methods

#### FormatQuerySQL(queryType string)
Format query SQL statement, replacing parameters in the query

##### SELECT Query Examples

```go
// Example: Basic query
users := db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10)

formattedSQL := users.FormatQuerySQL("SELECT")
// Output: SELECT * FROM users WHERE age > 18 ORDER BY created_at DESC LIMIT 10

// Example: Specify query fields
selectedUsers := db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC")

formattedSQL := selectedUsers.FormatQuerySQL("SELECT")
// Output: SELECT id, name, email FROM users WHERE status = 'active' ORDER BY created_at DESC

// Example: Complex query (JOIN)
complexQuery := db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5)

formattedSQL := complexQuery.FormatQuerySQL("SELECT")
// Output: SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > 1000 GROUP BY users.id HAVING COUNT(orders.id) > 5
```

##### COUNT Query Examples

```go
// Example: Basic count
userCount := db.M("users").
    Where("age > ?", 18)

formattedSQL := userCount.FormatQuerySQL("COUNT")
// Output: SELECT COUNT(*) FROM users WHERE age > 18

// Example: Conditional count
activeUserCount := db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35)

formattedSQL := activeUserCount.FormatQuerySQL("COUNT")
// Output: SELECT COUNT(*) FROM users WHERE status = 'active' AND age BETWEEN 18 AND 35

// Example: Distinct count
distinctEmailCount := db.M("users").
    Where("registered = ?", true)

formattedSQL := distinctEmailCount.FormatQuerySQL("COUNT DISTINCT")
// Output: SELECT COUNT(DISTINCT email) FROM users WHERE registered = 1
```

##### DELETE Query Examples

```go
// Example: Basic delete
deleteInactiveUsers := db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0))

formattedSQL := deleteInactiveUsers.FormatQuerySQL("DELETE")
// Output: DELETE FROM users WHERE last_login < '2024-08-06 15:18:10'

// Example: Delete with multiple conditions
deleteSpecificUsers := db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18)

formattedSQL := deleteSpecificUsers.FormatQuerySQL("DELETE")
// Output: DELETE FROM users WHERE status = 'inactive' AND age < 18

// Example: Limit delete quantity
limitedDelete := db.M("users").
    Where("is_test = ?", true).
    Limit(100)

formattedSQL := limitedDelete.FormatQuerySQL("DELETE")
// Output: DELETE FROM users WHERE is_test = 1 LIMIT 100
```

#### FormatInsertSQL(data interface{})
Format insert SQL statement

```go
// Example: Format insert
user := User{
    Name: "John Doe",
    Age: 25,
    Email: "johndoe@example.com",
}

formattedSQL, err := db.M("users").FormatInsertSQL(user)
// Output: INSERT INTO users (name, age, email) VALUES ('John Doe', 25, 'johndoe@example.com')
```

#### FormatBatchInsertSQL(data []interface{})
Format batch insert SQL statement

```go
// Example: Format batch insert
users := []User{
    {Name: "John Doe", Age: 25},
    {Name: "Jane Smith", Age: 30},
    {Name: "Bob Johnson", Age: 35},
}

formattedSQL, err := db.M("users").FormatBatchInsertSQL(users)
// Output: INSERT INTO users (name, age) VALUES ('John Doe', 25), ('Jane Smith', 30), ('Bob Johnson', 35)
```

#### FormatUpdateSQL(data interface{})
Format update SQL statement

```go
// Example: Format update
user := User{
    Name: "John Doe",
    Age: 26,
}

formattedSQL, err := db.M("users").
    Where("id = ?", 1).
    FormatUpdateSQL(user)
// Output: UPDATE users SET name = 'John Doe', age = 26 WHERE id = 1
```

### 3. SQL Statement Logging Methods

#### PrintQuerySQL(queryType string)
Print query SQL statement to log

##### SELECT Query Examples

```go
// Example: Basic query
db.M("users").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Limit(10).
    PrintQuerySQL("SELECT")
// Log output: Generated SQL statement sql=SELECT * FROM users WHERE age > 18 ORDER BY created_at DESC LIMIT 10

// Example: Specify query fields
db.M("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at DESC").
    PrintQuerySQL("SELECT")
// Log output: Generated SQL statement sql=SELECT id, name, email FROM users WHERE status = 'active' ORDER BY created_at DESC

// Example: Complex query (JOIN)
db.M("users").
    Join("orders", "users.id = orders.user_id").
    Where("orders.total > ?", 1000).
    GroupBy("users.id").
    Having("COUNT(orders.id) > ?", 5).
    PrintQuerySQL("SELECT")
// Log output: Generated SQL statement sql=SELECT * FROM users JOIN orders ON users.id = orders.user_id WHERE orders.total > 1000 GROUP BY users.id HAVING COUNT(orders.id) > 5
```

##### COUNT Query Examples

```go
// Example: Basic count
db.M("users").
    Where("age > ?", 18).
    PrintQuerySQL("COUNT")
// Log output: Generated SQL statement sql=SELECT COUNT(*) FROM users WHERE age > 18

// Example: Conditional count
db.M("users").
    Where("status = ?", "active").
    Where("age BETWEEN ? AND ?", 18, 35).
    PrintQuerySQL("COUNT")
// Log output: Generated SQL statement sql=SELECT COUNT(*) FROM users WHERE status = 'active' AND age BETWEEN 18 AND 35

// Example: Distinct count
db.M("users").
    Where("registered = ?", true).
    PrintQuerySQL("COUNT DISTINCT")
// Log output: Generated SQL statement sql=SELECT COUNT(DISTINCT email) FROM users WHERE registered = 1
```

##### DELETE Query Examples

```go
// Example: Basic delete
db.M("users").
    Where("last_login < ?", time.Now().AddDate(0, -6, 0)).
    PrintQuerySQL("DELETE")
// Log output: Generated SQL statement sql=DELETE FROM users WHERE last_login < '2024-08-06 15:18:10'

// Example: Delete with multiple conditions
db.M("users").
    Where("status = ?", "inactive").
    Where("age < ?", 18).
    PrintQuerySQL("DELETE")
// Log output: Generated SQL statement sql=DELETE FROM users WHERE status = 'inactive' AND age < 18

// Example: Limit delete quantity
db.M("users").
    Where("is_test = ?", true).
    Limit(100).
    PrintQuerySQL("DELETE")
// Log output: Generated SQL statement sql=DELETE FROM users WHERE is_test = 1 LIMIT 100
```

#### PrintInsertSQL(data interface{})
Print insert SQL statement to log

```go
// Example: Print insert
user := User{
    Name: "John Doe",
    Age: 25,
}

err := db.M("users").PrintInsertSQL(user)
// Log output: Generated insert SQL statement sql=INSERT INTO users (name, age) VALUES ('John Doe', 25)
```

#### PrintBatchInsertSQL(data []interface{})
Print batch insert SQL statement to log

```go
// Example: Print batch insert
users := []User{
    {Name: "John Doe", Age: 25},
    {Name: "Jane Smith", Age: 30},
}

err := db.M("users").PrintBatchInsertSQL(users)
// Log output: Generated batch insert SQL statement sql=INSERT INTO users (name, age) VALUES ('John Doe', 25), ('Jane Smith', 30)
```

#### PrintUpdateSQL(data interface{})
Print update SQL statement to log

```go
// Example: Print update
user := User{
    Name: "John Doe",
    Age: 26,
}

err := db.M("users").
    Where("id = ?", 1).
    PrintUpdateSQL(user)
// Log output: Generated update SQL statement sql=UPDATE users SET name = 'John Doe', age = 26 WHERE id = 1
```

## Precautions

- These methods are primarily for development and debugging stages
- Formatting methods are for display only and should not be used for actual database queries
- Methods return complete SQL statements, including parameters

## Performance Recommendations

- Use these methods cautiously in production environments
- For high-performance scenarios, use only when necessary

## License

These debugging tools are part of the XLORM framework and follow the framework's open-source license agreement.
