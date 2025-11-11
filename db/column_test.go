package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetType(t *testing.T) {
	col := Column{id: 1, name: "a"}
	//e := col.AddTypes("string")
	//对于nil的断言
	//assert.Nil(t, e)
	//fmt.Printf("e: %v\n", e)
	//fmt.Printf("col: %v\n", col)
	//断言相等
	assert.Equal(t, col.types, "string")
	//断言不相等
	//assert.NotEqual(t, col.GetType(), "string")

}
