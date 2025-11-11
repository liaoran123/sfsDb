package db

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试所有类型DbType和[]byte互相转换
func TestBytes(t *testing.T) {
	dt := DbTypes{}
	//测试String和[]byte互相转换
	str := "abc,de,"
	dt.String = String(str)
	bt := []byte(str)
	if !bytes.Equal(dt.String.ToBytes(), bt) {
		t.Errorf("dt.String.ToBytes()错误")
	}
	strb := Bytes(bt).String()
	fmt.Printf("bt: %v\n", bt)
	fmt.Printf("strb: %v\n", strb)
	assert.Equal(t, str, strb)

	//测试Time和[]byte互相转换
	timeStr := "2024-06-01 12:34:56"
	tm, _ := time.Parse("2006-01-02 15:04:05", timeStr)
	dt.Time = Time(tm)
	bt = dt.Time.ToBytes()
	tmb := Bytes(bt).Time()
	assert.Equal(t, tm, tmb)

	//测试Int和[]byte互相转换
	var intVal int64 = 1234567890
	dt.Int = Int(intVal)
	bt = dt.Int.ToBytes()
	intb := Bytes(bt).Int()
	assert.Equal(t, intVal, intb)
	//测试Float和[]byte互相转换
	var floatVal float64 = 12345.6789
	dt.Float = Float(floatVal)
	bt = dt.Float.ToBytes()
	floatb := Bytes(bt).Float()
	assert.Equal(t, floatVal, floatb)

	//测试Bool和[]byte互相转换
	dt.Bool = Bool(true)
	bt = dt.Bool.ToBytes()
	fmt.Printf("bt: %v\n", bt)
	Bytes(bt).Bool()
	assert.Equal(t, true, Bytes(bt).Bool())
	dt.Bool = Bool(false)
	bt = dt.Bool.ToBytes()
	assert.Equal(t, false, Bytes(bt).Bool())

	//测试合并Jion
	b1 := dt.String.ToBytes()
	b2 := dt.Time.ToBytes()
	b3 := dt.Int.ToBytes()
	b4 := dt.Float.ToBytes()
	b5 := dt.Bool.ToBytes()
	bj := Bytes(b1)
	bjJion := bj.Jion(b2, b3, b4, b5)
	fmt.Printf("合并后: %v\n", []byte(bjJion))

	//测试Split
	bs := Bytes(bjJion)
	splitBytes := bs.Split()
	assert.Equal(t, 5, len(splitBytes))
	assert.Equal(t, []byte("abc,de,"), splitBytes[0])
	assert.Equal(t, b2, splitBytes[1])
	assert.Equal(t, b3, splitBytes[2])
	assert.Equal(t, b4, splitBytes[3])
	assert.Equal(t, b5, splitBytes[4])
}
