package storage

import (
	"testing"
)

// TestRangeSet 测试RangeSet的基本功能
func TestRangeSet(t *testing.T) {
	// 创建RangeSet
	rangeSet := NewRangeSet()

	// 添加各种Range
	rangeSet = rangeSet.AddFullRange()
	rangeSet = rangeSet.AddPrefixRange([]byte("user_"))
	rangeSet = rangeSet.AddClosedRange([]byte("a"), []byte("z"))
	rangeSet = rangeSet.AddOpenRange([]byte("a"), []byte("z"))
	rangeSet = rangeSet.AddLeftClosedRange([]byte("a"), []byte("z"))
	rangeSet = rangeSet.AddRightClosedRange([]byte("a"), []byte("z"))

	// 验证RangeSet的长度
	expectedLength := 6
	if len(rangeSet) != expectedLength {
		t.Errorf("RangeSet长度错误，预期: %d, 实际: %d", expectedLength, len(rangeSet))
	}

	// 验证RangeSet中的Range
	for i, r := range rangeSet {
		if i == 0 {
			// 第一个Range应该是nil（全库扫描）
			if r != nil {
				t.Errorf("RangeSet[0]应该是nil，实际: %+v", r)
			}
		} else {
			// 其他Range不应该是nil
			if r == nil {
				t.Errorf("RangeSet[%d]不应该是nil", i)
			}
		}
	}
}

// TestGenerateJumpRanges 测试生成跳跃区间
func TestGenerateJumpRanges(t *testing.T) {
	// 测试1: 跳过特定前缀
	jumpRange1 := JumpRange{
		SkipStart: []byte("skip_"),
		SkipLimit: []byte("skip_0"),
	}
	rangeSet1 := GenerateJumpRanges(jumpRange1.SkipStart, jumpRange1.SkipLimit)

	// 验证生成的RangeSet长度
	if len(rangeSet1) != 2 {
		t.Errorf("Test1: RangeSet长度错误，预期: 2, 实际: %d", len(rangeSet1))
	}

	// 测试2: 跳过特定数值范围
	jumpRange2 := JumpRange{
		SkipStart: []byte("100"),
		SkipLimit: []byte("200"),
	}
	rangeSet2 := GenerateJumpRanges(jumpRange2.SkipStart, jumpRange2.SkipLimit)

	// 验证生成的RangeSet长度
	if len(rangeSet2) != 2 {
		t.Errorf("Test2: RangeSet长度错误，预期: 2, 实际: %d", len(rangeSet2))
	}

	// 测试3: 跳过空范围
	jumpRange3 := JumpRange{
		SkipStart: []byte("a"),
		SkipLimit: []byte("a"),
	}
	rangeSet3 := GenerateJumpRanges(jumpRange3.SkipStart, jumpRange3.SkipLimit)

	// 验证生成的RangeSet长度
	if len(rangeSet3) != 2 {
		t.Errorf("Test3: RangeSet长度错误，预期: 2, 实际: %d", len(rangeSet3))
	}
}

// TestGenerateMultiJumpRanges 测试生成多个跳跃区间
func TestGenerateMultiJumpRanges(t *testing.T) {
	// 测试1: 跳过多个不重叠区间
	multiJumpRanges1 := []JumpRange{
		{
			SkipStart: []byte("a"),
			SkipLimit: []byte("b"),
		},
		{
			SkipStart: []byte("c"),
			SkipLimit: []byte("d"),
		},
		{
			SkipStart: []byte("e"),
			SkipLimit: []byte("f"),
		},
	}
	rangeSet1 := GenerateMultiJumpRanges(multiJumpRanges1)

	// 验证生成的RangeSet长度
	expectedLength1 := 4 // 三个跳跃区间应该生成4个Range
	if len(rangeSet1) != expectedLength1 {
		t.Errorf("Test1: RangeSet长度错误，预期: %d, 实际: %d", expectedLength1, len(rangeSet1))
	}

	// 测试2: 空跳跃区间列表
	multiJumpRanges2 := []JumpRange{}
	rangeSet2 := GenerateMultiJumpRanges(multiJumpRanges2)

	// 验证生成的RangeSet长度
	expectedLength2 := 1 // 空跳跃区间列表应该生成1个全库扫描Range
	if len(rangeSet2) != expectedLength2 {
		t.Errorf("Test2: RangeSet长度错误，预期: %d, 实际: %d", expectedLength2, len(rangeSet2))
	}

	// 测试3: 单个跳跃区间
	multiJumpRanges3 := []JumpRange{
		{
			SkipStart: []byte("x"),
			SkipLimit: []byte("y"),
		},
	}
	rangeSet3 := GenerateMultiJumpRanges(multiJumpRanges3)

	// 验证生成的RangeSet长度
	expectedLength3 := 2 // 单个跳跃区间应该生成2个Range
	if len(rangeSet3) != expectedLength3 {
		t.Errorf("Test3: RangeSet长度错误，预期: %d, 实际: %d", expectedLength3, len(rangeSet3))
	}
}

// TestCompareKeys 测试CompareKeys函数
func TestCompareKeys(t *testing.T) {
	// 测试用例
	testCases := []struct {
		a        []byte
		b        []byte
		expected int
	}{{
		a:        []byte("a"),
		b:        []byte("b"),
		expected: -1,
	}, {
		a:        []byte("b"),
		b:        []byte("a"),
		expected: 1,
	}, {
		a:        []byte("a"),
		b:        []byte("a"),
		expected: 0,
	}, {
		a:        []byte("ab"),
		b:        []byte("abc"),
		expected: -1,
	}, {
		a:        []byte("abc"),
		b:        []byte("ab"),
		expected: 1,
	}, {
		a:        []byte("100"),
		b:        []byte("200"),
		expected: -1,
	}, {
		a:        []byte("200"),
		b:        []byte("100"),
		expected: 1,
	}, {
		a:        []byte("100"),
		b:        []byte("100"),
		expected: 0,
	}}

	// 执行测试
	for i, tc := range testCases {
		result := CompareKeys(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("TestCompareKeys[%d]: CompareKeys(%q, %q) = %d, expected %d", i, tc.a, tc.b, result, tc.expected)
		}
	}
}

// TestJumpRangeExample 测试跳跃区间示例
func TestJumpRangeExample(t *testing.T) {
	// 运行跳跃区间示例
	JumpRangeExample()
}

// TestAdvancedRangeUsage 测试高级Range使用示例
func TestAdvancedRangeUsage(t *testing.T) {
	// 运行高级Range使用示例
	AdvancedRangeUsage()
}
