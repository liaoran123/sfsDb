package util

import "strconv"

// StrToInt64 将字符串转换为int64类型
func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StrToUint64 将字符串转换为uint64类型
func StrToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// StrToInt 将字符串转换为int类型
func StrToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StrToInt32 将字符串转换为int32类型
func StrToInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

// StrToInt16 将字符串转换为int16类型
func StrToInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

// StrToInt8 将字符串转换为int8类型
func StrToInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

// StrToUint32 将字符串转换为uint32类型
func StrToUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

// StrToUint16 将字符串转换为uint16类型
func StrToUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

// StrToUint8 将字符串转换为uint8类型
func StrToUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}
