package engine

import (
	"sync"
)

var (
	// mapPool 用于管理 map[any]bool 类型的对象池
	mapPool = sync.Pool{
		New: func() any {

			return make(map[any]bool)
		},
	}
	/*小对象，对象池操作开销可能大于直接创建
	// fieldsBytesPool 用于管理 map[string][]byte 类型的对象池
	fieldsBytesPool = sync.Pool{
		New: func() any {
			return make(map[string][]byte)
		},
	}
	*/
	// stringSlicePool 用于管理 []string 类型的对象池
	stringSlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 10) // 预分配容量为 10
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
	return m
}

// PutMap 将 map[any]bool 对象归还到对象池
func PutMap(m map[any]bool) {
	// 清空 map 中的所有键值对，确保归还的对象干净
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

/*
// GetFieldsBytesMap 从对象池获取一个 map[string][]byte 对象
func GetFieldsBytesMap() map[string][]byte {
	m := fieldsBytesPool.Get().(map[string][]byte)
	// 清空 map 中的所有键值对，确保返回的数据干净
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutFieldsBytesMap 将 map[string][]byte 对象归还到对象池
func PutFieldsBytesMap(m map[string][]byte) {
	// 清空 map 中的所有键值对，确保归还的对象干净
	for k := range m {
		delete(m, k)
	}
	fieldsBytesPool.Put(m)
}

// ResetFieldsBytesPool 重置 fieldsBytesPool 对象池
func ResetFieldsBytesPool() {
	// 由于 sync.Pool 没有直接的重置方法，我们可以通过替换来实现
	fieldsBytesPool = sync.Pool{
		New: func() any {
			return make(map[string][]byte)
		},
	}
}
*/
// GetStringSlice 从对象池获取一个 []string 切片
func GetStringSlice() []string {
	s := stringSlicePool.Get().([]string)
	// 清空切片，确保返回的切片是空的
	s = s[:0]
	return s
}

// PutStringSlice 将 []string 切片归还到对象池
func PutStringSlice(s []string) {
	// 清空切片，确保归还的切片是空的
	s = s[:0]
	stringSlicePool.Put(s)
}

// ResetStringSlicePool 重置 stringSlicePool 对象池
func ResetStringSlicePool() {
	// 由于 sync.Pool 没有直接的重置方法，我们可以通过替换来实现
	stringSlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 10)
		},
	}
}
