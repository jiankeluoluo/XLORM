package xlorm

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 新增：SQL标识符转义函数
func escapeSQLIdentifier(name string) string {
	// 添加对保留字的过滤
	reservedWords := map[string]bool{
		"select": true,
		"insert": true,
		"update": true,
		"delete": true,
	}
	if reservedWords[strings.ToLower(name)] {
		return "`invalid`"
	}

	// 过滤非法字符，仅允许字母、数字、下划线和点
	var safeName strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			safeName.WriteRune(r)
		}
	}
	if safeName.Len() == 0 {
		return "``"
	}
	return "`" + safeName.String() + "`"
}

// safeTimeout 带最小值的超时时间
func safeTimeout(d time.Duration) string {
	if d <= 1 {
		return "1s"
	}
	return fmt.Sprintf("%vs", d.Seconds())
}

func getCachedPlaceholder(fieldCount int, placeholderCache *shardedCache) string {
	keyName := fmt.Sprintf("placeholder:%d", fieldCount)
	if v, ok := placeholderCache.Get(keyName); ok {
		return v[0] // 直接返回第一个元素
	}
	s := "(" + strings.Repeat("?,", fieldCount-1) + "?)"
	placeholderCache.Set(keyName, []string{s})
	return s
}

func parseLogLevel(level string) (slog.Level, error) {
	l, ok := logLevelMap[strings.ToLower(level)]
	if !ok || level == "" {
		return slog.LevelInfo, fmt.Errorf("无效的日志级别: %s,可选值:debug|info|warn|error", level)
	}
	return l, nil
}

// isValidFieldName 检查字段名是否合法
func isValidFieldName(field string) bool {
	// 快速预检查
	if len(field) == 0 {
		return false
	}
	// 使用位图加速判断
	var validChars [256]bool
	for i := 'a'; i <= 'z'; i++ {
		validChars[i] = true
	}
	for i := 'A'; i <= 'Z'; i++ {
		validChars[i] = true
	}
	for i := '0'; i <= '9'; i++ {
		validChars[i] = true
	}
	validChars['_'] = true
	validChars['.'] = true

	// 使用数组查表，避免多次比较
	for i := 0; i < len(field); i++ {
		if !validChars[field[i]] {
			return false
		}
	}
	return true
}

// isValidSafeOrderBy 检查是否是安全的OrderBy字符串是否只包含字母、数字、下划线、逗号或空格
func isValidSafeOrderBy(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != ',' && c != ' ' {
			return false
		}
	}
	return true
}

// extractFromMapSlice 从map切片提取字段
func extractFromMapSlice(maps []map[string]interface{}) ([]string, []interface{}, error) {
	if len(maps) == 0 {
		return nil, nil, errors.New("数据不能为空")
	}

	fields, _, err := extractFromMap(maps[0])
	if err != nil {
		return nil, nil, err
	}

	values := make([]interface{}, 0, len(maps)*len(fields))
	for _, m := range maps {
		for _, field := range fields {
			values = append(values, m[field])
		}
	}

	return fields, values, nil
}

func extractFromMap(m map[string]interface{}) ([]string, []interface{}, error) {
	fields := make([]string, 0, len(m))
	for k := range m {
		fields = append(fields, k) // 直接使用原始字段名
	}
	sort.Strings(fields)

	values := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		values = append(values, m[field])
	}

	return fields, values, nil
}

// convertTime 时间转换器
func convertTime(s string) (interface{}, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// isBasicType 判断是否为基本类型
func isBasicType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	case reflect.Struct:
		// 特殊处理 time.Time 类型
		return t.String() == "time.Time"
	default:
		return false
	}
}

// isEmptyValue 判断值是否为空
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
		// 针对常见类型的快速判断
	}
	return false
}

// 类型转换函数
func convertString(val string, _ reflect.Value) (interface{}, error) {
	return val, nil
}

func convertInt(val string, _ reflect.Value) (interface{}, error) {
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return nil, err
	}
	return intVal, nil
}

func convertInt64(val string, _ reflect.Value) (interface{}, error) {
	int64Val, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	return int64Val, nil
}

func convertBool(val string, _ reflect.Value) (interface{}, error) {
	return strconv.ParseBool(val)
}

func convertFloat32(val string, _ reflect.Value) (interface{}, error) {
	float32Val, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return nil, err
	}
	return float32(float32Val), nil
}

func convertFloat64(val string, _ reflect.Value) (interface{}, error) {
	float64Val, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return float64Val, nil
}
