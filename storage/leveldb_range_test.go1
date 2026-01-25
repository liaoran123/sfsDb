package storage

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb/util"
)

// TestLevelDBRange 测试LevelDB Range的各种用法
func TestLevelDBRange(t *testing.T) {
	// 运行区间取值示例
	ExampleLevelDBRange()

	// 运行实际使用示例
	LevelDBRangeUsage()
}

// TestKeyRangeHelper 测试键范围辅助函数
func TestKeyRangeHelper(t *testing.T) {
	// 测试闭区间辅助函数
	closedRange := KeyRange([]byte("a"), []byte("z"))
	if closedRange.Start == nil || closedRange.Limit == nil {
		t.Fatal("KeyRange返回了nil范围")
	}

	// 测试左开右开区间辅助函数
	exclusiveRange := ExclusiveKeyRange([]byte("a"), []byte("z"))
	if exclusiveRange.Start == nil || exclusiveRange.Limit == nil {
		t.Fatal("ExclusiveKeyRange返回了nil范围")
	}

	// 测试左开右闭区间辅助函数
	openClosedRange := OpenClosedKeyRange([]byte("a"), []byte("z"))
	if openClosedRange.Start == nil || openClosedRange.Limit == nil {
		t.Fatal("OpenClosedKeyRange返回了nil范围")
	}

	// 测试左闭右开区间辅助函数
	closedOpenRange := ClosedOpenKeyRange([]byte("a"), []byte("z"))
	if closedOpenRange.Start == nil || closedOpenRange.Limit == nil {
		t.Fatal("ClosedOpenKeyRange返回了nil范围")
	}

	// 验证辅助函数生成的范围是否符合预期
	if string(closedRange.Start) != "a" {
		t.Errorf("KeyRange Start错误，预期: 'a', 实际: '%s'", closedRange.Start)
	}

	if string(closedOpenRange.Limit) != "z" {
		t.Errorf("ClosedOpenKeyRange Limit错误，预期: 'z', 实际: '%s'", closedOpenRange.Limit)
	}
}

// TestBytesPrefix 测试BytesPrefix函数的使用
func TestBytesPrefix(t *testing.T) {
	// 测试常见前缀
	prefixes := [][]byte{
		[]byte("user_"),
		[]byte("order_"),
		[]byte("cache_"),
		[]byte("log_"),
	}

	for _, prefix := range prefixes {
		rangeObj := util.BytesPrefix(prefix)
		if rangeObj == nil {
			t.Errorf("BytesPrefix返回了nil范围: %s", prefix)
			continue
		}

		// 验证生成的范围是否符合预期
		if string(rangeObj.Start) != string(prefix) {
			t.Errorf("BytesPrefix Start错误，预期: '%s', 实际: '%s'", prefix, rangeObj.Start)
		}
	}
}
