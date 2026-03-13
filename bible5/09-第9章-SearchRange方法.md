# 第 7 章：SearchRange 方法

## 7.1 概述

在数据库查询中，范围搜索是一种常见的操作，SQL 中的 `BETWEEN` 语句是实现此类查询的标准方式。然而，在 sfsDb 数据库引擎中，`SearchRange` 方法提供了一种更高效、更灵活的范围搜索解决方案。

## 7.2 基本功能与设计理念

### 核心功能

`SearchRange` 方法是 sfsDb 中 `Table` 结构体的一个方法，用于执行高效的范围搜索操作：

```go
func (t *Table) SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*TableIter, error)
```

该方法接收三个参数：
- `funIter`：迭代器函数，用于遍历存储引擎中的数据
- `Start`：范围起始条件，是一个键值对映射
- `Limit`：范围结束条件，也是一个键值对映射

返回值是一个 `TableIter` 迭代器，用于遍历查询结果。

### 设计理念

`SearchRange` 方法的设计理念基于以下几点：
1. **基于索引的高效搜索**：利用表的索引结构快速定位范围边界
2. **灵活的范围定义**：支持多字段组合的范围查询
3. **空值处理机制**：通过 `nil` 值表示无边界（从最小值开始或到最大值结束）
4. **资源复用**：使用对象池管理临时资源，减少内存分配
5. **可定制的迭代器**：允许用户自定义迭代器行为

**代码来源**：`docs/zh/advanced/search_range.md`

## 7.3 实现原理与技术细节

### 核心实现流程

`SearchRange` 方法的核心实现逻辑封装在 `RangeForAny` 方法中，主要步骤如下：
1. **参数验证**：检查 `Start` 和 `Limit` 是否为 `nil`，并验证它们的键是否匹配
2. **索引匹配**：根据查询字段匹配合适的索引
3. **键值转换**：将查询条件转换为字节数组，用于索引查找
4. **范围计算**：根据转换后的键值计算搜索范围
5. **迭代器创建**：使用计算出的范围创建迭代器
6. **结果封装**：将迭代器封装为 `TableIter` 返回

### 关键技术点

#### 索引利用

`SearchRange` 方法会自动匹配最适合的索引，这是其性能优于 SQL BETWEEN 查询的关键因素之一：

```go
fieldname := GetStringSlice()
defer PutStringSlice(fieldname)
for k := range *Start {
    fieldname = append(fieldname, k)
}
idx := t.MatchIndexCached(fieldname)
if idx == nil {
    return nil, nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", fieldname, t.name)
}
```

#### 智能边界处理

当 `Limit` 的最后一个字段值为 `nil` 时，方法会使用前缀的下一个字节作为上限，表示到无穷大：

```go
if (*Limit)[fieldname[len(fieldname)-1]] == nil { // 当Limit最后一个字段值为nil时，使用前缀的下一个字节作为上限，表示到无穷大
    slice.Limit = util.BytesPrefix(pfx).Limit
}
```

#### 资源管理

方法使用对象池管理临时资源，减少内存分配和垃圾回收开销：

```go
fieldname := GetStringSlice()
defer PutStringSlice(fieldname)
// ...
fieldsBytes = t.FieldsToBytesNil(Start)
// ...
defer func() {
    if fieldsBytes != nil && *fieldsBytes != nil {
        GlobalFieldsBytesPool.Put(*fieldsBytes)
    }
}()
```

**代码来源**：`docs/zh/advanced/search_range.md`

## 7.4 与 SQL BETWEEN 查询的对比

| 特性 | SQL BETWEEN 查询 | sfsDb SearchRange 方法 |
|------|----------------|----------------------|
| 索引利用 | 依赖查询优化器选择索引 | 自动匹配最适合的索引 |
| 多字段范围 | 语法复杂，性能可能下降 | 原生支持，性能稳定 |
| 无边界查询 | 需要特殊处理（如使用 MIN/MAX） | 原生支持，通过 nil 值表示 |
| 自定义迭代 | 不支持 | 支持自定义迭代器函数 |
| 资源管理 | 由数据库引擎管理 | 显式资源池管理，减少开销 |
| 性能表现 | 受查询复杂度和数据量影响较大 | 始终保持高效，即使在大数据集上 |

**代码来源**：`docs/zh/advanced/search_range.md`

## 7.5 实际应用场景

### 1. 时序数据查询

`SearchRange` 方法特别适合处理时序数据，如传感器数据、日志数据等：

```go
// 查询某个时间范围内的传感器数据
startTime := time.Now().Add(-24 * time.Hour).Unix()
endTime := time.Now().Unix()

iter, err := table.SearchRange(nil, 
    &map[string]any{"timestamp": startTime}, 
    &map[string]any{"timestamp": endTime})
```

### 2. 多维度范围查询

支持多字段组合的范围查询，适用于复杂的业务场景：

```go
// 查询特定用户在特定时间段内的订单
iter, err := table.SearchRange(nil, 
    &map[string]any{"user_id": 123, "order_time": startTime}, 
    &map[string]any{"user_id": 123, "order_time": endTime})
```

### 3. 无边界查询

通过 `nil` 值实现无边界查询，简化代码：

```go
// 查询所有大于等于某个值的记录
iter, err := table.SearchRange(nil, 
    &map[string]any{"score": 90}, 
    &map[string]any{"score": nil})

// 查询所有记录（全表扫描）
iter, err := table.SearchRange(nil, 
    &map[string]any{"id": nil}, 
    &map[string]any{"id": nil})
```

**代码来源**：`docs/zh/advanced/search_range.md`

## 7.6 使用示例

### 基本使用示例

```go
// 创建表和索引
table, _ := TableNew("sensor_data")
fields := map[string]any{
    "timestamp": 0,
    "sensor_id": "",
    "value":     0.0,
}
table.SetFields(fields)

// 创建索引
idx, _ := DefaultPrimaryKeyNew("pk")
idx.AddFields("timestamp")
table.CreateIndex(idx)

// 执行范围搜索
start := map[string]any{"timestamp": 1609459200} // 2021-01-01 00:00:00
end := map[string]any{"timestamp": 1612137600}   // 2021-02-01 00:00:00

iter, err := table.SearchRange(nil, &start, &end)
if err != nil {
    // 处理错误
}
defer iter.Release()

// 遍历结果
records := iter.GetRecords(true)
defer records.Release()

for _, record := range records {
    fmt.Printf("时间戳: %d, 传感器ID: %s, 值: %f\n", 
        record["timestamp"], record["sensor_id"], record["value"])
}
```

**代码来源**：`docs/zh/advanced/search_range.md`、`engine/tableSearch_range_test.go`

### 完整测试示例

```go
package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTable_SearchRange 测试时序数据的区间搜索功能
func TestTable_SearchRange(t *testing.T) {
	// 清理旧数据
	tableName := "test_search_range_" + t.Name()
	defer os.RemoveAll("./test_search_range_db")

	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./test_search_range_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.GetDBManager().CloseDB()

	// 创建表，使用时间戳作为主键
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段，timestamp作为主键（int）
	fields := map[string]any{
		"timestamp": 0,   // 时间戳主键
		"value":     0.0, // 传感器值
		"sensor_id": "",  // 传感器ID
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("timestamp")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成测试数据（时序数据）
	now := int(time.Now().Unix())
	testData := make([]map[string]any, 100)

	for i := 0; i < 100; i++ {
		testData[i] = map[string]any{
			"timestamp": now + i,
			"value":     float64(i * 10),
			"sensor_id": fmt.Sprintf("sensor_%d", i%10),
		}
	}

	// 插入测试数据
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试1: 完全匹配的范围查询
	t.Run("ExactMatchRangeSearch", func(t *testing.T) {
		// 创建测试点
		testPoint := now + 15

		// 定义迭代器函数
		funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
			// 使用存储引擎的Iterator方法创建范围迭代器
			return table.kvStore.Iterator(start, limit)
		})

		// 执行完全匹配的范围查询
		iter, err := table.SearchRange(funIter, &map[string]any{"timestamp": testPoint}, &map[string]any{"timestamp": testPoint})
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果数量
		expectedCount := 1 // 只返回完全匹配的记录
		if len(records) != expectedCount {
			t.Errorf("期望 %d 条记录，实际得到 %d 条", expectedCount, len(records))
		}

		// 验证结果
		for _, record := range records {
			timestamp := record["timestamp"].(int)
			fmt.Printf("完全匹配测试 timestamp: %v\n", timestamp)
			if timestamp != testPoint {
				t.Errorf("记录时间戳 %d 不是期望的值 %d", timestamp, testPoint)
			}
		}
	})
}
```

**代码来源**：`engine/tableSearch_range_test.go`