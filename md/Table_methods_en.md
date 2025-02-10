# XLORM Table Methods Documentation

[中文版](Table_methods_zh.md "访问中文版")

[English](Table_methods_en.md "Access English Version")

## Overview

The `table.go` provides a comprehensive set of table operation methods that simplify database table operations such as create, read, update, and delete (CRUD). These methods encapsulate common database interaction patterns, offering an intuitive and efficient way to interact with databases.

### Key Features

- **Simplified Database Operations**: Provides direct and concise methods for database interactions
- **Type Safety**: Supports data operations with structs and maps
- **Flexibility**: Supports various query, insert, update, and delete scenarios
- **Transaction Support**: Seamlessly integrates transaction management
- **Performance Optimization**: Implements efficient database operation logic internally

### Use Cases

- Single table CRUD operations
- Batch data processing
- Complex conditional queries
- Data persistence
- Rapid prototype development

### Performance Considerations

- Methods are performance-optimized for common operations
- Supports batch operations to reduce database interactions
- For large-scale data processing, use batch operations and transactions

### Security

- Supports parameterized queries to prevent SQL injection
- Provides data validation and type conversion mechanisms
- Error handling and exception capturing

## Query Condition Methods

### Where
- Add query conditions
- Signature: `Where(condition string, args ...interface{}) *table`
- Example: `table.Where("id = ?", 1)`

### OrderBy
- Add sorting conditions
- Signature: `OrderBy(order string) *table`
- Example: `table.OrderBy("created_at desc")`

### Limit
- Limit the number of records
- Signature: `Limit(limit int64) *table`
- Example: `table.Limit(10)`

### Page
- Set pagination
- Signature: `Page(page, pageSize int64) *table`
- Example: `table.Page(1, 20)` // Page 1, 20 records per page

### Offset
- Add offset
- Signature: `Offset(offset int64) *table`
- Example: `table.Offset(10)` // Skip the first 10 records

### Fields
- Set query fields
- Signature: `Fields(fields string) *table`
- Example: `table.Fields("id, name, age")`

### Join
- Add table join
- Signature: `Join(join string) *table`
- Example: `table.Join("LEFT JOIN users ON users.id = orders.user_id")`

### GroupBy
- Add grouping conditions
- Signature: `GroupBy(groupBy string) *table`
- Example: `table.GroupBy("category")`

### Having
- Add group filtering conditions
- Signature: `Having(having string) *table`
- Example: `table.Having("count(*) > 10")`

## Query Methods

### Count
- Get record count
- Signature: `Count() (int64, error)`
- Example: `count, err := table.Count()`

### Find
- Query single record
- Signature: `Find() (map[string]interface{}, error)`
- Example: `record, err := table.Find()`

### FindAll
- Query multiple records
- Signature: `FindAll() ([]map[string]interface{}, error)`
- Example: `records, err := table.FindAll()`

### FindAllWithCursor
- Read data row by row using cursor
- Signature: `FindAllWithCursor(ctx context.Context, handler func(map[string]interface{}) error) error`
- Example:
```go
err := table.FindAllWithCursor(ctx, func(record map[string]interface{}) error {
    // Process each record
    return nil
})
```

## Context Methods

### WithContext
- Set context
- Signature: `WithContext(ctx context.Context) *table`
- Example: `table.WithContext(ctx)`

### FindAllWithContext
- Multi-record query with context
- Signature: `FindAllWithContext(ctx context.Context) ([]map[string]interface{}, error)`
- Example: `records, err := table.FindAllWithContext(ctx)`

## Total Count Control Methods

### HasTotal
- Set whether to get total count
- Signature: `HasTotal(need bool) *table`
- Example: `table.HasTotal(true)`

### GetTotal
- Get total record count
- Signature: `GetTotal() int64`
- Example: `total := table.GetTotal()`

## Data Operation Methods

### Insert
- Insert record
- Signature: `Insert(data interface{}) (lastInsertId int64, err error)`
- Example: `id, err := table.Insert(data)`

### InsertWithContext
- Insert record with context
- Signature: `InsertWithContext(ctx context.Context, data interface{}) (lastInsertId int64, err error)`
- Example: `id, err := table.InsertWithContext(ctx, data)`

### Update
- Update record
- Signature: `Update(data interface{}) (rowsAffected int64, err error)`
- Example: `affected, err := table.Update(data)`

### UpdateWithContext
- Update record with context
- Signature: `UpdateWithContext(ctx context.Context, data interface{}) (rowsAffected int64, err error)`
- Example: `affected, err := table.UpdateWithContext(ctx, data)`

### Delete
- Delete record
- Signature: `Delete() (rowsAffected int64, err error)`
- Example: `affected, err := table.Delete()`

### DeleteWithContext
- Delete record with context
- Signature: `DeleteWithContext(ctx context.Context) (rowsAffected int64, err error)`
- Example: `affected, err := table.DeleteWithContext(ctx)`

## Batch Operation Methods

### BatchInsert
- Batch insert records
- Signature: `BatchInsert(data []map[string]interface{}, batchSize int) (totalAffecteds int64, err error)`
- Example:
```go
users := []map[string]interface{}{
    {"name": "Alice", "age": 25},
    {"name": "Bob", "age": 30},
}
affected, err := table.BatchInsert(users, 100)
```

### BatchUpdate
- Batch update records
- Signature: `BatchUpdate(records []map[string]interface{}, keyField string, batchSize int) (totalAffecteds int64, err error)`
- Example:
```go
users := []map[string]interface{}{
    {"id": 1, "name": "Alice Updated", "age": 26},
    {"id": 2, "name": "Bob Updated", "age": 31},
}
affected, err := table.BatchUpdate(users, "id", 100)
```

## Transaction Methods

### Commit
- Commit transaction
- Signature: `Commit() error`
- Actual Usage:
```go
// Commit current transaction
err := transaction.Commit()
// Successfully committed transaction
```

### Rollback
- Rollback transaction
- Signature: `Rollback() error`
- Actual Usage:
```go
// Rollback current transaction
err := transaction.Rollback()
// Successfully rolled back transaction
```

### DB
- Get database instance
- Signature: `DB() *DB`
- Actual Usage:
```go
// Get the database instance associated with the transaction
db := transaction.DB()
```

## Precautions
- Most methods support method chaining
- Built-in SQL injection protection mechanism
- Supports context and non-context method versions
- Batch operations support large-scale data processing
- Customizable batch size
- Supports flexible data processing
