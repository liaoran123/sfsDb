package record

import (
	"testing"
)

// TestRecordPoolCleanliness 测试Record对象池返回的数据是否干净
func TestRecordPoolCleanliness(t *testing.T) {
	// 测试1：首次从对象池获取的Record应该是空的
	r1 := GetRecord()
	if len(r1) != 0 {
		t.Errorf("首次从对象池获取的Record长度应该为0，实际为%d", len(r1))
	}

	// 测试2：向Record中添加数据，然后放回对象池
	r1["name"] = "test"
	r1["age"] = 25
	if len(r1) != 2 {
		t.Errorf("添加数据后Record长度应该为2，实际为%d", len(r1))
	}

	// 放回对象池
	PutRecord(r1)

	// 测试3：再次从对象池获取Record，应该是空的
	r2 := GetRecord()
	if len(r2) != 0 {
		t.Errorf("从对象池获取的Record长度应该为0，实际为%d", len(r2))
	}

	// 测试4：多次获取和放回，确保数据始终干净
	for i := 0; i < 10; i++ {
		r := GetRecord()
		if len(r) != 0 {
			t.Errorf("第%d次从对象池获取的Record长度应该为0，实际为%d", i+1, len(r))
		}
		// 添加一些数据
		r["field1"] = i
		r["field2"] = "value"
		// 放回对象池
		PutRecord(r)
	}
}

// TestRecordPoolNilHandling 测试Record对象池对nil值的处理
func TestRecordPoolNilHandling(t *testing.T) {
	// 测试1：向对象池放回nil，应该不会出错
	PutRecord(nil)

	// 测试2：连续多次放回nil，应该不会出错
	for i := 0; i < 5; i++ {
		PutRecord(nil)
	}

	// 测试3：放回nil后，再次获取Record，应该仍然返回空的Record
	r := GetRecord()
	if len(r) != 0 {
		t.Errorf("放回nil后，从对象池获取的Record长度应该为0，实际为%d", len(r))
	}
	PutRecord(r)
}
