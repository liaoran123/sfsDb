package engine

import (
	"sync"

	"github.com/liaoran123/sfsDb/storage"
)

// 全局 TableIter 对象池实例
var GlobalTableIterPool *TableIterPool

// 初始化全局对象池
func init() {
	GlobalTableIterPool = NewTableIterPool()
}

// TableIterPool 是 TableIter 对象的池
type TableIterPool struct {
	pool sync.Pool
}

// NewTableIterPool 创建一个新的 TableIter 对象池
func NewTableIterPool() *TableIterPool {
	return &TableIterPool{
		pool: sync.Pool{
			New: func() interface{} {
				t := &TableIter{
					move: make(map[bool]func() bool),
					top:  make(map[bool]func() bool),
				}
				return t
			},
		},
	}
}

/*
// NewTableIterPool 创建一个新的 TableIter 对象池
func NewTableIterPool() *TableIterPool {
	return &TableIterPool{
		pool: sync.Pool{
			New: func() interface{} {
				t := &TableIter{
					move: make(map[bool]func() bool),
					top:  make(map[bool]func() bool),
				}
				// 添加 finalizer，当对象被 GC 回收时自动释放资源
				runtime.SetFinalizer(t, func(t *TableIter) {
					// 释放迭代器资源
					if t.iter != nil {
						t.iter.Release()
					}
					// 释放跳跃区间迭代器资源
					if t.jumpRanges != nil {
						for _, jr := range t.jumpRanges {
							if jr != nil {
								jr.Release()
							}
						}
					}
				})
				return t
			},
		},
	}
}

*/

// Get 从池中获取一个干净的 TableIter 对象
func (p *TableIterPool) Get(table *Table, iter storage.Iterator, index Index, selects ...string) *TableIter {
	if iter == nil {
		return nil
	}
	t := p.pool.Get().(*TableIter)
	// 释放迭代器资源
	if t.iter != nil {
		t.iter.Release()
	}
	// 释放跳跃区间迭代器资源
	if t.jumpRanges != nil {
		for _, jr := range t.jumpRanges {
			if jr != nil {
				jr.Release()
			}
		}
	}
	// 重置并初始化 TableIter 对象
	t.table = table
	t.selects = selects
	t.index = index
	t.iter = iter
	t.jumpRanges = nil
	t.match = nil
	// 重新初始化函数映射
	t.move[true] = iter.Next
	t.move[false] = iter.Prev
	t.top[true] = iter.First
	t.top[false] = iter.Last
	return t
}

// Put 将 TableIter 对象放回池中
func (p *TableIterPool) Put(t *TableIter) {
	if t == nil {
		return
	}
	// 释放迭代器资源
	if t.iter != nil {
		t.iter.Release()
	}
	// 释放跳跃区间迭代器资源
	if t.jumpRanges != nil {
		for _, jr := range t.jumpRanges {
			if jr != nil {
				jr.Release()
			}
		}
	}
	// 清理 TableIter 对象，确保下次获取时是干净的
	t.table = nil
	t.selects = nil
	t.index = nil
	t.iter = nil
	t.jumpRanges = nil
	t.match = nil
	// 清空函数映射，但保留映射结构
	for k := range t.move {
		delete(t.move, k)
	}
	for k := range t.top {
		delete(t.top, k)
	}
	// 将清理后的对象放回池中
	p.pool.Put(t)
}
