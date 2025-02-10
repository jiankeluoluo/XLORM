package xlorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	version string = "1.0.0.007"
)

// DB 数据库操作主结构体
type DB struct {
	*sql.DB
	dbName             string          // 数据库名称
	tablePre           string          // 表前缀
	wg                 sync.WaitGroup  // 等待组,用于等待所有任务携程退出
	ctxMu              *sync.RWMutex   // 改为指针类型
	logLevelVar        *slog.LevelVar  // 当前日志级别
	asyncDBMetrics     *asyncDBMetrics // 异步性能指标
	logger             *slog.Logger    // 日志记录器
	structFieldsCache  *shardedCache   // 结构体字段缓存
	placeholderCache   *shardedCache   // 占位符缓存
	StructMapper       *StructMapper   // 回调函数注册表
	startTime          time.Time       // 启动时间
	slowQueryThreshold time.Duration   // 慢查询阈值
	closed             atomic.Bool     // 是否已关闭
	ctx                context.Context
	cancel             context.CancelFunc
	poolStatsEnabled   atomic.Bool   // 原子状态标识
	poolStatsTicker    *time.Ticker  // 统计定时器
	poolStatsStop      chan struct{} // 停止信号
	poolStatsMutex     *sync.Mutex   // 互斥锁保护
	poolStatsInterval  time.Duration // 连接池统计间隔
	debug              bool          // 调试模式
}

// New 创建新的数据库连接
func New(cfg *Config) (*DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("数据库参数配置有误: %v", err)
	}
	// 设置默认值
	if cfg.DBName == "" {
		cfg.DBName = "master"
	}
	if cfg.Driver == "" {
		cfg.Driver = "mysql"
	} else {
		cfg.Driver = strings.ToLower(cfg.Driver)
	}
	if cfg.Charset == "" {
		cfg.Charset = "utf8mb4"
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = time.Hour * 1
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = time.Minute * 30
	}
	if cfg.ConnTimeout == 0 {
		cfg.ConnTimeout = time.Second * 1
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = time.Second * 30
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = time.Second * 30
	}
	if cfg.SlowQueryTime == 0 {
		cfg.SlowQueryTime = time.Second * 1
	}
	if cfg.EnablePoolStats {
		if cfg.PoolStatsInterval == 0 || cfg.PoolStatsInterval < time.Second {
			cfg.PoolStatsInterval = 60 * time.Second // 默认60秒
		}
	}
	if cfg.DBMetricsBufferSize == 0 {
		cfg.DBMetricsBufferSize = 1000 // 默认1000
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "./logs"
	}

	// 设置日志保留天数的默认值
	if cfg.LogRotationMaxAge <= 0 {
		cfg.LogRotationMaxAge = 30 // 默认保留30天
	}

	if cfg.LogBufferSize == 0 {
		cfg.LogBufferSize = 5000
	}

	switch cfg.Driver {
	case "mysql":
		return newMySQL(cfg)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}
}

// M Table的别名，返回一个表操作对象
func (db *DB) M(tableName string) *Table {
	return db.Table(tableName)
}

// Table 返回一个表操作对象
func (db *DB) Table(tableName string) *Table {
	t := tablePool.Get().(*Table)
	t.Reset()
	t.db = db
	if tableName == "" {
		db.logger.Error("tableName不能为空", "table", tableName)
		return t
	}
	// 检查SQL注入风险
	if strings.ContainsAny(tableName, ";\x00") {
		db.logger.Error("检测到可能的SQL注入尝试", "table", tableName)
		return t
	}
	t.tableName = db.GetTableName(tableName)
	return t
}

// GetTableName 获取数据库完整表名
func (db *DB) GetTableName(tableName string) string {
	var builder strings.Builder
	builder.WriteString("`")
	builder.WriteString(db.tablePre)
	builder.WriteString(tableName)
	builder.WriteString("`")
	return builder.String()
}

// WithContext 设置上下文
func (db *DB) WithContext(ctx context.Context) *DB {
	db.ctxMu.Lock()
	defer db.ctxMu.Unlock()
	db.ctx = ctx
	return db
}

// GetContext 获取上下文
func (db *DB) GetContext() context.Context {
	db.ctxMu.RLock()
	defer db.ctxMu.RUnlock()
	return db.ctx
}

// Begin 开始事务
func (db *DB) Begin() (*Transaction, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("数据库连接为空")
	}
	startTime := time.Now()
	traceID := uuid.New().String()
	if db.IsDebug() {
		db.logger.Debug("开始事务", "trace_id", traceID)
	}
	tx, err := db.DB.Begin()
	if err != nil {
		db.asyncDBMetrics.RecordError()
		return nil, fmt.Errorf("开始事务失败: %v, trace_id:%s", err, traceID)
	}

	db.asyncDBMetrics.RecordQueryDuration("begin_transaction", time.Since(startTime))
	return &Transaction{tx, db, traceID}, nil
}

// ExecTx 在事务中执行操作
func (db *DB) ExecTx(fn func(*Transaction) error) error {
	if db == nil || db.DB == nil {
		return errors.New("数据库连接为空")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			db.logger.Error("事务异常回滚",
				"error", "panic",
				"original_error", "",
				"trace_id", tx.traceID,
			)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			db.logger.Error("回滚事务失败",
				"error", rbErr,
				"original_error", err,
				"trace_id", tx.traceID,
			)
			return fmt.Errorf("执行事务失败: %v, 回滚失败: %v, trace_id:%s", err, rbErr, tx.traceID)
		}
		return fmt.Errorf("执行事务失败: %v, trace_id:%s", err, tx.traceID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v, trace_id:%s", err, tx.traceID)
	}
	if tx.db.IsDebug() {
		tx.db.logger.Debug("执行事务完成", "trace_id", tx.traceID)
	}
	return nil
}

// WithCache 使用缓存执行查询
func (db *DB) WithCache(cache Cache, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// 尝试从缓存获取
	if value, ok := cache.Get(key); ok {
		return value, nil
	}

	// 执行查询
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	if err := cache.Set(key, value, expiration); err != nil {
		db.logger.Error("设置缓存失败",
			"key", key,
			"error", err,
		)
	}

	return value, nil
}

// InvalidateCache 使缓存失效
func (db *DB) InvalidateCache(cache Cache, keys ...string) error {
	for _, key := range keys {
		if err := cache.Delete(key); err != nil {
			db.logger.Error("删除缓存失败",
				"key", key,
				"error", err,
			)
			return newDBError("InvalidateCache", err, "", nil)
		}
	}
	return nil
}

// PrepareContext 预处理SQL语句
func (db *DB) PrepareContext(query string) (*sql.Stmt, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("数据库连接为空")
	}

	startTime := time.Now()
	if db.IsDebug() {
		db.logger.Debug("预处理SQL语句",
			"query", query,
		)
	}

	stmt, err := db.DB.Prepare(query)
	duration := time.Since(startTime)
	if err != nil {
		db.asyncDBMetrics.RecordError()
		db.logger.Error("预处理SQL语句失败",
			"query", query,
			"error", err,
			"duration", duration.Seconds(),
		)
		return nil, fmt.Errorf("预处理SQL语句失败: %v", err)
	}

	db.asyncDBMetrics.RecordQueryDuration("prepare", duration)

	// 检查是否是慢查询
	if duration > db.slowQueryThreshold {
		db.asyncDBMetrics.RecordSlowQuery()
		db.logger.Warn("慢预处理",
			"query", query,
			"duration", duration.Seconds(),
		)
	}

	return stmt, nil
}

// Query 执行查询并返回行
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("数据库连接为空")
	}

	if query == "" {
		return nil, errors.New("执行查询失败，查询语句为空")
	}

	startTime := time.Now()
	db.logger.Debug("执行查询",
		"query", query,
		"args", args,
	)

	rows, err := db.DB.Query(query, args...)
	duration := time.Since(startTime)
	if err != nil {
		db.asyncDBMetrics.RecordError()
		db.logger.Error("查询失败",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration,
		)
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	db.asyncDBMetrics.RecordQueryDuration("query", duration)

	// 检查是否是慢查询
	if duration > db.slowQueryThreshold {
		db.asyncDBMetrics.RecordSlowQuery()
		db.logger.Warn("慢查询",
			"query", query,
			"args", args,
			"duration", duration.Seconds(),
		)
	}

	return rows, nil
}

// QueryWithContext 新增带Context的方法
func (db *DB) QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("数据库连接为空")
	}
	startTime := time.Now()
	if db.IsDebug() {
		db.logger.Debug("执行查询",
			"query", query,
			"args", args,
		)
	}
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(startTime)
	if err != nil {
		db.asyncDBMetrics.RecordError()
		db.logger.Error("查询失败",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration.Seconds(),
		)
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	db.asyncDBMetrics.RecordQueryDuration("queryWithContext", duration)

	// 检查是否是慢查询
	if duration > db.slowQueryThreshold {
		db.asyncDBMetrics.RecordSlowQuery()
		db.logger.Warn("慢查询",
			"query", query,
			"args", args,
			"duration", duration.Seconds(),
		)
	}

	return rows, nil
}

// Exec 执行更新操作
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("数据库连接为空")
	}
	if query == "" {
		return nil, errors.New("执行更新失败，查询语句为空")
	}
	startTime := time.Now()
	if db.IsDebug() {
		db.logger.Debug("执行更新",
			"query", query,
			"args", args,
		)
	}
	result, err := db.DB.Exec(query, args...)
	duration := time.Since(startTime)
	if err != nil {
		db.asyncDBMetrics.RecordError()
		db.logger.Error("更新失败",
			"query", query,
			"args", args,
			"error", err,
			"duration", duration.Seconds(),
		)
		return nil, fmt.Errorf("更新失败: %v", err)
	}

	db.asyncDBMetrics.RecordQueryDuration("exec", duration)

	// 检查是否是慢查询
	if duration > db.slowQueryThreshold {
		db.asyncDBMetrics.RecordSlowQuery()
		db.logger.Warn("慢更新",
			"query", query,
			"args", args,
			"duration", duration.Seconds(),
		)
	}

	return result, nil
}

// GetStructMapper 获取结构体映射器
func (db *DB) GetStructMapper() *StructMapper {
	return db.StructMapper
}

// GetStartTime 获取数据库连接开始时间
func (db *DB) GetStartTime() time.Time {
	return db.startTime
}

// GetDBName 获取数据库名称
func (db *DB) GetDBName() string {
	return db.dbName
}

// GetPoolStats 获取连接池统计
func (db *DB) GetPoolStats() *sql.DBStats {
	return poolStats.get()
}

// DBMetrics 获取性能指标
func (db *DB) DBMetrics() *dbMetrics {
	if db.asyncDBMetrics == nil {
		return nil
	}
	return db.asyncDBMetrics.dbMetrics
}

// SetDBMetricsEnable 统一控制所有指标收集
func (db *DB) SetDBMetricsEnable(enable bool) {
	db.poolStatsMutex.Lock()
	defer db.poolStatsMutex.Unlock()
	if db.poolStatsEnabled.Load() == enable {
		return
	}
	db.poolStatsEnabled.Store(enable)
	if enable {
		go db.collectPoolStats(db.poolStatsInterval)
	} else {
		// 安全停止
		if db.poolStatsTicker != nil {
			db.poolStatsTicker.Stop()
		}
		close(db.poolStatsStop)
		db.poolStatsStop = make(chan struct{})
		poolStats.init()
	}
}

// AsyncDBMetrics 获取异步性能指标
func (db *DB) AsyncDBMetrics() *asyncDBMetrics {
	return db.asyncDBMetrics
}

// GetDatabase 获取数据库连接
func (db *DB) GetDatabase() *sql.DB {
	return db.DB
}

// Logger 获取日志实例
func (db *DB) Logger() *slog.Logger {
	return db.logger
}

// AsyncLogger 获取异步日志实例
func (db *DB) AsyncLogger() *asyncLogger {
	if asyncLogger, ok := db.logger.Handler().(*asyncLogger); ok {
		return asyncLogger
	}
	return nil
}

func (db *DB) SetDebug(bool) *DB {
	db.ctxMu.Lock()
	defer db.ctxMu.Unlock()
	db.debug = true
	return db
}

// IsDebug 判断日志功能是否启用
func (db *DB) IsDebug() bool {
	return db.debug
}

// SetLogLevel 动态调整日志级别
func (db *DB) SetLogLevel(level string) error {
	db.ctxMu.Lock()
	defer db.ctxMu.Unlock()
	l, err := parseLogLevel(level)
	if err != nil {
		return err
	}
	db.logLevelVar.Set(l)
	return nil
}

// GetLogLevel 获取当前级别
func (db *DB) GetLogLevel() string {
	return strings.ToLower(db.logLevelVar.Level().String())
}

// Ping 测试数据库连接
func (db *DB) Ping(ctx context.Context) error {
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

// GetVersion 获取版本信息
func (db *DB) GetVersion() string {
	return version
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.closed.Load() {
		return nil
	}
	defer db.asyncDBMetrics.Stop()
	// 取消上下文，触发所有协程退出
	db.cancel()
	// 等待所有后台协程退出（探活、统计等）
	db.wg.Wait()

	var errs []error
	// 关闭数据库连接
	if err := db.DB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("关闭数据库连接失败: %w", err))
	}

	// 关闭日志文件
	if rotatingHandler, ok := db.logger.Handler().(*rotatingFileHandler); ok {
		if err := rotatingHandler.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭日志文件失败: %w", err))
		}
	}

	// 异步关闭日志处理器
	if handler, ok := db.logger.Handler().(*asyncLogger); ok {
		if err := handler.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭日志处理器失败: %w", err))
		}
	}
	// 停止统计协程
	db.SetDBMetricsEnable(false)
	// 停止指标收集
	db.structFieldsCache.Clear()
	db.placeholderCache.Clear()

	db.closed.Store(true)

	if len(errs) > 0 {
		return fmt.Errorf("关闭过程中发生错误: %v", errs)
	}
	return nil
}

// 添加定期Ping
func (db *DB) startKeepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	db.wg.Add(1)
	defer db.wg.Done()
	defer ticker.Stop()
	db.logger.Debug("开启连接探活协程")
	for {
		select {
		case <-ticker.C:
			// 执行探活逻辑
			ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
			err := db.PingContext(ctx)
			cancel()

			if err != nil && !errors.Is(err, context.Canceled) {
				db.logger.Error("数据库连接探活失败",
					"error", err,
				)
			}

		case <-db.ctx.Done():
			// 上下文已取消，退出循环
			db.logger.Debug("停止连接探活协程")
			return
		}
	}
}

// collectPoolStats 定期收集连接池统计信息
func (db *DB) collectPoolStats(poolStatsInterval time.Duration) {
	db.poolStatsMutex.Lock()
	defer db.poolStatsMutex.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	db.poolStatsTicker = time.NewTicker(poolStatsInterval)
	db.logger.Debug("开启连接池统计协程")
	poolStats.init()
	for {
		select {
		case <-db.poolStatsTicker.C:
			if !db.poolStatsEnabled.Load() {
				return
			}
			stats := db.DB.Stats()
			poolStats.update(&stats)
		case <-db.poolStatsStop:
			poolStats.init()
			db.logger.Debug("停止连接池统计协程")
			return
		case <-db.ctx.Done():
			poolStats.init()
			db.logger.Debug("结束连接池统计协程")
			return
		}
	}
}
