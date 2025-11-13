package db

import (
	"encoding/binary"
	"math"
	"time"
)

type DbType interface {
	ToBytes() []byte
}

type String string

// 处理分隔符DEFAULT_SPLIT=‘-’字符
// 遇到分隔符时，重复写入两次作为转义。
// DEFAULT_SPLIT不能是系统默认的分隔符/否则会导致无法区分是数据中的分隔符还是实际的分隔符
func (s String) ToBytes() []byte {
	return []byte(s)

}

type Time time.Time

// 时间类存储仍然是字符串
func (t Time) ToBytes() []byte {
	dt := time.Time(t).Format("2006-01-02 15:04:05")
	return []byte(dt)
}

type Bool bool

func (b Bool) ToBytes() []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

// 整型统一用int64存储
type Int int64

func (i Int) ToBytes() []byte {
	bytesBuffer := make([]byte, 8)
	if IsLitEndian {
		binary.LittleEndian.PutUint64(bytesBuffer, uint64(i))
	} else {
		binary.BigEndian.PutUint64(bytesBuffer, uint64(i))
	}
	return bytesBuffer
}

// 浮点型统一用float64存储
type Float float64

func (f Float) ToBytes() []byte {
	bits := math.Float64bits(float64(f))
	bytes := make([]byte, 8)
	if IsLitEndian {
		binary.LittleEndian.PutUint64(bytes, bits)
	} else {
		binary.BigEndian.PutUint64(bytes, bits)
	}
	return bytes
}

type DbTypes struct {
	Bool   Bool
	Int    Int
	Float  Float
	Time   Time
	String String
}
