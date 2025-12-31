package table

import (
	"github.com/liaoran123/sfsDb/engine/types"
	"github.com/liaoran123/sfsDb/storage"
)

// Table 表接口
type Table interface {
	// GetName 获取表名
	GetName() string

	// GetSchema 获取表结构
	GetSchema() *types.TableSchema

	// Insert 插入记录
	Insert(values map[string]any) error

	// Update 更新记录
	Update(where any, values map[string]any) error

	// Delete 删除记录
	Delete(where any) error

	// Select 查询记录
	Select(fields []string, where any) (Result, error)

	// Get 获取单条记录
	Get(primaryKey any) (map[string]any, error)

	// Close 关闭表
	Close() error
}

// Result 查询结果接口
type Result interface {
	// Next 移动到下一条记录
	Next() bool

	// Scan 将当前记录扫描到指定的变量中
	Scan(dest ...any) error

	// Map 将当前记录转换为map
	Map() (map[string]any, error)

	// Count 获取结果集大小
	Count() int64

	// Close 关闭结果集
	Close() error
}

// TableImpl 表实现
type TableImpl struct {
	name         string
	schema       *types.TableSchema
	kvStore      storage.Store
	indexes      map[string]Index
	primaryIndex Index
}

// Index 索引接口
type Index interface {
	// GetName 获取索引名
	GetName() string

	// GetType 获取索引类型
	GetType() types.IndexType

	// Add 添加索引项
	Add(key []byte, value []byte) error

	// Delete 删除索引项
	Delete(key []byte) error

	// Get 查询索引项
	Get(key []byte) ([]byte, error)

	// Iterator 创建索引迭代器
	Iterator(prefix []byte, reverse bool) storage.Iterator

	// Close 关闭索引
	Close() error
}

// NewTable 创建新的表实例
func NewTable(name string, schema *types.TableSchema, kvStore storage.Store) (Table, error) {
	// 验证schema
	if err := validateSchema(schema); err != nil {
		return nil, err
	}

	// 创建表实例
	table := &TableImpl{
		name:    name,
		schema:  schema,
		kvStore: kvStore,
		indexes: make(map[string]Index),
	}

	// 创建主键索引

	primaryIndex, err := NewPrimaryIndex(table)
	if err != nil {
		return nil, err
	}
	table.primaryIndex = primaryIndex
	table.indexes["primary"] = primaryIndex

	// 创建其他索引
	for _, indexDef := range schema.Indexes {
		index, err := NewSecondaryIndex(table, indexDef)
		if err != nil {
			return nil, err
		}
		table.indexes[indexDef.Name] = index
	}

	return table, nil
}

// validateSchema 验证表结构
func validateSchema(schema *types.TableSchema) error {
	// 检查表名
	if schema.Name == "" {
		return &TableError{msg: "table name cannot be empty"}
	}

	// 检查字段
	if len(schema.Fields) == 0 {
		return &TableError{msg: "table must have at least one field"}
	}

	// 检查主键
	if len(schema.PrimaryKey) == 0 {
		return &TableError{msg: "table must have a primary key"}
	}

	return nil
}

// GetName 获取表名
func (t *TableImpl) GetName() string {
	return t.name
}

// GetSchema 获取表结构
func (t *TableImpl) GetSchema() *types.TableSchema {
	return t.schema
}

// Insert 插入记录
func (t *TableImpl) Insert(values map[string]any) error {
	// TODO: 实现插入记录逻辑
	return nil
}

// Update 更新记录
func (t *TableImpl) Update(where any, values map[string]any) error {
	// TODO: 实现更新记录逻辑
	return nil
}

// Delete 删除记录
func (t *TableImpl) Delete(where any) error {
	// TODO: 实现删除记录逻辑
	return nil
}

// Select 查询记录
func (t *TableImpl) Select(fields []string, where any) (Result, error) {
	// TODO: 实现查询记录逻辑
	return nil, nil
}

// Get 获取单条记录
func (t *TableImpl) Get(primaryKey any) (map[string]any, error) {
	// TODO: 实现获取单条记录逻辑
	return nil, nil
}

// Close 关闭表
func (t *TableImpl) Close() error {
	// 关闭所有索引
	for _, idx := range t.indexes {
		if err := idx.Close(); err != nil {
			return err
		}
	}
	return nil
}

// TableError 表错误类型
type TableError struct {
	msg string
}

// NewTableError 创建表错误
func NewTableError(msg string) *TableError {
	return &TableError{msg: msg}
}

// Error 返回错误消息
func (e *TableError) Error() string {
	return e.msg
}
