package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jiankeluoluo/xlorm"
)

func main() {
	cfg := &xlorm.Config{
		DBName:             "master", // 数据库
		Driver:             "MySQL",
		Host:               "localhost",
		Port:               3306,
		Database:           "test_db",
		Username:           "root",
		Password:           "root",
		Charset:            "utf8mb4",
		TablePrefix:        "test_",
		LogDir:             "./logs",
		LogLevel:           "debug",
		Debug:              true,
		EnablePoolStats:    true, // 是否启用性能指标收集
		LogRotationEnabled: true, // 启用日志轮转
		LogRotationMaxAge:  30,   // 日志保留30天
	}

	e, err := xlorm.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer e.Close()

	// // 测试数据库
	TestDb(e)

	// 测试结构体形式
	TestStruct(e)

	// // 执行日志测试
	TestLoger(e)

	// // // // 获取日志指标
	TestLogMetrics(e)

	// // // 获取连接池统计
	TestPoolStats(e)

	// // // 测试数据库指标
	TestDbMetrics(e)

}

// 测试日志
func TestLoger(e *xlorm.DB) {
	// 记录初始日志级别
	fmt.Println("初始日志级别：", e.GetLogLevel())

	// 修改日志级别并打印
	_ = e.SetLogLevel("info")
	fmt.Println("修改后的级别：", e.GetLogLevel())

	// 修改日志级别并打印
	_ = e.SetLogLevel("Debug")
	fmt.Println("修改后的级别：", e.GetLogLevel())
	// 获取日志实例
	logger := e.Logger()
	// 记录不同级别的日志
	logger.Debug("这是一条调试日志")
	logger.Info("这是一条信息日志")
	logger.Warn("这是一条警告日志")
	logger.Error("这是一条错误日志")

	// 使用LogAttrs提升性能
	logger.LogAttrs(context.Background(), slog.LevelDebug,
		"SQL详情",
		slog.String("query", "ababa"),
		slog.Any("args", "balla"),
	)

	logger.Error("这是一条错误日志lalalalla")

	// 记录不同级别的日志
	logger.Info("数据库操作",
		slog.String("operation", "query"),
		slog.Int("user_id", 123),
	)

	logger.Error("查询失败",
		slog.String("error", "查询失败"),
		slog.String("query", "2222"),
	)

	// 使用LogAttrs提升性能
	logger.LogAttrs(context.Background(), slog.LevelDebug,
		"SQL详情",
		slog.String("query", "ababa"),
		slog.Any("args", "balla"),
	)
	time.Sleep(100 * time.Millisecond) // 给异步日志处理一些时间
}

// 测试日志指标
func TestLogMetrics(e *xlorm.DB) {
	logMetrics := e.AsyncLogger().GetLogMetrics()

	// 打印日志指标
	fmt.Printf("日志指标:\n")
	fmt.Printf("总日志数: %d\n", logMetrics["total_logs"])
	fmt.Printf("丢弃的日志数: %d\n", logMetrics["dropped_logs"])
	fmt.Printf("日志通道深度: %d\n", logMetrics["channel_depth"])

	asyncLogger := e.AsyncLogger()
	// 获取总处理日志数量
	totalLogsCount := asyncLogger.GetTotalLogsCount()
	// 获取丢弃的日志数量(异步阻塞处理不及时被丢弃的日志)
	droppedLogsCount := asyncLogger.GetDroppedLogsCount()

	fmt.Printf("总日志数: %d\n", totalLogsCount)
	fmt.Printf("丢弃的日志数: %d\n", droppedLogsCount)
}

// 测试连接池指标
func TestPoolStats(e *xlorm.DB) {
	stats := e.GetPoolStats()
	// 打印连接池信息
	fmt.Printf("连接池状态:\n")
	fmt.Printf("最大连接数: %d\n", stats.MaxOpenConnections)
	fmt.Printf("当前打开连接数: %d\n", stats.OpenConnections)
	fmt.Printf("正在使用的连接数: %d\n", stats.InUse)
	fmt.Printf("空闲连接数: %d\n", stats.Idle)
	fmt.Printf("等待连接次数: %d\n", stats.WaitCount)
	fmt.Printf("等待连接总时间: %v\n", stats.WaitDuration)
	fmt.Printf("因超过最大空闲时间关闭的连接数: %d\n", stats.MaxIdleClosed)
	fmt.Printf("因超过最大生命周期关闭的连接数: %d\n", stats.MaxLifetimeClosed)
}

// 测试数据库指标
func TestDbMetrics(e *xlorm.DB) {
	stats := e.DBMetrics()
	metrics := stats.GetDBMetrics()

	// 打印性能指标
	// 打印基本指标
	fmt.Printf("数据库性能指标:\n")
	fmt.Printf("数据库名称: %s\n", metrics["db_name"])
	fmt.Printf("总查询数: %d\n", metrics["total_queries"])
	fmt.Printf("慢查询数: %d\n", metrics["slow_queries"])
	fmt.Printf("查询错误数: %d\n", metrics["total_errors"])
	fmt.Printf("影响的总行数: %d\n", metrics["total_affected_rows"])

	// 打印查询耗时统计
	queryStats := metrics["query_stats"].(map[string]interface{})
	for queryType, stat := range queryStats {
		statMap := stat.(map[string]interface{})
		fmt.Printf("查询类型 %s 的耗时统计:\n", queryType)
		fmt.Printf("查询次数: %d\n", statMap["count"])
		fmt.Printf("总耗时: %v\n", statMap["total_time"])
		fmt.Printf("平均耗时: %v\n", statMap["average_time"])
	}
}

func TestDb(e *xlorm.DB) {
	// 测试插入
	insert(e)

	// 测试更新
	update(e)

	// 测试删除
	delete(e)

	// 测试查询
	find(e)

	// 测试查询全部
	findAll(e)

	// 测试批量插入
	batchInsert(e)

	// 测试批量更新
	batchUpdate(e)
	// 事务处理
	transaction(e)

}

// 插入数据
func insert(e *xlorm.DB) {
	data := map[string]interface{}{
		"name": "测试用户",
		"age":  25,
	}
	id, err := e.M("users").Insert(data)
	if err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}
	fmt.Printf("最后插入的ID: %d\n", id)
}

// // 批量插入
func batchInsert(e *xlorm.DB) {
	records := []map[string]interface{}{
		{"name": "用户1", "age": 20},
		{"name": "用户2", "age": 21},
		{"name": "用户3", "age": 22},
	}
	result, err := e.M("users").BatchInsert(records, 500)
	if err != nil {
		fmt.Printf("批量插入失败: %v\n", err)
		return
	}
	fmt.Printf("批量插入成功，影响行数: %d\n", result)
}

// 批量更新
func batchUpdate(e *xlorm.DB) {
	// 示例2：批量更新
	updates := []map[string]interface{}{
		{
			"id":   1,
			"name": "张三(已更新)",
			"age":  30,
		},
		{
			"id":   2,
			"name": "李四(已更新)",
			"age":  31,
		},
	}

	affected, err := e.M("users").BatchUpdate(updates, "id", 500)
	if err != nil {
		log.Printf("批量更新失败: %v", err)
	} else {
		log.Printf("成功更新 %d 条记录", affected)
	}
}

// 查询单条
func find(e *xlorm.DB) {
	result, err := e.M("users").Fields("id,name,age").Where("id = ?", 11).Find()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("查询结果: %+v\n", result)
}

// 查询多条
func findAll(e *xlorm.DB) {
	users, err := e.M("users").
		Where("age > ?", 30).
		OrderBy("age DESC").
		FindAll()

	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		for _, user := range users {
			fmt.Printf("用户ID: %d, 姓名: %s, 年龄: %d\n",
				user["id"], user["name"], user["age"])
		}
	}
}

// 更新数据
func update(e *xlorm.DB) {
	data := map[string]interface{}{
		"age": 26,
	}
	result, err := e.M("users").Where("id = ?", 1).Update(data)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}
	fmt.Printf("更新成功，影响行数: %d\n", result)
}

// 删除数据
func delete(e *xlorm.DB) {
	result, err := e.M("users").Where("id = ?", 1).Delete()
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Printf("删除成功，影响行数: %d\n", result)
}

// 事务处理
func transaction(e *xlorm.DB) {
	// 示例3：事务中的批量操作
	tx, err := e.Begin()
	if err != nil {
		log.Fatalf("开启事务失败: %v", err)
	}

	// 在事务中执行批量插入
	newRecords := []map[string]interface{}{
		{
			"name": "赵六",
			"age":  23,
		},
		{
			"name": "钱七",
			"age":  24,
		},
	}

	affected, err := tx.DB().M("users").BatchInsert(newRecords, 500)
	if err != nil {
		tx.Rollback()
		log.Printf("事务中批量插入失败: %v", err)
	} else {
		log.Printf("事务中成功插入 %d 条记录", affected)
		// 在事务中执行批量更新
		newUpdates := []map[string]interface{}{
			{
				"id":   1,
				"name": "张三(再次更新)",
				"age":  40,
			},
		}

		affected, err = tx.DB().M("users").BatchUpdate(newUpdates, "id", 500)
		if err != nil {
			tx.Rollback()
			log.Printf("事务中批量更新失败: %v", err)
		} else {
			// 提交事务
			err = tx.Commit()
			if err != nil {
				log.Printf("提交事务失败: %v", err)
			} else {
				log.Printf("事务提交成功")
			}
		}
		log.Printf("事务中成功更新 %d 条记录", affected)
		log.Printf("事务结束")
	}
}

// 测试结构体形式
func TestStruct(e *xlorm.DB) {
	// User 示例结构体
	type User struct {
		Id   int    `db:"id,primary"`
		Name string `db:"name,callback=upper"`
		Age  int    `db:"age"`
	}
	// 示例1：插入
	data := &User{
		Name: "测试用户",
		Age:  25,
	}
	id, err := e.M("users").Insert(data)
	if err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}
	fmt.Printf("最后插入的ID: %d\n", id)
}
