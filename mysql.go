package xlorm

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// newMySQL 创建新的MySQL数据库连接
func newMySQL(cfg *Config) (*DB, error) {
	// 构建 DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		safeTimeout(cfg.ConnTimeout),  // 带最小值的超时
		safeTimeout(cfg.ReadTimeout),  // 带最小值的读超时
		safeTimeout(cfg.WriteTimeout), // 带最小值的写超时
	)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %v", err)
	}

	logLevelVar := new(slog.LevelVar)
	logLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("日志级别设置失败: %v", err)
	}
	logLevelVar.Set(logLevel)

	// 创建异步处理器
	asyncHandler := NewAsyncLogger(NewRotatingFileHandler(
		cfg.LogDir,
		"db",
		time.Duration(cfg.LogRotationMaxAge)*24*time.Hour,
		logLevelVar,
		cfg.LogRotationEnabled,
	).handler, cfg.LogBufferSize)

	// 创建 DB 实例
	xdb := &DB{
		ctxMu:              new(sync.RWMutex),
		ctx:                ctx,
		cancel:             cancel,
		dbName:             cfg.DBName,
		DB:                 db,
		tablePre:           cfg.TablePrefix,
		asyncDBMetrics:     newAsyncDBMetrics(cfg.DBName, cfg.DBMetricsBufferSize),
		structFieldsCache:  newShardedCache(),
		placeholderCache:   newShardedCache(),
		StructMapper:       NewStructMapper(),
		logger:             slog.New(asyncHandler),
		logLevelVar:        logLevelVar,
		startTime:          time.Now(),
		poolStatsStop:      make(chan struct{}),
		poolStatsInterval:  cfg.PoolStatsInterval,
		poolStatsMutex:     new(sync.Mutex), // 互斥锁保护
		poolStatsTicker:    nil,             // 统计定时器
		slowQueryThreshold: cfg.SlowQueryTime,
		debug:              cfg.Debug,
	}

	// 启动连接池统计信息收集
	if cfg.EnablePoolStats {
		xdb.poolStatsEnabled.Store(true)
		go xdb.collectPoolStats(cfg.PoolStatsInterval)
	}

	// 启动连接探活
	go xdb.startKeepAlive()

	return xdb, nil
}
