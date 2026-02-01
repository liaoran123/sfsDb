package types

import (
	"errors"
)

// Type 数据类型接口
type Type interface {
	// Name 获取类型名称
	Name() string

	// Size 获取类型大小
	Size() int

	// Encode 编码值
	Encode(value interface{}) ([]byte, error)

	// Decode 解码值
	Decode(data []byte) (interface{}, error)

	// Validate 验证值
	Validate(value interface{}) error
}

// TypeRegistry 类型注册中心
type TypeRegistry struct {
	types map[string]Type
}

// NewTypeRegistry 创建类型注册中心
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]Type),
	}
}

// RegisterType 注册类型
func (r *TypeRegistry) RegisterType(name string, typ Type) error {
	if _, exists := r.types[name]; exists {
		return errors.New("type already registered")
	}

	r.types[name] = typ
	return nil
}

// UnregisterType 注销类型
func (r *TypeRegistry) UnregisterType(name string) error {
	if _, exists := r.types[name]; !exists {
		return errors.New("type not found")
	}

	delete(r.types, name)
	return nil
}

// GetType 获取类型
func (r *TypeRegistry) GetType(name string) (Type, error) {
	typ, exists := r.types[name]
	if !exists {
		return nil, errors.New("type not found")
	}

	return typ, nil
}

// ListTypes 列出所有类型
func (r *TypeRegistry) ListTypes() []string {
	types := make([]string, 0, len(r.types))
	for name := range r.types {
		types = append(types, name)
	}

	return types
}

// Encode 编码值
func (r *TypeRegistry) Encode(typeName string, value interface{}) ([]byte, error) {
	typ, err := r.GetType(typeName)
	if err != nil {
		return nil, err
	}

	return typ.Encode(value)
}

// Decode 解码值
func (r *TypeRegistry) Decode(typeName string, data []byte) (interface{}, error) {
	typ, err := r.GetType(typeName)
	if err != nil {
		return nil, err
	}

	return typ.Decode(data)
}

// Validate 验证值
func (r *TypeRegistry) Validate(typeName string, value interface{}) error {
	typ, err := r.GetType(typeName)
	if err != nil {
		return err
	}

	return typ.Validate(value)
}
