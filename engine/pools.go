package engine

import (
	"sync"

	"github.com/liaoran123/sfsDb/storage"
)

var (
	// mapPool 用于管理 map[any]bool 类型的对象池
	mapPool = sync.Pool{
		New: func() any {

			return make(map[any]bool)
		},
	}
	/*小对象，对象池操作开销可能大于直接创建
	// fieldsBytesPool 用于管理 map[string][]byte 类型的对象池
	fieldsBytesPool = sync.Pool{
		New: func() any {
			return make(map[string][]byte)
		},
	}
	*/
	// stringSlicePool 用于管理 []string 类型的对象池
	stringSlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 10) // 预分配容量为 10
		},
	}
)

// GetMap 从对象池获取一个 map[any]bool 对象
// sync.Pool不会导致内存泄漏，确保返回的数据干净即可。
func GetMap() map[any]bool {
	m := mapPool.Get().(map[any]bool)
	// 清空 map 中的所有键值对，确保返回的数据干净。保险操作，避免数据不干净
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMap 将 map[any]bool 对象归还到对象池
func PutMap(m map[any]bool) {
	// 清空 map 中的所有键值对，确保归还的对象干净
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

// GetStringSlice 从对象池获取一个 []string 切片
func GetStringSlice() []string {
	s := stringSlicePool.Get().([]string)
	// 清空切片，确保返回的切片是空的
	s = s[:0]
	return s
}

// PutStringSlice 将 []string 切片归还到对象池
func PutStringSlice(s []string) {
	if s == nil {
		return
	}
	// 清空切片，确保归还的切片是空的
	s = s[:0]
	stringSlicePool.Put(s)
}

// ResetStringSlicePool 重置 stringSlicePool 对象池
func ResetStringSlicePool() {
	// 由于 sync.Pool 没有直接的重置方法，我们可以通过替换来实现
	stringSlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 10)
		},
	}
}

// batchContainerPool 是 batchContainer 的对象池
var batchContainerPool = &sync.Pool{
	New: func() any {
		return &batchContainer{
			values: make(map[uint8][]byte, 3),
		}
	},
}

// GetBatchContainer 从对象池中获取一个 batchContainer
func GetBatchContainer(batch storage.Batch, indexs *Indexs, tbid uint8, kvStore storage.Store) *batchContainer {
	// 从对象池中获取一个 batchContainer
	c := batchContainerPool.Get().(*batchContainer)
	// 设置 batchContainer 的状态
	c.indexs = indexs
	c.batch = batch
	c.tbid = tbid
	c.kvStore = kvStore
	c.maxBatchSize = -1
	c.values[0] = nil
	c.values[1] = nil
	c.values[2] = nil
	return c
}

// PutBatchContainer 将 batchContainer 归还到对象池
func PutBatchContainer(c *batchContainer) {
	// 重置 batchContainer 的状态
	c.indexs = nil
	c.batch = nil
	c.tbid = 0
	c.kvStore = nil
	c.maxBatchSize = -1
	// 清空 values 映射
	for k := range c.values {
		delete(c.values, k)
	}
	// 重新初始化 values 映射
	c.values[0] = nil
	c.values[1] = nil
	c.values[2] = nil
	// 归还到对象池
	batchContainerPool.Put(c)
}
