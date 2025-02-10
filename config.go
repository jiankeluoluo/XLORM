package xlorm

import (
	"errors"
	"time"
)

// Config 数据库配置结构体
type Config struct {
	DBName              string        //数据库别名称、用于区分不同数据库
	Driver              string        // 数据库驱动
	Host                string        // 主机地址
	Username            string        // 用户名
	Password            string        // 密码
	Database            string        // 数据库名称
	Charset             string        // 字符集
	TablePrefix         string        // 表前缀
	LogDir              string        // 日志目录
	LogLevel            string        // 日志级别（支持：debug|info|warn|error）
	ConnMaxLifetime     time.Duration // 连接最大生命周期
	ConnMaxIdleTime     time.Duration // 连接最大空闲时间
	ConnTimeout         time.Duration // 连接超时时间
	ReadTimeout         time.Duration // 读取超时时间
	WriteTimeout        time.Duration // 写入超时时间
	SlowQueryTime       time.Duration // 慢查询阈值
	PoolStatsInterval   time.Duration // 连接池统计频率
	Port                int
	LogBufferSize       int  // 日志缓冲区数量（默认5000）
	MaxOpenConns        int  // 最大打开连接数（默认0）
	MaxIdleConns        int  // 最大空闲连接数（默认0）
	LogRotationMaxAge   int  // 日志保留天数，默认30天
	DBMetricsBufferSize int  // 异步指标缓冲区数量（默认1000）
	LogRotationEnabled  bool // 是否启用日志轮转
	EnablePoolStats     bool // 是否启用性能指标（默认false）
	Debug               bool // 是否开启调试模式（默认false）
}

// Validate 验证配置
func (cfg *Config) Validate() error {
	if cfg == nil {
		return errors.New("配置不能为空")
	}
	if cfg.Host == "" {
		return errors.New("数据库主机不能为空")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return errors.New("无效端口号")
	}
	if cfg.Username == "" {
		return errors.New("数据库用户名不能为空")
	}
	if cfg.Database == "" {
		return errors.New("数据库名称不能为空")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "debug"
	}
	if _, err := parseLogLevel(cfg.LogLevel); err != nil {
		return err
	}
	return nil
}
