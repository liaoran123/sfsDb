package engine

import (
	"runtime"
	"sync"
)

var (
	// mapPool 用于管理 map[any]bool 类型的对象池
	mapPool = sync.Pool{
		New: func() any {

			return make(map[any]bool)
		},
	}
)

// GetMap 从对象池获取一个 map[any]bool 对象
// sync.Pool不会导致内存泄漏，确保返回的数据干净即可。
func GetMap() map[any]bool {
	m := mapPool.Get().(map[any]bool)
	// 清空 map 中的所有键值对，确保返回的数据干净。保险操作，避免数据不干净
	for k := range m {
		delete(m, k)
	}
	// 创建一个指向 map 的指针，用于设置 finalizer
	// 注意：由于 Go 中 map 是引用类型，我们需要确保 finalizer 能够正确触发
	mapPtr := &m
	//确保在对象被垃圾回收时清除其中的键值对
	runtime.SetFinalizer(mapPtr, func(ptr *map[any]bool) {
		if *ptr != nil {
			// 清空 map 中的所有键值对，确保对象池中的对象始终是干净的
			for k := range *ptr {
				delete(*ptr, k)
			}
		}
	})
	return m
}
