package xlorm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// converterFunc 定义类型转换函数，用于将字符串转换为特定类型
type converterFunc func(string, reflect.Value) (interface{}, error)

// structDialect 定义数据库方言接口，用于处理不同数据库的特殊标识符和时间格式
type structDialect interface {
	QuoteIdentifier(string) string
	FormatTime(time.Time) string
}

// standardDialect 标准方言实现，适用于大多数关系型数据库
type standardDialect struct{}

// QuoteIdentifier 使用双引号包裹标识符(暂不启用)
func (d *standardDialect) QuoteIdentifier(s string) string {
	return s
}

// FormatTime 使用RFC3339格式化时间
func (d *standardDialect) FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// fieldMeta 存储字段的元数据信息
type fieldMeta struct {
	dbName     string
	sqlType    string
	defaultVal string
	callbacks  map[string]func(interface{}) (interface{}, error)
	ignored    bool
	prefix     string
	required   bool
	omitempty  bool
	isPK       bool
	hasDefault bool
}

// structMeta 存储结构体的元数据
type structMeta struct {
	fields     map[string]fieldMeta
	fieldOrder []string
	converter  map[string]converterFunc
	pkFields   []string
}

// structConfig 存储处理选项的配置
type structConfig struct {
	SkipDefault   bool
	SkipCallbacks map[string]bool
	dialect       structDialect
}

// structOption 定义配置选项的函数类型
type structOption func(*structConfig)

// StructMapper 提供结构体映射和转换的高级功能
type StructMapper struct {
	stageBefore    string
	stageAfter     string
	stageGlobal    string
	metaCache      sync.Map
	converters     map[reflect.Kind]converterFunc
	defaultDialect structDialect

	// 回调相关字段
	callbacks sync.Map

	skipDefault   bool
	skipCallbacks map[string]bool
}

// NewStructMapper 创建一个新的 StructMapper 实例
func NewStructMapper() *StructMapper {
	return &StructMapper{
		stageBefore: "_before",
		stageAfter:  "_after",
		stageGlobal: "_global",
		metaCache:   sync.Map{},
		converters: map[reflect.Kind]converterFunc{
			reflect.String:  convertString,
			reflect.Int:     convertInt,
			reflect.Int64:   convertInt64,
			reflect.Bool:    convertBool,
			reflect.Float32: convertFloat32,
			reflect.Float64: convertFloat64,
		},
		callbacks:      sync.Map{},
		defaultDialect: &standardDialect{},
		skipCallbacks:  make(map[string]bool),
	}
}

// GetPrimaryKeys 获取结构体的所有主键值
func (sm *StructMapper) GetPrimaryKeys(obj interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	meta := sm.getStructMeta(val.Type())
	result := make(map[string]interface{})

	// 获取所有主键字段的值
	for _, pkField := range meta.pkFields {
		field := val.FieldByName(pkField)
		result[pkField] = field.Interface()
	}

	return result, nil
}

// GetPrimaryKey 获取结构体的第一个主键值
func (sm *StructMapper) GetPrimaryKey(obj interface{}) (string, interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("input must be a struct")
	}

	meta := sm.getStructMeta(val.Type())

	if len(meta.pkFields) == 0 {
		return "", nil, fmt.Errorf("primary key not found")
	}
	field := val.FieldByName(meta.pkFields[0])
	return meta.pkFields[0], field.Interface(), nil
}

func (sm *StructMapper) StructToMap(s interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	t := val.Type()

	// 获取结构体元数据
	meta := sm.getStructMeta(t)

	// 处理结构体字段
	result := make(map[string]interface{})
	for _, fieldName := range meta.fieldOrder {
		field := val.FieldByName(fieldName)
		fieldMeta := meta.fields[fieldName]

		// 跳过忽略的字段
		if fieldMeta.ignored {
			continue
		}

		// 递归处理嵌套结构体
		if field.Kind() == reflect.Struct && !isBasicType(field.Type()) {
			nestedMap, err := sm.StructToMap(field.Interface())
			if err != nil {
				return nil, err
			}
			for k, v := range nestedMap {
				result[k] = v
			}
			continue
		}

		// 处理默认值和空值
		if isEmptyValue(field) && fieldMeta.hasDefault {
			defaultVal, err := sm.convertValue(fieldMeta.defaultVal, field.Type())
			if err != nil {
				return nil, err
			}
			field = reflect.ValueOf(defaultVal)
		}

		// 将字段值添加到结果map
		quotedName := sm.defaultDialect.QuoteIdentifier(fieldMeta.dbName)
		result[quotedName] = field.Interface()
	}

	return result, nil
}

// ToMapWithOptions 将结构体转换为map，支持自定义选项
func (sm *StructMapper) ToMapWithOptions(obj interface{}, options ...structOption) (map[string]interface{}, error) {
	// 创建配置，设置默认值
	cfg := &structConfig{
		SkipDefault:   sm.skipDefault,
		SkipCallbacks: sm.skipCallbacks,
		dialect:       sm.defaultDialect,
	}

	// 应用用户提供的选项
	for _, opt := range options {
		opt(cfg)
	}

	val := reflect.ValueOf(obj)
	return sm.processValue(val, cfg)
}

// RegisterConverter 注册自定义类型转换器
func (sm *StructMapper) RegisterConverter(kind reflect.Kind, fn converterFunc) {
	sm.converters[kind] = fn
}

// RegisterCallback 注册回调函数
func (sm *StructMapper) RegisterCallback(name string, callback func(interface{}) (interface{}, error)) error {
	if _, exists := sm.callbacks.Load(name); exists {
		return fmt.Errorf("回调函数 %s 已存在", name)
	}

	sm.callbacks.Store(name, callback)
	return nil
}

// GetCallback 获取回调函数
func (sm *StructMapper) GetCallback(name string) (func(interface{}) (interface{}, error), bool) {
	if callback, exists := sm.callbacks.Load(name); exists {
		return callback.(func(interface{}) (interface{}, error)), true
	}
	return nil, false
}

// DelCallback 删除回调函数
func (sm *StructMapper) DelCallback(name string) {
	sm.callbacks.Delete(name)
}

// ToMap 将结构体转换为 map
func (sm *StructMapper) ToMap(obj interface{}, options ...structOption) (map[string]interface{}, error) {
	cfg := &structConfig{
		SkipDefault:   sm.skipDefault,
		SkipCallbacks: sm.skipCallbacks,
		dialect:       sm.defaultDialect,
	}

	// 应用用户提供的选项
	for _, opt := range options {
		opt(cfg)
	}

	return sm.processValue(reflect.ValueOf(obj), cfg)
}

// SkipDefault 设置是否跳过默认值
func (sm *StructMapper) SkipDefault() structOption {
	return func(c *structConfig) {
		c.SkipDefault = true
	}
}

// SkipCallback 设置是否跳过回调
func (sm *StructMapper) SkipCallback() structOption {
	return func(c *structConfig) {
		c.SkipCallbacks[sm.stageBefore] = true
		c.SkipCallbacks[sm.stageAfter] = true
	}
}

// WithDialect 设置数据库方言
func (sm *StructMapper) WithDialect(d structDialect) structOption {
	return func(c *structConfig) {
		c.dialect = d
	}
}

// 内部方法：获取结构体元数据
func (sm *StructMapper) getStructMeta(t reflect.Type) *structMeta {
	// 处理指针类型
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 尝试从缓存中获取元数据
	if v, ok := sm.metaCache.Load(t); ok {
		return v.(*structMeta)
	}

	meta := &structMeta{
		fields:     make(map[string]fieldMeta),
		fieldOrder: make([]string, 0),
		converter:  make(map[string]converterFunc),
		pkFields:   make([]string, 0),
	}

	// 遍历结构体的所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 解析字段元数据
		fieldMeta := sm.parseFieldMeta(&field)
		if fieldMeta.ignored {
			continue
		}

		// 记录主键字段
		if fieldMeta.isPK {
			meta.pkFields = append(meta.pkFields, field.Name)
		}

		meta.fields[field.Name] = fieldMeta
		meta.fieldOrder = append(meta.fieldOrder, field.Name)
	}

	// 缓存元数据
	sm.metaCache.Store(t, meta)
	return meta
}

// 解析字段元数据的内部方法
func (sm *StructMapper) parseFieldMeta(field *reflect.StructField) fieldMeta {
	dbTag := field.Tag.Get("db")
	if dbTag == "-" {
		return fieldMeta{ignored: true}
	}

	parts := strings.Split(dbTag, ",")
	fieldMeta := fieldMeta{
		dbName:     parts[0],
		callbacks:  make(map[string]func(interface{}) (interface{}, error)),
		ignored:    false,
		prefix:     "",
		required:   false,
		omitempty:  false,
		isPK:       false,
		hasDefault: false,
	}

	for _, part := range parts[1:] {
		switch {
		case part == "pk":
			fieldMeta.isPK = true
		case part == "required":
			fieldMeta.required = true
		case part == "omitempty":
			fieldMeta.omitempty = true
		case strings.HasPrefix(part, "type="):
			fieldMeta.sqlType = strings.TrimPrefix(part, "type=")
		case strings.HasPrefix(part, "default="):
			fieldMeta.hasDefault = true
			fieldMeta.defaultVal = strings.TrimPrefix(part, "default=")
		case part == "ignore":
			fieldMeta.ignored = true
		}
	}

	return fieldMeta
}

// processValue 递归处理结构体的值，转换为map
func (sm *StructMapper) processValue(val reflect.Value, cfg *structConfig) (map[string]interface{}, error) {
	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	// 只处理结构体类型
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	result := make(map[string]interface{})
	meta := sm.getStructMeta(val.Type())

	// 执行全局前置回调
	if cb, ok := sm.GetCallback(sm.stageGlobal); ok {
		res, err := cb(val.Interface())
		if err != nil {
			return nil, err
		}
		val = reflect.ValueOf(res)
	}

	// 遍历结构体字段
	for _, fieldName := range meta.fieldOrder {
		field := val.FieldByName(fieldName)
		fieldMeta := meta.fields[fieldName]

		// 安全地处理未导出字段
		if !field.CanInterface() {
			continue
		}

		// 执行字段级别前置回调
		for _, cb := range fieldMeta.callbacks {
			if !cfg.SkipCallbacks[sm.stageBefore] {
				res, err := cb(field.Interface())
				if err != nil {
					return nil, err
				}
				field = reflect.ValueOf(res)
			}
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Struct {
			// 对于嵌套结构体，特殊处理
			if field.Type().Name() == "" || !isBasicType(field.Type()) {
				nestedMap, err := sm.processValue(field, cfg)
				if err != nil {
					return nil, err
				}
				for k, v := range nestedMap {
					result[k] = v
				}
				continue
			}
		}

		// 处理 omitempty 标签
		if fieldMeta.omitempty && isEmptyValue(field) {
			// 如果没有默认值，则跳过
			if !fieldMeta.hasDefault {
				continue
			}
		}

		// 处理默认值
		if isEmptyValue(field) && fieldMeta.hasDefault {
			defaultVal, err := sm.convertValue(fieldMeta.defaultVal, field.Type())
			if err != nil {
				return nil, err
			}
			field = reflect.ValueOf(defaultVal)
		}

		// 执行字段级别后置回调
		for _, cb := range fieldMeta.callbacks {
			if !cfg.SkipCallbacks[sm.stageAfter] {
				res, err := cb(field.Interface())
				if err != nil {
					return nil, err
				}
				field = reflect.ValueOf(res)
			}
		}

		// 将字段值添加到结果map
		quotedName := cfg.dialect.QuoteIdentifier(fieldMeta.dbName)
		result[quotedName] = field.Interface()
	}

	// 执行全局后置回调
	if cb, ok := sm.GetCallback(sm.stageGlobal); ok {
		res, err := cb(result)
		if err != nil {
			return nil, err
		}
		result = res.(map[string]interface{})
	}

	return result, nil
}

// convertValue 根据字段类型转换默认值
func (sm *StructMapper) convertValue(defaultVal string, fieldType reflect.Type) (interface{}, error) {
	if converter, ok := sm.converters[fieldType.Kind()]; ok {
		return converter(defaultVal, reflect.Value{})
	}
	if fieldType == reflect.TypeOf(time.Time{}) {
		return convertTime(defaultVal)
	}
	return nil, fmt.Errorf("unsupported type conversion for %v", fieldType)
}
