package record

import (
	"fmt"
	"testing"
)

// TestGetRecordWithCapacity 测试GetRecordWithCapacity函数是否能正确获取指定初始容量的Record对象
func TestGetRecordWithCapacity(t *testing.T) {
	// 测试1：获取指定容量的Record，应该是空的
	capacity := 20
	r1 := GetRecordWithCapacity(capacity)
	if len(r1) != 0 {
		t.Errorf("从对象池获取的Record长度应该为0，实际为%d", len(r1))
	}

	// 测试2：向Record中添加数据，验证容量是否足够
	for i := 0; i < capacity; i++ {
		r1[fmt.Sprintf("field%d", i)] = i
	}
	if len(r1) != capacity {
		t.Errorf("添加数据后Record长度应该为%d，实际为%d", capacity, len(r1))
	}

	// 放回对象池
	PutRecord(r1)

	// 测试3：再次获取指定容量的Record，应该是空的
	r2 := GetRecordWithCapacity(capacity)
	if len(r2) != 0 {
		t.Errorf("从对象池获取的Record长度应该为0，实际为%d", len(r2))
	}

	// 测试4：多次获取不同容量的Record，确保数据始终干净
	capacities := []int{5, 10, 15, 25, 50}
	for _, cap := range capacities {
		r := GetRecordWithCapacity(cap)
		if len(r) != 0 {
			t.Errorf("获取容量为%d的Record长度应该为0，实际为%d", cap, len(r))
		}
		// 向Record中添加数据
		for i := 0; i < cap; i++ {
			r[fmt.Sprintf("field%d", i)] = i
		}
		if len(r) != cap {
			t.Errorf("添加数据后Record长度应该为%d，实际为%d", cap, len(r))
		}
		// 放回对象池
		PutRecord(r)
	}
}
