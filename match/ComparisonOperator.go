package match

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
	// Prefix 前缀匹配
	Prefix
	// Suffix 后缀匹配
	Suffix
	// Contains 包含匹配
	Contains
)
