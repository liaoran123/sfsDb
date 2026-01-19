package engine

import (
	"github.com/liaoran123/sfsDb/util"
)

// UtilMatch 实现了 Match 接口，使用 util.Match 进行比较操作
type UtilMatch struct {
	fieldName string                  // 要比较的字段名
	op        util.ComparisonOperator // 比较操作符
	value     any                     // 比较值
	utilMatch util.Match              // 内部使用的 util.Match 实例
}

// NewUtilMatch 创建一个新的 UtilMatch 实例
func NewUtilMatch(fieldName string, op util.ComparisonOperator, value any) *UtilMatch {
	return &UtilMatch{
		fieldName: fieldName,
		op:        op,
		value:     value,
		utilMatch: util.NewMatch(value),
	}
}

// Match 实现了 Match 接口，使用 util.Match 进行比较操作
func (um *UtilMatch) Match(fields *map[string]any) bool {
	if fields == nil {
		return false
	}

	// 获取要比较的字段值
	fieldValue, exists := (*fields)[um.fieldName]
	if !exists {
		return false
	}

	// 创建一个新的 util.Match 实例，使用字段值作为比较的基础
	fieldMatch := util.NewMatch(fieldValue)
	// 使用正确的比较顺序：fieldValue 与 um.value 进行比较
	return fieldMatch.Compare(um.op, um.value)
}

// FieldMatch 是一个更简单的匹配器，用于直接比较字段值
// 它实现了 Match 接口，内部使用 util.Match 进行比较
type FieldMatch struct {
	fieldName string         // 要比较的字段名
	matchFunc func(any) bool // 匹配函数
}

// NewFieldMatch 创建一个新的 FieldMatch 实例
func NewFieldMatch(fieldName string, matchFunc func(any) bool) *FieldMatch {
	return &FieldMatch{
		fieldName: fieldName,
		matchFunc: matchFunc,
	}
}

// Match 实现了 Match 接口，使用自定义的匹配函数
func (fm *FieldMatch) Match(fields *map[string]any) bool {
	if fields == nil {
		return false
	}

	// 获取要比较的字段值
	fieldValue, exists := (*fields)[fm.fieldName]
	if !exists {
		return false
	}

	// 使用自定义的匹配函数进行比较
	return fm.matchFunc(fieldValue)
}

// NewEqualMatch 创建一个相等匹配器
func NewEqualMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.Equal, value)
}

// NewNotEqualMatch 创建一个不相等匹配器
func NewNotEqualMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.NotEqual, value)
}

// NewGreaterThanMatch 创建一个大于匹配器
func NewGreaterThanMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.GreaterThan, value)
}

// NewGreaterThanOrEqualMatch 创建一个大于等于匹配器
func NewGreaterThanOrEqualMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.GreaterThanOrEqual, value)
}

// NewLessThanMatch 创建一个小于匹配器
func NewLessThanMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.LessThan, value)
}

// NewLessThanOrEqualMatch 创建一个小于等于匹配器
func NewLessThanOrEqualMatch(fieldName string, value any) *UtilMatch {
	return NewUtilMatch(fieldName, util.LessThanOrEqual, value)
}

// NewLikeMatch 创建一个 LIKE 匹配器
func NewLikeMatch(fieldName string, pattern string) *UtilMatch {
	return NewUtilMatch(fieldName, util.Like, pattern)
}

// NewPrefixMatch 创建一个前缀匹配器
func NewPrefixMatch(fieldName string, prefix string) *UtilMatch {
	return NewUtilMatch(fieldName, util.Prefix, prefix)
}

// NewSuffixMatch 创建一个后缀匹配器
func NewSuffixMatch(fieldName string, suffix string) *UtilMatch {
	return NewUtilMatch(fieldName, util.Suffix, suffix)
}

// NewContainsMatch 创建一个包含匹配器
func NewContainsMatch(fieldName string, substr string) *UtilMatch {
	return NewUtilMatch(fieldName, util.Contains, substr)
}
