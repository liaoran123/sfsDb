package util

import (
	"encoding/binary"
)

// binary.BigEndian.PutUint32 大端
// binary.LittleEndian.PutUint32 小端
// Int16ToBytes 将 int16 按指定字节序转为 2 字节切片
func Int16ToBytes(v int16, order binary.ByteOrder) []byte {
	b := make([]byte, 2)
	order.PutUint16(b, uint16(v))
	return b
}

// Int32ToBytes 将 int32 按指定字节序转为 4 字节切片
func Int32ToBytes(v int32, order binary.ByteOrder) []byte {
	b := make([]byte, 4)
	order.PutUint32(b, uint32(v))
	return b
}

// Int64ToBytes 将 int64 按指定字节序转为 8 字节切片
func Int64ToBytes(v int64, order binary.ByteOrder) []byte {
	b := make([]byte, 8)
	order.PutUint64(b, uint64(v))
	return b
}
func IntToBytes(v int, order binary.ByteOrder) []byte {
	if intSize == 32 {
		return Int32ToBytes(int32(v), order)
	}
	return Int64ToBytes(int64(v), order)
}

// Uint16ToBytes 将 uint16 按指定字节序转为 2 字节切片
func Uint16ToBytes(v uint16, order binary.ByteOrder) []byte {
	b := make([]byte, 2)
	order.PutUint16(b, v)
	return b
}

// Uint32ToBytes 将 uint32 按指定字节序转为 4 字节切片
func Uint32ToBytes(v uint32, order binary.ByteOrder) []byte {
	b := make([]byte, 4)
	order.PutUint32(b, v)
	return b
}

// Uint64ToBytes 将 uint64 按指定字节序转为 8 字节切片
func Uint64ToBytes(v uint64, order binary.ByteOrder) []byte {
	b := make([]byte, 8)
	order.PutUint64(b, v)
	return b
}
