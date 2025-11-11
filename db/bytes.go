package db

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

type Bytes []byte

func (b Bytes) Int() int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int
	if IsLitEndian {
		binary.Read(bytesBuffer, binary.LittleEndian, &x)
	} else {
		binary.Read(bytesBuffer, binary.BigEndian, &x)
	}
	return x
}
func (b Bytes) Float32() float32 {
	var bits uint32
	if IsLitEndian {
		bits = binary.LittleEndian.Uint32(b)
	} else {
		bits = binary.BigEndian.Uint32(b)
	}
	return math.Float32frombits(bits)
}
func (b Bytes) Float64() float64 {
	var bits uint64
	if IsLitEndian {
		bits = binary.LittleEndian.Uint64(b)
	} else {
		bits = binary.BigEndian.Uint64(b)
	}
	return math.Float64frombits(bits)
}
func (b Bytes) String() string {
	return string(b)
}

// 如果时间格式错误，返回2006-01-02 15:04:05
func (b Bytes) Time() time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", string(b))
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	}
	return t
}

// 合并多个
func (b Bytes) Jion(split string, ib ...[]byte) {
	for _, v := range ib {
		b = bytes.Join([][]byte{b, v}, []byte(split))
	}
}

/*
// 数据按DEFAULT_SPLIT分隔组合
func (b Bytes) JionDefault(ib ...[]byte) {
	for _, v := range ib {
		b = bytes.Join([][]byte{b, v}, []byte(DEFAULT_SPLIT))
	}
}

// 数据按INDEX_SPLIT分隔组合
func (b Bytes) JionIndex(ib ...[]byte) {
	for _, v := range ib {
		b = bytes.Join([][]byte{b, v}, []byte(INDEX_SPLIT))
	}
}


bytes.Equal 函数用于比较两个字节切片是否相等。
bytes.Compare 函数用于比较两个字节切片的大小。如果 b1 小于 b2，则返回 -1；如果 b1 等于 b2，则返回 0；如果 b1 大于 b2，则返回 1。
reflect.DeepEqual 函数可以比较两个任意类型的值是否相等。对于字节切片，也可以使用该函数。
*/
