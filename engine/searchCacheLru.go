package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LRUCache 基于 LRU 算法的缓存实现
type LRUCache struct {
	cache      map[string]cacheItem
	accessList []string          // 访问顺序列表，最近访问的在末尾
	capacity   int               // 缓存容量
	mutex      sync.RWMutex      // 读写锁
	enabled    bool              // 缓存是否启用
	size       int64             // 缓存大小
	accesses   int64             // 访问次数
	hits       int64             // 命中次数
}

// cacheItem 缓存项
type cacheItem struct {
	value      *map[string][]byte
	timestamp  int64
}

// NewLRUCache 创建新的 LRU 缓存
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		cache:      make(map[string]cacheItem),
		accessList: make([]string, 0, capacity),
		capacity:   capacity,
		enabled:    true,
	}
}

// Enable 启用缓存
func (c *LRUCache) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = true
}

// Disable 禁用缓存
func (c *LRUCache) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = false
	// 禁用时清空缓存
	c.clear()
}

// IsEnabled 检查缓存是否启用
func (c *LRUCache) IsEnabled() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.enabled
}

// Get 获取缓存项
func (c *LRUCache) Get(key string) (*map[string][]byte, bool) {
	// 先使用读锁检查缓存是否启用
	c.mutex.RLock()
	enabled := c.enabled
	c.mutex.RUnlock()

	// 如果缓存未启用，直接返回未命中
	if !enabled {
		return nil, false
	}

	// 使用写锁进行实际操作（因为需要更新访问顺序）
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 再次检查缓存是否启用（防止在获取锁期间状态发生变化）
	if !c.enabled {
		return nil, false
	}

	atomic.AddInt64(&c.accesses, 1)

	// 查找缓存项
	item, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	atomic.AddInt64(&c.hits, 1)

	// 更新访问顺序
	c.updateAccessOrder(key)

	return item.value, true
}

// Set 设置缓存项
func (c *LRUCache) Set(key string, value *map[string][]byte) {
	// 先使用读锁检查缓存是否启用
	c.mutex.RLock()
	enabled := c.enabled
	c.mutex.RUnlock()

	// 如果缓存未启用，直接返回
	if !enabled {
		return
	}

	// 使用写锁进行实际操作
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 再次检查缓存是否启用（防止在获取锁期间状态发生变化）
	if !c.enabled {
		return
	}

	// 检查缓存是否已满
	if len(c.cache) >= c.capacity {
		// 删除最久未使用的项
		c.evict()
	}

	// 添加新项
	c.cache[key] = cacheItem{
		value:     value,
		timestamp: time.Now().UnixNano(),
	}

	// 更新访问顺序
	c.updateAccessOrder(key)

	// 更新缓存大小
	atomic.StoreInt64(&c.size, int64(len(c.cache)))
}

// Clear 清空缓存
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clear()
}

// clear 内部清空缓存的方法
func (c *LRUCache) clear() {
	c.cache = make(map[string]cacheItem)
	c.accessList = make([]string, 0, c.capacity)
	atomic.StoreInt64(&c.size, 0)
	atomic.StoreInt64(&c.accesses, 0)
	atomic.StoreInt64(&c.hits, 0)
}

// evict 驱逐最久未使用的项
func (c *LRUCache) evict() {
	if len(c.accessList) == 0 {
		return
	}

	// 删除最久未使用的项（列表开头）
	oldestKey := c.accessList[0]
	delete(c.cache, oldestKey)
	c.accessList = c.accessList[1:]
}

// updateAccessOrder 更新访问顺序
func (c *LRUCache) updateAccessOrder(key string) {
	// 从访问列表中移除当前键
	for i, k := range c.accessList {
		if k == key {
			c.accessList = append(c.accessList[:i], c.accessList[i+1:]...)
			break
		}
	}

	// 将当前键添加到访问列表末尾
	c.accessList = append(c.accessList, key)
}

// Size 获取缓存大小
func (c *LRUCache) Size() int64 {
	return atomic.LoadInt64(&c.size)
}

// Stats 获取缓存统计信息
func (c *LRUCache) Stats() (size, accesses, hits int64) {
	return atomic.LoadInt64(&c.size), atomic.LoadInt64(&c.accesses), atomic.LoadInt64(&c.hits)
}

// 全局 LRU 缓存实例
var (
	fieldsBytesLRUCache = NewLRUCache(1000) // 默认容量 1000
)

// GetFieldsBytesLRUCache 获取 LRU 缓存实例
func GetFieldsBytesLRUCache() *LRUCache {
	return fieldsBytesLRUCache
}

// EnableFieldsBytesLRUCache 启用 LRU 缓存
func EnableFieldsBytesLRUCache() {
	fieldsBytesLRUCache.Enable()
}

// DisableFieldsBytesLRUCache 禁用 LRU 缓存
func DisableFieldsBytesLRUCache() {
	fieldsBytesLRUCache.Disable()
}

// FieldsToBytesNilLRU 使用 LRU 缓存的字段转换
func (t *Table) FieldsToBytesNilLRU(fields *map[string]any) *map[string][]byte {
	// 生成缓存键，基于字段名和字段值
	fieldCount := len(*fields)
	estimatedSize := fieldCount * 40 // 每个键值对预估40字节，增加一些冗余空间
	var cacheKey strings.Builder
	cacheKey.Grow(estimatedSize)

	// 对字段名排序，确保相同字段集合生成相同的缓存键
	fieldNames := make([]string, 0, fieldCount)
	for k := range *fields {
		fieldNames = append(fieldNames, k)
	}
	sort.Strings(fieldNames)

	// 生成缓存键
	for i, k := range fieldNames {
		v := (*fields)[k]
		cacheKey.WriteString(k)
		cacheKey.WriteByte(':')
		// 针对常见类型进行优化，减少 fmt.Sprintf 的开销
		switch val := v.(type) {
		case string:
			cacheKey.WriteString(val)
		case int:
			cacheKey.WriteString(strconv.Itoa(val))
		case int8:
			cacheKey.WriteString(strconv.Itoa(int(val)))
		case int16:
			cacheKey.WriteString(strconv.Itoa(int(val)))
		case int32:
			cacheKey.WriteString(strconv.FormatInt(int64(val), 10))
		case int64:
			cacheKey.WriteString(strconv.FormatInt(val, 10))
		case uint:
			cacheKey.WriteString(strconv.FormatUint(uint64(val), 10))
		case uint8:
			cacheKey.WriteString(strconv.FormatUint(uint64(val), 10))
		case uint16:
			cacheKey.WriteString(strconv.FormatUint(uint64(val), 10))
		case uint32:
			cacheKey.WriteString(strconv.FormatUint(uint64(val), 10))
		case uint64:
			cacheKey.WriteString(strconv.FormatUint(val, 10))
		case float32:
			cacheKey.WriteString(strconv.FormatFloat(float64(val), 'g', -1, 32))
		case float64:
			cacheKey.WriteString(strconv.FormatFloat(val, 'g', -1, 64))
		case bool:
			if val {
				cacheKey.WriteString("true")
			} else {
				cacheKey.WriteString("false")
			}
		case nil:
			cacheKey.WriteString("nil")
		default:
			// 对于复杂类型，使用 fmt.Sprintf
			cacheKey.WriteString(fmt.Sprintf("%v", val))
		}
		if i < len(fieldNames)-1 {
			cacheKey.WriteByte(',')
		}
	}

	cacheKeyStr := cacheKey.String()

	// 尝试从 LRU 缓存获取
	if fieldsBytes, ok := fieldsBytesLRUCache.Get(cacheKeyStr); ok {
		return fieldsBytes
	}

	// 计算字段转换
	fieldsBytes := t.FieldsToBytesNil(fields)

	// 缓存结果
	fieldsBytesLRUCache.Set(cacheKeyStr, fieldsBytes)

	return fieldsBytes
}
