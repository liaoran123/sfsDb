package engine

import (
	"fmt"
	"testing"
)

// TestGetFieldsBytesMapWithCapacity 测试GetFieldsBytesMapWithCapacity函数是否能正确获取指定初始容量的map[string][]byte对象
func TestGetFieldsBytesMapWithCapacity(t *testing.T) {
	// 测试1：获取指定容量的map，应该是空的
	capacity := 20
	m1 := GetFieldsBytesMap()
	if len(m1) != 0 {
		t.Errorf("获取的map长度应该为0，实际为%d", len(m1))
	}

	// 测试2：向map中添加数据，验证容量是否足够
	for i := 0; i < capacity; i++ {
		key := fmt.Sprintf("field%d", i)
		m1[key] = []byte{byte(i)}
	}
	if len(m1) != capacity {
		t.Errorf("添加数据后map长度应该为%d，实际为%d", capacity, len(m1))
	}

	// 测试3：多次获取不同容量的map，确保数据始终干净
	capacities := []int{5, 10, 15, 25, 50}
	for _, cap := range capacities {
		m := GetFieldsBytesMap()
		if len(m) != 0 {
			t.Errorf("获取容量为%d的map长度应该为0，实际为%d", cap, len(m))
		}
		// 向map中添加数据
		for i := 0; i < cap; i++ {
			key := fmt.Sprintf("field%d", i)
			m[key] = []byte{byte(i)}
		}
		if len(m) != cap {
			t.Errorf("添加数据后map长度应该为%d，实际为%d", cap, len(m))
		}
	}
}
