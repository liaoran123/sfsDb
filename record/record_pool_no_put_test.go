package record

import (
	"testing"
)

// TestRecordPoolCleanlinessWithoutPut 测试在没有显示调用PutRecord()时，Record对象池是否依然保持数据干净
func TestRecordPoolCleanlinessWithoutPut(t *testing.T) {
	// 测试1：首次从对象池获取的Record应该是空的
	r1 := GetRecord()
	if len(r1) != 0 {
		t.Errorf("首次从对象池获取的Record长度应该为0，实际为%d", len(r1))
	}

	// 测试2：向Record中添加数据，但不放回对象池
	r1["name"] = "test"
	r1["age"] = 25
	if len(r1) != 2 {
		t.Errorf("添加数据后Record长度应该为2，实际为%d", len(r1))
	}

	// 注意：这里没有调用PutRecord()

	// 测试3：再次从对象池获取Record，应该是空的
	r2 := GetRecord()
	if len(r2) != 0 {
		t.Errorf("再次从对象池获取的Record长度应该为0，实际为%d", len(r2))
	}

	// 测试4：多次获取，确保每次都返回空的Record
	for i := 0; i < 10; i++ {
		r := GetRecord()
		if len(r) != 0 {
			t.Errorf("第%d次从对象池获取的Record长度应该为0，实际为%d", i+1, len(r))
		}
		// 向Record中添加数据，但不放回对象池
		r["field1"] = i
		r["field2"] = "value"
		if len(r) != 2 {
			t.Errorf("添加数据后Record长度应该为2，实际为%d", len(r))
		}
	}
}
