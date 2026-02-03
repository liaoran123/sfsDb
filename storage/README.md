# Storage Layer

存储层是sfsDb数据库的核心组件，负责数据的持久化存储和检索。该层采用了模块化设计，支持多种存储引擎，包括LevelDB、RocksDB和ToplingDB。

## 目录结构

```
storage/
├── kv.go            # KV存储接口定义
├── leveldb.go       # LevelDB存储引擎实现
├── rocksdb.go       # RocksDB存储引擎实现
├── toplingdb.go     # ToplingDB存储引擎实现
└── README.md        # 存储层文档
```

## 核心组件

### 1. KV存储接口 (kv.go)

定义了存储层的核心接口，包括：

- `Store`：KV存储的主要接口，提供数据的增删改查、批量操作、迭代器和快照功能
- `Batch`：批量操作接口，支持批量Put和Delete操作
- `Iterator`：迭代器接口，支持顺序遍历数据
- `Snapshot`：快照接口，提供一致读功能

### 2. LevelDB实现 (leveldb.go)

基于LevelDB的存储引擎实现，特点：

- 高效的键值存储
- 支持范围查询
- 支持快照功能
- 线程安全的读取操作
- 批量操作支持

### 3. RocksDB实现 (rocksdb.go)

基于RocksDB的存储引擎实现，特点：

- 高性能的键值存储
- 支持列族（Column Families）
- 支持多种压缩算法
- 支持快照功能
- 批量操作支持
- 可配置的缓存和内存管理

### 4. ToplingDB实现 (toplingdb.go)

基于ToplingDB的存储引擎实现，特点：

- 超高性能的键值存储
- 支持分布式部署
- 支持多种索引类型
- 支持快照功能
- 批量操作支持
- 自适应的压缩算法

## 接口设计

### Store接口

```go
type Store interface {
    // Get 获取指定key的值
    Get(key []byte) ([]byte, error)

    // Put 设置key-value对
    Put(key []byte, value []byte) error

    // Delete 删除指定key
    Delete(key []byte) error

    // Batch 创建批量操作对象
    Batch() Batch

    // Iterator 创建迭代器
    Iterator(para ...[]byte) Iterator

    // Snapshot 创建快照
    Snapshot() (Snapshot, error)

    // Close 关闭存储
    Close() error
}
```

### Batch接口

```go
type Batch interface {
    // Put 添加put操作
    Put(key []byte, value []byte)

    // Delete 添加delete操作
    Delete(key []byte)

    // Commit 提交批量操作
    Commit() error
}
```

### Iterator接口

```go
type Iterator interface {
    // First 移动到第一个元素
    First() bool

    // Last 移动到最后一个元素
    Last() bool

    // Seek 移动到大于等于指定key的位置
    Seek(key []byte) bool

    // Next 移动到下一个元素
    Next() bool

    // Prev 移动到前一个元素
    Prev() bool

    // Key 获取当前元素的key
    Key() []byte

    // Value 获取当前元素的value
    Value() []byte

    // Valid 检查迭代器是否有效
    Valid() bool

    // Release 释放迭代器资源
    Release()
}
```

### Snapshot接口

```go
type Snapshot interface {
    // Get 从快照中获取指定key的值
    Get(key []byte) ([]byte, error)

    // Iterator 从快照中创建迭代器
    Iterator(para ...[]byte) Iterator

    // Release 释放快照资源
    Release() error
}
```

## 使用示例

### 基本使用

```go
// 导入存储层包
import "github.com/liaoran123/sfsDb/storage"

// 创建存储实例
config := storage.StoreConfig{Path: "./test_db"}
store, err := storage.NewStore(config)
if err != nil {
    // 处理错误
}
defer store.Close()

// 写入数据
err = store.Put([]byte("key1"), []byte("value1"))
if err != nil {
    // 处理错误
}

// 读取数据
value, err := store.Get([]byte("key1"))
if err != nil {
    // 处理错误
}

// 删除数据
err = store.Delete([]byte("key1"))
if err != nil {
    // 处理错误
}
```

### 批量操作

```go
// 创建批量操作对象
batch := store.Batch()

// 添加操作
batch.Put([]byte("key1"), []byte("value1"))
batch.Put([]byte("key2"), []byte("value2"))
batch.Delete([]byte("key3"))

// 提交批量操作
err = batch.Commit()
if err != nil {
    // 处理错误
}
```

### 使用迭代器

```go
// 创建迭代器
iter := store.Iterator()
defer iter.Release()

// 遍历所有数据
for iter.First(); iter.Valid(); iter.Next() {
    key := iter.Key()
    value := iter.Value()
    // 处理数据
}

// 前缀扫描
iter = store.Iterator([]byte("prefix"))
defer iter.Release()

// 范围扫描
iter = store.Iterator([]byte("start"), []byte("end"))
defer iter.Release()
```

### 使用快照

```go
// 创建快照
snapshot, err := store.Snapshot()
if err != nil {
    // 处理错误
}
defer snapshot.Release()

// 从快照中读取数据
value, err := snapshot.Get([]byte("key"))
if err != nil {
    // 处理错误
}

// 从快照中创建迭代器
iter := snapshot.Iterator()
defer iter.Release()
```

### 使用外部存储实例

sfsDb 支持使用外部实现的 Store 实例，只需通过 `SetStore` 函数设置即可：

```go
// 导入存储层包
import "github.com/liaoran123/sfsDb/storage"

// 创建自定义存储实例
customStore := NewCustomStore()

// 设置外部存储实例
storage.SetStore(customStore)

// 现在整个 sfsdb 项目都会使用这个自定义存储实例
// 例如，BackupDb 等函数会自动使用这个实例

// 注意：使用外部存储实例时，需要自行管理其生命周期
// 确保在不再使用时正确关闭存储实例
// customStore.Close()
```

**使用场景**：
- 集成第三方存储实现
- 为特定场景定制存储逻辑
- 在测试中使用内存存储或模拟存储
- 实现特殊的存储功能，如加密、压缩等

## 支持的存储引擎

| 存储引擎 | 特点 | 适用场景 |
|---------|------|----------|
| LevelDB | 轻量级，性能稳定 | 小规模数据，对性能要求不高的场景 |
| RocksDB | 高性能，功能丰富 | 中等规模数据，对性能要求较高的场景 |
| ToplingDB | 超高性能，分布式支持 | 大规模数据，高并发场景 |

## 如何添加新的存储引擎

要添加新的存储引擎，只需要实现`Store`、`Batch`、`Iterator`和`Snapshot`接口即可。

1. 创建新的引擎实现文件（如`newengine.go`）
2. 实现`Store`接口的所有方法
3. 实现`Batch`、`Iterator`和`Snapshot`接口
4. 添加创建新引擎实例的函数

## 最佳实践

1. **关闭资源**：使用完存储实例、迭代器和快照后，一定要调用`Close()`或`Release()`方法释放资源
2. **批量操作**：对于大量的写操作，使用批量操作可以提高性能
3. **使用迭代器**：对于范围查询，使用迭代器比多次Get操作更高效
4. **合理使用快照**：快照可以提供一致读，但会增加内存使用，使用后及时释放
5. **选择合适的存储引擎**：根据数据规模和性能要求选择合适的存储引擎
6. **错误处理**：妥善处理存储操作返回的错误，避免数据丢失

## 性能优化

1. **调整缓存大小**：根据可用内存调整存储引擎的缓存大小
2. **优化压缩算法**：根据数据特点选择合适的压缩算法
3. **使用前缀扫描**：对于前缀查询，使用前缀扫描比全表扫描更高效
4. **批量提交**：合并多个小的写操作，减少磁盘IO
5. **避免频繁创建和销毁迭代器**：复用迭代器可以减少资源消耗

## 监控和调试

1. **添加日志**：在关键操作点添加日志，便于调试和监控
2. **收集性能指标**：收集存储引擎的性能指标，如读写延迟、吞吐量等
3. **使用存储引擎自带的工具**：如LevelDB的ldb工具、RocksDB的rocksdb工具
4. **定期检查数据完整性**：使用校验和等机制检查数据完整性

## 未来规划

1. 支持更多存储引擎
2. 实现分布式存储
3. 支持数据分片和复制
4. 实现自动故障恢复
5. 提供更丰富的监控和调试工具
6. 优化存储引擎的性能和资源使用
