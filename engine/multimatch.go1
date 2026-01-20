package engine

// LogicOperator 定义多字段匹配的逻辑操作符
type LogicOperator int

// 逻辑操作符枚举值
const (
	// LogicAND 表示逻辑与，所有匹配器都必须匹配
	LogicAND LogicOperator = iota
	// LogicOR 表示逻辑或，至少一个匹配器必须匹配
	LogicOR
)

// MultiFieldMatch 实现了 Match 接口，用于多字段匹配
type MultiFieldMatch struct {
	operator LogicOperator // 逻辑操作符
	matchers []Match       // 匹配器列表
}

// NewMultiFieldMatch 创建一个新的 MultiFieldMatch 实例
func NewMultiFieldMatch(operator LogicOperator, matchers ...Match) *MultiFieldMatch {
	return &MultiFieldMatch{
		operator: operator,
		matchers: matchers,
	}
}

// Match 实现了 Match 接口，用于多字段匹配
func (mm *MultiFieldMatch) Match(fields *map[string]any) bool {
	if fields == nil || len(mm.matchers) == 0 {
		return false
	}

	switch mm.operator {
	case LogicAND:
		// 逻辑与：所有匹配器都必须匹配
		for _, matcher := range mm.matchers {
			if !matcher.Match(fields) {
				return false
			}
		}
		return true
	case LogicOR:
		// 逻辑或：至少一个匹配器必须匹配
		for _, matcher := range mm.matchers {
			if matcher.Match(fields) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// AddMatcher 添加一个匹配器到多字段匹配器中
func (mm *MultiFieldMatch) AddMatcher(matcher Match) {
	mm.matchers = append(mm.matchers, matcher)
}

// NewAndMatch 创建一个逻辑与的多字段匹配器
func NewAndMatch(matchers ...Match) *MultiFieldMatch {
	return NewMultiFieldMatch(LogicAND, matchers...)
}

// NewOrMatch 创建一个逻辑或的多字段匹配器
func NewOrMatch(matchers ...Match) *MultiFieldMatch {
	return NewMultiFieldMatch(LogicOR, matchers...)
}

// ComplexMatch 是一个更复杂的匹配器，允许嵌套的逻辑操作
type ComplexMatch struct {
	matchers []any // 可以是 Match 或 ComplexMatch
	operator LogicOperator
}

// NewComplexMatch 创建一个新的 ComplexMatch 实例
func NewComplexMatch(operator LogicOperator, matchers ...any) *ComplexMatch {
	return &ComplexMatch{
		matchers: matchers,
		operator: operator,
	}
}

// Match 实现了 Match 接口，用于复杂的嵌套逻辑匹配
func (cm *ComplexMatch) Match(fields *map[string]any) bool {
	if fields == nil || len(cm.matchers) == 0 {
		return false
	}

	result := cm.operator == LogicAND // LogicAND 默认结果为 true，LogicOR 默认结果为 false

	for _, item := range cm.matchers {
		var matched bool

		// 检查 item 是 Match 还是 ComplexMatch
		if match, ok := item.(Match); ok {
			matched = match.Match(fields)
		} else if complexMatch, ok := item.(*ComplexMatch); ok {
			matched = complexMatch.Match(fields)
		} else {
			continue // 忽略无效类型
		}

		switch cm.operator {
		case LogicAND:
			result = result && matched
			if !result {
				return false // 短路：只要一个不匹配就返回 false
			}
		case LogicOR:
			result = result || matched
			if result {
				return true // 短路：只要一个匹配就返回 true
			}
		}
	}

	return result
}

// AddMatcher 添加一个匹配器或复杂匹配器到复杂匹配器中
func (cm *ComplexMatch) AddMatcher(matcher any) {
	cm.matchers = append(cm.matchers, matcher)
}

// NewAndComplexMatch 创建一个逻辑与的复杂匹配器
func NewAndComplexMatch(matchers ...any) *ComplexMatch {
	return NewComplexMatch(LogicAND, matchers...)
}

// NewOrComplexMatch 创建一个逻辑或的复杂匹配器
func NewOrComplexMatch(matchers ...any) *ComplexMatch {
	return NewComplexMatch(LogicOR, matchers...)
}
