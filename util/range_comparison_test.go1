package util

import (
	"testing"
)

func TestFromComparisonWithNewOperators(t *testing.T) {
	// 创建 RangeHelper 实例
	h := NewRangeHelper([]byte(""))

	// 测试 Prefix 操作符
	prefixRange := h.FromComparison(Prefix, []byte("prefix"))
	expectedStart := []byte("prefix")
	expectedLimit := BytesPrefix([]byte("prefix")).Limit

	if string(prefixRange.Start) != string(expectedStart) {
		t.Errorf("Prefix start range mismatch: expected %q, got %q", expectedStart, prefixRange.Start)
	}

	if string(prefixRange.Limit) != string(expectedLimit) {
		t.Errorf("Prefix limit range mismatch: expected %q, got %q", expectedLimit, prefixRange.Limit)
	}

	// 测试 Suffix 操作符
	suffixRange := h.FromComparison(Suffix, []byte("suffix"))
	if suffixRange != FullScanRange {
		t.Error("Suffix should return FullScanRange")
	}

	// 测试 Contains 操作符
	containsRange := h.FromComparison(Contains, []byte("contains"))
	if containsRange != FullScanRange {
		t.Error("Contains should return FullScanRange")
	}

	// 测试静态方法 FromComparison 对新操作符的支持
	staticPrefixRange := FromComparison(Prefix, []byte("prefix"), []byte(""))
	if string(staticPrefixRange.Start) != string(expectedStart) {
		t.Errorf("Static Prefix start range mismatch: expected %q, got %q", expectedStart, staticPrefixRange.Start)
	}

	if string(staticPrefixRange.Limit) != string(expectedLimit) {
		t.Errorf("Static Prefix limit range mismatch: expected %q, got %q", expectedLimit, staticPrefixRange.Limit)
	}
}

func TestBytesPrefix(t *testing.T) {
	// 测试 BytesPrefix 函数
	prefix := []byte("test")
	rangeObj := BytesPrefix(prefix)

	// 预期的范围应该是 ["test", "tesu")
	expectedStart := []byte("test")
	expectedLimit := []byte("tesu")

	if string(rangeObj.Start) != string(expectedStart) {
		t.Errorf("BytesPrefix start mismatch: expected %q, got %q", expectedStart, rangeObj.Start)
	}

	if string(rangeObj.Limit) != string(expectedLimit) {
		t.Errorf("BytesPrefix limit mismatch: expected %q, got %q", expectedLimit, rangeObj.Limit)
	}

	// 测试边界情况：空前缀
	emptyRange := BytesPrefix([]byte(""))
	if len(emptyRange.Start) != 0 {
		t.Errorf("Empty prefix start should be empty slice, got length %d", len(emptyRange.Start))
	}

	if len(emptyRange.Limit) != 0 {
		t.Errorf("Empty prefix limit should be empty slice, got length %d", len(emptyRange.Limit))
	}
}
