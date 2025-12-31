package util

import (
	"encoding/binary"
	"strconv"
	"unsafe"
)

var EndianOrder = Endian()

// intSize 表示 int 类型的位数，32 或 64
const intSize = strconv.IntSize

// 判断大小端
func Endian() binary.ByteOrder {
	var i int = 0x1
	ptr := unsafe.Pointer(&i)
	b := *(*byte)(ptr)
	if b == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}
