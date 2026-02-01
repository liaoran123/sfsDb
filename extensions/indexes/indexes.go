package indexes

import (
	"errors"
)

// Index 索引接口
type Index interface {
	// Name 获取索引名称
	Name() string

	// SetName 设置索引名称
	SetName(name string) error

	// AddFields 添加索引字段
	AddFields(fields ...string)

	// GetFields 获取索引字段
	GetFields() []string

	// Len 获取索引字段数量
	Len() int

	// SetId 设置索引ID
	SetId(id uint8)

	// GetId 获取索引ID
	GetId() uint8

	// Prefix 获取索引前缀
	Prefix(tbid uint8) []byte

	// Join 连接索引字段
	Join(fieldsBytes *map[string][]byte) []byte

	// JoinPrefix 连接索引前缀
	JoinPrefix(tbid uint8, val []byte) []byte

	// JoinValue 连接索引值
	JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte

	// MatchFields 匹配索引字段
	MatchFields(fields ...string) bool

	// Parse 解析索引值
	Parse(fields []string, value []byte) *map[string][]byte
}

// PrimaryKey 主键接口
type PrimaryKey interface {
	Index

	// GetID 获取主键ID
	GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte

	// GetfieldTypeLen 获取字段类型长度
	GetfieldTypeLen(tablefields *map[string]any) *map[string]uint8

	// ParsePrimaryKey 解析主键值
	ParsePrimaryKey(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
}

// FullTextIndex 全文索引接口
type FullTextIndex interface {
	Index

	// SetFullField 设置全文索引字段
	SetFullField(field string, len int) error

	// JoinFullValues 连接全文索引值
	JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte

	// Tokenize 分词
	Tokenize(nr string, ftlen int) (tokens []string)
}

// IndexFactory 索引工厂接口
type IndexFactory interface {
	// CreateIndex 创建索引
	CreateIndex(name string, fields ...string) (Index, error)

	// CreatePrimaryKey 创建主键
	CreatePrimaryKey(name string, fields ...string) (PrimaryKey, error)

	// CreateFullTextIndex 创建全文索引
	CreateFullTextIndex(name string, field string, ftlen int) (FullTextIndex, error)
}

// IndexRegistry 索引注册中心
type IndexRegistry struct {
	factories map[string]IndexFactory
}

// NewIndexRegistry 创建索引注册中心
func NewIndexRegistry() *IndexRegistry {
	return &IndexRegistry{
		factories: make(map[string]IndexFactory),
	}
}

// RegisterFactory 注册索引工厂
func (r *IndexRegistry) RegisterFactory(name string, factory IndexFactory) error {
	if _, exists := r.factories[name]; exists {
		return errors.New("index factory already registered")
	}

	r.factories[name] = factory
	return nil
}

// UnregisterFactory 注销索引工厂
func (r *IndexRegistry) UnregisterFactory(name string) error {
	if _, exists := r.factories[name]; !exists {
		return errors.New("index factory not found")
	}

	delete(r.factories, name)
	return nil
}

// GetFactory 获取索引工厂
func (r *IndexRegistry) GetFactory(name string) (IndexFactory, error) {
	factory, exists := r.factories[name]
	if !exists {
		return nil, errors.New("index factory not found")
	}

	return factory, nil
}

// ListFactories 列出所有索引工厂
func (r *IndexRegistry) ListFactories() []string {
	factories := make([]string, 0, len(r.factories))
	for name := range r.factories {
		factories = append(factories, name)
	}

	return factories
}

// CreateIndex 创建索引
func (r *IndexRegistry) CreateIndex(factoryName, name string, fields ...string) (Index, error) {
	factory, err := r.GetFactory(factoryName)
	if err != nil {
		return nil, err
	}

	return factory.CreateIndex(name, fields...)
}

// CreatePrimaryKey 创建主键
func (r *IndexRegistry) CreatePrimaryKey(factoryName, name string, fields ...string) (PrimaryKey, error) {
	factory, err := r.GetFactory(factoryName)
	if err != nil {
		return nil, err
	}

	return factory.CreatePrimaryKey(name, fields...)
}

// CreateFullTextIndex 创建全文索引
func (r *IndexRegistry) CreateFullTextIndex(factoryName, name string, field string, ftlen int) (FullTextIndex, error) {
	factory, err := r.GetFactory(factoryName)
	if err != nil {
		return nil, err
	}

	return factory.CreateFullTextIndex(name, field, ftlen)
}
