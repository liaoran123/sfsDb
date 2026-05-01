# 第 5 章：Storage 存储包

---

## 5.1 完整的可运行示例

让我们通过一个完整的示例来演示 sfsDb 的 Storage 存储包功能：

```go
package main

import (
	"fmt"
	"os"

	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("=== sfsDb 数据库实战 - 第 5 章：Storage存储包示例 ===")
	fmt.Println()

	dbPath := "./bible_examples_chapter05_db"
	os.RemoveAll(dbPath)
	defer os.RemoveAll(dbPath)

	fmt.Println("1. 使用DBManager打开数据库")
	db, err := storage.GetDBManager().OpenDB(dbPath)
	if err != nil {
		fmt.Printf("数据库打开失败: %v\n", err)
		return
	}
	defer storage.GetDBManager().CloseDB()
	fmt.Println("   ✓ 数据库打开成功")
	fmt.Println()

	fmt.Println("2. 基本KV操作 - Put和Get")
	key1 := []byte("user:1")
	value1 := []byte(`{"name":"Alice","age":25}`)
	err = db.Put(key1, value1)
	if err != nil {
		fmt.Printf("Put失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 成功写入 key: user:1")

	key2 := []byte("user:2")
	value2 := []byte(`{"name":"Bob","age":30}`)
	err = db.Put(key2, value2)
	if err != nil {
		fmt.Printf("Put失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 成功写入 key: user:2")

	getValue1, err := db.Get(key1)
	if err != nil {
		fmt.Printf("Get失败: %v\n", err)
		return
	}
	fmt.Printf("   ✓ 读取 user:1: %s\n", getValue1)
	fmt.Println()

	fmt.Println("3. 批量操作")
	batch := db.GetBatch()
	batch.Put([]byte("product:1"), []byte("Laptop"))
	batch.Put([]byte("product:2"), []byte("Mouse"))
	batch.Put([]byte("product:3"), []byte("Keyboard"))
	batch.Delete([]byte("user:2"))

	err = db.WriteBatch(batch)
	if err != nil {
		fmt.Printf("批量操作失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 批量操作成功（添加3个product，删除user:2）")
	fmt.Println()

	fmt.Println("4. 使用迭代器遍历数据")
	iter := db.Iterator([]byte("product:"), []byte("product:zzzz"))
	fmt.Println("   遍历 product 开头的key:")
	for iter.First(); iter.Valid(); iter.Next() {
		fmt.Printf("   Key: %s, Value: %s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	fmt.Println()

	fmt.Println("5. 创建快照")
	snapshot, err := db.Snapshot()
	if err != nil {
		fmt.Printf("创建快照失败: %v\n", err)
		return
	}
	defer snapshot.Release()
	fmt.Println("   ✓ 快照创建成功")

	err = db.Put([]byte("product:4"), []byte("Monitor"))
	if err != nil {
		fmt.Printf("Put失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 在主数据库添加 product:4")

	snapValue, err := snapshot.Get([]byte("product:4"))
	if err == storage.ErrNotFound {
		fmt.Println("   ✓ 快照中看不到 product:4（符合预期）")
	} else if err != nil {
		fmt.Printf("快照读取失败: %v\n", err)
		return
	} else {
		fmt.Printf("   快照中 product:4 的值: %s\n", snapValue)
	}
	fmt.Println()

	fmt.Println("6. 删除操作")
	err = db.Delete([]byte("user:1"))
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 成功删除 user:1")

	_, err = db.Get([]byte("user:1"))
	if err == storage.ErrNotFound {
		fmt.Println("   ✓ 确认 user:1 已删除")
	} else if err != nil {
		fmt.Printf("读取失败: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Println("=== 示例运行完成 ===")
}
```

---

## 5.2 KV 存储接口

Storage 包定义了统一的 KV 存储接口，支持多种后端实现。

### Store 接口

```go
type Store interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	GetBatch() Batch
	WriteBatch(batch Batch, put ...bool) error
	Iterator(start, limit []byte) Iterator
	Snapshot() (Snapshot, error)
	SwitchToSnapshot() error
	SwitchToDB() error
	Close() error
}
```

### 主要方法说明：

#### Get

获取指定 key 的值。

```go
func Get(key []byte) ([]byte, error)
```

**示例：**
```go
value, err := db.Get([]byte("key"))
if err == storage.ErrNotFound {
    fmt.Println("key不存在")
} else if err != nil {
    fmt.Printf("读取错误: %v\n", err)
}
```

#### Put

设置 key-value 对。

```go
func Put(key []byte, value []byte) error
```

**示例：**
```go
err := db.Put([]byte("key"), []byte("value"))
if err != nil {
    fmt.Printf("写入错误: %v\n", err)
}
```

#### Delete

删除指定 key。

```go
func Delete(key []byte) error
```

**示例：**
```go
err := db.Delete([]byte("key"))
if err != nil {
    fmt.Printf("删除错误: %v\n", err)
}
```

---

## 5.3 批量操作

### Batch 接口

```go
type Batch interface {
	Put(key []byte, value []byte)
	Delete(key []byte)
	Len() int
	Reset()
}
```

### 使用示例：

```go
batch := db.GetBatch()
batch.Put([]byte("key1"), []byte("value1"))
batch.Put([]byte("key2"), []byte("value2"))
batch.Delete([]byte("key3"))

err := db.WriteBatch(batch)
if err != nil {
    fmt.Printf("批量操作失败: %v\n", err)
}
```

### GetBatch

获取批处理操作对象。

```go
func GetBatch() Batch
```

### WriteBatch

执行批量写入操作。

```go
func WriteBatch(batch Batch, put ...bool) error
```

**参数说明：**
- `batch`: 批量操作对象
- `put`: 是否将 batch 放回对象池，默认是 true

---

## 5.4 迭代器

### Iterator 接口

```go
type Iterator interface {
	First() bool
	Last() bool
	Seek(key []byte) bool
	Next() bool
	Prev() bool
	Key() []byte
	Value() []byte
	Valid() bool
	Release()
}
```

### 主要方法说明：

#### First

移动到第一个元素。

```go
func First() bool
```

#### Last

移动到最后一个元素。

```go
func Last() bool
```

#### Seek

移动到大于等于指定 key 的位置。

```go
func Seek(key []byte) bool
```

#### Next

移动到下一个元素。

```go
func Next() bool
```

#### Prev

移动到前一个元素。

```go
func Prev() bool
```

#### Key

获取当前元素的 key。

```go
func Key() []byte
```

#### Value

获取当前元素的 value。

```go
func Value() []byte
```

#### Valid

检查迭代器是否有效。

```go
func Valid() bool
```

#### Release

释放迭代器资源。

```go
func Release()
```

### 使用示例：

```go
iter := db.Iterator([]byte("user:"), []byte("user:zzzz"))
defer iter.Release()

for iter.First(); iter.Valid(); iter.Next() {
    fmt.Printf("Key: %s, Value: %s\n", iter.Key(), iter.Value())
}
```

---

## 5.5 快照功能

### Snapshot 接口

```go
type Snapshot interface {
	Get(key []byte) ([]byte, error)
	Iterator(start, limit []byte) Iterator
	Release() error
}
```

### 主要方法说明：

#### Snapshot

创建快照。

```go
func Snapshot() (Snapshot, error)
```

#### SwitchToSnapshot

切换到快照模式。

```go
func SwitchToSnapshot() error
```

#### SwitchToDB

切换回数据库模式。

```go
func SwitchToDB() error
```

### 使用示例：

```go
snapshot, err := db.Snapshot()
if err != nil {
    fmt.Printf("创建快照失败: %v\n", err)
    return
}
defer snapshot.Release()

value, err := snapshot.Get([]byte("key"))
if err != nil {
    fmt.Printf("快照读取失败: %v\n", err)
}
```

---

## 5.6 LevelDB 实现

sfsDb 默认使用 LevelDB 作为存储后端。

### LevelDBStore 结构体

```go
type LevelDBStore struct {
	ldb        LevelDBGetter
	originalDB *leveldb.DB
	isSnapshot bool
	opts       *opt.Options
}
```

### 特点：

- 支持快照功能
- 自动恢复机制
- 高效的读写性能
- 对象池优化

---

## 5.7 DBManager 数据库管理器

DBManager 负责管理数据库实例的创建、打开和关闭。

### 主要函数：

#### GetDBManager

获取数据库管理器实例。

```go
func GetDBManager() *DBManager
```

#### OpenDB

打开存储数据库，使用管理器内部的 db 变量，保证全局唯一实例。

```go
func (dm *DBManager) OpenDB(path string, encryptConfig ...*EncryptionConfig) (Store, error)
```

**示例：**
```go
db, err := storage.GetDBManager().OpenDB("./mydb")
if err != nil {
    fmt.Printf("打开数据库失败: %v\n", err)
    return
}
defer storage.GetDBManager().CloseDB()
```

#### OpenDBWithScenario

根据场景打开存储数据库。

```go
func (dm *DBManager) OpenDBWithScenario(path string, scenario string, encryptConfig ...*EncryptionConfig) (Store, error)
```

**示例：**
```go
db, err := storage.GetDBManager().OpenDBWithScenario("./mydb", storage.ScenarioIoT)
if err != nil {
    fmt.Printf("打开数据库失败: %v\n", err)
    return
}
```

#### CloseDB

关闭数据库。

```go
func (dm *DBManager) CloseDB() error
```

#### GetDB

获取当前数据库实例。

```go
func (dm *DBManager) GetDB() Store
```

#### SetDB

设置数据库实例。

```go
func (dm *DBManager) SetDB(store Store)
```

---

## 5.8 错误定义

Storage 包定义了以下错误类型：

```go
var (
	ErrNotFound     = NewError("key not found")
	ErrInvalidKey   = NewError("invalid key")
	ErrInvalidValue = NewError("invalid value")
	ErrStoreClosed  = NewError("store is closed")
)
```

### Error 类型

```go
type Error struct {
	msg string
}
```

### NewError

创建新的 KV 存储错误。

```go
func NewError(msg string) *Error
```

---

## 5.9 最佳实践

1. **始终使用 DBManager**：通过 GetDBManager() 获取数据库实例，确保全局唯一性
2. **及时关闭数据库**：使用 defer 确保数据库被正确关闭
3. **合理使用批量操作**：对于多个写操作，优先使用 Batch
4. **及时释放迭代器和快照**：使用 defer Release() 确保资源被释放
5. **处理 ErrNotFound**：Get 操作时要检查 ErrNotFound 错误
6. **根据场景选择配置**：使用 OpenDBWithScenario 根据应用场景选择合适的配置

---

## 5.10 小结

Storage 存储包是 sfsDb 的底层存储层，它提供了：

- **统一的 KV 存储接口**：Store 接口定义了标准的存储操作
- **高效的 LevelDB 实现**：基于 LevelDB 的高性能存储后端
- **批量操作支持**：Batch 接口支持高效的批量写入
- **强大的迭代器**：Iterator 接口支持灵活的数据遍历
- **快照功能**：Snapshot 接口支持数据的一致性读取
- **DBManager 管理**：全局数据库管理器，确保实例唯一性

通过 Storage 包，sfsDb 实现了高效、可靠的底层数据存储。

---

**本章版本**：4.0.0
**最后更新**：2026-03-11
