package engine

import (
	"sync"
)

// 用于管理 map[string]any 类型的对象池
var anyMapPool = sync.Pool{
	New: func() any {
		// 创建一个新的 map[string]any，初始容量设为 10，适用于大多数场景
		return make(map[string]any, 10)
	},
}

// GetAnyMap 从对象池获取一个 map[string]any 对象
func GetAnyMap() map[string]any {
	m := anyMapPool.Get().(map[string]any)
	// 清空 map，确保返回的数据是干净的
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutAnyMap 将 map[string]any 对象归还到对象池
func PutAnyMap(m map[string]any) {
	if m == nil {
		return
	}
	
	// 清空 map，确保归还的对象是干净的
	for k := range m {
		delete(m, k)
	}
	
	anyMapPool.Put(m)
}

// GetAnyMapPointer 从对象池获取一个 map[string]any 并返回其指针
func GetAnyMapPointer() *map[string]any {
	m := GetAnyMap()
	return &m
}

// PutAnyMapPointer 将 map[string]any 指针指向的对象归还到对象池
func PutAnyMapPointer(mp *map[string]any) {
	if mp == nil || *mp == nil {
		return
	}
	
	PutAnyMap(*mp)
	// 注意：这里不释放指针本身，因为指针是局部变量，会由编译器自动处理
}
