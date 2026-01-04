package storage

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"
)

// RangeExample 展示LevelDB中util.Range的各种区间取值示例
// 注意：在LevelDB中，util.Range的Start是包含的，Limit是不包含的
// 即区间为 [Start, Limit)

// ExampleLevelDBRange 展示各种区间取值示例
func ExampleLevelDBRange() {
	fmt.Println("=== LevelDB Range取值示例 ===")

	// 1. 全库扫描
	fmt.Println("\n1. 全库扫描:")
	// 当slice为nil时，迭代器会遍历整个数据库
	var fullScan *util.Range
	fmt.Printf("   Range: %+v\n", fullScan)
	fmt.Println("   描述: 遍历整个数据库，从第一个key到最后一个key")

	// 2. 前缀扫描
	fmt.Println("\n2. 前缀扫描:")
	// 使用BytesPrefix函数创建前缀扫描范围
	prefixScan := util.BytesPrefix([]byte("user_"))
	fmt.Printf("   Range: %+v\n", prefixScan)
	fmt.Println("   描述: 遍历所有以'user_'为前缀的key")
	fmt.Println("   等价于: ['user_', 'user_'+1)")

	// 3. 闭区间扫描 ["a", "z"]
	fmt.Println("\n3. 闭区间扫描 [\"a\", \"z\"]:")
	// 在LevelDB中，Limit是不包含的，所以要包含"z"，需要设置为"z"+1
	closedRange := &util.Range{
		Start: []byte("a"),
		Limit: append([]byte("z"), 0), // "z"+1，确保包含"z"
	}
	fmt.Printf("   Range: %+v\n", closedRange)
	fmt.Println("   描述: 遍历所有大于等于'a'且小于'z'+1的key，即包含所有以a-z开头的key")

	// 4. 开区间扫描 ("a", "z")
	fmt.Println("\n4. 开区间扫描 ('a', 'z'):")
	// 不包含"a"，也不包含"z"
	openRange := &util.Range{
		Start: append([]byte("a"), 0), // "a"+1，不包含"a"
		Limit: []byte("z"),            // 不包含"z"
	}
	fmt.Printf("   Range: %+v\n", openRange)
	fmt.Println("   描述: 遍历所有大于'a'且小于'z'的key")

	// 5. 左闭右开区间扫描 ["a", "z")
	fmt.Println("\n5. 左闭右开区间扫描 [\"a\", \"z\"):")
	// 包含"a"，不包含"z"
	leftClosedRange := &util.Range{
		Start: []byte("a"),
		Limit: []byte("z"),
	}
	fmt.Printf("   Range: %+v\n", leftClosedRange)
	fmt.Println("   描述: 遍历所有大于等于'a'且小于'z'的key")

	// 6. 左开右闭区间扫描 ("a", "z"]
	fmt.Println("\n6. 左开右闭区间扫描 ('a', \"z\"]:")
	// 不包含"a"，包含"z"
	rightClosedRange := &util.Range{
		Start: append([]byte("a"), 0), // "a"+1，不包含"a"
		Limit: append([]byte("z"), 0), // "z"+1，包含"z"
	}
	fmt.Printf("   Range: %+v\n", rightClosedRange)
	fmt.Println("   描述: 遍历所有大于'a'且小于'z'+1的key，即包含'a'之后到'z'的所有key")

	// 7. 单值扫描 ["key1", "key1"+1)
	fmt.Println("\n7. 单值扫描:")
	// 只扫描单个key "key1"
	singleValueRange := &util.Range{
		Start: []byte("key1"),
		Limit: append([]byte("key1"), 0), // "key1"+1，只包含"key1"
	}
	fmt.Printf("   Range: %+v\n", singleValueRange)
	fmt.Println("   描述: 只扫描key为'key1'的记录")

	// 8. 数值范围扫描 [100, 200)
	fmt.Println("\n8. 数值范围扫描 [100, 200):")
	// 假设key是数值的字符串形式，如"100", "101", ..., "199"
	numericRange := &util.Range{
		Start: []byte("100"),
		Limit: []byte("200"), // 不包含"200"
	}
	fmt.Printf("   Range: %+v\n", numericRange)
	fmt.Println("   描述: 遍历所有大于等于'100'且小于'200'的key，即100-199的数值范围")

	// 9. 从某个key开始的所有记录 ["start", +∞)
	fmt.Println("\n9. 从某个key开始的所有记录:")
	// Limit为nil表示没有上限
	fromStartRange := &util.Range{
		Start: []byte("start"),
		Limit: nil, // 没有上限
	}
	fmt.Printf("   Range: %+v\n", fromStartRange)
	fmt.Println("   描述: 遍历所有大于等于'start'的key")

	// 10. 到某个key结束的所有记录 (-∞, "end")
	fmt.Println("\n10. 到某个key结束的所有记录:")
	// Start为nil表示没有下限
	toEndRange := &util.Range{
		Start: nil, // 没有下限
		Limit: []byte("end"),
	}
	fmt.Printf("   Range: %+v\n", toEndRange)
	fmt.Println("   描述: 遍历所有小于'end'的key")

	// 11. 特定范围扫描 ["key10", "key20")
	fmt.Println("\n11. 特定范围扫描 [\"key10\", \"key20\"):")
	specificRange := &util.Range{
		Start: []byte("key10"),
		Limit: []byte("key20"),
	}
	fmt.Printf("   Range: %+v\n", specificRange)
	fmt.Println("   描述: 遍历所有大于等于'key10'且小于'key20'的key")

	// 12. 空范围扫描
	fmt.Println("\n12. 空范围扫描:")
	// Start等于Limit时，范围为空
	emptyRange := &util.Range{
		Start: []byte("same"),
		Limit: []byte("same"),
	}
	fmt.Printf("   Range: %+v\n", emptyRange)
	fmt.Println("   描述: 空范围，不会返回任何记录")

	fmt.Println("\n=== 区间取值示例结束 ===")
}

// LevelDBRangeUsage 在实际场景中的使用示例
func LevelDBRangeUsage() {
	fmt.Println("=== LevelDB Range实际使用示例 ===")

	// 场景1: 用户管理系统中，遍历所有用户
	fmt.Println("\n场景1: 用户管理系统")
	userPrefix := util.BytesPrefix([]byte("user_"))
	fmt.Printf("   遍历所有用户: %+v\n", userPrefix)

	// 场景2: 订单系统中，查询特定日期范围内的订单
	fmt.Println("\n场景2: 订单系统")
	// 订单格式：order_20231201_12345
	orderRange := &util.Range{
		Start: []byte("order_20231201"),
		Limit: append([]byte("order_20231231"), 0), // 包含12月31日的订单
	}
	fmt.Printf("   查询12月份订单: %+v\n", orderRange)

	// 场景3: 缓存系统中，清理特定前缀的缓存
	fmt.Println("\n场景3: 缓存系统")
	cachePrefix := util.BytesPrefix([]byte("cache_product_"))
	fmt.Printf("   清理产品缓存: %+v\n", cachePrefix)

	// 场景4: 日志系统中，查询特定级别和时间段的日志
	fmt.Println("\n场景4: 日志系统")
	// 日志格式：log_20231201_debug_123
	logRange := &util.Range{
		Start: []byte("log_20231201_debug"),
		Limit: []byte("log_20231202_info"),
	}
	fmt.Printf("   查询12月1日debug日志: %+v\n", logRange)

	fmt.Println("\n=== 实际使用示例结束 ===")
}
