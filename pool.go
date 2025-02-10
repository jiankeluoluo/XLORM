package xlorm

import (
	"database/sql"
	"sync"
	"sync/atomic"
)

// 对象池定义
var tablePool = sync.Pool{
	New: func() interface{} {
		return &Table{
			where: make([]string, 0, 4),
			args:  make([]interface{}, 0, 4),
			joins: make([]string, 0, 2),
		}
	},
}

var builderPool = sync.Pool{
	New: func() interface{} {
		return &builder{
			fields: make([]string, 0, 8),
			where:  make([]string, 0, 4),
			args:   make([]interface{}, 0, 4),
			joins:  make([]string, 0, 2),
		}
	},
}

// dbPoolStats 连接池统计信息
type dbPoolStats struct {
	stats atomic.Pointer[sql.DBStats]
}

// init 初始化连接池统计信息
func (p *dbPoolStats) init() {
	defaultStats := &sql.DBStats{}
	p.update(defaultStats)
}

// update 更新连接池统计信息
func (p *dbPoolStats) update(newStats *sql.DBStats) {
	p.stats.Store(newStats)
}

// get 获取连接池统计信息
func (p *dbPoolStats) get() *sql.DBStats {
	return p.stats.Load()
}

var poolStats = &dbPoolStats{
	stats: atomic.Pointer[sql.DBStats]{},
}
