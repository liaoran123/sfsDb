package monitor

import "sync"

type QueryPerformance struct {
	Count     int   // 扫描次数
	TotalTime int64 // 查询总耗时（纳秒）
	Index     uint8 // 搜索时使用的索引的id
}

// QueryPerformancePool 是QueryPerformance对象池，用于减少频繁创建和销毁对象的性能开销
var QueryPerformancePool = sync.Pool{
	New: func() any {
		return &QueryPerformance{}
	},
}

// GetQueryPerformance 从对象池获取一个QueryPerformance对象
func GetQueryPerformance() *QueryPerformance {
	return QueryPerformancePool.Get().(*QueryPerformance)
}

// PutQueryPerformance 将QueryPerformance对象放回对象池，并重置其字段值
func PutQueryPerformance(perf *QueryPerformance) {
	// 重置对象字段值
	perf.Count = 0
	perf.TotalTime = 0
	QueryPerformancePool.Put(perf)
}

/*
查询性能监控
Count: 查询次数
TotalTime: 查询总耗时（纳秒）
// 执行查询前记录开始时间
startTime := time.Now()

// 执行查询操作...
result := db.Query(...)

// 计算查询耗时并更新性能数据
duration := time.Since(startTime).Nanoseconds()
*/
