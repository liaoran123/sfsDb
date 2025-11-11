package db

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

// 列，字段
type Column struct {
	tbId    uint8 //表的id
	id      uint8
	name    string   //字段名
	types   string   //数据类型
	primary *Primary //主键，默认：nil，即非主键
	/*
		索引可以为多个，包括单索引和组合索引以及全文索引
		如按filed1索引查询，但是排序按filed2。如：select * from tb where filed1=3 order by filed2
		又按filed1索引查询，但是分组按filed2。如：select count(*) from tb where filed1=3 group by filed2
		这个时候必须创建filed1和filed2的组合索引。同时设置查询时使用该组合索引。
		不支持没有索引的查询和排序。以更规范的创建数据库和避免复杂的程序。
	*/
	index        []*Index //索引、全文索引、组合索引的数组
	CurrentIndex int      //当前查询使用的索引，index的下标
	DefultValue  []byte   //默认值，将值直接保存为对应的[]byte，则不需要每次都进行转换。
	Value        any      //用于添加，修改等赋值
	isdel        bool     //逻辑删除字段，保证大表删除字段不产生负担
}

// 设置字段名，不能包含标点符号
func (c *Column) Name(name string) error {
	//判断name是否合法
	for _, s := range name {
		if unicode.IsPunct(s) {
			return errors.New("字段名，不能包含标点符号。Field names cannot contain punctuation marks.")
		}
	}
	c.name = name
	return nil
}

// 更改字段名
func (c *Column) ChangeName(name string) {
	c.name = name
}

// 更改主键
func (c *Column) ChangePrimary(primary *Primary) {
	c.primary = primary
}

// 取消删除主键
func (c *Column) DeletePrimary() {
	c.primary = nil
}

// 检查类型是否匹配
func (c *Column) CheckValue() bool {
	types := strings.ToLower(c.types)
	switch c.Value.(type) {
	case string:
		if types == "string" {
			return true
		}
	case int:
		if types == "int" {
			return true
		}
	case bool:
		if types == "bool" {
			return true
		}
	case time.Time:
		if types == "time" {
			return true
		}
	case float32:
		if types == "float32" {
			return true
		}
	case float64:
		if types == "float64" {
			return true
		}

	}
	return false
}

// 添加索引
func (c *Column) Addindex(index *Index) {
	c.index = append(c.index, index)
	//对表原有数据加上索引kv
}

// 删除字段，逻辑删除
func (c *Column) Del() {
	c.isdel = true
}

// 删除字段，物理删除
func (c *Column) Clear() {
	//
}
