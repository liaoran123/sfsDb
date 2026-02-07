package engine

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// StrategyType 策略类型
type StrategyType int

const (
	// StrategyStatic 静态分类策略
	StrategyStatic StrategyType = iota
	// StrategyDynamic 动态分析策略
	StrategyDynamic
	// StrategyHybrid 混合策略（静态+动态）
	StrategyHybrid
)

// StrategyConfig 策略配置
type StrategyConfig struct {
	StrategyType       StrategyType
	ConcurrencyThreshold int32
	MemoryThreshold    uint64
	RequestThreshold   int
}

// 默认配置
var defaultConfig = StrategyConfig{
	StrategyType:       StrategyStatic,
	ConcurrencyThreshold: 1000,
	MemoryThreshold:    500, // MB
	RequestThreshold:   1000, // 每分钟请求数
}

// 全局配置和锁
var (
	strategyConfig StrategyConfig
	configMutex    sync.RWMutex
)

// SetStrategyConfig 设置策略配置
func SetStrategyConfig(config StrategyConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	strategyConfig = config
}

// GetStrategyConfig 获取策略配置
func GetStrategyConfig() StrategyConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return strategyConfig
}

// 系统负载监控
var (
	concurrentRequests int32
	peakRequests      int32
	lastResetTime     time.Time
)

// 时间窗口统计
type timeWindowCounter struct {
	requests    []int
	windowSize  int
	currentIdx  int
	total       int
	lastUpdated time.Time
	mu          sync.Mutex
}

var requestCounter = timeWindowCounter{
	requests:    make([]int, 60), // 60秒窗口
	windowSize:  60,
	currentIdx:  0,
	total:       0,
	lastUpdated: time.Now(),
}

// 更新时间窗口
func (tc *timeWindowCounter) update() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
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

// 记录请求
func (tc *timeWindowCounter) record() {
	tc.update()
	
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	tc.requests[tc.currentIdx]++
	tc.total++
}

// 获取时间窗口内的总请求数
func (tc *timeWindowCounter) getTotal() int {
	tc.update()
	
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	return tc.total
}

// 记录请求开始
func recordRequestStart() {
	current := atomic.AddInt32(&concurrentRequests, 1)
	if current > atomic.LoadInt32(&peakRequests) {
		atomic.StoreInt32(&peakRequests, current)
	}
	
	// 定期重置统计
	if time.Since(lastResetTime) > time.Minute {
		atomic.StoreInt32(&concurrentRequests, 0)
		atomic.StoreInt32(&peakRequests, 0)
		lastResetTime = time.Now()
	}
	
	// 记录到时间窗口
	requestCounter.record()
}

// 记录请求结束
func recordRequestEnd() {
	atomic.AddInt32(&concurrentRequests, -1)
}

// 获取当前内存使用
func getCurrentMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc / (1024 * 1024) // 转换为MB
}

// 获取当前系统状态
func getSystemState() (int32, uint64, int) {
	concurrency := atomic.LoadInt32(&concurrentRequests)
	memory := getCurrentMemoryUsage()
	requests := requestCounter.getTotal()
	return concurrency, memory, requests
}

// GetStringSliceWithDynamicStrategy 动态策略获取 []string 切片
func GetStringSliceWithDynamicStrategy() []string {
	recordRequestStart()
	defer recordRequestEnd()
	
	concurrency, memory, requests := getSystemState()
	config := GetStrategyConfig()
	
	// 基于系统状态判断
	if concurrency > config.ConcurrencyThreshold || 
	   memory > config.MemoryThreshold || 
	   requests > config.RequestThreshold {
		return GetStringSlice() // 高负载时使用对象池
	}
	return make([]string, 0, 10) // 低负载时直接创建
}

// PutStringSliceWithDynamicStrategy 动态策略归还 []string 切片
func PutStringSliceWithDynamicStrategy(s []string) {
	concurrency, memory, requests := getSystemState()
	config := GetStrategyConfig()
	
	// 基于系统状态判断
	if concurrency > config.ConcurrencyThreshold || 
	   memory > config.MemoryThreshold || 
	   requests > config.RequestThreshold {
		PutStringSlice(s) // 高负载时归还到对象池
	}
	// 低负载时无需处理，由垃圾回收自动回收
}

// GetStringSliceWithConfigStrategy 根据配置选择策略
func GetStringSliceWithConfigStrategy() []string {
	config := GetStrategyConfig()
	
	switch config.StrategyType {
	case StrategyStatic:
		return GetStringSliceWithStrategy()
	case StrategyDynamic:
		return GetStringSliceWithDynamicStrategy()
	case StrategyHybrid:
		// 混合策略：默认使用静态策略，高负载时切换到动态策略
		concurrency, _, _ := getSystemState()
		if concurrency > config.ConcurrencyThreshold {
			return GetStringSliceWithDynamicStrategy()
		}
		return GetStringSliceWithStrategy()
	default:
		return GetStringSliceWithStrategy()
	}
}

// PutStringSliceWithConfigStrategy 根据配置选择归还策略
func PutStringSliceWithConfigStrategy(s []string) {
	config := GetStrategyConfig()
	
	switch config.StrategyType {
	case StrategyStatic:
		PutStringSliceWithStrategy(s)
	case StrategyDynamic:
		PutStringSliceWithDynamicStrategy(s)
	case StrategyHybrid:
		// 混合策略：默认使用静态策略，高负载时切换到动态策略
		concurrency, _, _ := getSystemState()
		if concurrency > config.ConcurrencyThreshold {
			PutStringSliceWithDynamicStrategy(s)
		} else {
			PutStringSliceWithStrategy(s)
		}
	default:
		PutStringSliceWithStrategy(s)
	}
}

// SwitchToStaticStrategy 切换到静态策略
func SwitchToStaticStrategy() {
	config := GetStrategyConfig()
	config.StrategyType = StrategyStatic
	SetStrategyConfig(config)
}

// SwitchToDynamicStrategy 切换到动态策略
func SwitchToDynamicStrategy() {
	config := GetStrategyConfig()
	config.StrategyType = StrategyDynamic
	SetStrategyConfig(config)
}

// SwitchToHybridStrategy 切换到混合策略
func SwitchToHybridStrategy() {
	config := GetStrategyConfig()
	config.StrategyType = StrategyHybrid
	SetStrategyConfig(config)
}

// UpdateStrategyThresholds 更新策略阈值
func UpdateStrategyThresholds(concurrencyThreshold int32, memoryThreshold uint64, requestThreshold int) {
	config := GetStrategyConfig()
	config.ConcurrencyThreshold = concurrencyThreshold
	config.MemoryThreshold = memoryThreshold
	config.RequestThreshold = requestThreshold
	SetStrategyConfig(config)
}

// 初始化配置
func init() {
	strategyConfig = defaultConfig
	lastResetTime = time.Now()
}
