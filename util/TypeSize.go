package util

import (
	"reflect"
	"strconv"
	"sync"
)

// TypeSizeFunc 定义计算类型大小的函数类型
type TypeSizeFunc func(value any) int

// typeSizeRegistry 存储类型到大小计算函数的映射
var (
	typeSizeRegistry = make(map[string]TypeSizeFunc)
	typeSizeMutex    sync.RWMutex
)

// 初始化默认类型大小计算函数
func init() {
	// 注册基本类型
	RegisterTypeSize("bool", func(value any) int {
		return 1
	})

	RegisterTypeSize("int", func(value any) int {
		return strconv.IntSize / 8
	})

	RegisterTypeSize("int8", func(value any) int {
		return 1
	})

	RegisterTypeSize("uint8", func(value any) int {
		return 1
	})

	RegisterTypeSize("int16", func(value any) int {
		return 2
	})

	RegisterTypeSize("uint16", func(value any) int {
		return 2
	})

	RegisterTypeSize("int32", func(value any) int {
		return 4
	})

	RegisterTypeSize("uint32", func(value any) int {
		return 4
	})

	RegisterTypeSize("int64", func(value any) int {
		return 8
	})

	RegisterTypeSize("uint64", func(value any) int {
		return 8
	})

	RegisterTypeSize("uint", func(value any) int {
		return strconv.IntSize / 8
	})

	RegisterTypeSize("float32", func(value any) int {
		return 4
	})

	RegisterTypeSize("float64", func(value any) int {
		return 8
	})

	RegisterTypeSize("complex64", func(value any) int {
		return 8
	})

	RegisterTypeSize("complex128", func(value any) int {
		return 16
	})

	RegisterTypeSize("time.Time", func(value any) int {
		// time.DateTime 格式的长度是固定的，例如 "2024-01-01 12:00:00"
		return len("2006-01-02 15:04:05")
	})
}

// RegisterTypeSize 注册新类型的大小计算函数
func RegisterTypeSize(typeName string, sizeFunc TypeSizeFunc) {
	typeSizeMutex.Lock()
	defer typeSizeMutex.Unlock()
	typeSizeRegistry[typeName] = sizeFunc
}

// TypeSize 根据传入的值的类型返回相应的字节长度
func TypeSize(value any) int {
	if value == nil {
		return 0
	}

	// 获取类型名称
	typeName := reflect.TypeOf(value).String()

	// 检查是否在注册表中
	typeSizeMutex.RLock()
	sizeFunc, exists := typeSizeRegistry[typeName]
	typeSizeMutex.RUnlock()

	if exists {
		return sizeFunc(value)
	}

	// 检查是否为结构体类型
	if reflect.TypeOf(value).Kind() == reflect.Struct {
		return 0
	}

	// 默认返回 0
	return 0
}
