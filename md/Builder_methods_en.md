# XLORM Builder Methods Documentation

[中文版](Builder_methods_zh.md "访问中文版")

[English](Builder_methods_en.md "Access English Version")

## Overview

The `builder.go` provides a powerful set of query building tools to help developers construct complex database queries in a flexible and intuitive way. These methods allow you to dynamically create SELECT, INSERT, UPDATE, and DELETE queries through method chaining and combination.

### Key Features

- **Method Chaining**: Supports continuous method calls to gradually build complex queries
- **Dynamic Conditions**: Dynamically add query conditions based on runtime logic
- **Type Safety**: Provides type-safe query building methods
- **High Readability**: Uses intuitive method names and syntax to improve code readability

### Use Cases

- Complex conditional queries
- Dynamically generating query statements
- Building queries based on user input
- Reducing repetitive query code

### Performance Considerations

- Builder methods aim to minimize memory allocation and performance overhead
- For high-performance scenarios, it is recommended to optimize query building logic in advance

## Creating a Builder

### NewBuilder
- Create a query builder
- Signature: `NewBuilder(table string) *Builder`
- Example: `builder := db.NewBuilder("users")`
- Actual Usage:
```go
// Create a query builder for the users table
userBuilder := db.NewBuilder("users")
// Generated SQL statement: (no SQL statement generated)
```

## Query Configuration Methods

### Fields
- Set query fields
- Signature: `Fields(fields ...string) *Builder`
- Example: `builder.Fields("id", "name", "age")`
- Actual Usage:
```go
// Query only id, name, and email fields of users
query, args := db.NewBuilder("users").
    Fields("id", "name", "email").
    Build()
// Generated SQL statement: SELECT id, name, email FROM users
```

### Where
- Add query conditions
- Signature: `Where(condition string, args ...interface{}) *Builder`
- Example: `builder.Where("age > ?", 18)`
- Actual Usage:
```go
// Query users older than 18
query, args := db.NewBuilder("users").
    Where("age > ?", 18).
    Build()
// Generated SQL statement: SELECT * FROM users WHERE age > 18
```

### OrWhere
- Add OR query conditions
- Signature: `OrWhere(condition string, args ...interface{}) *Builder`
- Example: `builder.Where("age > ?", 18).OrWhere("status = ?", "active")`
- Actual Usage:
```go
// Query users older than 18 or with active status
query, args := db.NewBuilder("users").
    Where("age > ?", 18).
    OrWhere("status = ?", "active").
    Build()
// Generated SQL statement: SELECT * FROM users WHERE age > 18 OR status = 'active'
```

### NotWhere
- Add NOT query conditions
- Signature: `NotWhere(condition string, args ...interface{}) *Builder`
- Example: `builder.NotWhere("status = ?", "deleted")`
- Actual Usage:
```go
// Query users whose status is not deleted
query, args := db.NewBuilder("users").
    NotWhere("status = ?", "deleted").
    Build()
// Generated SQL statement: SELECT * FROM users WHERE status != 'deleted'
```

### Join
- Add table join
- Signature: `Join(join string) *Builder`
- Example: `builder.Join("LEFT JOIN orders ON users.id = orders.user_id")`
- Actual Usage:
```go
// Join users and orders tables
query, args := db.NewBuilder("users").
    Fields("users.id", "users.name", "orders.order_id").
    Join("LEFT JOIN orders ON users.id = orders.user_id").
    Build()
// Generated SQL statement: SELECT users.id, users.name, orders.order_id FROM users LEFT JOIN orders ON users.id = orders.user_id
```

### GroupBy
- Add grouping
- Signature: `GroupBy(groupBy string) *Builder`
- Example: `builder.GroupBy("category")`
- Actual Usage:
```go
// Group users by city and count
query, args := db.NewBuilder("users").
    Fields("city", "COUNT(*) as user_count").
    GroupBy("city").
    Build()
// Generated SQL statement: SELECT city, COUNT(*) as user_count FROM users GROUP BY city
```

### Having
- Add group filtering conditions
- Signature: `Having(having string) *Builder`
- Example: `builder.Having("count(*) > 10")`
- Actual Usage:
```go
// Find cities with more than 10 users
query, args := db.NewBuilder("users").
    Fields("city", "COUNT(*) as user_count").
    GroupBy("city").
    Having("user_count > 10").
    Build()
// Generated SQL statement: SELECT city, COUNT(*) as user_count FROM users GROUP BY city HAVING user_count > 10
```

### OrderBy
- Add sorting
- Signature: `OrderBy(orderBy string) *Builder`
- Example: `builder.OrderBy("created_at desc")`
- Actual Usage:
```go
// Sort users by creation time in descending order
query, args := db.NewBuilder("users").
    OrderBy("created_at desc").
    Limit(10).
    Build()
// Generated SQL statement: SELECT * FROM users ORDER BY created_at desc LIMIT 10
```

### Limit
- Limit the number of records
- Signature: `Limit(limit int64) *Builder`
- Example: `builder.Limit(10)`
- Actual Usage:
```go
// Retrieve first 10 user records
query, args := db.NewBuilder("users").
    Limit(10).
    Build()
// Generated SQL statement: SELECT * FROM users LIMIT 10
```

### Page
- Set pagination
- Signature: `Page(page, pageSize int64) *Builder`
- Example: `builder.Page(1, 20)` // Page 1, 20 records per page
- Actual Usage:
```go
// Query page 2 with 15 records per page
query, args := db.NewBuilder("users").
    Page(2, 15).
    Build()
// Generated SQL statement: SELECT * FROM users LIMIT 15 OFFSET 15
```

### Offset
- Add offset
- Signature: `Offset(offset int64) *Builder`
- Example: `builder.Offset(10)` // Skip the first 10 records
- Actual Usage:
```go
// Skip first 10 records and retrieve next 5
query, args := db.NewBuilder("users").
    Offset(10).
    Limit(5).
    Build()
// Generated SQL statement: SELECT * FROM users LIMIT 5 OFFSET 10
```

### ForUpdate
- Add row lock
- Signature: `ForUpdate() *Builder`
- Example: `builder.ForUpdate()`
- Actual Usage:
```go
// Query and lock specific user record to prevent concurrent modification
query, args := db.NewBuilder("users").
    Where("id = ?", 123).
    ForUpdate().
    Build()
// Generated SQL statement: SELECT * FROM users WHERE id = 123 FOR UPDATE
```

## Build Methods

### Build
- Build SQL statement
- Signature: `Build() (string, []interface{})`
- Example:
```go
query, args := builder.Build()
// query is the generated SQL statement
// args are the corresponding parameters
```
- Actual Usage:
```go
// Complex query example: combining multiple conditions
query, args := db.NewBuilder("users").
    Fields("id", "name", "email").
    Where("age > ?", 18).
    OrWhere("status = ?", "active").
    OrderBy("created_at desc").
    Limit(10).
    Build()
// Generated SQL statement: SELECT id, name, email FROM users WHERE age > 18 OR status = 'active' ORDER BY created_at desc LIMIT 10
```

### ReleaseBuilder
- Release Builder object to the pool
- Signature: `ReleaseBuilder()`
- Example: `builder.ReleaseBuilder()`
- Note: Usually, you don't need to call this method manually, as the `Build()` method already includes automatic release of the Builder object to the object pool
- Actual Usage:
```go
// In most cases, you don't need to manually call ReleaseBuilder
query, args := db.NewBuilder("users").
    Where("status = ?", "active").
    Build()
// Build method automatically handles the release of the Builder object
```

## Usage Example

```go
query, args := db.NewBuilder("users").
    Fields("id", "name").
    Where("age > ?", 18).
    OrderBy("created_at desc").
    Limit(10).
    Build()
// Generated SQL statement: SELECT id, name FROM users WHERE age > ? ORDER BY created_at desc LIMIT 10
```

## Precautions
- Supports method chaining
- Flexible query condition configuration
- Uses object pool to manage Builder objects for performance optimization
- Automatically handles fields, conditions, grouping, sorting, and other query parameters
