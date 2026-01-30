package record

import (
	"testing"
)

// TestGetRecordsCleanliness 测试GetRecords函数返回的数据是否干净
func TestGetRecordsCleanliness(t *testing.T) {
	// 测试1：首次获取的Records应该是空的
	rs1 := GetRecords()
	if len(rs1) != 0 {
		t.Errorf("首次获取的Records长度应该为0，实际为%d", len(rs1))
	}

	// 测试2：向Records中添加数据，然后放回对象池
	rs1 = append(rs1, make(Record, 8))
	if len(rs1) != 1 {
		t.Errorf("添加数据后Records长度应该为1，实际为%d", len(rs1))
	}

	// 放回对象池
	PutRecords(rs1)

	// 测试3：再次从对象池获取Records，应该是空的
	rs2 := GetRecords()
	if len(rs2) != 0 {
		t.Errorf("从对象池获取的Records长度应该为0，实际为%d", len(rs2))
	}

	// 测试4：多次获取和放回，确保数据始终干净
	for i := 0; i < 10; i++ {
		rs := GetRecords()
		if len(rs) != 0 {
			t.Errorf("第%d次获取的Records长度应该为0，实际为%d", i+1, len(rs))
		}
		// 添加一些数据
		for j := 0; j < 5; j++ {
			rs = append(rs, make(Record, 8))
		}
		// 放回对象池
		PutRecords(rs)
	}
}

// TestGetRecordsWithCapacityCleanliness 测试GetRecordsWithCapacity函数返回的数据是否干净
func TestGetRecordsWithCapacityCleanliness(t *testing.T) {
	// 测试1：获取指定容量的Records应该是空的
	capacity := 10
	rs1 := GetRecordsWithCapacity(capacity)
	if len(rs1) != 0 {
		t.Errorf("获取的Records长度应该为0，实际为%d", len(rs1))
	}

	// 测试2：向Records中添加数据，然后放回对象池
	rs1 = append(rs1, make(Record, 8))
	if len(rs1) != 1 {
		t.Errorf("添加数据后Records长度应该为1，实际为%d", len(rs1))
	}

	// 放回对象池
	PutRecords(rs1)

	// 测试3：再次获取指定容量的Records，应该是空的
	rs2 := GetRecordsWithCapacity(capacity)
	if len(rs2) != 0 {
		t.Errorf("从对象池获取的Records长度应该为0，实际为%d", len(rs2))
	}
}

// TestMakeRecordCleanliness 测试make(Record, 8)返回的数据是否干净
func TestMakeRecordCleanliness(t *testing.T) {
	// 测试1：创建的Record应该是空的
	r := make(Record, 8)
	if len(r) != 0 {
		t.Errorf("创建的Record长度应该为0，实际为%d", len(r))
	}

	// 测试2：多次创建，确保每次都是新的空Record
	for i := 0; i < 10; i++ {
		r := make(Record, 8)
		if len(r) != 0 {
			t.Errorf("第%d次创建的Record长度应该为0，实际为%d", i+1, len(r))
		}
	}
}
