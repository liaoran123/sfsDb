package storage

import (
	"errors"
)

// Store 存储接口
type Store interface {
	// Put 存储键值对
	Put(key, value []byte) error

	// Get 获取值
	Get(key []byte) ([]byte, error)

	// Delete 删除键值对
	Delete(key []byte) error

	// Batch 获取批处理对象
	Batch() Batch

	// GetBatch 获取批处理对象
	GetBatch() Batch

	// WriteBatch 写入批处理
	WriteBatch(batch Batch) error

	// Iterator 创建迭代器
	Iterator(start, limit []byte) Iterator

	// Snapshot 创建快照
	Snapshot() (Snapshot, error)

	// Close 关闭存储
	Close() error
}

// Batch 批处理接口
type Batch interface {
	// Put 存储键值对
	Put(key, value []byte) error

	// Delete 删除键值对
	Delete(key []byte) error

	// Reset 重置批处理
	Reset()
}

// Iterator 迭代器接口
type Iterator interface {
	// First 移动到第一个键
	First() bool

	// Last 移动到最后一个键
	Last() bool

	// Seek 移动到指定键
	Seek(key []byte) bool

	// Next 移动到下一个键
	Next() bool

	// Prev 移动到上一个键
	Prev() bool

	// Key 获取当前键
	Key() []byte

	// Value 获取当前值
	Value() []byte

	// Valid 检查迭代器是否有效
	Valid() bool

	// Error 获取错误
	Error() error

	// Release 释放迭代器
	Release()
}

// Snapshot 快照接口
type Snapshot interface {
	// Get 获取值
	Get(key []byte) ([]byte, error)

	// Iterator 创建迭代器
	Iterator(start, limit []byte) Iterator

	// Release 释放快照
	Release()
}

// StoreFactory 存储工厂接口
type StoreFactory interface {
	// CreateStore 创建存储
	CreateStore(path string) (Store, error)

	// CreateStoreWithOptions 创建存储（带选项）
	CreateStoreWithOptions(path string, options map[string]any) (Store, error)

	// Name 获取存储名称
	Name() string

	// Version 获取存储版本
	Version() string
}

// StoreRegistry 存储注册中心
type StoreRegistry struct {
	factories map[string]StoreFactory
}

// NewStoreRegistry 创建存储注册中心
func NewStoreRegistry() *StoreRegistry {
	return &StoreRegistry{
		factories: make(map[string]StoreFactory),
	}
}

// RegisterFactory 注册存储工厂
func (r *StoreRegistry) RegisterFactory(name string, factory StoreFactory) error {
	if _, exists := r.factories[name]; exists {
		return errors.New("store factory already registered")
	}

	r.factories[name] = factory
	return nil
}

// UnregisterFactory 注销存储工厂
func (r *StoreRegistry) UnregisterFactory(name string) error {
	if _, exists := r.factories[name]; !exists {
		return errors.New("store factory not found")
	}

	delete(r.factories, name)
	return nil
}

// GetFactory 获取存储工厂
func (r *StoreRegistry) GetFactory(name string) (StoreFactory, error) {
	factory, exists := r.factories[name]
	if !exists {
		return nil, errors.New("store factory not found")
	}

	return factory, nil
}

// ListFactories 列出所有存储工厂
func (r *StoreRegistry) ListFactories() []string {
	factories := make([]string, 0, len(r.factories))
	for name := range r.factories {
		factories = append(factories, name)
	}

	return factories
}

// CreateStore 创建存储
func (r *StoreRegistry) CreateStore(factoryName, path string) (Store, error) {
	factory, err := r.GetFactory(factoryName)
	if err != nil {
		return nil, err
	}

	return factory.CreateStore(path)
}

// CreateStoreWithOptions 创建存储（带选项）
func (r *StoreRegistry) CreateStoreWithOptions(factoryName, path string, options map[string]any) (Store, error) {
	factory, err := r.GetFactory(factoryName)
	if err != nil {
		return nil, err
	}

	return factory.CreateStoreWithOptions(path, options)
}
