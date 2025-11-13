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

	str = "(abc(,)de(,)"
	b1 = []byte(str)
	fb := byte(45)
	eb := Bytes(b1).Format(45)
	fmt.Printf(" %v\n", []byte(eb))
	beb := Bytes(eb).String()
	fmt.Printf("转义后:beb: %v\n", beb)
	i, ub := Bytes(eb).UnFormat()
	fmt.Printf("id: %v\n", i)
	assert.Equal(t, i, fb)
	fmt.Printf("反转义后: %v\n", ub)
	sub := Bytes(ub).String()
	fmt.Printf("sub: %v\n", sub)
	assert.Equal(t, str, sub)

	//测试Split
	dtbool := dt.Bool.ToBytes()
	fbool := Bytes(dtbool).Format(1)
	dtint := dt.Int.ToBytes()
	fint := Bytes(dtint).Format(2)
	dtfloat := dt.Float.ToBytes()
	ffloat := Bytes(dtfloat).Format(3)
	dttime := dt.Time.ToBytes()
	ftime := Bytes(dttime).Format(4)
	dtstring := dt.String.ToBytes()
	fstring := Bytes(dtstring).Format(5)
	merged := Bytes(fbool).Jion(fint, ffloat, ftime, fstring)
	fmt.Printf("合并后: %v\n", []byte(merged))
	splited := Bytes(merged).Split()
	id, data := Bytes(splited[0]).UnFormat()
	assert.Equal(t, byte(1), id)
	assert.Equal(t, dt.Bool.ToBytes(), data)
	id, data = Bytes(splited[1]).UnFormat()
	assert.Equal(t, byte(2), id)
	assert.Equal(t, dt.Int.ToBytes(), data)
	id, data = Bytes(splited[2]).UnFormat()
	assert.Equal(t, byte(3), id)
	assert.Equal(t, dt.Float.ToBytes(), data)
	id, data = Bytes(splited[3]).UnFormat()
	assert.Equal(t, byte(4), id)
	assert.Equal(t, dt.Time.ToBytes(), data)
	id, data = Bytes(splited[4]).UnFormat()
	assert.Equal(t, byte(5), id)
	assert.Equal(t, dt.String.ToBytes(), data)

	for _, v := range splited {
		id, data := Bytes(v).UnFormat()
		fmt.Printf("分割后 %d : %v:%s\n", id, data, string(data))
	}
}

// 递归计算斐波那契数列
func Fib(n int) int {
	if n <= 1 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

// 基准测试函数，内存性能测试
func BenchmarkFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fib(30) // 测试计算第30个斐波那契数
	}
}
