package storage

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"
)

// RangeSet 表示多个Range的集合，用于处理复杂的区间需求
type RangeSet []*util.Range

// JumpRange 表示需要跳跃的区间，用于生成跳跃区间的RangeSet
type JumpRange struct {
	// 需要跳跃的起始位置（包含）
	SkipStart []byte
	// 需要跳跃的结束位置（不包含）
	SkipLimit []byte
}

// NewRangeSet 创建一个新的RangeSet
func NewRangeSet() RangeSet {
	return make(RangeSet, 0)
}

// AddRange 向RangeSet中添加一个Range
func (rs RangeSet) AddRange(r *util.Range) RangeSet {
	return append(rs, r)
}

// AddFullRange 添加全库扫描Range
func (rs RangeSet) AddFullRange() RangeSet {
	return append(rs, nil)
}

// AddPrefixRange 添加前缀扫描Range
func (rs RangeSet) AddPrefixRange(prefix []byte) RangeSet {
	return append(rs, util.BytesPrefix(prefix))
}

// KeyRange 创建一个闭区间 [start, end]
// start: 范围开始（包含）
// end: 范围结束（包含）
// 返回一个util.Range，确保包含end
func KeyRange(start, end []byte) *util.Range {
	return &util.Range{
		Start: start,
		Limit: append(end, 0), // end+1，确保包含end
	}
}

// ExclusiveKeyRange 定义一个左开右开的键范围
// start: 范围开始（不包含）
// end: 范围结束（不包含）
func ExclusiveKeyRange(start, end []byte) *util.Range {
	return &util.Range{
		Start: append(start, 0), // start+1，不包含start
		Limit: end,              // 不包含end
	}
}

// ClosedOpenKeyRange 定义一个左闭右开的键范围
// start: 范围开始（包含）
// end: 范围结束（不包含）
func ClosedOpenKeyRange(start, end []byte) *util.Range {
	return &util.Range{
		Start: start,
		Limit: end,
	}
}

// OpenClosedKeyRange 定义一个左开右闭的键范围
// start: 范围开始（不包含）
// end: 范围结束（包含）
func OpenClosedKeyRange(start, end []byte) *util.Range {
	return &util.Range{
		Start: append(start, 0), // start+1，不包含start
		Limit: append(end, 0),   // end+1，包含end
	}
}

// AddClosedRange 添加闭区间 [start, end]
func (rs RangeSet) AddClosedRange(start, end []byte) RangeSet {
	return append(rs, KeyRange(start, end))
}

// AddOpenRange 添加开区间 (start, end)
func (rs RangeSet) AddOpenRange(start, end []byte) RangeSet {
	return append(rs, ExclusiveKeyRange(start, end))
}

// AddLeftClosedRange 添加左闭右开区间 [start, end)
func (rs RangeSet) AddLeftClosedRange(start, end []byte) RangeSet {
	return append(rs, ClosedOpenKeyRange(start, end))
}

// AddRightClosedRange 添加左开右闭区间 (start, end]
func (rs RangeSet) AddRightClosedRange(start, end []byte) RangeSet {
	return append(rs, OpenClosedKeyRange(start, end))
}

// GenerateJumpRanges 生成跳跃区间的RangeSet
// 跳跃区间表示：遍历除了[skipStart, skipLimit)之外的所有数据
func GenerateJumpRanges(skipStart, skipLimit []byte) RangeSet {
	rs := NewRangeSet()

	// 第一部分：(-∞, skipStart)
	beforeSkip := &util.Range{
		Limit: skipStart,
	}
	rs = rs.AddRange(beforeSkip)

	// 第二部分：[skipLimit, +∞)
	afterSkip := &util.Range{
		Start: skipLimit,
	}
	rs = rs.AddRange(afterSkip)

	return rs
}

// GenerateMultiJumpRanges 生成多个跳跃区间的RangeSet
// 跳跃多个不重叠的区间
func GenerateMultiJumpRanges(jumpRanges []JumpRange) RangeSet {
	// 简单实现：假设jumpRanges是有序且不重叠的
	// 实际应用中可能需要更复杂的区间合并逻辑
	rs := NewRangeSet()

	// 第一部分：(-∞, firstSkipStart)
	if len(jumpRanges) > 0 {
		firstJump := jumpRanges[0]
		beforeFirst := &util.Range{
			Limit: firstJump.SkipStart,
		}
		rs = rs.AddRange(beforeFirst)

		// 中间部分：[prevSkipLimit, nextSkipStart)
		for i := 1; i < len(jumpRanges); i++ {
			prevJump := jumpRanges[i-1]
			currentJump := jumpRanges[i]
			middleRange := &util.Range{
				Start: prevJump.SkipLimit,
				Limit: currentJump.SkipStart,
			}
			rs = rs.AddRange(middleRange)
		}

		// 最后部分：[lastSkipLimit, +∞)
		lastJump := jumpRanges[len(jumpRanges)-1]
		afterLast := &util.Range{
			Start: lastJump.SkipLimit,
		}
		rs = rs.AddRange(afterLast)
	} else {
		// 没有跳跃区间，添加全库扫描
		rs = rs.AddFullRange()
	}

	return rs
}

// IteratorWithJumpRange 使用跳跃区间遍历数据库
// 示例函数，展示如何使用RangeSet进行跳跃区间遍历
func IteratorWithJumpRange(store Store, jumpRange JumpRange) error {
	// 生成跳跃区间的RangeSet
	rangeSet := GenerateJumpRanges(jumpRange.SkipStart, jumpRange.SkipLimit)

	fmt.Printf("\n使用跳跃区间遍历，跳过: [%s, %s)\n", jumpRange.SkipStart, jumpRange.SkipLimit)

	// 遍历每个Range
	for i, r := range rangeSet {
		fmt.Printf("\n处理Range %d: %+v\n", i, r)

		// 使用当前Range创建迭代器
		var iter Iterator
		if r != nil {
			if r.Start != nil && r.Limit != nil {
				// 有Start和Limit，使用范围扫描
				iter = store.Iterator(r)
			} else if r.Start != nil {
				// 只有Start，从Start开始到结尾
				iter = store.Iterator(r)
				// 使用Seek定位到Start位置
				iter.Seek(r.Start)
			} else if r.Limit != nil {
				// 只有Limit，从开头到Limit
				iter = store.Iterator(r)
				// 从第一个元素开始
				iter.First()
			} else {
				// 没有Start和Limit，全库扫描
				iter = store.Iterator(r)
				iter.First()
			}
		} else {
			// r为nil，全库扫描
			iter = store.Iterator(r)
			iter.First()
		}

		// 遍历迭代器
		for iter.Valid() {
			key := iter.Key()
			value := iter.Value()

			// 检查是否超出当前Range的Limit
			if r != nil && r.Limit != nil {
				// 比较key是否小于Limit
				if CompareKeys(key, r.Limit) >= 0 {
					// 超出当前Range，结束遍历
					break
				}
			}

			// 处理当前key-value
			fmt.Printf("  Key: %s, Value: %s\n", key, value)

			// 移动到下一个元素
			iter.Next()
		}

		// 释放迭代器
		iter.Release()
	}

	return nil
}

// JumpRangeExample 展示跳跃区间的使用示例
func JumpRangeExample() {
	fmt.Println("=== 跳跃区间使用示例 ===")

	// 示例1：跳过特定前缀
	fmt.Println("\n1. 跳过特定前缀 'skip_':")
	jumpRange1 := JumpRange{
		SkipStart: []byte("skip_"),
		SkipLimit: append([]byte("skip_"), 0), // "skip_"+1
	}
	rangeSet1 := GenerateJumpRanges(jumpRange1.SkipStart, jumpRange1.SkipLimit)
	fmt.Printf("   生成的RangeSet: %+v\n", rangeSet1)
	fmt.Println("   描述: 遍历除了所有以'skip_'为前缀的key之外的所有数据")

	// 示例2：跳过特定数值范围
	fmt.Println("\n2. 跳过数值范围 [100, 200):")
	jumpRange2 := JumpRange{
		SkipStart: []byte("100"),
		SkipLimit: []byte("200"),
	}
	rangeSet2 := GenerateJumpRanges(jumpRange2.SkipStart, jumpRange2.SkipLimit)
	fmt.Printf("   生成的RangeSet: %+v\n", rangeSet2)
	fmt.Println("   描述: 遍历除了数值范围100-199之外的所有数据")

	// 示例3：跳过特定时间段
	fmt.Println("\n3. 跳过特定时间段:")
	jumpRange3 := JumpRange{
		SkipStart: []byte("log_20231201"),
		SkipLimit: []byte("log_20240101"),
	}
	rangeSet3 := GenerateJumpRanges(jumpRange3.SkipStart, jumpRange3.SkipLimit)
	fmt.Printf("   生成的RangeSet: %+v\n", rangeSet3)
	fmt.Println("   描述: 遍历除了2023年12月份之外的所有日志")

	// 示例4：跳过多个不重叠区间
	fmt.Println("\n4. 跳过多个不重叠区间:")
	multiJumpRanges := []JumpRange{
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
	rangeSet4 := GenerateMultiJumpRanges(multiJumpRanges)
	fmt.Printf("   生成的RangeSet: %+v\n", rangeSet4)
	fmt.Println("   描述: 遍历除了a-b, c-d, e-f之外的所有数据")

	// 示例5：使用RangeSet组合多个Range
	fmt.Println("\n5. 组合多个Range:")
	rangeSet5 := NewRangeSet()
	// 添加前缀扫描
	rangeSet5 = rangeSet5.AddPrefixRange([]byte("user_"))
	// 添加特定范围
	rangeSet5 = rangeSet5.AddLeftClosedRange([]byte("order_20231201"), []byte("order_20231231"))
	// 添加全库扫描的一部分
	rangeSet5 = rangeSet5.AddRightClosedRange([]byte("log_20231201"), []byte("log_20231202"))
	fmt.Printf("   生成的RangeSet: %+v\n", rangeSet5)
	fmt.Println("   描述: 组合了前缀扫描、特定范围和右闭区间")

	fmt.Println("\n=== 跳跃区间使用示例结束 ===")
}

// AdvancedRangeUsage 高级Range使用示例
func AdvancedRangeUsage() {
	fmt.Println("=== 高级Range使用示例 ===")

	// 场景1: 电商系统中，查询除了特定类别的商品
	fmt.Println("\n场景1: 电商系统")
	// 商品格式：product_类别_12345
	jumpRange := JumpRange{
		SkipStart: []byte("product_electronics_"),
		SkipLimit: append([]byte("product_electronics_"), 0),
	}
	rangeSet := GenerateJumpRanges(jumpRange.SkipStart, jumpRange.SkipLimit)
	fmt.Printf("   查询除电子产品外的所有商品: %+v\n", rangeSet)

	// 场景2: 金融系统中，查询除了特定时间段的交易
	fmt.Println("\n场景2: 金融系统")
	// 交易格式：transaction_20231201_12345
	jumpRange = JumpRange{
		SkipStart: []byte("transaction_20231201"),
		SkipLimit: []byte("transaction_20231202"),
	}
	rangeSet = GenerateJumpRanges(jumpRange.SkipStart, jumpRange.SkipLimit)
	fmt.Printf("   查询除2023年12月1日外的所有交易: %+v\n", rangeSet)

	// 场景3: 社交系统中，查询除了特定用户组的用户
	fmt.Println("\n场景3: 社交系统")
	// 用户格式：user_admin_12345, user_normal_12345
	jumpRange = JumpRange{
		SkipStart: []byte("user_admin_"),
		SkipLimit: append([]byte("user_admin_"), 0),
	}
	rangeSet = GenerateJumpRanges(jumpRange.SkipStart, jumpRange.SkipLimit)
	fmt.Printf("   查询除管理员外的所有用户: %+v\n", rangeSet)

	// 场景4: 游戏系统中，查询除了特定等级范围的玩家
	fmt.Println("\n场景4: 游戏系统")
	// 玩家格式：player_level_10_12345
	jumpRange = JumpRange{
		SkipStart: []byte("player_level_10"),
		SkipLimit: []byte("player_level_20"),
	}
	rangeSet = GenerateJumpRanges(jumpRange.SkipStart, jumpRange.SkipLimit)
	fmt.Printf("   查询除10-19级外的所有玩家: %+v\n", rangeSet)

	fmt.Println("\n=== 高级Range使用示例结束 ===")
}

// CompareKeys 比较两个key的大小
// 返回值：
//
//	-1: a < b
//	 0: a == b
//	 1: a > b
func CompareKeys(a, b []byte) int {
	// 比较两个byte slice的大小
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}

	// 比较长度
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}

	return 0
}
