package engine

import (
	"sync"
)

// FieldsBytesPool 是用于管理 map[string][]byte 对象的池
type FieldsBytesPool struct {
	pool sync.Pool
}

// 全局 FieldsBytesPool 实例
var GlobalFieldsBytesPool *FieldsBytesPool

// 初始化全局对象池
func init() {
	GlobalFieldsBytesPool = NewFieldsBytesPool()
}

// NewFieldsBytesPool 创建一个新的 FieldsBytesPool
func NewFieldsBytesPool() *FieldsBytesPool {
	return &FieldsBytesPool{
		pool: sync.Pool{
			New: func() interface{} {
				// 创建一个新的 map[string][]byte，初始容量设为 10，适用于大多数场景
				return make(map[string][]byte, 10)
			},
		},
	}
}

// Get 从池中获取一个 map[string][]byte 对象
func (p *FieldsBytesPool) Get() map[string][]byte {
	m := p.pool.Get().(map[string][]byte)
	// 清空 map，确保返回的数据是干净的
	for k := range m {
		delete(m, k)
	}
	return m
}

// Put 将 map[string][]byte 对象放回池中
func (p *FieldsBytesPool) Put(m map[string][]byte) {
	if m == nil {
		return
	}

	// 清空 map，确保下次获取时是干净的
	for k := range m {
		delete(m, k)
	}

	// 将清空后的 map 放回池中
	p.pool.Put(m)
}

// GetMapPointer 从池中获取一个 map[string][]byte 并返回其指针
func (p *FieldsBytesPool) GetMapPointer() *map[string][]byte {
	m := p.Get()
	return &m
}

// PutMapPointer 将 map[string][]byte 指针指向的对象放回池中
func (p *FieldsBytesPool) PutMapPointer(mp *map[string][]byte) {
	if mp == nil || *mp == nil {
		return
	}
	p.Put(*mp)
	// 注意：这里不释放指针本身，因为指针是局部变量，会由编译器自动处理
}
