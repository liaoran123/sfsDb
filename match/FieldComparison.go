package match

// 使用 FieldComparison.Match 进行比较操作
// 主要是用于主键迭代器对于无索引字段的匹配。
type FieldComparison struct {
	fieldName   string             // 要比较的字段名
	op          ComparisonOperator // 比较操作符
	value       any                // 比较值
	innerValues ValueComparer      // 内部使用的 ValueComparer 实例
}

// NewFieldComparison 创建一个新的 FieldComparison 实例
func NewFieldComparison(fieldName string, op ComparisonOperator, value any) *FieldComparison {
	return &FieldComparison{
		fieldName:   fieldName,
		op:          op,
		value:       value,
		innerValues: NewValueComparer(value),
	}
}

// Match 实现了 Match 接口，使用 FieldComparison.Match 进行比较操作
func (um *FieldComparison) Match(fields *map[string]any) bool {
	if fields == nil {
		return false
	}

	// 获取要比较的字段值
	fieldValue, exists := (*fields)[um.fieldName]
	if !exists {
		return false
	}

	// 创建一个新的 ValueComparer 实例，使用字段值作为比较的基础
	fieldMatch := NewValueComparer(fieldValue)
	// 使用正确的比较顺序：fieldValue 与 um.value 进行比较
	return fieldMatch.Compare(um.op, um.value)
}

// FieldMatch 是一个更简单的匹配器，用于直接比较字段值
// 它实现了 Match 接口，内部使用 ValueComparer 进行比较
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
func NewEqualMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, Equal, value)
}

// NewNotEqualMatch 创建一个不相等匹配器
func NewNotEqualMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, NotEqual, value)
}

// NewGreaterThanMatch 创建一个大于匹配器
func NewGreaterThanMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, GreaterThan, value)
}

// NewGreaterThanOrEqualMatch 创建一个大于等于匹配器
func NewGreaterThanOrEqualMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, GreaterThanOrEqual, value)
}

// NewLessThanMatch 创建一个小于匹配器
func NewLessThanMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, LessThan, value)
}

// NewLessThanOrEqualMatch 创建一个小于等于匹配器
func NewLessThanOrEqualMatch(fieldName string, value any) *FieldComparison {
	return NewFieldComparison(fieldName, LessThanOrEqual, value)
}

// NewLikeMatch 创建一个 LIKE 匹配器
func NewLikeMatch(fieldName string, pattern string) *FieldComparison {
	return NewFieldComparison(fieldName, Like, pattern)
}

// NewPrefixMatch 创建一个前缀匹配器
func NewPrefixMatch(fieldName string, prefix string) *FieldComparison {
	return NewFieldComparison(fieldName, Prefix, prefix)
}

// NewSuffixMatch 创建一个后缀匹配器
func NewSuffixMatch(fieldName string, suffix string) *FieldComparison {
	return NewFieldComparison(fieldName, Suffix, suffix)
}

// NewContainsMatch 创建一个包含匹配器
func NewContainsMatch(fieldName string, substr string) *FieldComparison {
	return NewFieldComparison(fieldName, Contains, substr)
}
