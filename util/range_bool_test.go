package util

import (
	"bytes"
	"testing"
)

// TestRangeHelper_Bool 测试RangeHelper对bool类型值的范围处理
func TestRangeHelper_Bool(t *testing.T) {
	// bool类型的字节表示
	trueBytes := []byte{1}  // true
	falseBytes := []byte{0} // false

	// 创建RangeHelper，使用空前缀
	rh := NewRangeHelper([]byte{})

	// 测试1: Equal操作符
	t.Run("Equal", func(t *testing.T) {
		// 测试true
		rangeTrue := rh.FromComparison(Equal, trueBytes)
		t.Logf("Equal(true) range: Start=%v, Limit=%v", rangeTrue.Start, rangeTrue.Limit)

		// 测试false
		rangeFalse := rh.FromComparison(Equal, falseBytes)
		t.Logf("Equal(false) range: Start=%v, Limit=%v", rangeFalse.Start, rangeFalse.Limit)
	})

	// 测试2: GreaterThan操作符
	t.Run("GreaterThan", func(t *testing.T) {
		// 测试true
		rangeTrue := rh.FromComparison(GreaterThan, trueBytes)
		t.Logf("GreaterThan(true) range: Start=%v, Limit=%v", rangeTrue.Start, rangeTrue.Limit)

		// 测试false
		rangeFalse := rh.FromComparison(GreaterThan, falseBytes)
		t.Logf("GreaterThan(false) range: Start=%v, Limit=%v", rangeFalse.Start, rangeFalse.Limit)
	})

	// 测试3: GreaterThanOrEqual操作符
	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		// 测试true
		rangeTrue := rh.FromComparison(GreaterThanOrEqual, trueBytes)
		t.Logf("GreaterThanOrEqual(true) range: Start=%v, Limit=%v", rangeTrue.Start, rangeTrue.Limit)

		// 测试false
		rangeFalse := rh.FromComparison(GreaterThanOrEqual, falseBytes)
		t.Logf("GreaterThanOrEqual(false) range: Start=%v, Limit=%v", rangeFalse.Start, rangeFalse.Limit)
	})

	// 测试4: LessThan操作符
	t.Run("LessThan", func(t *testing.T) {
		// 测试true
		rangeTrue := rh.FromComparison(LessThan, trueBytes)
		t.Logf("LessThan(true) range: Start=%v, Limit=%v", rangeTrue.Start, rangeTrue.Limit)

		// 测试false
		rangeFalse := rh.FromComparison(LessThan, falseBytes)
		t.Logf("LessThan(false) range: Start=%v, Limit=%v", rangeFalse.Start, rangeFalse.Limit)
	})

	// 测试5: LessThanOrEqual操作符
	t.Run("LessThanOrEqual", func(t *testing.T) {
		// 测试true
		rangeTrue := rh.FromComparison(LessThanOrEqual, trueBytes)
		t.Logf("LessThanOrEqual(true) range: Start=%v, Limit=%v", rangeTrue.Start, rangeTrue.Limit)

		// 测试false
		rangeFalse := rh.FromComparison(LessThanOrEqual, falseBytes)
		t.Logf("LessThanOrEqual(false) range: Start=%v, Limit=%v", rangeFalse.Start, rangeFalse.Limit)
	})

	// 测试6: 分析bool类型的逻辑关系
	t.Run("BoolLogicalAnalysis", func(t *testing.T) {
		// 在布尔逻辑中，true > false
		// 所以：
		// - GreaterThan(false) 应该包含 true
		// - LessThan(true) 应该包含 false
		// - GreaterThan(true) 应该不包含任何值
		// - LessThan(false) 应该不包含任何值

		t.Logf("Bool logical analysis:")
		t.Logf("trueBytes: %v", trueBytes)
		t.Logf("falseBytes: %v", falseBytes)
		// 使用bytes.Compare比较字节切片
		comparisonResult := bytes.Compare(trueBytes, falseBytes)
		t.Logf("bytes.Compare(trueBytes, falseBytes): %v", comparisonResult)
		t.Logf("trueBytes > falseBytes: %v", comparisonResult > 0)
		t.Logf("trueBytes == falseBytes: %v", comparisonResult == 0)
		t.Logf("trueBytes < falseBytes: %v", comparisonResult < 0)
	})
}
