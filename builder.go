package xlorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// builder SQL查询构建器结构体
type builder struct {
	groupBy   string        // GROUP BY 子句
	having    string        // HAVING 子句
	orderBy   string        // ORDER BY 子句
	table     string        // 表名
	fields    []string      // 字段列表
	where     []string      // WHERE 条件
	joins     []string      // JOIN 子句
	args      []interface{} // 查询参数
	limit     int64         // 查询限制
	offset    int64         // 查询偏移
	forUpdate bool          // 是否为 FOR UPDATE 查询
	errs      []error       // 错误列表

	// 新增位运算相关字段
	conditionFlags uint64
	conditionIndex int
}

// NewBuilder 创建查询构建器
func (db *DB) NewBuilder(table string) *builder {
	b := builderPool.Get().(*builder)
	b.Reset()
	if table == "" {
		b.errs = append(b.errs, errors.New("table名称不能为空"))
		return b
	}
	// 检查SQL注入风险
	if strings.ContainsAny(table, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("table检测到可能的SQL注入尝试: %s", table))
		return b
	}
	b.table = table
	return b
}

// 重置查询构建器
func (b *builder) Reset() *builder {
	b.table = ""
	b.fields = nil
	b.where = nil
	b.args = nil
	b.joins = nil
	b.groupBy = ""
	b.having = ""
	b.orderBy = ""
	b.limit = 0
	b.offset = 0
	b.forUpdate = false
	b.errs = nil
	b.conditionFlags = 0
	b.conditionIndex = 0
	return b
}

// Fields 设置查询字段
func (b *builder) Fields(fields ...string) *builder {
	if len(fields) == 0 {
		return b
	}
	for _, field := range fields {
		if field == "" {
			continue
		}
		// 检查SQL注入
		if !isValidFieldName(field) {
			b.errs = append(b.errs, fmt.Errorf("fields包含非法字符: %s", field))
			continue
		}
		b.fields = append(b.fields, field)
	}
	return b
}

// Where 添加查询条件
func (b *builder) Where(condition string, args ...interface{}) *builder {
	if condition == "" {
		return b
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		b.errs = append(b.errs, fmt.Errorf("where条件参数数量不匹配: condition:%s,args_count:%d", condition, len(args)))
		return b
	}

	// 检查SQL注入
	if strings.ContainsAny(condition, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("where检测到可能的SQL注入尝试: condition:%s", condition))
		return b
	}

	b.where = append(b.where, condition)
	b.args = append(b.args, args...)

	// 更新位标记和索引
	if b.conditionIndex == 0 {
		b.conditionFlags |= condAND // 第一个条件默认为 AND
	}
	b.conditionIndex++

	return b
}

// OrWhere 添加 OR 查询条件
func (b *builder) OrWhere(condition string, args ...interface{}) *builder {
	if condition == "" {
		return b
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		b.errs = append(b.errs, fmt.Errorf("OrWhere条件参数数量不匹配: condition:%s,args_count:%d", condition, len(args)))
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(condition, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("OrWhere检测到可能的SQL注入尝试: %s", condition))
		return b
	}

	b.where = append(b.where, fmt.Sprintf("OR %s", condition))
	b.args = append(b.args, args...)

	// 更新位标记和索引
	b.conditionFlags |= condOR
	b.conditionIndex++
	return b
}

// NotWhere 添加 NOT 查询条件
func (b *builder) NotWhere(condition string, args ...interface{}) *builder {
	if condition == "" {
		return b
	}

	// 增强校验：检查是否有未参数化的值
	if strings.Count(condition, "?") != len(args) {
		b.errs = append(b.errs, fmt.Errorf("NotWhere条件参数数量不匹配: condition:%s,args_count:%d", condition, len(args)))
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(condition, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("NotWhere检测到可能的SQL注入尝试: %s", condition))
		return b
	}

	// 为 NOT 条件添加 NOT 前缀
	notCondition := "NOT (" + condition + ")"
	b.where = append(b.where, notCondition)
	b.args = append(b.args, args...)
	// 更新位标记和索引
	b.conditionFlags |= condNOT
	b.conditionIndex++

	return b
}

// Join 添加表连接
func (b *builder) Join(join string) *builder {
	if join == "" {
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(join, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("Join检测到可能的SQL注入尝试: %s", join))
		return b
	}

	b.joins = append(b.joins, join)
	return b
}

// GroupBy 添加分组条件
func (b *builder) GroupBy(groupBy string) *builder {
	if groupBy == "" {
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(groupBy, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("GroupBy检测到可能的SQL注入尝试: %s", groupBy))
		return b
	}

	b.groupBy = groupBy
	return b
}

// Having 添加分组过滤条件
func (b *builder) Having(having string) *builder {
	if having == "" {
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(having, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("Having检测到可能的SQL注入尝试: %s", having))
		return b
	}

	b.having = having
	return b
}

// OrderBy 添加排序条件
func (b *builder) OrderBy(order string) *builder {
	if order == "" {
		return b
	}

	if !isValidSafeOrderBy(order) {
		b.errs = append(b.errs, fmt.Errorf("OrderBy检测到不可用的排序字段: %s", order))
		return b
	}

	// 检查SQL注入风险
	if strings.ContainsAny(order, ";\x00") {
		b.errs = append(b.errs, fmt.Errorf("OrderBy检测到可能的SQL注入尝试: %s", order))
		return b
	}

	b.orderBy = order
	return b
}

// Limit 添加限制条件
func (b *builder) Limit(limit int64) *builder {
	if limit <= 0 {
		return b
	}

	b.limit = limit
	return b
}

// Offset 添加偏移条件
func (b *builder) Offset(offset int64) *builder {
	if offset < 0 {
		return b
	}

	b.offset = offset
	return b
}

// ForUpdate 设置是否为 FOR UPDATE 查询
func (b *builder) ForUpdate(forUpdate bool) *builder {
	b.forUpdate = forUpdate
	return b
}

// Page 设置分页
func (b *builder) Page(page, pageSize int64) *builder {
	if page <= 0 || pageSize <= 0 {
		b.errs = append(b.errs, fmt.Errorf("page和pageSize必须为正数: page=%d, pageSize=%d", page, pageSize))
		return b
	}
	b.limit = pageSize
	b.offset = (page - 1) * pageSize
	return b
}

// Build 构建SQL语句
func (b *builder) Build() (string, []interface{}, error) {
	defer b.ReleaseBuilder()
	var query strings.Builder
	query.WriteString("SELECT ")

	// 处理字段
	if len(b.fields) == 0 {
		query.WriteString("*")
	} else {
		query.WriteString("`")
		query.WriteString(strings.Join(b.fields, "`, `"))
		query.WriteString("`")
	}

	// 添加表名
	query.WriteString(" FROM ")
	query.WriteString(b.table)

	// 添加连接
	if len(b.joins) > 0 {
		query.WriteByte(' ')
		query.WriteString(strings.Join(b.joins, " "))
	}

	// 添加条件
	if len(b.where) > 0 {
		whereString, _ := b.GetWhere(true)
		if whereString != "" {
			query.WriteString(whereString)
		}
	}

	// 添加分组
	if b.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(b.groupBy)
	}

	// 添加分组条件
	if b.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(b.having)
	}

	// 添加排序
	if b.orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(b.orderBy)
	}

	// 添加限制
	if b.limit > 0 {
		query.WriteString(" LIMIT ")
		query.WriteString(strconv.FormatInt(b.limit, 10))
	}

	// 添加偏移
	if b.offset > 0 {
		query.WriteString(" OFFSET ")
		query.WriteString(strconv.FormatInt(b.offset, 10))
	}

	// 添加行锁
	if b.forUpdate {
		query.WriteString(" FOR UPDATE")
	}

	return query.String(), b.args, errors.Join(b.errs...)
}

// GetWhere 获取WHERE子句
func (b *builder) GetWhere(addPreStr bool) (string, []interface{}) {
	// 添加条件
	if len(b.where) > 0 {
		// 预估SQL长度，避免频繁扩容
		query := strings.Builder{}
		query.Grow(256)

		if addPreStr {
			query.WriteString(" WHERE ")
		}

		// 使用位运算快速判断条件类型
		switch {
		case b.conditionFlags&condOR != 0:
			// 存在 OR 条件，使用括号确保正确性
			query.WriteByte('(')
			for i, condition := range b.where {
				if i > 0 {
					query.WriteString(" OR ")
				}
				query.WriteString(condition)
			}
			query.WriteByte(')')

		case b.conditionFlags&condNOT != 0:
			// 存在 NOT 条件，使用括号确保正确性
			query.WriteByte('(')
			for i, condition := range b.where {
				if i > 0 {
					query.WriteString(" AND ")
				}
				query.WriteString(condition)
			}
			query.WriteByte(')')

		default:
			// 纯 AND 条件，直接连接
			query.WriteString(strings.Join(b.where, " AND "))
		}
		// 重置条件标志（重要！）
		b.conditionFlags = 0
		b.conditionIndex = 0
		return query.String(), b.args
	}

	return "", nil
}

// ReleaseBuilder 手动释放Builder对象到池中
// 注意：Build方法已经内置了释放Builder对象到池中的功能
func (b *builder) ReleaseBuilder() {
	b.Reset()
	builderPool.Put(b)
}
