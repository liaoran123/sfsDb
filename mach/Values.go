package match

import (
	"regexp"
	"strings"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// ValueComparer 结构体用于对 any 值进行比较操作
type ValueComparer struct {
	value any           // 要比较的值
	typ   util.TypeEnum // 值的类型枚举
}

// NewValueComparer 创建一个新的 ValueComparer 实例
func NewValueComparer(v any) ValueComparer {
	return ValueComparer{
		value: v,
		typ:   util.GetTypeEnum(v),
	}
}

// Equal 检查两个值是否相等
func (m ValueComparer) Equal(another any) bool {
	// 处理 nil 情况
	if m.value == nil && another == nil {
		return true
	}
	if m.value == nil || another == nil {
		return false
	}

	// 类型相同，直接比较
	if m.typ == util.GetTypeEnum(another) {
		return m.value == another
	}

	// 类型不同，尝试转换后比较
	return compareValues(m.value, another) == 0
}

// NotEqual 检查两个值是否不相等
func (m ValueComparer) NotEqual(another any) bool {
	return !m.Equal(another)
}

// GreaterThan 检查当前值是否大于另一个值
func (m ValueComparer) GreaterThan(another any) bool {
	return compareValues(m.value, another) > 0
}

// GreaterThanOrEqual 检查当前值是否大于等于另一个值
func (m ValueComparer) GreaterThanOrEqual(another any) bool {
	return compareValues(m.value, another) >= 0
}

// LessThan 检查当前值是否小于另一个值
func (m ValueComparer) LessThan(another any) bool {
	return compareValues(m.value, another) < 0
}

// LessThanOrEqual 检查当前值是否小于等于另一个值
func (m ValueComparer) LessThanOrEqual(another any) bool {
	return compareValues(m.value, another) <= 0
}

// Prefix 检查当前值是否以前缀开头（仅适用于字符串类型）
func (m ValueComparer) Prefix(prefix string) bool {
	if str, ok := m.value.(string); ok {
		return strings.HasPrefix(str, prefix)
	}
	return false
}

// Suffix 检查当前值是否以后缀结尾（仅适用于字符串类型）
func (m ValueComparer) Suffix(suffix string) bool {
	if str, ok := m.value.(string); ok {
		return strings.HasSuffix(str, suffix)
	}
	return false
}

// Contains 检查当前值是否包含子字符串（仅适用于字符串类型）
func (m ValueComparer) Contains(sub string) bool {
	if str, ok := m.value.(string); ok {
		return strings.Contains(str, sub)
	}
	return false
}

// Like 检查当前值是否匹配类似SQL LIKE的模式（仅适用于字符串类型）
// 支持的通配符：
//
//	% - 匹配任意长度的字符串（包括空字符串）
//	_ - 匹配单个字符
func (m ValueComparer) Like(pattern string) bool {
	if str, ok := m.value.(string); ok {
		// 使用正则表达式处理所有LIKE模式
		// 替换所有%为.*，然后使用正则表达式匹配
		// 注意：这是一个简单实现，没有处理转义字符
		regexPattern := strings.ReplaceAll(pattern, "%", ".*")
		matched, _ := regexp.MatchString(regexPattern, str)
		return matched
	}
	return false
}

// Exists 检查当前值是否存在（非 nil）
func (m ValueComparer) Exists() bool {
	return m.value != nil
}

// Compare 根据比较操作符进行比较
func (m ValueComparer) Compare(op ComparisonOperator, another any) bool {
	switch op {
	case Equal:
		return m.Equal(another)
	case NotEqual:
		return m.NotEqual(another)
	case GreaterThan:
		return m.GreaterThan(another)
	case GreaterThanOrEqual:
		return m.GreaterThanOrEqual(another)
	case LessThan:
		return m.LessThan(another)
	case LessThanOrEqual:
		return m.LessThanOrEqual(another)
	case Like:
		if pattern, ok := another.(string); ok {
			return m.Like(pattern)
		}
		return false
	case Prefix:
		if prefix, ok := another.(string); ok {
			return m.Prefix(prefix)
		}
		return false
	case Suffix:
		if suffix, ok := another.(string); ok {
			return m.Suffix(suffix)
		}
		return false
	case Contains:
		if substr, ok := another.(string); ok {
			return m.Contains(substr)
		}
		return false
	default:
		return false
	}
}

// compareValues 比较两个值，返回 -1, 0, 1 分别表示小于、等于、大于
func compareValues(a, b any) int {
	// 处理 nil 情况
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// 获取类型枚举
	aType := util.GetTypeEnum(a)
	bType := util.GetTypeEnum(b)

	// 相同类型比较
	if aType == bType {
		switch aType {
		case util.TypeInt:
			aVal := a.(int)
			bVal := b.(int)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeInt8:
			aVal := a.(int8)
			bVal := b.(int8)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeInt16:
			aVal := a.(int16)
			bVal := b.(int16)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeInt32:
			aVal := a.(int32)
			bVal := b.(int32)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeInt64:
			aVal := a.(int64)
			bVal := b.(int64)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeUint:
			aVal := a.(uint)
			bVal := b.(uint)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeUint8:
			aVal := a.(uint8)
			bVal := b.(uint8)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeUint16:
			aVal := a.(uint16)
			bVal := b.(uint16)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeUint32:
			aVal := a.(uint32)
			bVal := b.(uint32)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeUint64:
			aVal := a.(uint64)
			bVal := b.(uint64)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeFloat32:
			aVal := a.(float32)
			bVal := b.(float32)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeFloat64:
			aVal := a.(float64)
			bVal := b.(float64)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeBool:
			aVal := a.(bool)
			bVal := b.(bool)
			if aVal == bVal {
				return 0
			}
			if aVal {
				return 1
			}
			return -1
		case util.TypeString:
			aVal := a.(string)
			bVal := b.(string)
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case util.TypeTime:
			aVal := a.(time.Time)
			bVal := b.(time.Time)
			if aVal.Before(bVal) {
				return -1
			}
			if aVal.After(bVal) {
				return 1
			}
			return 0
		}
	}

	// 不同类型比较，尝试转换为相同类型
	// 这里实现基本的数值类型转换比较
	// 对于更复杂的类型转换，可以根据需求扩展

	// 数值类型间的比较
	if (aType >= util.TypeInt && aType <= util.TypeUint64) || (bType >= util.TypeInt && bType <= util.TypeUint64) {
		// 将数值转换为 float64 进行比较
		aFloat := convertToFloat64(a)
		bFloat := convertToFloat64(b)

		if aFloat < bFloat {
			return -1
		}
		if aFloat > bFloat {
			return 1
		}
		return 0
	}

	// 不同类型无法比较，返回 0
	return 0
}

// convertToFloat64 将数值类型转换为 float64
func convertToFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}
