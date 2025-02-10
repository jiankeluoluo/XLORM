package xlorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// asyncLogger 异步日志处理器
type asyncLogger struct {
	baseHandler slog.Handler       // 实际处理器
	ch          chan slog.Record   // 缓冲通道
	wg          *sync.WaitGroup    // 使用指针避免复制
	ctx         context.Context    // 上下文
	cancel      context.CancelFunc // 取消函数
	dropped     atomic.Uint64      // 丢弃的日志计数
	total       atomic.Uint64      // 总处理日志数
	errCh       chan error         // 错误通道
	closed      atomic.Bool        // 是否已关闭
}

// rotatingFileHandler 日志文件旋转处理器
type rotatingFileHandler struct {
	handler            slog.Handler // 实际处理器
	dir                string       // 日志目录
	baseFileName       string       // 基础文件名
	currentDate        string       // 当前日期
	currentFile        *os.File     // 当前日志文件
	mu                 *sync.Mutex
	maxAge             time.Duration  // 日志文件最大保留时间
	logLevel           *slog.LevelVar // 日志级别
	logRotationEnabled bool           // 日志轮转是否启用
}

// NewAsyncLogger 创建异步日志处理器
func NewAsyncLogger(h slog.Handler, bufferSize int) *asyncLogger {
	ctx, cancel := context.WithCancel(context.Background())
	al := &asyncLogger{
		baseHandler: h,
		ch:          make(chan slog.Record, bufferSize),
		wg:          &sync.WaitGroup{}, // 使用指针初始化
		ctx:         ctx,
		cancel:      cancel,
		errCh:       make(chan error, 100), // 增加错误通道
	}

	// 启动处理协程
	al.wg.Add(1)
	go al.process()

	return al
}

// Enabled 实现 slog.Handler 接口
func (al *asyncLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return al.baseHandler.Enabled(ctx, level)
}

// Handle 实现 slog.Handler 接口
func (al *asyncLogger) Handle(ctx context.Context, r slog.Record) error {
	// 快速检查是否已关闭
	if al.closed.Load() {
		return errors.New("日志处理器已关闭")
	}
	select {
	case al.ch <- r: // 尝试非阻塞写入
		al.total.Add(1)
		return nil
	case <-al.ctx.Done():
		return al.ctx.Err() // 已关闭
	default:
		al.dropped.Add(1)
		// 通道满时记录警告
		select {
		case al.errCh <- fmt.Errorf("日志通道已满，丢弃日志记录"):
		default:
			// 错误通道也满了，直接忽略
		}
		return nil
	}
}

// WithAttrs 实现 slog.Handler 接口
func (al *asyncLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &asyncLogger{
		baseHandler: al.baseHandler.WithAttrs(attrs),
		ch:          al.ch,
		wg:          al.wg,
		ctx:         al.ctx,
		cancel:      al.cancel,
	}
}

// WithGroup 实现 slog.Handler 接口
func (al *asyncLogger) WithGroup(name string) slog.Handler {
	return &asyncLogger{
		baseHandler: al.baseHandler.WithGroup(name),
		ch:          al.ch,
		wg:          al.wg,
		ctx:         al.ctx,
		cancel:      al.cancel,
	}
}

func (al *asyncLogger) Close() error {
	if al.closed.Load() {
		return nil
	}
	if !al.closed.CompareAndSwap(false, true) {
		return errors.New("日志处理器已关闭")
	}

	close(al.ch) // 关闭通道
	al.cancel()  // 关闭上下文，触发 process() 退出

	// 创建带超时的等待通道
	done := make(chan struct{}, 1)
	go func() {
		// 设置处理剩余日志的总超时时间
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		defer func() {
			al.wg.Wait()
			done <- struct{}{}
			close(done)
		}()

		// 非阻塞地处理剩余日志
		for {
			select {
			case r, ok := <-al.ch:
				if !ok {
					return
				}
				// 尝试处理剩余日志，但要受超时控制
				_ = al.baseHandler.Handle(ctx, r)
			default:
				return
			}
		}
	}()

	// 等待处理或超时
	select {
	case <-done:
		return al.collectErrors()
	case <-time.After(5 * time.Second):
		log.Printf("日志处理器关闭超时")
		return errors.New("日志处理器关闭超时")
	}
}

// GetDroppedLogsCount 获取丢弃的日志数量
func (al *asyncLogger) GetDroppedLogsCount() uint64 {
	return al.dropped.Load()
}

// GetTotalLogsCount 获取总处理日志数量
func (al *asyncLogger) GetTotalLogsCount() uint64 {
	return al.total.Load()
}

// GetLogMetrics 获取当前日志状态
func (al *asyncLogger) GetLogMetrics() map[string]uint64 {
	return map[string]uint64{
		"total_logs":    al.total.Load(),
		"dropped_logs":  al.dropped.Load(),
		"channel_depth": uint64(len(al.ch)),
	}
}

func (al *asyncLogger) collectErrors() error {
	var errs []error

	for {
		select {
		case err, ok := <-al.errCh:
			if !ok {
				// 错误通道已关闭
				if len(errs) == 0 {
					return nil
				}
				return errors.Join(errs...)
			}
			errs = append(errs, err)
		case <-time.After(5 * time.Second):
			// 超时处理
			if len(errs) > 0 {
				return fmt.Errorf("日志处理错误（部分）: %v", errors.Join(errs...))
			}
			return errors.New("收集日志错误超时")
		default:
			// 没有更多错误，立即返回
			if len(errs) == 0 {
				return nil
			}
			return errors.Join(errs...)
		}
	}
}

// process 日志处理协程
func (al *asyncLogger) process() {
	defer al.wg.Done()
	defer close(al.errCh)

	for {
		select {
		case r, ok := <-al.ch:
			if !ok {
				return
			}
			// 调试：打印完整的日志记录信息
			// log.Printf("接收到日志记录: Message='%s', Level=%v", r.Message, r.Level)
			// 统一处理日志和超时
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := al.baseHandler.Handle(ctx, r); err != nil {
				select {
				case al.errCh <- err:
				default:
					log.Printf("错误通道已满，丢弃错误: %v", err)
				}
			}
			cancel()

		case <-al.ctx.Done():
			// 上下文取消，退出
			return
		}
	}
}

func NewRotatingFileHandler(dir, baseFileName string, maxAge time.Duration, logLevel *slog.LevelVar, LogRotationEnabled bool) *rotatingFileHandler {
	r := &rotatingFileHandler{
		mu:                 new(sync.Mutex),
		dir:                dir,
		baseFileName:       baseFileName,
		maxAge:             maxAge,
		logLevel:           logLevel,
		logRotationEnabled: LogRotationEnabled,
	}
	r.openNewFileIfNeeded()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handler = slog.NewJSONHandler(r.currentFile, &slog.HandlerOptions{Level: r.logLevel})
	go r.startLogRotationCleanup()
	return r
}

// 实现 io.Writer 接口
func (r *rotatingFileHandler) Write(p []byte) (n int, err error) {
	if err := r.openNewFileIfNeeded(); err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentFile.Write(p)
}

// 实现 slog.Handler 接口
func (r *rotatingFileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true // 根据实际需求调整
}

func (r *rotatingFileHandler) Handle(ctx context.Context, record slog.Record) error {
	return r.handler.Handle(ctx, record)
}

func (r *rotatingFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return r
}

func (r *rotatingFileHandler) WithGroup(name string) slog.Handler {
	return r
}

func (r *rotatingFileHandler) openNewFileIfNeeded() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 创建日志处理器
	if r.logRotationEnabled {
		currentDate := time.Now().Format("2006-01-02")
		if currentDate != r.currentDate {

			// 关闭旧文件
			if r.currentFile != nil {
				_ = r.currentFile.Sync() // 强制刷新
				_ = r.currentFile.Close()
			}

			// 确保日志目录存在
			if err := os.MkdirAll(r.dir, 0755); err != nil {
				return err
			}

			// 创建新文件
			filename := filepath.Join(r.dir, fmt.Sprintf("%s_%s.log", r.baseFileName, currentDate))
			file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}

			r.currentFile = file
			r.currentDate = currentDate
		}
		return nil
	}
	if r.currentFile != nil {
		return nil
	}
	// 确保日志目录存在
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return err
	}
	// 非轮转模式下明确使用
	file, err := os.OpenFile(filepath.Join(r.dir, fmt.Sprintf("%s.log", r.baseFileName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("无法打开日志文件: %v", err)
	}
	r.currentFile = file
	r.currentDate = ""
	return nil
}

// startLogRotationCleanup 开始日志轮转清理
func (r *rotatingFileHandler) startLogRotationCleanup() {
	// 如果日志轮转未启用，直接返回
	if !r.logRotationEnabled {
		return
	}
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanupOldLogs()
	}
}

// cleanupOldLogs 清理旧日志
func (r *rotatingFileHandler) cleanupOldLogs() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 如果日志轮转未启用，直接返回
	if !r.logRotationEnabled {
		return
	}

	files, err := os.ReadDir(r.dir)
	if err != nil {
		fmt.Printf("读取日志目录失败: %v\n", err)
		return
	}

	cutoffTime := time.Now().Add(-r.maxAge)
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), r.baseFileName) && strings.HasSuffix(file.Name(), ".log") {
			parts := strings.Split(file.Name(), "_")
			if len(parts) < 2 || !strings.HasSuffix(parts[1], ".log") {
				continue // 忽略格式不匹配的文件
			}
			// 检查日期部分是否为有效格式
			datePart := strings.TrimSuffix(parts[1], ".log")
			if _, err := time.Parse("2006-01-02", datePart); err != nil {
				continue
			}
			info, err := file.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoffTime) {
				os.Remove(filepath.Join(r.dir, file.Name()))
			}
		}
	}
}

func (r *rotatingFileHandler) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentFile != nil {
		// 先同步
		if err := r.currentFile.Sync(); err != nil {
			return fmt.Errorf("同步日志文件失败: %v", err)
		}

		// 再关闭
		if err := r.currentFile.Close(); err != nil {
			return fmt.Errorf("关闭日志文件失败: %v", err)
		}

		r.currentFile = nil
		r.currentDate = ""
	}
	return nil
}
