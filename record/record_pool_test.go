package record

import (
	"testing"
)

// TestRecordPoolDataClean 测试从对象池获取的 Record 对象数据是否干净
func TestRecordPoolDataClean(t *testing.T) {
	// 测试 GetRecord() 函数
	for i := 0; i < 10; i++ {
		r := GetRecord()
		defer PutRecord(r)

		// 检查返回的 Record 是否为空
		if len(r) != 0 {
			t.Fatalf("GetRecord() should return an empty Record, got %d fields", len(r))
		}

		// 向 Record 中添加一些数据
		r["id"] = i
		r["name"] = "test"

		// 将 Record 放回对象池
		PutRecord(r)
	}

	// 再次从对象池获取 Record，确保返回的对象是空的
	for i := 0; i < 5; i++ {
		r := GetRecord()
		defer PutRecord(r)

		// 检查返回的 Record 是否为空
		if len(r) != 0 {
			t.Fatalf("GetRecord() should return an empty Record after being put back, got %d fields", len(r))
		}
	}
}

// TestRecordsPoolDataClean 测试从对象池获取的 Records 对象数据是否干净
func TestRecordsPoolDataClean(t *testing.T) {
	// 测试 GetRecords() 函数
	for i := 0; i < 10; i++ {
		rs := GetRecords()
		defer PutRecords(rs)

		// 检查返回的 Records 是否为空
		if len(rs) != 0 {
			t.Fatalf("GetRecords() should return an empty Records, got %d records", len(rs))
		}

		// 向 Records 中添加一些数据
		for j := 0; j < 5; j++ {
			r := GetRecord()
			r["id"] = j
			r["name"] = "test"
			rs = append(rs, r)
		}

		// 将 Records 放回对象池
		PutRecords(rs)
	}

	// 再次从对象池获取 Records，确保返回的对象是空的
	for i := 0; i < 5; i++ {
		rs := GetRecords()
		defer PutRecords(rs)

		// 检查返回的 Records 是否为空
		if len(rs) != 0 {
			t.Fatalf("GetRecords() should return an empty Records after being put back, got %d records", len(rs))
		}
	}
}

// TestRecordsWithCapacityPoolDataClean 测试从对象池获取的指定容量的 Records 对象数据是否干净
func TestRecordsWithCapacityPoolDataClean(t *testing.T) {
	capacity := 10

	// 测试 GetRecordsWithCapacity() 函数
	for i := 0; i < 10; i++ {
		rs := GetRecordsWithCapacity(capacity)
		defer PutRecords(rs)

		// 检查返回的 Records 是否为空
		if len(rs) != 0 {
			t.Fatalf("GetRecordsWithCapacity() should return an empty Records, got %d records", len(rs))
		}

		// 检查容量是否至少为指定值
		if cap(rs) < capacity {
			t.Fatalf("GetRecordsWithCapacity() should return a Records with capacity at least %d, got %d", capacity, cap(rs))
		}

		// 向 Records 中添加一些数据
		for j := 0; j < capacity; j++ {
			r := GetRecord()
			r["id"] = j
			r["name"] = "test"
			rs = append(rs, r)
		}

		// 将 Records 放回对象池
		PutRecords(rs)
	}

	// 再次从对象池获取 Records，确保返回的对象是空的
	for i := 0; i < 5; i++ {
		rs := GetRecordsWithCapacity(capacity)
		defer PutRecords(rs)

		// 检查返回的 Records 是否为空
		if len(rs) != 0 {
			t.Fatalf("GetRecordsWithCapacity() should return an empty Records after being put back, got %d records", len(rs))
		}

		// 检查容量是否至少为指定值
		if cap(rs) < capacity {
			t.Fatalf("GetRecordsWithCapacity() should return a Records with capacity at least %d, got %d", capacity, cap(rs))
		}
	}
}
