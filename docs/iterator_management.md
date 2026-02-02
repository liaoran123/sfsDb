# 迭代器资源管理指南

## 概述

本指南说明了 sfsDb 中迭代器资源的管理方式和生命周期，帮助开发者正确使用和管理迭代器，避免资源泄漏和并发冲突。

## 迭代器类型

### 1. storage.Iterator

**底层迭代器**，直接与存储引擎交互，由 `kvStore.Iterator()` 方法创建。

**管理方式**：
- **自动管理**：通过 `IteratorCache` 缓存管理
- **手动管理**：当不使用缓存时，需要手动调用 `Release()` 方法

**生命周期**：
- 创建：`kvStore.Iterator()`
- 使用：遍历数据
- 释放：`iter.Release()` 或 `IteratorCache` 自动释放

### 2. *engine.TableIter

**高级迭代器**，包装了 `storage.Iterator`，提供更多功能，由 `Table.Search()` 方法返回。

**管理方式**：
- **手动管理**：需要手动调用 `Release()` 方法
- **原因**：包含更多状态和逻辑，需要更精细的管理

**生命周期**：
- 创建：`Table.Search()`
- 使用：遍历和处理数据
- 释放：`iter.Release()`

## IteratorCache 使用指南

### 什么是 IteratorCache

`IteratorCache` 是一个用于缓存和管理 `storage.Iterator` 资源的组件，提供以下功能：

- **自动资源管理**：自动释放过期或不再使用的迭代器资源
- **并发安全**：支持高并发场景
- **内存效率**：使用 LRU 算法管理缓存大小
- **资源复用**：缓存和复用频繁使用的迭代器

### 使用示例

```go
import (
	"github.com/liaoran123/sfsDb/indexCache"
	"github.com/liaoran123/sfsDb/storage"
)

// 创建迭代器缓存，容量为 100
cache := indexCache.NewIteratorCache(100)

// 缓存迭代器
key := []byte("query_key")
iter := kvStore.Iterator(start, limit)
cache.Set(key, iter)

// 获取缓存的迭代器
cachedIter, found := cache.Get(key)
if found {
	// 使用缓存的迭代器
	// ...
	// 注意：不需要手动调用 Release()，缓存会自动管理
}

// 当不再需要缓存时，清空缓存
cache.Clear()
```

### 适用场景

`IteratorCache` 特别适合：

- **频繁重复查询**：相同查询条件的迭代器可以复用
- **高并发场景**：支持并发访问和管理
- **资源敏感场景**：自动释放资源，避免泄漏
- **内存受限环境**：LRU 算法管理缓存大小

## 最佳实践

### 1. storage.Iterator 管理

**推荐方式**：使用 `IteratorCache` 自动管理

```go
// 推荐：使用 IteratorCache
cache := indexCache.NewIteratorCache(100)
key := []byte("query_key")
iter := kvStore.Iterator(start, limit)
cache.Set(key, iter)
// 使用后不需要手动释放
```

**备选方式**：手动管理

```go
// 备选：手动管理
iter := kvStore.Iterator(start, limit)
defer iter.Release() // 手动释放
// 使用迭代器
// ...
```

### 2. *engine.TableIter 管理

**必须手动管理**：

```go
// 必须手动管理 TableIter
iter := table.Search(&conditions)
defer iter.Release() // 必须手动释放
// 使用迭代器
// ...
```

### 3. 并发安全

- **IteratorCache**：内部实现了线程安全，支持并发访问
- **手动管理**：需要自己处理并发安全

### 4. 避免的错误

- **双重释放**：不要对同一个迭代器既放入缓存又手动释放
- **资源泄漏**：不要忘记释放手动管理的迭代器
- **并发冲突**：在多线程环境中注意迭代器的使用

## 常见问题

### Q: IteratorCache 是否会导致并发冲突？

**A: 不会**。`IteratorCache` 内部使用了 `sync.Map` 和 `sync.RWMutex` 来保证并发安全，所有操作都是线程安全的。

### Q: 为什么 TableIter 需要手动释放？

**A: 因为 TableIter 包含了更多的状态和逻辑**，如跳转区间、匹配条件等，需要更精细的管理，不适合完全由缓存自动管理。

### Q: 如何判断一个迭代器是否需要手动释放？

**A: 根据迭代器类型判断**：
- `storage.Iterator`：可以使用 `IteratorCache` 自动管理
- `*engine.TableIter`：必须手动释放

## 结论

正确管理迭代器资源是确保 sfsDb 应用稳定运行的重要因素。通过本指南的指导，开发者可以：

- 选择合适的迭代器管理方式
- 避免资源泄漏和并发冲突
- 提高应用的性能和可靠性

请根据具体的使用场景，选择最适合的迭代器管理方式，确保应用的稳定运行。
