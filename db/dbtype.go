package db

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

type DbType interface {
	ToBytes() []byte
}

type String string

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

type Int int

func (i Int) ToBytes() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if IsLitEndian {
		binary.Write(bytesBuffer, binary.LittleEndian, i)
	} else {
		binary.Write(bytesBuffer, binary.BigEndian, i)
	}
	return bytesBuffer.Bytes()
}

type Float64 float64

func (f Float64) ToBytes() []byte {
	bits := math.Float64bits(float64(f))
	bytes := make([]byte, 8)
	if IsLitEndian {
		binary.LittleEndian.PutUint64(bytes, bits)
	} else {
		binary.BigEndian.PutUint64(bytes, bits)
	}
	return bytes
}

type Float32 float32

func (f Float32) ToBytes() []byte {
	bits := math.Float32bits(float32(f))
	bytes := make([]byte, 4)
	if IsLitEndian {
		binary.LittleEndian.PutUint32(bytes, bits)
	} else {
		binary.BigEndian.PutUint32(bytes, bits)
	}
	return bytes
}

type DbTypes struct {
	Bool    Bool
	Int     Int
	Float32 Float32
	Float64 Float64
	Time    Time
	String  String
}
