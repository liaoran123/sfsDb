package record

import (
	"fmt"
	"strconv"
)

// Operation 运算接口，
// 对记录进行各种运算，并返回运算结果生成一个新的字段名称
// NewField()和返回的any，将会在record map[string]any中添加一个新的字段
type Operation interface {
	Evaluate(record map[string]any) any
	//运算后添加的字段名
	NewField() string //运算后添加的字段名
}

//添加常用字段的运算的接口实现

// CommonOperation 常用字段运算的实现
// 支持字段间的加减乘除等常用运算
// 运算结果会生成一个新的字段添加到记录中
type CommonOperation struct {
	opType   string         // 运算类型：add, sub, mul, div, avg, sum
	fields   []string       // 参与运算的字段列表
	newField string         // 运算后生成的新字段名
	args     map[string]any // 额外参数
}

// NewCommonOperation 创建一个新的常用字段运算实例
// opType: 运算类型，如 add, sub, mul, div, avg, sum
// fields: 参与运算的字段列表
// newField: 运算后生成的新字段名
// args: 额外参数，如除法运算的除数默认值等
func NewCommonOperation(opType string, fields []string, newField string, args map[string]any) *CommonOperation {
	if args == nil {
		args = make(map[string]any)
	}
	return &CommonOperation{
		opType:   opType,
		fields:   fields,
		newField: newField,
		args:     args,
	}
}

// Evaluate 执行字段运算并返回结果
// record: 要运算的记录数据
func (op *CommonOperation) Evaluate(record map[string]any) any {
	if len(op.fields) == 0 {
		return nil
	}

	switch op.opType {
	case "add", "sum":
		return op.sumFields(record)
	case "sub":
		return op.subFields(record)
	case "mul":
		return op.mulFields(record)
	case "div":
		return op.divFields(record)
	case "avg":
		return op.avgFields(record)
	case "max":
		return op.maxFields(record)
	case "min":
		return op.minFields(record)
	case "concat":
		return op.concatFields(record)
	default:
		return nil
	}
}

// NewField 返回运算后要添加的新字段名
func (op *CommonOperation) NewField() string {
	return op.newField
}

// sumFields 计算多个字段的和
func (op *CommonOperation) sumFields(record map[string]any) any {
	var sum float64
	for _, field := range op.fields {
		if val, ok := record[field]; ok {
			sum += op.toFloat64(val)
		}
	}
	return sum
}

// subFields 计算多个字段的差
// 规则：fields[0] - fields[1] - fields[2] - ...
func (op *CommonOperation) subFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	result := op.toFloat64(record[op.fields[0]])
	for i := 1; i < len(op.fields); i++ {
		result -= op.toFloat64(record[op.fields[i]])
	}
	return result
}

// mulFields 计算多个字段的乘积
func (op *CommonOperation) mulFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	result := op.toFloat64(record[op.fields[0]])
	for i := 1; i < len(op.fields); i++ {
		result *= op.toFloat64(record[op.fields[i]])
	}
	return result
}

// divFields 计算多个字段的商
// 规则：fields[0] / fields[1] / fields[2] / ...
// 支持设置默认除数，防止除零错误
func (op *CommonOperation) divFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	result := op.toFloat64(record[op.fields[0]])
	for i := 1; i < len(op.fields); i++ {
		divisor := op.toFloat64(record[op.fields[i]])
		if divisor == 0 {
			// 检查是否有默认除数设置
			if defaultDiv, ok := op.args["default_divisor"].(float64); ok {
				divisor = defaultDiv
			} else {
				// 默认返回0，避免除零错误
				return 0.0
			}
		}
		result /= divisor
	}
	return result
}

// avgFields 计算多个字段的平均值
func (op *CommonOperation) avgFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	sum := 0.0
	count := 0
	for _, field := range op.fields {
		if _, ok := record[field]; ok {
			sum += op.toFloat64(record[field])
			count++
		}
	}

	if count == 0 {
		return 0.0
	}
	return sum / float64(count)
}

// maxFields 计算多个字段中的最大值
func (op *CommonOperation) maxFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	maxVal := op.toFloat64(record[op.fields[0]])
	for i := 1; i < len(op.fields); i++ {
		val := op.toFloat64(record[op.fields[i]])
		if val > maxVal {
			maxVal = val
		}
	}
	return maxVal
}

// minFields 计算多个字段中的最小值
func (op *CommonOperation) minFields(record map[string]any) any {
	if len(op.fields) == 0 {
		return 0.0
	}

	minVal := op.toFloat64(record[op.fields[0]])
	for i := 1; i < len(op.fields); i++ {
		val := op.toFloat64(record[op.fields[i]])
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}

// concatFields 连接多个字段的值
// 支持设置连接符，默认为空字符串
func (op *CommonOperation) concatFields(record map[string]any) any {
	separator, ok := op.args["separator"].(string)
	if !ok {
		separator = ""
	}

	result := ""
	for i, field := range op.fields {
		if val, ok := record[field]; ok {
			if i > 0 {
				result += separator
			}
			result += op.toString(val)
		}
	}
	return result
}

// toFloat64 将任意类型转换为float64
func (op *CommonOperation) toFloat64(val any) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case string:
		// 尝试将字符串转换为数值
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0.0
	default:
		return 0.0
	}
}

// toString 将任意类型转换为字符串
func (op *CommonOperation) toString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 辅助函数：创建常用运算实例

// NewAddOperation 创建加法运算实例
func NewAddOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("add", fields, newField, nil)
}

// NewSubOperation 创建减法运算实例
func NewSubOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("sub", fields, newField, nil)
}

// NewMulOperation 创建乘法运算实例
func NewMulOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("mul", fields, newField, nil)
}

// NewDivOperation 创建除法运算实例
func NewDivOperation(fields []string, newField string, defaultDivisor float64) *CommonOperation {
	return NewCommonOperation("div", fields, newField, map[string]any{
		"default_divisor": defaultDivisor,
	})
}

// NewAvgOperation 创建平均值运算实例
func NewAvgOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("avg", fields, newField, nil)
}

// NewSumOperation 创建求和运算实例
func NewSumOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("sum", fields, newField, nil)
}

// NewMaxOperation 创建最大值运算实例
func NewMaxOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("max", fields, newField, nil)
}

// NewMinOperation 创建最小值运算实例
func NewMinOperation(fields []string, newField string) *CommonOperation {
	return NewCommonOperation("min", fields, newField, nil)
}

// NewConcatOperation 创建字符串连接运算实例
func NewConcatOperation(fields []string, newField string, separator string) *CommonOperation {
	return NewCommonOperation("concat", fields, newField, map[string]any{
		"separator": separator,
	})
}

//添加采用字段的运算的接口实现
