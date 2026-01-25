package monitor

import (
	"sync"
	"sync/atomic"
)

// 初始化函数
func init() {
	AtomicInt = make(map[string]*atomic.Int64)
	AtomicIntDec = make(map[string]*atomic.Int64)
}

// key: string，表id,index_id组成；格式：table_id,index_id
// *atomic.Int64 记录put的键值数
var AtomicInt map[string]*atomic.Int64

// *atomic.Int64 记录delete的键值数
var AtomicIntDec map[string]*atomic.Int64

// AtomicMap 类型用于记录键值变化
type AtomicMap map[string]*atomic.Int64

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
	defer atomicMapMutex.Unlock()
	// 确保键存在，如果不存在则创建
	if m[key] == nil {
		m[key] = &atomic.Int64{}
	}
	m[key].Add(1)
}

/*
	func (m AtomicMap) Dec(tableID, indexID byte) {
		key := formatKey(tableID, indexID)
		// 加锁保护map操作
		atomicMapMutex.Lock()
		defer atomicMapMutex.Unlock()
		// 确保键存在，如果不存在则创建
		if m[key] == nil {
			m[key] = &atomic.Int64{}
		}
		m[key].Add(-1)
	}
*/
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
