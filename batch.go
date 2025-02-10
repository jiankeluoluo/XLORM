package xlorm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	defaultBatchSize = 1000
)

// BatchInsert 批量插入数据，使用事务确保原子性和性能
// data 批量插入的数据
// batchSize 单词批量插入的数据量，默认：1000
// totalAffecteds 返回影响的行数
// err 返回错误信息
func (t *Table) BatchInsert(data []map[string]interface{}, batchSize int) (totalAffecteds int64, err error) {
	if batchSize == 0 {
		batchSize = defaultBatchSize
	}
	dataLen := len(data)
	// 检查数据是否为空
	if dataLen == 0 {
		return 0, nil
	}

	// 记录开始时间
	startTime := time.Now()

	// 开启单个事务
	tx, err := t.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 预校验字段
	firstBatchEnd := batchSize
	if firstBatchEnd > dataLen {
		firstBatchEnd = dataLen
	}
	checkFields, err := t.extractBatchFields(data[0:firstBatchEnd])
	if err != nil {
		return 0, err
	}
	checkFieldsLen := len(checkFields)

	// 预计算参数总容量
	fieldCount := len(checkFields)
	totalArgs := dataLen * fieldCount
	args := make([]interface{}, 0, totalArgs)

	// 预生成占位符
	placeholder := getCachedPlaceholder(fieldCount, t.db.placeholderCache)

	// 构建基础SQL
	baseQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ",
		t.tableName,
		strings.Join(checkFields, ", "),
	)

	var totalAffected int64

	if t.db.IsDebug() {
		t.db.logger.Debug("批量插入开始",
			"table", t.tableName,
			"SQL", baseQuery,
			"count", dataLen,
			"batchSize", batchSize,
		)
	}

	// 分批处理
	for i := 0; i < dataLen; i += batchSize {
		end := i + batchSize
		if end > dataLen {
			end = dataLen
		}
		batchData := data[i:end]

		// 快速校验字段数量
		if len(batchData[0]) != checkFieldsLen {
			return totalAffected, errors.New("字段数量不匹配")
		}

		// 构建当前批次的占位符
		placeholders := make([]string, len(batchData))
		for j := range placeholders {
			placeholders[j] = placeholder
		}

		// 填充参数
		for _, item := range batchData {
			for _, field := range checkFields {
				cleanField := strings.Trim(field, "`")
				args = append(args, item[cleanField])
			}
		}

		// 执行批次插入
		query := baseQuery + strings.Join(placeholders, ",")
		result, err := tx.Exec(query, args...)
		if err != nil {
			t.db.logger.Error("批量插入失败",
				"batchStart", i,
				"batchEnd", end,
				"error", err,
			)
			t.db.asyncDBMetrics.RecordError()
			return totalAffected, fmt.Errorf("批次插入失败: %v", err)
		}

		// 更新影响行数
		rowsAffected, _ := result.RowsAffected()
		totalAffected += rowsAffected
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return totalAffected, fmt.Errorf("提交事务失败: %v", err)
	}

	// 记录性能指标
	duration := time.Since(startTime)
	t.db.asyncDBMetrics.RecordQueryDuration("batch_insert", duration)
	t.db.asyncDBMetrics.RecordAffectedRows(totalAffected)

	if t.db.IsDebug() {
		t.db.logger.Debug("批量插入完成",
			"table", t.tableName,
			"affected", totalAffected,
			"duration", duration.Seconds(),
		)
	}

	return totalAffected, nil

}

// BatchUpdate 批量更新数据
// 返回更新的行数和错误
func (t *Table) BatchUpdate(records []map[string]interface{}, keyField string, batchSize int) (totalAffecteds int64, err error) {
	if batchSize == 0 {
		batchSize = defaultBatchSize
	}
	recordsLen := len(records)
	if recordsLen == 0 {
		return 0, nil
	}
	if keyField == "" {
		return 0, errors.New("必须指定主键字段")
	}

	startTime := time.Now()
	if t.db.IsDebug() {
		t.db.logger.Debug("开始批量更新",
			"table", t.tableName,
			"count", recordsLen,
		)
	}
	// 开启事务
	tx, err := t.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		} else if err != nil {
			tx.Rollback()
		}
	}()

	var totalAffected int64
	for i := 0; i < recordsLen; i += batchSize {
		end := i + batchSize
		if end > recordsLen {
			end = recordsLen
		}

		batch := records[i:end]
		affected, err := t.updateBatch(tx, batch, keyField)
		if err != nil {
			return totalAffected, err
		}
		totalAffected += affected
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return totalAffected, fmt.Errorf("提交事务失败: %v", err)
	}

	duration := time.Since(startTime)
	// 记录性能指标
	t.db.asyncDBMetrics.RecordQueryDuration("batch_update", duration)
	t.db.asyncDBMetrics.RecordAffectedRows(totalAffected)

	if t.db.IsDebug() {
		t.db.logger.Info("批量更新完成",
			"table", t.tableName,
			"count", recordsLen,
			"affected", totalAffected,
			"duration", duration.Seconds(),
		)
	}
	return totalAffected, nil
}

// updateBatch 更新一批数据
func (t *Table) updateBatch(tx *Transaction, records []map[string]interface{}, keyField string) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}

	// 提取更新字段
	var updateFields []string
	for field := range records[0] {
		if field != keyField {
			updateFields = append(updateFields, field)
		}
	}
	if len(updateFields) == 0 {
		return 0, errors.New("没有要更新的字段")
	}

	// 构建CASE语句
	var query strings.Builder
	query.WriteString("UPDATE")
	query.WriteString(t.tableName)
	query.WriteString(" SET ")

	var args []interface{}
	for i, field := range updateFields {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString("`")
		query.WriteString(field)
		query.WriteString("` = CASE `")
		query.WriteString(keyField)
		query.WriteString("`")

		for _, record := range records {
			keyValue, ok := record[keyField]
			if !ok {
				return 0, fmt.Errorf("记录缺少主键字段: %s", keyField)
			}

			value, ok := record[field]
			if !ok {
				return 0, fmt.Errorf("记录缺少更新字段: %s", field)
			}

			query.WriteString(" WHEN ? THEN ? ")
			args = append(args, keyValue, value)
		}
		query.WriteString(" END")
	}

	// 添加WHERE条件
	query.WriteString(" WHERE `")
	query.WriteString(keyField)
	query.WriteString("` IN (")

	for i, record := range records {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString("?")
		args = append(args, record[keyField])
	}
	query.WriteString(")")

	// 执行SQL
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if t.db.IsDebug() {
		t.db.logger.Debug("执行SQL", "updateBatch", query.String(), "args", args)
	}

	result, err := tx.ExecContext(ctx, query.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("执行SQL失败: %v", err)
	}

	return result.RowsAffected()
}

// extractBatchFields 从批量数据中提取字段
func (t *Table) extractBatchFields(data []map[string]interface{}) ([]string, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	// 从第一条记录提取字段
	fields := make([]string, 0, len(data[0]))
	for field := range data[0] {
		// 转义字段名
		escapedField := escapeSQLIdentifier(field)
		fields = append(fields, escapedField)
	}

	// 验证所有记录的字段一致性
	for _, item := range data[1:] {
		if len(item) != len(fields) {
			return nil, fmt.Errorf("批量插入数据字段不一致：第一条记录有 %d 个字段，当前记录有 %d 个字段", len(fields), len(item))
		}
	}

	return fields, nil
}
