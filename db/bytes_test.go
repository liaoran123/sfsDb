package db

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试所有类型和[]byte互相转换
func TestBytes(t *testing.T) {
	dt := DbTypes{}
	//测试String和[]byte互相转换
	str := "abcde"
	dt.String = String(str)
	bt := []byte(str)
	if !bytes.Equal(dt.String.ToBytes(), bt) {
		t.Errorf("dt.String.ToBytes()错误")
	}
	strb := Bytes(bt).String()
	fmt.Printf("bt: %v\n", bt)
	fmt.Printf("strb: %v\n", strb)
	assert.Equal(t, str, strb)

}
