package engine

import (
	"testing"
)

// 基准测试：使用对象池获取和归还 map[string][]byte
func BenchmarkGetPutFieldsBytesMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 从对象池获取
		m := GetFieldsBytesMap()
		// 模拟使用
		m["key1"] = []byte("value1")
		m["key2"] = []byte("value2")
		m["key3"] = []byte("value3")
		// 归还到对象池
		PutFieldsBytesMap(m)
	}
}

// 基准测试：不使用对象池，每次创建新的 map[string][]byte
func BenchmarkMakeFieldsBytesMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次创建新的 map
		m := make(map[string][]byte)
		// 模拟使用
		m["key1"] = []byte("value1")
		m["key2"] = []byte("value2")
		m["key3"] = []byte("value3")
		// 不需要归还，由垃圾回收处理
	}
}

// 基准测试：使用对象池获取和归还 map[any]bool
func BenchmarkGetPutMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 从对象池获取
		m := GetMap()
		// 模拟使用
		m["key1"] = true
		m["key2"] = false
		m["key3"] = true
		// 归还到对象池
		PutMap(m)
	}
}

// 基准测试：不使用对象池，每次创建新的 map[any]bool
func BenchmarkMakeMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次创建新的 map
		m := make(map[any]bool)
		// 模拟使用
		m["key1"] = true
		m["key2"] = false
		m["key3"] = true
		// 不需要归还，由垃圾回收处理
	}
}

// 基准测试：使用对象池获取和归还 []string
func BenchmarkGetPutStringSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 从对象池获取
		s := GetStringSlice()
		// 模拟使用
		s = append(s, "key1")
		s = append(s, "key2")
		s = append(s, "key3")
		// 归还到对象池
		PutStringSlice(s)
	}
}

// 基准测试：不使用对象池，每次创建新的 []string
func BenchmarkMakeStringSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次创建新的 []string
		s := make([]string, 0)
		// 模拟使用
		s = append(s, "key1")
		s = append(s, "key2")
		s = append(s, "key3")
		// 不需要归还，由垃圾回收处理
	}
}

// 基准测试：不使用对象池，每次创建新的 map[string]any
func BenchmarkMakeStringAnyMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次创建新的 map[string]any
		m := make(map[string]any)
		// 模拟使用
		m["key1"] = "value1"
		m["key2"] = 123
		m["key3"] = true
		// 不需要归还，由垃圾回收处理
	}
}
