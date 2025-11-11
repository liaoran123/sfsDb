package db

import "unsafe"

// 判断大小端
func IsLittleEndian() bool {
	var i int = 0x1
	ptr := unsafe.Pointer(&i)
	b := *(*byte)(ptr)
	return b == 1
}
