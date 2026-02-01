package metrics

import (
	"sync/atomic"
	"time"
)

// Metrics 指标收集器接口
type Metrics interface {
	// 事务指标
	RecordTransaction(duration time.Duration, success bool)
	GetTransactionCount() int64
	GetTransactionSuccessRate() float64
	GetAverageTransactionDuration() time.Duration

	// 查询指标
	RecordQuery(duration time.Duration, queryType string)
	GetQueryCount() int64
	GetAverageQueryDuration() time.Duration

	// 存储指标
	RecordStorageOperation(opType string, duration time.Duration)
	GetStorageOperationCount() int64

	// 系统指标
	RecordSystemState(concurrency int32, memoryUsage uint64, requests int)
	GetCurrentSystemState() (int32, uint64, int)

	// 重置指标
	Reset()
}

// timeWindowCounter 时间窗口计数器
type timeWindowCounter struct {
	requests    []int
	windowSize  int
	currentIdx  int
	total       int
	lastUpdated time.Time
}

// NewTimeWindowCounter 创建新的时间窗口计数器
func NewTimeWindowCounter(windowSize int) *timeWindowCounter {
	return &timeWindowCounter{
		requests:    make([]int, windowSize),
		windowSize:  windowSize,
		currentIdx:  0,
		total:       0,
		lastUpdated: time.Now(),
	}
}

// Update 更新时间窗口
func (tc *timeWindowCounter) Update() {
	now := time.Now()
	elapsed := int(now.Sub(tc.lastUpdated).Seconds())

	if elapsed > 0 {
		// 清除过期的计数
		for i := 0; i < elapsed && i < tc.windowSize; i++ {
			tc.total -= tc.requests[tc.currentIdx]
			tc.requests[tc.currentIdx] = 0
			tc.currentIdx = (tc.currentIdx + 1) % tc.windowSize
		}
		tc.lastUpdated = now
	}
}

// Record 记录一个事件
func (tc *timeWindowCounter) Record() {
	tc.Update()
	tc.requests[tc.currentIdx]++
	tc.total++
}

// GetTotal 获取窗口内的总计数
func (tc *timeWindowCounter) GetTotal() int {
	tc.Update()
	return tc.total
}

// defaultMetrics 默认指标收集器实现
type defaultMetrics struct {
	// 事务指标
	transactionCount       int64
	transactionSuccessCount int64
	totalTransactionDuration int64 // 纳秒

	// 查询指标
	queryCount            int64
	totalQueryDuration    int64 // 纳秒
	queryTypeCounts       map[string]int64

	// 存储指标
	storageOperationCount int64

	// 系统指标
	concurrency          int32
	memoryUsage          uint64
	requests             int
	requestCounter       *timeWindowCounter
}

// NewDefaultMetrics 创建默认指标收集器
func NewDefaultMetrics() Metrics {
	return &defaultMetrics{
		queryTypeCounts: make(map[string]int64),
		requestCounter:  NewTimeWindowCounter(60), // 60秒窗口
	}
}

// RecordTransaction 记录事务指标
func (m *defaultMetrics) RecordTransaction(duration time.Duration, success bool) {
	atomic.AddInt64(&m.transactionCount, 1)
	if success {
		atomic.AddInt64(&m.transactionSuccessCount, 1)
	}
	atomic.AddInt64(&m.totalTransactionDuration, duration.Nanoseconds())
}

// GetTransactionCount 获取事务计数
func (m *defaultMetrics) GetTransactionCount() int64 {
	return atomic.LoadInt64(&m.transactionCount)
}

// GetTransactionSuccessRate 获取事务成功率
func (m *defaultMetrics) GetTransactionSuccessRate() float64 {
	count := atomic.LoadInt64(&m.transactionCount)
	if count == 0 {
		return 0
	}
	success := atomic.LoadInt64(&m.transactionSuccessCount)
	return float64(success) / float64(count)
}

// GetAverageTransactionDuration 获取平均事务持续时间
func (m *defaultMetrics) GetAverageTransactionDuration() time.Duration {
	count := atomic.LoadInt64(&m.transactionCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalTransactionDuration)
	return time.Duration(total / count)
}

// RecordQuery 记录查询指标
func (m *defaultMetrics) RecordQuery(duration time.Duration, queryType string) {
	atomic.AddInt64(&m.queryCount, 1)
	atomic.AddInt64(&m.totalQueryDuration, duration.Nanoseconds())
	// 简单的并发安全处理，实际生产环境可能需要更复杂的方案
	m.queryTypeCounts[queryType]++
}

// GetQueryCount 获取查询计数
func (m *defaultMetrics) GetQueryCount() int64 {
	return atomic.LoadInt64(&m.queryCount)
}

// GetAverageQueryDuration 获取平均查询持续时间
func (m *defaultMetrics) GetAverageQueryDuration() time.Duration {
	count := atomic.LoadInt64(&m.queryCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalQueryDuration)
	return time.Duration(total / count)
}

// RecordStorageOperation 记录存储操作指标
func (m *defaultMetrics) RecordStorageOperation(opType string, duration time.Duration) {
	atomic.AddInt64(&m.storageOperationCount, 1)
}

// GetStorageOperationCount 获取存储操作计数
func (m *defaultMetrics) GetStorageOperationCount() int64 {
	return atomic.LoadInt64(&m.storageOperationCount)
}

// RecordSystemState 记录系统状态
func (m *defaultMetrics) RecordSystemState(concurrency int32, memoryUsage uint64, requests int) {
	atomic.StoreInt32(&m.concurrency, concurrency)
	atomic.StoreUint64(&m.memoryUsage, memoryUsage)
	m.requests = requests
	m.requestCounter.Record()
}

// GetCurrentSystemState 获取当前系统状态
func (m *defaultMetrics) GetCurrentSystemState() (int32, uint64, int) {
	return atomic.LoadInt32(&m.concurrency),
		atomic.LoadUint64(&m.memoryUsage),
		m.requestCounter.GetTotal()
}

// Reset 重置指标
func (m *defaultMetrics) Reset() {
	atomic.StoreInt64(&m.transactionCount, 0)
	atomic.StoreInt64(&m.transactionSuccessCount, 0)
	atomic.StoreInt64(&m.totalTransactionDuration, 0)
	atomic.StoreInt64(&m.queryCount, 0)
	atomic.StoreInt64(&m.totalQueryDuration, 0)
	atomic.StoreInt64(&m.storageOperationCount, 0)
	atomic.StoreInt32(&m.concurrency, 0)
	atomic.StoreUint64(&m.memoryUsage, 0)
	m.requests = 0
	m.queryTypeCounts = make(map[string]int64)
	m.requestCounter = NewTimeWindowCounter(60)
}
