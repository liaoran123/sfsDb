package engine

import (
	"fmt"
	"strings"
)

/*
对象键（sys-table-name ,sys-tableid-idx-name，sys-tableid-field-name）读取已有ID
对象键不存在，则从对应类型的自动增长计数器（sys-table,sys-tableid-idx，sys-tableid-field）获取新ID
*/

// 常量定义
const (
	// SysPrefix 系统前缀
	SysPrefix = "sys"
	// 键类型常量
	KeyTypeTable = "table"
	KeyTypeIndex = "idx"
	KeyTypeField = "field"
	// Separator 键分隔符
	Separator = "-"
)

// ValueJoiner 用于实现系统ID键的生成和拼接
// 支持生成以下格式的键：
// 1. 对象键：sys-table-name, sys-tableid-idx-name, sys-tableid-field-name
// 2. 计数器键：sys-table, sys-tableid-idx, sys-tableid-field
type ValueJoiner struct{}

// NewValueJoiner 创建一个新的值拼接器
func NewValueJoiner() *ValueJoiner {
	return &ValueJoiner{}
}

// Join 实现基本的值拼接功能
// 根据输入的参数生成拼接后的字符串
func (vj *ValueJoiner) Join(parts ...string) string {
	return strings.Join(parts, Separator)
}

// GenerateObjectKey 生成基本对象键
// 示例：sys-type-name
func (vj *ValueJoiner) GenerateObjectKey(prefix, objType, name string) string {
	return vj.Join(prefix, objType, name)
}

// GenerateTableKey 生成表对象键
// 示例：sys-table-tableName
func (vj *ValueJoiner) GenerateTableKey(tableName string) string {
	return vj.GenerateObjectKey(SysPrefix, KeyTypeTable, tableName)
}

// formatTableID 将表ID转换为字符串
func formatTableID(tableID uint8) string {
	return fmt.Sprintf("%d", tableID)
}

// GenerateIndexKey 生成索引对象键
// 示例：sys-tableid-idx-indexName
func (vj *ValueJoiner) GenerateIndexKey(tableID uint8, indexName string) string {
	return vj.Join(SysPrefix, formatTableID(tableID), KeyTypeIndex, indexName)
}

// GenerateFieldKey 生成字段对象键
// 示例：sys-tableid-field-fieldName
func (vj *ValueJoiner) GenerateFieldKey(tableID uint8, fieldName string) string {
	return vj.Join(SysPrefix, formatTableID(tableID), KeyTypeField, fieldName)
}

// GenerateCounterKey 生成计数器键
// 示例：sys-type
func (vj *ValueJoiner) GenerateCounterKey(prefix, objType string) string {
	return vj.Join(prefix, objType)
}

// GenerateTableCounterKey 生成表计数器键
// 示例：sys-table
func (vj *ValueJoiner) GenerateTableCounterKey() string {
	return vj.GenerateCounterKey(SysPrefix, KeyTypeTable)
}

// GenerateIndexCounterKey 生成索引计数器键
// 示例：sys-tableid-idx
func (vj *ValueJoiner) GenerateIndexCounterKey(tableID uint8) string {
	return vj.Join(SysPrefix, formatTableID(tableID), KeyTypeIndex)
}

// GenerateFieldCounterKey 生成字段计数器键
// 示例：sys-tableid-field
func (vj *ValueJoiner) GenerateFieldCounterKey(tableID uint8) string {
	return vj.Join(SysPrefix, formatTableID(tableID), KeyTypeField)
}
