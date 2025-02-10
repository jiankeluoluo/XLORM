package xlorm

import (
	"sync"
	"sync/atomic"
	"time"
)

// dbMetrics 性能指标结构体
type dbMetrics struct {
	dbname         string
	queryDurations sync.Map
	affectedRows   atomic.Int64
	totalQueries   atomic.Int64
	slowQueries    atomic.Int64
	errors         atomic.Int64
}

// asyncDBMetrics 异步性能指标结构体
type asyncDBMetrics struct {
	buffer   *ringBuffer
	stopChan chan struct{}
	wg       sync.WaitGroup
	*dbMetrics
	droppedMetrics atomic.Uint64 //丢弃的指标数量
}

// ringBuffer 线程安全的环形缓冲区
type ringBuffer struct {
	buffer []func(*dbMetrics)
	size   int
	head   int
	tail   int
	count  int
	mu     sync.Mutex
}

// newRingBuffer 创建一个新的环形缓冲区
func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		buffer: make([]func(*dbMetrics), size),
		size:   size,
	}
}

// Enqueue 向环形缓冲区添加元素
func (rb *ringBuffer) Enqueue(item func(*dbMetrics)) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == rb.size {
		// 缓冲区已满，覆盖最旧的元素
		rb.head = (rb.head + 1) % rb.size
		rb.buffer[rb.tail] = item
		rb.tail = (rb.tail + 1) % rb.size
		return false
	}

	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++
	return true
}

// Dequeue 从环形缓冲区取出元素
func (rb *ringBuffer) Dequeue() (func(*dbMetrics), bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil, false
	}

	item := rb.buffer[rb.head]
	rb.head = (rb.head + 1) % rb.size
	rb.count--
	return item, true
}

// newMetrics 创建新的性能指标实例
func newDBMetrics(dbname string) *dbMetrics {
	return &dbMetrics{dbname: dbname}
}

// newAsyncMetrics 创建新的异步性能指标实例
func newAsyncDBMetrics(dbname string, bufferSize int) *asyncDBMetrics {
	defaultBufferSize := 1000
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	am := &asyncDBMetrics{
		buffer:    newRingBuffer(bufferSize),
		stopChan:  make(chan struct{}),
		dbMetrics: newDBMetrics(dbname),
	}
	am.start()
	return am
}

// GetDBMetrics 获取性能指标统计
func (m *dbMetrics) GetDBMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	metrics["db_name"] = m.dbname
	// 收集查询时间统计
	queryStats := make(map[string]interface{})
	m.queryDurations.Range(func(key, value interface{}) bool {
		durations := value.([]time.Duration)
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		queryStats[key.(string)] = map[string]interface{}{
			"count":        len(durations),
			"total_time":   total,
			"average_time": total / time.Duration(len(durations)),
		}
		return true
	})

	metrics["query_stats"] = queryStats
	metrics["total_affected_rows"] = m.affectedRows.Load()
	metrics["total_queries"] = m.totalQueries.Load()
	metrics["slow_queries"] = m.slowQueries.Load()
	metrics["total_errors"] = m.errors.Load()

	return metrics
}

// ResetDBMetrics 重置性能指标
func (m *dbMetrics) ResetDBMetrics() {
	m.queryDurations = sync.Map{}
	m.affectedRows.Store(0)
	m.totalQueries.Store(0)
	m.slowQueries.Store(0)
	m.errors.Store(0)
}

// RecordQueryDuration 记录查询耗时
func (m *dbMetrics) RecordQueryDuration(queryType string, duration time.Duration) {
	if queryType == "" {
		queryType = "unknown"
	}
	m.totalQueries.Add(1)
	if durations, ok := m.queryDurations.Load(queryType); ok {
		durs := durations.([]time.Duration)
		durs = append(durs, duration)
		m.queryDurations.Store(queryType, durs)
	} else {
		m.queryDurations.Store(queryType, []time.Duration{duration})
	}
}

// RecordAffectedRows 记录影响的行数
func (m *dbMetrics) RecordAffectedRows(rows int64) {
	m.affectedRows.Add(rows)
}

// RecordError 记录错误
func (m *dbMetrics) RecordError() {
	m.errors.Add(1)
}

// RecordSlowQuery 记录慢查询
func (m *dbMetrics) RecordSlowQuery() {
	m.slowQueries.Add(1)
}

func (am *asyncDBMetrics) start() {
	am.wg.Add(1)
	go func() {
		defer am.wg.Done()
		for {
			select {
			case <-am.stopChan:
				return
			default:
				// 尝试从环形缓冲区获取并处理指标
				if metricFunc, ok := am.buffer.Dequeue(); ok {
					metricFunc(am.dbMetrics)
				} else {
					// 如果缓冲区为空，短暂休眠以避免过度自旋
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	}()
}

// Stop 停止异步指标收集
func (am *asyncDBMetrics) Stop() {
	close(am.stopChan)
	am.wg.Wait()
}

// recordMetric 记录指标的通用方法
func (am *asyncDBMetrics) recordMetric(metricFunc func(*dbMetrics)) {
	if !am.buffer.Enqueue(metricFunc) {
		// 缓冲区已满，记录丢弃的指标
		am.droppedMetrics.Add(1)
	}
}

// RecordQueryDuration 记录查询耗时
func (am *asyncDBMetrics) RecordQueryDuration(queryType string, duration time.Duration) {
	am.recordMetric(func(m *dbMetrics) {
		m.RecordQueryDuration(queryType, duration)
	})
}

// RecordError 记录错误
func (am *asyncDBMetrics) RecordError() {
	am.recordMetric(func(m *dbMetrics) {
		m.RecordError()
	})
}

// RecordSlowQuery 记录慢查询
func (am *asyncDBMetrics) RecordSlowQuery() {
	am.recordMetric(func(m *dbMetrics) {
		m.RecordSlowQuery()
	})
}

// GetDroppedMetricsCount 获取丢弃的指标数量
func (am *asyncDBMetrics) GetDroppedMetricsCount() uint64 {
	return am.droppedMetrics.Load()
}
