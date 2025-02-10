package xlorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 条件位标记常量
const (
	condAND uint64 = 1 << iota // AND 条件
	condOR                     // OR 条件
	condNOT                    // NOT 条件
)

// Table 表操作结构体
type Table struct {
	db        *DB
	tableName string
	orderBy   string
	groupBy   string
	having    string
	fields    []string
	where     []string
	joins     []string
	args      []interface{}
	total     int64 // 记录集总数
	limit     int64
	offset    int64
	hasTotal  bool // 是否需要获取总数

	// 新增位运算相关字段
	conditionFlags uint64
	conditionIndex int
}

// Release 释放Table对象到池中
func (t *Table) Release() {
	if t.db.IsDebug() {
		t.db.logger.Debug("释放Table对象", "table", t.tableName)
	}
	t.Reset()
	tablePool.Put(t)
}

// Reset 重置Table对象的状态
func (t *Table) Reset() {
	t.db = nil
	t.tableName = ""
	t.orderBy = ""
	t.limit = 0
	t.offset = 0
	t.fields = nil
	t.groupBy = ""
	t.having = ""
	t.where = nil
	t.args = nil
	t.joins = nil
	t.hasTotal = false
	t.total = 0

	// 重置新增字段
	t.conditionFlags = 0
	t.conditionIndex = 0
}

func (t *Table) WithContext(ctx context.Context) *Table {
	t.db.ctxMu.Lock()
	defer t.db.ctxMu.Unlock()
	t.db.ctx = ctx
	return t
}

// Insert 插入记录
// lastInsertId 返回插入的记录的ID
// err 返回错误信息
func (t *Table) Insert(data interface{}) (lastInsertId int64, err error) {
	return t.insert(context.Background(), data, "INSERT")
}

// InsertWithContext 插入记录
// lastInsertId 返回插入的记录的ID
// err 返回错误信息
func (t *Table) InsertWithContext(ctx context.Context, data interface{}) (lastInsertId int64, err error) {
	return t.insert(ctx, data, "INSERT")
}

// Update 更新记录
func (t *Table) Update(data interface{}) (rowsAffected int64, err error) {
	return t.update(context.Background(), data)
}

// UpdateWithContext 更新记录
func (t *Table) UpdateWithContext(ctx context.Context, data interface{}) (rowsAffected int64, err error) {
	return t.update(ctx, data)
}

// Delete 删除记录
func (t *Table) Delete() (rowsAffected int64, err error) {
	return t.delete(context.Background())
}

// DeleteWithContext 删除记录
func (t *Table) DeleteWithContext(ctx context.Context) (rowsAffected int64, err error) {
	return t.delete(ctx)
}

// Find 查询单条记录
func (t *Table) Find() (map[string]interface{}, error) {
	t.limit = 1
	t.hasTotal = false
	records, err := t.findAllWithContext(context.Background(), "find")
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, sql.ErrNoRows
	}
	return records[0], nil
}

// FindAll 查询多条记录
// 如果之前调用了HasTotal(true)，会先执行一次Count查询获取总数
// 返回值：
//   - []map[string]interface{}: 查询结果集，每个map代表一条记录，key为字段名，value为字段值
//   - error: 如果发生错误，返回具体的错误信息
//
// 可能的错误：
//   - "构建查询SQL失败": 生成SQL语句时发生错误，通常是由于查询条件不正确
//   - "执行查询失败": 执行SQL时发生错误，可能是由于网络问题或SQL语法错误
//   - "获取列信息失败": 无法获取结果集的列信息，可能是由于表结构发生变化
//   - "扫描数据失败": 将数据库返回的数据转换为Go类型时失败
func (t *Table) FindAll() ([]map[string]interface{}, error) {
	return t.findAllWithContext(context.Background(), "findAll")
}

// FindAllWithContext 带上下文的FindAll
func (t *Table) FindAllWithContext(ctx context.Context) ([]map[string]interface{}, error) {
	return t.findAllWithContext(ctx, "findAllWithContext")
}

// FindAllWithCursor 使用游标逐行读取数据，减少内存占用
// handler 是处理每一行记录的回调函数，返回error时会中止处理
func (t *Table) FindAllWithCursor(ctx context.Context, handler func(map[string]interface{}) error) error {
	defer t.Release()
	startTime := time.Now()
	// 如果需要获取总数，先执行 Count 查询
	if t.hasTotal {
		// 创建一个新的Table对象用于Count查询，避免影响当前查询
		countTable := t.db.M(t.tableName)
		// 复制查询条件
		t.copyQueryConditions(countTable)

		// 执行Count查询
		total, err := countTable.Count()
		if err != nil {
			return fmt.Errorf("获取记录总数失败: %v", err)
		}
		t.total = total
	}

	// 构建查询SQL
	query, args := t.buildQuery("SELECT")

	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "findAllWithContext", query, "args", args)
	}

	// 执行查询
	rows, err := t.db.QueryContext(ctx, query, args...)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行查询失败", "findAllWithContext", query, "args", args, "error", err)
		return fmt.Errorf("执行查询失败: %v", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("获取列信息失败", "findAllWithContext", query, "args", args, "error", err)
		return fmt.Errorf("获取列信息失败: %v", err)
	}

	columnsLen := len(columns)

	// 准备扫描缓冲
	values := make([]interface{}, columnsLen)
	scanArgs := make([]interface{}, columnsLen)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 逐行处理
	for rows.Next() {
		// 扫描数据
		if err := rows.Scan(scanArgs...); err != nil {
			t.db.asyncDBMetrics.RecordError()
			t.db.logger.Error("扫描数据失败", "findAllWithContext", query, "args", args, "error", err)
			return fmt.Errorf("扫描数据失败: %v", err)
		}

		// 转换为map
		record := make(map[string]interface{}, columnsLen)
		for i, col := range columns {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				record[col] = string(v)
			default:
				record[col] = v
			}
		}

		// 调用处理函数
		if err := handler(record); err != nil {
			return err // 允许调用方中止处理流程
		}
	}

	// 检查遍历错误
	if err := rows.Err(); err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("遍历结果集失败", "findAllWithContext", query, "args", args, "error", err)
		return fmt.Errorf("遍历结果集失败: %v", err)
	}

	// 记录慢查询
	duration := time.Since(startTime)
	t.db.asyncDBMetrics.RecordQueryDuration("findAllWithContext", duration)

	if duration >= t.db.slowQueryThreshold {
		t.db.asyncDBMetrics.RecordSlowQuery()
		t.db.logger.Warn("慢查询",
			"query", query,
			"args", args,
			"duration", duration.Seconds(),
			"threshold", t.db.slowQueryThreshold,
		)
	}

	return nil
}

// Count 获取记录数
func (t *Table) Count() (int64, error) {
	defer t.Release()
	startTime := time.Now()
	query, args := t.buildQuery("COUNT")
	var count int64
	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "count", query, "args", args)
	}
	err := t.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行查询失败", "count", query, "args", args, "error", err)
		return 0, fmt.Errorf("执行查询失败: %v", err)
	}
	t.db.asyncDBMetrics.RecordQueryDuration("count", time.Since(startTime))
	return count, nil
}

// GetTotal 获取记录集总数
// 仅当在执行FindAll之前调用HasTotal(true)时才会返回有效值否则返回0
func (t *Table) GetTotal() int64 {
	if !t.hasTotal {
		return 0
	}
	return t.total
}

// GetWhere 获取WHERE子句
func (t *Table) GetWhere(addPreStr bool) (string, []interface{}) {
	// 添加条件
	if len(t.where) > 0 {
		// 预估SQL长度，避免频繁扩容
		query := strings.Builder{}
		query.Grow(256)

		if addPreStr {
			query.WriteString(" WHERE ")
		}

		// 使用位运算快速判断条件类型
		switch {
		case t.conditionFlags&condOR != 0:
			// 存在 OR 条件，使用括号确保正确性
			query.WriteByte('(')
			for i, condition := range t.where {
				if i > 0 {
					query.WriteString(" OR ")
				}
				query.WriteString(condition)
			}
			query.WriteByte(')')

		case t.conditionFlags&condNOT != 0:
			// 存在 NOT 条件，使用括号确保正确性
			query.WriteByte('(')
			for i, condition := range t.where {
				if i > 0 {
					query.WriteString(" AND ")
				}
				query.WriteString(condition)
			}
			query.WriteByte(')')

		default:
			// 纯 AND 条件，直接连接
			query.WriteString(strings.Join(t.where, " AND "))
		}
		// 重置条件标志（重要！）
		t.conditionFlags = 0
		t.conditionIndex = 0
		return query.String(), t.args
	}

	return "", nil
}

// Where 添加查询条件
func (t *Table) Where(condition string, args ...interface{}) *Table {
	if condition == "" {
		return t
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		t.db.logger.Error("条件参数数量不匹配",
			"condition", condition,
			"args_count", len(args),
		)
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(condition, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "condition", condition)
		return t
	}

	t.where = append(t.where, condition)
	t.args = append(t.args, args...)

	// 更新位标记和索引
	if t.conditionIndex == 0 {
		t.conditionFlags |= condAND // 第一个条件默认为 AND
	}
	t.conditionIndex++

	return t
}

// OrWhere 添加 OR 查询条件
func (t *Table) OrWhere(condition string, args ...interface{}) *Table {
	if condition == "" {
		return t
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		t.db.logger.Error("条件参数数量不匹配",
			"condition", condition,
			"args_count", len(args),
		)
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(condition, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "condition", condition)
		return t
	}

	t.where = append(t.where, condition)
	t.args = append(t.args, args...)

	// 更新位标记和索引
	t.conditionFlags |= condOR
	t.conditionIndex++

	return t
}

// NotWhere 添加 NOT 查询条件
func (t *Table) NotWhere(condition string, args ...interface{}) *Table {
	if condition == "" {
		return t
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		t.db.logger.Error("条件参数数量不匹配",
			"condition", condition,
			"args_count", len(args),
		)
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(condition, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "condition", condition)
		return t
	}

	// 为 NOT 条件添加 NOT 前缀
	notCondition := "NOT (" + condition + ")"
	t.where = append(t.where, notCondition)
	t.args = append(t.args, args...)

	// 更新位标记和索引
	t.conditionFlags |= condNOT
	t.conditionIndex++

	return t
}

// OrderBy 添加排序条件
func (t *Table) OrderBy(order string) *Table {
	if order == "" {
		return t
	}
	if !isValidSafeOrderBy(order) {
		t.db.logger.Error("非法排序字段", "order", order)
		return t
	}
	// 检查SQL注入
	if strings.ContainsAny(order, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "order", order)
		return t
	}

	t.orderBy = order
	return t
}

// Limit 添加限制条件
func (t *Table) Limit(limit int64) *Table {
	if limit < 0 {
		t.db.logger.Error("limit不能为负数", "limit", limit)
		return t
	}
	t.limit = limit
	return t
}

// Page 设置分页
func (t *Table) Page(page, pageSize int64) *Table {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	t.limit = pageSize
	t.offset = (page - 1) * pageSize
	return t
}

// Offset 添加偏移量
func (t *Table) Offset(offset int64) *Table {
	if offset < 0 {
		t.db.logger.Error("offset不能为负数", "offset", offset)
		return t
	}
	t.offset = offset
	return t
}

// Fields 设置查询字段
func (t *Table) Fields(fields ...string) *Table {
	if len(fields) == 0 {
		return t
	}

	for _, field := range fields {
		if field == "" {
			continue
		}
		// 检查SQL注入
		if !isValidFieldName(field) {
			t.db.logger.Error("检测到可能的SQL注入尝试", "field", field)
			return t
		}
		t.fields = append(t.fields, field)
	}
	return t
}

// Join 添加表连接
func (t *Table) Join(join string) *Table {
	if join == "" {
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(join, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "join", join)
		return t
	}

	t.joins = append(t.joins, join)
	return t
}

// GroupBy 添加分组条件
func (t *Table) GroupBy(groupBy string) *Table {
	if groupBy == "" {
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(groupBy, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "groupBy", groupBy)
		return t
	}

	t.groupBy = groupBy
	return t
}

// Having 添加分组过滤条件
func (t *Table) Having(having string) *Table {
	if having == "" {
		return t
	}

	// 检查SQL注入
	if strings.ContainsAny(having, ";\x00") {
		t.db.logger.Error("检测到可能的SQL注入尝试", "having", having)
		return t
	}

	t.having = having
	return t
}

// HasTotal 设置是否需要获取总数
// 当设置为true时，在执行FindAll时会自动执行一次Count查询获取符合条件的记录总数
// 可以通过GetTotal方法获取查询结果
func (t *Table) HasTotal(need bool) *Table {
	t.hasTotal = need
	return t
}

// findAllWithContext 实际执行带上下文的FindAll
func (t *Table) findAllWithContext(ctx context.Context, findType string) ([]map[string]interface{}, error) {
	defer t.Release()
	startTime := time.Now()
	if findType == "" {
		findType = "findAllWithContext"
	}
	// 如果需要获取总数，先执行 Count 查询
	if t.hasTotal {
		// 创建一个新的Table对象用于Count查询，避免影响当前查询
		countTable := t.db.M(t.tableName)
		// 复制查询条件
		t.copyQueryConditions(countTable)

		// 执行Count查询
		total, err := countTable.Count()
		if err != nil {
			return nil, fmt.Errorf("获取记录总数失败: %v", err)
		}
		t.total = total
	}

	// 构建查询SQL
	query, args := t.buildQuery("SELECT")

	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", findType, query, "args", args)
	}

	// 执行查询
	rows, err := t.db.QueryContext(ctx, query, args...)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行查询失败", findType, query, "args", args, "error", err)
		return nil, fmt.Errorf("执行查询失败: %v", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("获取列信息失败", findType, query, "args", args, "error", err)
		return nil, fmt.Errorf("获取列信息失败: %v", err)
	}

	columnsLen := len(columns)

	// 预分配结果集切片，减少扩容
	var results []map[string]interface{}
	if t.limit > 0 {
		results = make([]map[string]interface{}, 0, t.limit)
	} else {
		// 如果没有limit，给一个合理的初始容量
		results = make([]map[string]interface{}, 0, 64)
	}

	// 准备扫描目标
	values := make([]interface{}, columnsLen)
	scanArgs := make([]interface{}, columnsLen)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 扫描结果
	for rows.Next() {
		// 扫描数据
		if err := rows.Scan(scanArgs...); err != nil {
			t.db.asyncDBMetrics.RecordError()
			t.db.logger.Error("扫描数据失败", findType, query, "args", args, "error", err)
			return nil, fmt.Errorf("扫描数据失败: %v", err)
		}

		row := make(map[string]interface{}, columnsLen)
		for i, col := range columns {
			val := values[i]
			if val == nil {
				row[col] = nil
				continue
			}

			// 处理特殊类型
			switch v := val.(type) {
			case []byte:
				// 尝试将[]byte转换为字符串
				row[col] = string(v)
			default:
				row[col] = v
			}
		}

		results = append(results, row)
	}

	// 检查遍历错误
	if err = rows.Err(); err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("遍历结果集失败", findType, query, "args", args, "error", err)
		return nil, fmt.Errorf("遍历结果集失败: %v", err)
	}

	// 记录慢查询
	duration := time.Since(startTime)

	// 记录查询耗时
	t.db.asyncDBMetrics.RecordQueryDuration(findType, duration)

	if duration >= t.db.slowQueryThreshold {
		t.db.asyncDBMetrics.RecordSlowQuery()
		t.db.logger.Warn("慢查询",
			"query", query,
			"args", args,
			"duration", duration.Seconds(),
			"threshold", t.db.slowQueryThreshold,
			"rows", len(results),
		)
	}

	return results, nil
}

// insert 内部插入方法
func (t *Table) insert(ctx context.Context, data interface{}, insertType string) (int64, error) {
	defer t.Release()
	startTime := time.Now()
	fields, values, err := t.extractFieldsAndValues(data)
	if err != nil {
		return 0, err
	}

	if len(fields) == 0 {
		return 0, errors.New("插入的数据不能为空，字段名为空")
	}

	query, err := t.buildInsertSQL(insertType, fields)
	if err != nil {
		return 0, err
	}

	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "insert", query, "args", values)
	}

	// 执行SQL
	result, err := t.db.ExecContext(ctx, query, values...)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行SQL失败", "insert", query, "args", values, "error", err)
		return 0, err
	}

	// 获取最后插入的ID
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	t.db.asyncDBMetrics.RecordQueryDuration("insert", time.Since(startTime))
	return lastInsertId, nil
}

func (t *Table) update(ctx context.Context, data interface{}) (int64, error) {
	defer t.Release()
	startTime := time.Now()
	fields, values, err := t.extractFieldsAndValues(data)
	if err != nil {
		return 0, err
	}

	// 构建SQL语句
	query, whereArgs, err := t.buildUpdateSQL(fields)
	if err != nil {
		return 0, err
	}

	// 合并参数
	args := append(values, whereArgs...)

	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "update", query, "args", args)
	}

	// 执行SQL
	result, err := t.db.ExecContext(ctx, query, args...)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行SQL失败", "update", query, "args", args, "error", err)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	if t.db.IsDebug() {
		t.db.logger.Debug("更新操作结果", "rowsAffected", rowsAffected)
	}

	t.db.asyncDBMetrics.RecordQueryDuration("update", time.Since(startTime))
	return rowsAffected, nil
}

func (t *Table) delete(ctx context.Context) (int64, error) {
	defer t.Release()
	startTime := time.Now()
	query, args := t.buildQuery("DELETE")
	if query == "" || args == nil {
		return 0, errors.New("构建查询语句失败，查询语句或参数为空")
	}
	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "delete", query, "args", args)
	}
	// 执行SQL
	result, err := t.db.ExecContext(ctx, query, args...)
	if err != nil {
		t.db.asyncDBMetrics.RecordError()
		t.db.logger.Error("执行SQL失败", "delete", query, "args", args, "error", err)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	if t.db.IsDebug() {
		t.db.logger.Debug("删除操作结果", "rowsAffected", rowsAffected)
	}
	t.db.asyncDBMetrics.RecordQueryDuration("delete", time.Since(startTime))
	return rowsAffected, nil
}

// buildPlaceholders 构建占位符
func (t *Table) buildPlaceholders(fieldCount, recordCount int) []string {
	// 2. 直接创建目标切片
	placeholders := make([]string, recordCount)

	// 3. 并行填充（Go 1.21+ 新增清析语法）
	clear(placeholders) // 显式初始化（非必须）

	// 4. 内存预分配优化
	if recordCount > 0 {
		placeholders[0] = getCachedPlaceholder(fieldCount, t.db.placeholderCache) //生成带括号的单记录占位符
		for i := 1; i < recordCount; i *= 2 {
			copy(placeholders[i:], placeholders[:i])
		}
	}

	return placeholders
}

// copyQueryConditions 复制查询条件到目标Table对象
// 用于在不影响原查询的情况下执行Count等操作
func (t *Table) copyQueryConditions(target *Table) {
	if len(t.where) > 0 {
		target.where = make([]string, len(t.where))
		copy(target.where, t.where)
	}

	if len(t.args) > 0 {
		target.args = make([]interface{}, len(t.args))
		copy(target.args, t.args)
	}

	if len(t.joins) > 0 {
		target.joins = make([]string, len(t.joins))
		copy(target.joins, t.joins)
	}

	target.groupBy = t.groupBy
	target.having = t.having
}

// extractFieldsAndValues 提取字段和值
func (t *Table) extractFieldsAndValues(data interface{}) ([]string, []interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return extractFromMap(v)
	case []map[string]interface{}:
		return extractFromMapSlice(v)
	default:
		// 使用增强版StructToMap处理结构体
		m, err := t.db.StructMapper.StructToMap(data)
		if err != nil {
			return nil, nil, err
		}
		return extractFromMap(m)
	}
}

// buildQuery 构建查询语句
func (t *Table) buildQuery(queryType string) (string, []interface{}) {
	// 预估SQL长度，避免频繁扩容
	query := strings.Builder{}
	query.Grow(256)

	var args []interface{}

	// 构建基础查询
	switch queryType {
	case "SELECT":
		query.WriteString("SELECT ")
		if len(t.fields) > 0 {
			query.WriteString("`")
			query.WriteString(strings.Join(t.fields, "`, `"))
			query.WriteString("`")
		} else {
			query.WriteByte('*')
		}
		query.WriteString(" FROM ")
		query.WriteString(t.tableName)

	case "COUNT":
		query.WriteString("SELECT COUNT(*) FROM ")
		query.WriteString(t.tableName)

	case "DELETE":
		query.WriteString("DELETE FROM ")
		query.WriteString(t.tableName)

	default:
		t.db.logger.Error("不支持的查询类型", "type", queryType)
		return "", nil
	}

	// 添加连接
	if len(t.joins) > 0 {
		for _, join := range t.joins {
			query.WriteByte(' ')
			query.WriteString(join)
		}
	}

	// 添加条件
	if len(t.where) > 0 {
		whereString, whereArgs := t.GetWhere(true)
		if whereString != "" {
			args = make([]interface{}, 0, len(whereArgs))
			query.WriteString(whereString)
			args = append(args, whereArgs...)
		}
	}

	// 添加分组
	if t.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(t.groupBy)

		if t.having != "" {
			query.WriteString(" HAVING ")
			query.WriteString(t.having)
		}
	}

	// 添加排序
	if t.orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(t.orderBy)
	}

	// 添加限制和偏移
	if t.limit > 0 {
		query.WriteString(" LIMIT ")
		query.WriteString(strconv.FormatInt(t.limit, 10))

		if t.offset > 0 {
			query.WriteString(" OFFSET ")
			query.WriteString(strconv.FormatInt(t.offset, 10))
		}
	}

	return query.String(), args
}

// 生成插入SQL语句
func (t *Table) buildInsertSQL(insertType string, fields []string) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("插入的数据不能为空")
	}
	// 构建插入SQL语句
	var sql strings.Builder
	sql.WriteString(insertType)
	sql.WriteString(" INTO ")
	sql.WriteString(t.tableName)
	sql.WriteString(" (`")
	sql.WriteString(strings.Join(fields, "`,`"))
	sql.WriteString("`) VALUES ")
	sql.WriteString(strings.Join(t.buildPlaceholders(len(fields), 1), ","))
	return sql.String(), nil
}

// buildUpdateSQL 构建更新SQL语句
func (t *Table) buildUpdateSQL(fields []string) (string, []interface{}, error) {

	if len(fields) == 0 {
		return "", nil, fmt.Errorf("更新操作必须指定字段")
	}

	whereClause, whereArgs := t.GetWhere(true)
	if whereClause == "" {
		t.db.logger.Warn("更新操作未指定 WHERE 条件，拒绝执行")
		return "", nil, fmt.Errorf("更新操作必须指定 WHERE 条件")
	}

	// 构建SET子句
	var clause strings.Builder
	for _, field := range fields {
		clause.WriteString("`")
		clause.WriteString(field)
		clause.WriteString("` = ?,")
	}

	var sql strings.Builder
	sql.WriteString("UPDATE ")
	sql.WriteString(t.tableName)
	sql.WriteString(" SET ")
	sql.WriteString(strings.TrimSuffix(clause.String(), ","))
	sql.WriteString(whereClause)
	return sql.String(), whereArgs, nil
}
