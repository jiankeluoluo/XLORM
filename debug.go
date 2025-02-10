package xlorm

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GetQuerySQL 获取最终生成的SQL语句和参数
// 这个方法可以帮助开发者查看实际的SQL查询，便于调试和日志记录
func (t *Table) GetQuerySQL(queryType string) (string, []interface{}) {
	// 克隆当前的 Table 对象，避免修改原始对象
	clonedTable := *t

	// 构建查询语句
	query, args := clonedTable.buildQuery(queryType)

	return query, args
}

// FormatQuerySQL 格式化SQL语句，将参数替换到查询中
// 注意：这个方法仅用于调试，不应用于实际的数据库查询，因为存在SQL注入风险
func (t *Table) FormatQuerySQL(queryType string) string {
	query, args := t.GetQuerySQL(queryType)

	// 使用正则表达式替换 ? 占位符
	re := regexp.MustCompile(`\?`)

	// 替换参数
	formattedQuery := re.ReplaceAllStringFunc(query, func(_ string) string {
		if len(args) == 0 {
			return "NULL"
		}

		arg := args[0]
		args = args[1:]

		switch v := arg.(type) {
		case string:
			return fmt.Sprintf("'%s'", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v)
		case float32, float64:
			return fmt.Sprintf("%f", v)
		case bool:
			if v {
				return "TRUE"
			}
			return "FALSE"
		case time.Time:
			return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
		case nil:
			return "NULL"
		default:
			return fmt.Sprintf("'%v'", v)
		}
	})

	return formattedQuery
}

// PrintQuerySQL 打印SQL语句到日志
func (t *Table) PrintQuerySQL(queryType string) {
	query := t.FormatQuerySQL(queryType)
	t.db.logger.Info("生成的SQL语句", "sql", query)
}

// GetInsertSQL 获取插入语句的SQL
func (t *Table) GetInsertSQL(data interface{}) (string, []interface{}, error) {
	// 提取字段和值
	fields, values, err := t.extractFieldsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	// 构建插入语句
	query, err := t.buildInsertSQL("INSERT", fields)
	if err != nil {
		return "", nil, err
	}

	return query, values, nil
}

// FormatInsertSQL 格式化插入SQL语句
func (t *Table) FormatInsertSQL(data interface{}) (string, error) {
	query, args, err := t.GetInsertSQL(data)
	if err != nil {
		return "", err
	}

	return t.formatSQL(query, args), nil
}

// PrintInsertSQL 打印插入SQL语句到日志
func (t *Table) PrintInsertSQL(data interface{}) error {
	query, err := t.FormatInsertSQL(data)
	if err != nil {
		return err
	}
	t.db.logger.Info("生成的插入SQL语句", "sql", query)
	fmt.Println("生成的插入SQL语句:", query)
	return nil
}

// GetBatchInsertSQL 获取批量插入语句的SQL
func (t *Table) GetBatchInsertSQL(data []interface{}) (string, []interface{}, error) {
	if len(data) == 0 {
		return "", nil, nil
	}

	// 提取字段和值
	fields, _, err := t.extractFieldsAndValues(data[0])
	if err != nil {
		return "", nil, err
	}

	// 构建批量插入语句
	query := strings.Builder{}
	query.WriteString("INSERT INTO ")
	query.WriteString(t.tableName)
	query.WriteString(" (")
	query.WriteString(strings.Join(fields, ", "))
	query.WriteString(") VALUES ")

	// 生成批量插入的占位符
	placeholders := make([]string, len(data))
	allValues := make([]interface{}, 0, len(data)*len(fields))

	for i, item := range data {
		_, itemValues, err := t.extractFieldsAndValues(item)
		if err != nil {
			return "", nil, err
		}

		// 生成每条记录的占位符
		placeholderGroup := make([]string, len(fields))
		for j := range placeholderGroup {
			placeholderGroup[j] = "?"
		}
		placeholders[i] = "(" + strings.Join(placeholderGroup, ", ") + ")"

		allValues = append(allValues, itemValues...)
	}

	query.WriteString(strings.Join(placeholders, ", "))

	return query.String(), allValues, nil
}

// FormatBatchInsertSQL 格式化批量插入SQL语句
func (t *Table) FormatBatchInsertSQL(data []interface{}) (string, error) {
	query, args, err := t.GetBatchInsertSQL(data)
	if err != nil {
		return "", err
	}

	return t.formatSQL(query, args), nil
}

// PrintBatchInsertSQL 打印批量插入SQL语句到日志
func (t *Table) PrintBatchInsertSQL(data []interface{}) error {
	query, err := t.FormatBatchInsertSQL(data)
	if err != nil {
		return err
	}
	t.db.logger.Info("生成的批量插入SQL语句", "sql", query)
	fmt.Println("生成的批量插入SQL语句:", query)
	return nil
}

// GetUpdateSQL 获取更新语句的SQL
func (t *Table) GetUpdateSQL(data interface{}) (string, []interface{}, error) {
	// 提取字段和值
	fields, values, err := t.extractFieldsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	// 构建SQL语句
	query, whereArgs, err := t.buildUpdateSQL(fields)
	if err != nil {
		return "", nil, err
	}

	// 合并参数
	args := append(values, whereArgs...)

	return query, args, nil
}

// FormatUpdateSQL 格式化更新SQL语句
func (t *Table) FormatUpdateSQL(data interface{}) (string, error) {
	query, args, err := t.GetUpdateSQL(data)
	if err != nil {
		return "", err
	}

	return t.formatSQL(query, args), nil
}

// PrintUpdateSQL 打印更新SQL语句到日志
func (t *Table) PrintUpdateSQL(data interface{}) error {
	query, err := t.FormatUpdateSQL(data)
	if err != nil {
		return err
	}
	t.db.logger.Info("生成的更新SQL语句", "sql", query)
	fmt.Println("生成的更新SQL语句:", query)
	return nil
}

// formatSQL 内部方法，用于格式化SQL语句（复用之前的格式化逻辑）
func (t *Table) formatSQL(query string, args []interface{}) string {
	// 使用正则表达式替换 ? 占位符
	re := regexp.MustCompile(`\?`)

	// 替换参数
	formattedQuery := re.ReplaceAllStringFunc(query, func(_ string) string {
		if len(args) == 0 {
			return "NULL"
		}

		arg := args[0]
		args = args[1:]

		switch v := arg.(type) {
		case string:
			return fmt.Sprintf("'%s'", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v)
		case float32, float64:
			return fmt.Sprintf("%f", v)
		case bool:
			if v {
				return "TRUE"
			}
			return "FALSE"
		case time.Time:
			return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
		case nil:
			return "NULL"
		default:
			return fmt.Sprintf("'%v'", v)
		}
	})

	return formattedQuery
}
