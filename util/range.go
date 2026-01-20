package util

// ComparisonOperator 定义比较操作符枚举
type ComparisonOperator int

// 比较操作符枚举值
const (
	// Equal 等于 (=)
	Equal ComparisonOperator = iota
	// NotEqual 不等于 (!=)
	NotEqual
	// GreaterThan 大于 (>)
	GreaterThan
	// GreaterThanOrEqual 大于等于 (>=)
	GreaterThanOrEqual
	// LessThan 小于 (<)
	LessThan
	// LessThanOrEqual 小于等于 (<=)
	LessThanOrEqual
	// Like 类似于SQL LIKE操作
	Like
)

// String 返回比较操作符的字符串表示
func (op ComparisonOperator) String() string {
	switch op {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case Like:
		return "LIKE"

	default:
		return "unknown"
	}
}

// RangeHelper 提供LevelDB Range的辅助函数和常量
// 注意：在LevelDB中，Range的Start是包含的，Limit是不包含的
// 即区间为 [Start, Limit)
type RangeHelper struct {
	pfx []byte
}

// NewRangeHelper 创建一个新的RangeHelper实例
func NewRangeHelper(pfx []byte) *RangeHelper {
	return &RangeHelper{
		pfx: pfx,
	}
}

// 常用的Range变量
var (
	// FullScanRange 表示全库扫描的Range
	FullScanRange *Range
)

// FullScan 返回全库扫描的Range
// 当slice为nil时，迭代器会遍历整个数据库
func (h *RangeHelper) FullScan() *Range {
	return FullScanRange
}

// Prefix 返回前缀扫描的Range
// 例如：h.Prefix([]byte("user_")) 将匹配所有以"user_"为前缀的key
func (h *RangeHelper) Prefix(prefix []byte) *Range {
	return BytesPrefix(prefix)
}

/*
// LikeRange 返回类似SQL LIKE操作的Range
// 支持的通配符：
//
//	% - 匹配任意长度的字符串（包括空字符串）
//	_ - 匹配单个字符
//
// 示例：
//
//	h.LikeRange([]byte("prefix%")) → ["prefix", "prefix\xff") - 前缀匹配
//	h.LikeRange([]byte("%suffix")) → 不支持，返回nil
//	h.LikeRange([]byte("%middle%")) → 不支持，返回nil
//	h.LikeRange([]byte("prefix_%_suffix")) → 不支持，返回nil
//
// 注意：目前只支持前缀匹配，即通配符%只出现在末尾的情况
func (h *RangeHelper) LikeRange(pattern []byte) *Range {
	patternStr := string(pattern)

	// 检查是否为前缀匹配模式：%只出现在末尾
	if lastPercent := strings.LastIndex(patternStr, "%"); lastPercent != -1 {
		if lastPercent == len(patternStr)-1 {
			// 前缀匹配：prefix%
			prefix := pattern[:lastPercent]
			return h.Prefix(prefix)
		}
	}

	// 检查是否为精确匹配（没有通配符）
	if !strings.Contains(patternStr, "%") && !strings.Contains(patternStr, "_") {
		// 精确匹配，使用Equal
		return h.FromComparison(Equal, pattern)
	}

	// 其他模式暂不支持
	return nil
}

// LikeRange 返回类似SQL LIKE操作的Range（静态方法）
func LikeRange(pattern []byte) *Range {
	h := NewRangeHelper()
	return h.LikeRange(pattern)
}
*/
// KeyRange 创建一个闭区间 [start, end]
// start: 范围开始（包含）
// end: 范围结束（包含）
// 返回一个Range，确保包含end
func (h *RangeHelper) KeyRange(start, end []byte) *Range {
	return &Range{
		Start: start,
		Limit: append(end, 0), // end+1，确保包含end
	}
}

// util.BytesPrefix的函数
// 功能：返回前缀的下一个字节
// 例如：BytesPrefix([]byte("user_")) 将返回 []byte("user_")
func bytesPrefix(prefix []byte) []byte {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return limit
}

// FromComparison 创建基于比较操作符的Range
// op: 比较操作符
// value: 比较值
// 返回对应的Range
// 示例：
//
//	h.FromComparison(Equal, []byte("key1")) → ["key1", "key1"+1)
//	h.FromComparison(GreaterThan, []byte("key1")) → ["key1"+1, +∞)
//	h.FromComparison(GreaterThanOrEqual, []byte("key1")) → ["key1", +∞)
//	h.FromComparison(LessThan, []byte("key1")) → (-∞, "key1")
//	h.FromComparison(LessThanOrEqual, []byte("key1")) → (-∞, "key1"+1)
//	h.FromComparison(Like, []byte("prefix%")) → ["prefix", "prefix\xff")
//	h.FromComparison(Prefix, []byte("prefix")) → ["prefix", "prefix\xff")
func (h *RangeHelper) FromComparison(op ComparisonOperator, value []byte) *Range {
	switch op {
	case Equal:
		// [value, value+1)
		return &Range{
			Start: value,
			Limit: append(value, 0),
		}
	case GreaterThan:
		// (value, +∞) → [value+1, +∞)
		return &Range{
			Start: append(value, 0),
			Limit: bytesPrefix(h.pfx),
		}
	case GreaterThanOrEqual:
		// [value, +∞)
		return &Range{
			Start: value,
			Limit: bytesPrefix(h.pfx),
		}
	case LessThan:
		// (-∞, value)
		return &Range{
			Start: h.pfx,
			Limit: value,
		}
	case LessThanOrEqual:
		// (-∞, value] → (-∞, value+1)
		return &Range{
			Start: h.pfx,
			Limit: append(value, 0),
		}
	case Like:
		// 使用Prefix处理LIKE操作，仅支持前缀匹配
		return h.Prefix(value)

	case NotEqual:
		// 不等于操作使用全库扫描，具体过滤在应用层处理
		// 或者调用FromComparisonNotEqual获取两个Range
		return FullScanRange
	default:
		// 默认返回全库扫描
		return FullScanRange
	}
}

// FromComparisonNotEqual 处理不等于操作，返回两个Range
// 对于不等于value的情况，返回两个Range：
// 1. (-∞, value) - 小于value的范围
// 2. (value, +∞) - 大于value的范围
func (h *RangeHelper) FromComparisonNotEqual(value []byte) [2]*Range {
	return [2]*Range{
		// (-∞, value) - 小于value的范围
		{
			Start: h.pfx,
			Limit: value,
		},
		// (value, +∞) - 大于value的范围
		{
			Start: append(value, 0),
			Limit: bytesPrefix(h.pfx),
		},
	}
}

// FromComparisonNotEqual 处理不等于操作，返回两个Range（静态方法）
func FromComparisonNotEqual(value []byte, pfx []byte) [2]*Range {
	h := NewRangeHelper(pfx)
	return h.FromComparisonNotEqual(value)
}

// FromComparison 创建基于比较操作符的Range（静态方法）
// op: 比较操作符
// value: 比较值
// pfx: 前缀值，用于替代无穷范围
// 返回对应的Range
func FromComparison(op ComparisonOperator, value []byte, pfx []byte) *Range {
	h := NewRangeHelper(pfx)
	return h.FromComparison(op, value)
}

// ExclusiveRange 创建一个左开右开的键范围 (start, end)
// start: 范围开始（不包含）
// end: 范围结束（不包含）
func (h *RangeHelper) ExclusiveRange(start, end []byte) *Range {
	return &Range{
		Start: append(start, 0), // start+1，不包含start
		Limit: end,              // 不包含end
	}
}

// OpenClosedRange 创建一个左开右闭的键范围 (start, end]
// start: 范围开始（不包含）
// end: 范围结束（包含）
func (h *RangeHelper) OpenClosedRange(start, end []byte) *Range {
	return &Range{
		Start: append(start, 0), // start+1，不包含start
		Limit: append(end, 0),   // end+1，包含end
	}
}

// ClosedOpenRange 创建一个左闭右开的键范围 [start, end)
// start: 范围开始（包含）
// end: 范围结束（不包含）
func (h *RangeHelper) ClosedOpenRange(start, end []byte) *Range {
	return &Range{
		Start: start,
		Limit: end,
	}
}

// SingleValue 创建一个单值扫描的Range [value, value+1)
// 只扫描key等于value的记录
func (h *RangeHelper) SingleValue(value []byte) *Range {
	return &Range{
		Start: value,
		Limit: append(value, 0), // value+1，只包含value
	}
}

// FromStart 创建一个从某个key开始的所有记录的Range [start, +∞)
// start: 范围开始（包含）
// 返回的Range将匹配所有大于等于start的key
func (h *RangeHelper) FromStart(start []byte) *Range {
	return &Range{
		Start: start,
		Limit: bytesPrefix(h.pfx), // 没有上限，使用前缀替代
	}
}

// ToEnd 创建一个到某个key结束的所有记录的Range (-∞, end)
// end: 范围结束（不包含）
// 返回的Range将匹配所有小于end的key
func (h *RangeHelper) ToEnd(end []byte) *Range {
	return &Range{
		Start: h.pfx, // 没有下限，使用前缀替代
		Limit: end,
	}
}

// NumericRange 创建一个数值范围的Range [start, end)
// 假设key是数值的字符串形式，如"100", "101", ..., "199"
// start: 数值范围开始（包含）
// end: 数值范围结束（不包含）
func (h *RangeHelper) NumericRange(start, end []byte) *Range {
	return &Range{
		Start: start,
		Limit: end,
	}
}

// EmptyRange 创建一个空范围的Range
// 当Start等于Limit时，范围为空，不会返回任何记录
func (h *RangeHelper) EmptyRange(value []byte) *Range {
	return &Range{
		Start: value,
		Limit: value,
	}
}

// Range is a key range.
type Range struct {
	// Start of the key range, include in the range.
	Start []byte

	// Limit of the key range, not include in the range.
	Limit []byte
}

// BytesPrefix returns key range that satisfy the given prefix.
// This only applicable for the standard 'bytes comparer'.
func BytesPrefix(prefix []byte) *Range {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return &Range{prefix, limit}
}
