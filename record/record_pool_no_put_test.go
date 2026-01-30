package record

import (
	"testing"
)

// TestGetRecordsCleanlinessWithoutPut 测试在外部没有调用PutRecords的情况下，GetRecords函数依然能够返回干净的数据
func TestGetRecordsCleanlinessWithoutPut(t *testing.T) {
	// 测试1：首次获取的Records应该是空的
	rs1 := GetRecords()
	if len(rs1) != 0 {
		t.Errorf("首次获取的Records长度应该为0，实际为%d", len(rs1))
	}

	// 测试2：向Records中添加数据，但不放回对象池
	rs1 = append(rs1, make(Record, 8))
	if len(rs1) != 1 {
		t.Errorf("添加数据后Records长度应该为1，实际为%d", len(rs1))
	}

	// 注意：这里没有调用PutRecords

	// 测试3：再次获取Records，应该是空的
	rs2 := GetRecords()
	if len(rs2) != 0 {
		t.Errorf("再次获取的Records长度应该为0，实际为%d", len(rs2))
	}

	// 测试4：多次获取，确保每次都返回空的Records
	for i := 0; i < 10; i++ {
		rs := GetRecords()
		if len(rs) != 0 {
			t.Errorf("第%d次获取的Records长度应该为0，实际为%d", i+1, len(rs))
		}
		// 向Records中添加数据，但不放回对象池
		rs = append(rs, make(Record, 8))
		if len(rs) != 1 {
			t.Errorf("添加数据后Records长度应该为1，实际为%d", len(rs))
		}
	}
}
