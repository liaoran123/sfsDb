package util

import (
	"math"
	"reflect"
	"strconv"
)

// AnyToInt 使用 reflect 包将任意类型转换为 int，处理所有数值类型和字符串类型
func AnyToInt(v any) int {
	if v == nil {
		return int(0)
	}
	// 使用 reflect 包获取值的类型和值
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// 检查是否超出 int64 的范围
		if val.Uint() > math.MaxInt64 {
			return int(math.MaxInt64)
		}
		return int(val.Uint())
	case reflect.Float32, reflect.Float64:
		// 检查是否超出 int64 的范围
		floatVal := val.Float()
		if floatVal > math.MaxInt64 {
			return int(math.MaxInt64)
		}
		if floatVal < math.MinInt64 {
			return int(math.MinInt64)
		}
		return int(floatVal)
	//字符串
	case reflect.String:
		if i, err := strconv.Atoi(val.String()); err == nil {
			return i
		}
		return 0

	default:
		return int(0)
	}
}
