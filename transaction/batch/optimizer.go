package batch

import (
	"sync"
	"time"
)

// Optimizer 批量操作优化器
type Optimizer struct {
	// 配置参数
	minBatchSize        int           // 最小批量大小
	maxBatchSize        int           // 最大批量大小
	targetLatency       time.Duration // 目标延迟
	batchMergeThreshold int           // 批量合并阈值

	// 状态参数
	currentBatchSize     int           // 当前批量大小
	lastAdjustmentTime   time.Time     // 上次调整时间
	lastOperationLatency time.Duration // 上次操作延迟

	// 统计信息
	totalOperations  int64   // 总操作数
	totalBatches     int64   // 总批次数
	averageBatchSize float64 // 平均批量大小

	// 并发控制
	mu sync.RWMutex // 读写锁
}

// NewOptimizer 创建新的批量操作优化器
func NewOptimizer() *Optimizer {
	return &Optimizer{
		minBatchSize:        1,
		maxBatchSize:        1000,
		targetLatency:       10 * time.Millisecond,
		batchMergeThreshold: 50,
		currentBatchSize:    100,
		lastAdjustmentTime:  time.Now(),
	}
}

// GetBatchSize 获取当前批量大小
func (o *Optimizer) GetBatchSize() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentBatchSize
}

// RecordOperation 记录操作结果，用于调整批量大小
func (o *Optimizer) RecordOperation(operationCount int, latency time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// 更新统计信息
	o.totalOperations += int64(operationCount)
	o.totalBatches++
	o.averageBatchSize = float64(o.totalOperations) / float64(o.totalBatches)
	o.lastOperationLatency = latency

	// 每100次操作调整一次批量大小
	if o.totalBatches%100 == 0 {
		o.adjustBatchSize()
	}
}

// adjustBatchSize 调整批量大小
func (o *Optimizer) adjustBatchSize() {
	// 根据延迟调整批量大小
	if o.lastOperationLatency > o.targetLatency {
		// 延迟过高，减小批量大小
		o.currentBatchSize = max(o.minBatchSize, o.currentBatchSize/2)
	} else if o.lastOperationLatency < o.targetLatency/2 {
		// 延迟过低，增大批量大小
		o.currentBatchSize = min(o.maxBatchSize, o.currentBatchSize*2)
	}

	o.lastAdjustmentTime = time.Now()
}

// ShouldMergeBatches 判断是否应该合并批量
func (o *Optimizer) ShouldMergeBatches(batchSize int) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return batchSize < o.batchMergeThreshold
}

// MergeBatches 合并多个批量操作
func (o *Optimizer) MergeBatches(batches [][]interface{}) []interface{} {
	var merged []interface{}
	for _, batch := range batches {
		merged = append(merged, batch...)
	}
	return merged
}

// GetStats 获取优化器统计信息
func (o *Optimizer) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return map[string]interface{}{
		"minBatchSize":         o.minBatchSize,
		"maxBatchSize":         o.maxBatchSize,
		"currentBatchSize":     o.currentBatchSize,
		"targetLatency":        o.targetLatency,
		"lastOperationLatency": o.lastOperationLatency,
		"totalOperations":      o.totalOperations,
		"totalBatches":         o.totalBatches,
		"averageBatchSize":     o.averageBatchSize,
	}
}

// SetTargetLatency 设置目标延迟
func (o *Optimizer) SetTargetLatency(latency time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.targetLatency = latency
}

// SetBatchSizeRange 设置批量大小范围
func (o *Optimizer) SetBatchSizeRange(minVal, maxVal int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.minBatchSize = minVal
	o.maxBatchSize = maxVal
	o.currentBatchSize = min((minVal+maxVal)/2, maxVal)
}
