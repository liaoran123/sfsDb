package engine

import (
	"strings"
	"sync"
	"sync/atomic"
)

// searchCache.go 文件包含搜索相关的缓存管理功能
// 主要包括：
// 1. 索引匹配缓存 (indexMatchCache)

// 3. 缓存统计功能
// 4. 缓存清除功能

// ConcurrentCache 并发安全的缓存实现
type ConcurrentCache struct {
	cache    *sync.Map
	size     int64
	accesses int64
	hits     int64
	maxSize  int64 // 最大缓存大小
}

// NewConcurrentCache 创建新的并发缓存
func NewConcurrentCache() *ConcurrentCache {
	return &ConcurrentCache{
		cache:   &sync.Map{},
		maxSize: 1000, // 默认最大缓存大小
	}
}

// Get 获取缓存条目
func (c *ConcurrentCache) Get(key string) (interface{}, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}
	atomic.AddInt64(&c.accesses, 1)
	if entry, ok := c.cache.Load(key); ok {
		atomic.AddInt64(&c.hits, 1)
		return entry, true
	}
	return nil, false
}

// Set 设置缓存条目
func (c *ConcurrentCache) Set(key string, entry interface{}) {
	if c == nil || c.cache == nil {
		return
	}
	// 检查缓存大小，如果超过限制，先清理
	if atomic.LoadInt64(&c.size) >= c.maxSize {
		c.evictOldEntries()
	}
	// 存储新条目
	c.cache.Store(key, entry)
	atomic.AddInt64(&c.size, 1)
}

// evictOldEntries 清理旧条目，实现简单的 FIFO 策略
func (c *ConcurrentCache) evictOldEntries() {
	if c == nil || c.cache == nil {
		return
	}
	// 清理策略：保留 80% 的缓存大小
	targetSize := int64(float64(c.maxSize) * 0.8)
	if targetSize < 0 {
		targetSize = 0
	}
	currentSize := atomic.LoadInt64(&c.size)
	if currentSize <= targetSize {
		return
	}
	// 需要清理的条目数
	toEvict := currentSize - targetSize
	if toEvict <= 0 {
		return
	}
	// 实现简单的 FIFO 策略：删除前 toEvict 个条目
	evicted := int64(0)
	c.cache.Range(func(key, value interface{}) bool {
		if evicted >= toEvict {
			return false
		}
		c.cache.Delete(key)
		evicted++
		return true
	})
	// 更新缓存大小
	atomic.AddInt64(&c.size, -evicted)
}

// Clear 清除缓存
func (c *ConcurrentCache) Clear() {
	if c == nil {
		return
	}
	c.cache = &sync.Map{}
	atomic.StoreInt64(&c.size, 0)
	atomic.StoreInt64(&c.accesses, 0)
	atomic.StoreInt64(&c.hits, 0)
}

// Size 获取缓存大小
func (c *ConcurrentCache) Size() int64 {
	return atomic.LoadInt64(&c.size)
}

// Stats 获取缓存统计信息
func (c *ConcurrentCache) Stats() (size, accesses, hits int64) {
	return atomic.LoadInt64(&c.size), atomic.LoadInt64(&c.accesses), atomic.LoadInt64(&c.hits)
}

// 全局缓存实例
var (
	indexMatchCache     = NewConcurrentCache()
	fieldsBytesCache    = NewConcurrentCache()
	indexCacheSizeLimit = 1000
)

// SetIndexCacheSizeLimit 设置索引缓存大小限制
func (t *Table) SetIndexCacheSizeLimit(limit int) {
	//控制一个合理数值，防止缓存大小过大
	if limit <= 0 {
		limit = 1000
	}
	indexCacheSizeLimit = limit
}

// 优化后的索引匹配
func (t *Table) MatchIndexCached(fields []string) Index {
	// 生成缓存键，不排序以提高性能
	cacheKey := strings.Join(fields, ",")

	// 尝试从缓存获取
	if idx, ok := indexMatchCache.Get(cacheKey); ok {
		return idx.(Index)
	}

	// 计算索引匹配
	idx := t.MatchIndex(fields...)

	// 缓存结果，ConcurrentCache.Set 已经是并发安全的
	indexMatchCache.Set(cacheKey, idx)

	return idx
}

// CacheStats 缓存统计信息结构体
type CacheStats struct {
	Size     int64   // 缓存大小
	Accesses int64   // 访问次数
	Hits     int64   // 命中次数
	HitRate  float64 // 命中率
}

// GetIndexMatchCacheStats 获取索引匹配缓存统计信息
func GetIndexMatchCacheStats() CacheStats {
	size, accesses, hits := indexMatchCache.Stats()
	hitRate := 0.0
	if accesses > 0 {
		hitRate = float64(hits) / float64(accesses)
	}
	return CacheStats{
		Size:     size,
		Accesses: accesses,
		Hits:     hits,
		HitRate:  hitRate,
	}
}

// GetAllCacheStats 获取所有缓存统计信息
func GetAllCacheStats() map[string]CacheStats {
	return map[string]CacheStats{
		"indexMatchCache": GetIndexMatchCacheStats(),
	}
}

// ClearIndexMatchCache 清除索引匹配缓存
func ClearIndexMatchCache() {
	indexMatchCache.Clear()
}

// ClearFieldsBytesCache 清除字段转换缓存
func ClearFieldsBytesCache() {
	fieldsBytesCache.Clear()
}

// ClearAllCaches 清除所有缓存
func ClearAllCaches() {
	ClearIndexMatchCache()
}
