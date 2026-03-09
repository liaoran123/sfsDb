package util

import (
	"strconv"
	"time"
)

/*
// TypeSizeFunc 定义计算类型大小的函数类型
type TypeSizeFunc func(value any) int

// 初始化默认类型大小计算函数
func init() {
	// 静态类型判断，无需动态注册
}
*/
// TypeSize 根据传入的值的类型返回相应的字节长度
func TypeSize(value any) int {
	if value == nil {
		return 0
	}

	// 静态类型判断，提高性能，减少内存使用
	switch value.(type) {
	case bool:
		return 1
	case int:
		return strconv.IntSize / 8
	case int8:
		return 1
	case uint8:
		return 1
	case int16:
		return 2
	case uint16:
		return 2
	case int32:
		return 4
	case uint32:
		return 4
	case int64:
		return 8
	case uint64:
		return 8
	case uint:
		return strconv.IntSize / 8
	case float32:
		return 4
	case float64:
		return 8
	case complex64:
		return 8
	case complex128:
		return 16
	case time.Time:
		// time.DateTime 格式的长度是固定的，例如 "2024-01-01 12:00:00"
		return len("2006-01-02 15:04:05")
	}

	// 默认返回 0
	return defaultTypeSize
}

var defaultTypeSize = 0

// 可以创建一个全局默认大小。用于不定长类型一般就是特指字符串类型作为组合主键时使用。
// 如果需要在表的局部使用，可以使用表的函数GetfieldTypeLen 精准设定。
func SetDefaultTypeSize(size int) {
	defaultTypeSize = size
}
