package util

import (
	"time"
)

// TypeEnum 表示支持的数据类型枚举
type TypeEnum int

// 定义类型枚举值
const (
	TypeInt TypeEnum = iota
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeBool
	TypeString
	TypeTime
	TypeComplex64
	TypeComplex128
)

// GetTypeInstance 根据类型枚举返回相应的类型实例
func GetTypeInstance(t TypeEnum) any {
	switch t {
	case TypeInt:
		return int(0)
	case TypeInt8:
		return int8(0)
	case TypeInt16:
		return int16(0)
	case TypeInt32:
		return int32(0)
	case TypeInt64:
		return int64(0)
	case TypeUint:
		return uint(0)
	case TypeUint8:
		return uint8(0)
	case TypeUint16:
		return uint16(0)
	case TypeUint32:
		return uint32(0)
	case TypeUint64:
		return uint64(0)
	case TypeFloat32:
		return float32(0)
	case TypeFloat64:
		return float64(0)
	case TypeBool:
		return false
	case TypeString:
		return ""
	case TypeTime:
		return time.Time{}
	case TypeComplex64:
		return complex64(0)
	case TypeComplex128:
		return complex128(0)
	default: //序列化json
		return nil
	}
}

// GetTypeEnum 根据值的类型返回相应的类型枚举
func GetTypeEnum(v any) TypeEnum {
	switch v.(type) {
	case int:
		return TypeInt
	case int8:
		return TypeInt8
	case int16:
		return TypeInt16
	case int32:
		return TypeInt32
	case int64:
		return TypeInt64
	case uint:
		return TypeUint
	case uint8:
		return TypeUint8
	case uint16:
		return TypeUint16
	case uint32:
		return TypeUint32
	case uint64:
		return TypeUint64
	case float32:
		return TypeFloat32
	case float64:
		return TypeFloat64
	case bool:
		return TypeBool
	case string:
		return TypeString
	case time.Time:
		return TypeTime
	case complex64:
		return TypeComplex64
	case complex128:
		return TypeComplex128
	default: //序列化json
		return -1
	}
}

// ParseByType 根据类型枚举和字符串值返回解析后的值
func ParseByType(t TypeEnum, s string) (any, error) {
	return StrToAny(s, GetTypeInstance(t))
}

// ConvertToString 根据类型枚举和值返回字符串表示
func ConvertToString(v any) string {
	return AnyToStr(v)
}

// ConvertToBytes 根据类型枚举和值返回字节表示
func ConvertToBytes(v any) []byte {
	return AnyToBytes(v)
}
