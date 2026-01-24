package monitor

import "sync/atomic"

// 初始化函数
func init() {
	AtomicInt = make(map[string]*atomic.Int64)
	AtomicIntDec = make(map[string]*atomic.Int64)
}

//key: string，表id,索引id组成；格式：table_id,index_id
//*atomic.Int64 记录put的键值数
var AtomicInt map[string]*atomic.Int64

//*atomic.Int64 记录delete的键值数
var AtomicIntDec map[string]*atomic.Int64

type AtomicMap map[string]*atomic.Int64

// 格式化键名：table_id,index_id
func formatKey(tableID, indexID byte) string {
	return string(tableID) + "," + string(indexID)
}

func (m AtomicMap) Inc(tableID, indexID byte) {
	key := formatKey(tableID, indexID)
	// 确保键存在，如果不存在则创建
	if m[key] == nil {
		m[key] = &atomic.Int64{}
	}
	m[key].Add(1)
}

/*
func (m AtomicMap) Dec(tableID, indexID byte) {
	key := formatKey(tableID, indexID)
	// 确保键存在，如果不存在则创建
	if m[key] == nil {
		m[key] = &atomic.Int64{}
	}
	m[key].Add(-1)
}
*/
func (m AtomicMap) Get(tableID, indexID byte) int64 {
	key := formatKey(tableID, indexID)
	if m[key] == nil {
		return 0
	}
	return m[key].Load()
}
