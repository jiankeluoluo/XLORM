package xlorm

import (
	"fmt"
	"runtime/debug"
	"time"
)

// dbError 数据库错误结构体
type dbError struct {
	Query string        // 错误的 SQL 查询
	Stack string        // 错误堆栈信息
	Op    string        // 操作名称
	Time  time.Time     // 错误发生的时间
	Err   error         // 原始错误
	Args  []interface{} // 查询参数
}

// newDBError 创建数据库错误
func newDBError(op string, err error, query string, args []interface{}) *dbError {
	return &dbError{
		Op:    op,
		Err:   err,
		Query: query,
		Args:  args,
		Stack: string(debug.Stack()),
		Time:  time.Now(),
	}
}

// Error 实现error接口
func (e *dbError) Error() string {
	return fmt.Sprintf("[%s] %s: %v (Query: %s, Args: %v)",
		e.Time.Format("2006-01-02 15:04:05"),
		e.Op,
		e.Err,
		e.Query,
		e.Args,
	)
}

// Unwrap 实现errors.Unwrap接口
func (e *dbError) Unwrap() error {
	return e.Err
}
