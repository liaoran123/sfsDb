package util

import (
	"sync"
)

// -----------------------------------------------------
// []byte 对象池，用于复用 []byte
var BytesPool = &sync.Pool{
	New: func() any {
		return make([]byte, 0, 128) // 初始容量 1024，可按需调整
	},
}

// 从池中获取一个 []byte
func GetBytes() []byte {
	return BytesPool.Get().([]byte)
}

// 将 []byte 归还池中，重置长度但不释放底层数组
func PutBytes(b []byte) {
	b = b[:0] // 仅截断长度，保留容量
	BytesPool.Put(b)
}

// -----------------------------------------------------
// [][]byte 对象池，用于复用二维字节数组
var BytesArrayPool = &sync.Pool{
	New: func() any {
		return make([][]byte, 0, 64) // 初始容量 64，可按需调整
	},
}

// 从池中获取一个 [][]byte
func GetBytesArray() [][]byte {
	return BytesArrayPool.Get().([][]byte)
}

// 将 [][]byte 归还池中，重置长度但不释放底层数组
func PutBytesArray(b [][]byte) {
	b = b[:0] // 仅截断长度，保留容量
	BytesArrayPool.Put(b)
}

// -----------------------------------------------------
// map[string]any 对象池，用于复用 map[string]any
var MapAnyPool = &sync.Pool{
	New: func() any {
		return make(map[string]any)
	},
}

// 从池中获取一个 map[string]any
func GetMapAny() map[string]any {
	return MapAnyPool.Get().(map[string]any)
}

// 将 map[string]any 归还池中，清空 map 内容
func PutMapAny(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	MapAnyPool.Put(m)
}

// -----------------------------------------------------
// []map[string]any 对象池，用于复用切片 map[string]any
var MapsAnyPool = &sync.Pool{
	New: func() any {
		return make([]map[string]any, 0, 16) // 初始容量 16，可按需调整
	},
}

// 从池中获取一个 []map[string]any
func GetMapsAny() []map[string]any {
	return MapsAnyPool.Get().([]map[string]any)
}

// 将 []map[string]any 归还池中，清空切片内容
func PutMapsAny(m []map[string]any) {
	// 先逐个清空内部 map
	for i := range m {
		for k := range m[i] {
			delete(m[i], k)
		}
	}
	// 截断切片长度，保留容量
	m = m[:0]
	MapsAnyPool.Put(m)
}

// -----------------------------------------------------
// []string 对象池，用于复用字符串切片
var StringSlicePool = &sync.Pool{
	New: func() any {
		return make([]string, 0, 64) // 初始容量 64，可按需调整
	},
}

// 从池中获取一个 []string
func GetStringSlice() []string {
	return StringSlicePool.Get().([]string)
}

// 将 []string 归还池中，重置长度但不释放底层数组
func PutStringSlice(s []string) {
	s = s[:0] // 仅截断长度，保留容量
	StringSlicePool.Put(s)
}
