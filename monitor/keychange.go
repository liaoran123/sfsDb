package monitor

import (
	"sync"
	"sync/atomic"
)

// 初始化函数
func init() {
	AtomicInt = make(map[string]*atomic.Int64)
	AtomicDec = make(map[string]*atomic.Int64)
}

// AtomicMap 类型用于记录键值变化
type AtomicMap map[string]*atomic.Int64

// key: string，表id,index_id组成；格式：table_id,index_id
// *atomic.Int64 记录put的键值数
var AtomicInt AtomicMap

// *atomic.Int64 记录delete的键值数
var AtomicDec AtomicMap

// 全局互斥锁，保护所有map的并发访问
var atomicMapMutex sync.RWMutex

// 格式化键名：table_id,index_id
func formatKey(tableID, indexID byte) string {
	return string(tableID) + "," + string(indexID)
}

func (m AtomicMap) Inc(tableID, indexID byte) {
	key := formatKey(tableID, indexID)
	// 加锁保护map操作
	atomicMapMutex.Lock()
	// 确保键存在，如果不存在则创建
	if m[key] == nil {
		m[key] = &atomic.Int64{}
	}
	// 获取计数器指针
	counter := m[key]
	atomicMapMutex.Unlock()
	// atomic.Int64.Add 本身是原子操作，不需要在锁内执行
	counter.Add(1)
	// 注意：不能直接在解锁后使用 m[key].Add(1)，因为：
	// 1. 解锁后 map 的访问不再受锁保护，可能出现并发安全问题
	// 2. 通过先获取指针，确保在锁内安全获取计数器引用，然后在锁外原子操作
}

func (m AtomicMap) Dec(tableID, indexID byte) {
	key := formatKey(tableID, indexID)
	// 加锁保护map操作
	atomicMapMutex.Lock()
	// 确保键存在，如果不存在则创建
	if m[key] == nil {
		m[key] = &atomic.Int64{}
	}
	// 获取计数器指针
	counter := m[key]
	atomicMapMutex.Unlock()
	// atomic.Int64.Add 本身是原子操作，不需要在锁内执行
	counter.Add(-1)
}
func (m AtomicMap) Get(tableID, indexID byte) int64 {
	key := formatKey(tableID, indexID)
	// 加读锁保护map操作
	atomicMapMutex.RLock()
	defer atomicMapMutex.RUnlock()
	if m[key] == nil {
		return 0
	}
	return m[key].Load()
}

// GetAllCounters 获取所有计数器数据
// 返回:
//   map[string]int64: put操作的计数器数据
//   map[string]int64: delete操作的计数器数据

func GetAllCounters() (map[string]int64, map[string]int64) {
	atomicMapMutex.RLock()
	defer atomicMapMutex.RUnlock()

	putCounters := make(map[string]int64)
	for key, counter := range AtomicInt {
		putCounters[key] = counter.Load()
	}

	deleteCounters := make(map[string]int64)
	for key, counter := range AtomicDec {
		deleteCounters[key] = counter.Load()
	}

	return putCounters, deleteCounters
}

// IncPut 增加 put 计数器
// 参数:
//   tableID: 表ID
//   indexID: 索引ID

func IncPut(tableID, indexID byte) {
	AtomicInt.Inc(tableID, indexID)
}

// IncDelete 增加 delete 计数器
// 参数:
//   tableID: 表ID
//   indexID: 索引ID

func IncDelete(tableID, indexID byte) {
	AtomicDec.Inc(tableID, indexID)
}

// GetPutCount 获取 put 计数器值
// 参数:
//   tableID: 表ID
//   indexID: 索引ID
// 返回:
//   int64: 计数器值

func GetPutCount(tableID, indexID byte) int64 {
	return AtomicInt.Get(tableID, indexID)
}

// GetDeleteCount 获取 delete 计数器值
// 参数:
//   tableID: 表ID
//   indexID: 索引ID
// 返回:
//   int64: 计数器值

func GetDeleteCount(tableID, indexID byte) int64 {
	return AtomicDec.Get(tableID, indexID)
}
