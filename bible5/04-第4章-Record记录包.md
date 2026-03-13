# 第 4 章：Record 记录包

---

## 4.1 完整的可运行示例

让我们通过一个完整的示例来演示 sfsDb 的 Record 记录包功能：

```go
package main

import (
	"fmt"
	"os"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("=== sfsDb 数据库实战 - 第 4 章：Record记录包示例 ===")
	fmt.Println()

	dbPath := "./bible_examples_chapter04_db"
	os.RemoveAll(dbPath)
	defer os.RemoveAll(dbPath)

	fmt.Println("1. 初始化数据库")
	_, err := storage.GetDBManager().OpenDB(dbPath)
	if err != nil {
		fmt.Printf("数据库打开失败: %v\n", err)
		return
	}
	defer storage.GetDBManager().CloseDB()
	fmt.Println("   ✓ 数据库打开成功")
	fmt.Println()

	fmt.Println("2. 创建产品表")
	productTable, err := engine.TableNew("products")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	fields := map[string]any{
		"id":       0,
		"name":     "",
		"price":    0.0,
		"quantity": 0,
		"category": "",
	}
	productTable.SetFields(fields)

	pk, _ := engine.DefaultPrimaryKeyNew("pk_id")
	pk.AddFields("id")
	productTable.CreateIndex(pk)
	fmt.Println("   ✓ 产品表创建成功")
	fmt.Println()

	fmt.Println("3. 插入测试数据")
	products := []map[string]any{
		{"id": 1, "name": "笔记本电脑", "price": 5999.0, "quantity": 50, "category": "电子产品"},
		{"id": 2, "name": "无线鼠标", "price": 129.0, "quantity": 200, "category": "电子产品"},
		{"id": 3, "name": "机械键盘", "price": 399.0, "quantity": 100, "category": "电子产品"},
		{"id": 4, "name": "显示器", "price": 1599.0, "quantity": 30, "category": "电子产品"},
		{"id": 5, "name": "USB数据线", "price": 19.9, "quantity": 500, "category": "配件"},
	}

	for _, p := range products {
		_, err = productTable.Insert(&p)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
	}
	fmt.Println("   ✓ 成功插入", len(products), "条产品数据")
	fmt.Println()

	fmt.Println("4. 查询所有产品并转换为Records")
	iter := productTable.Iter()
	records := record.GetRecords()
	defer records.Release()

	for iter.Next() {
		r := record.GetRecord()
		row := iter.Row()
		for k, v := range *row {
			r[k] = v
		}
		records = append(records, r)
	}
	iter.Release()
	fmt.Println("   ✓ 查询到", len(records), "条记录")
	fmt.Println()

	fmt.Println("5. 使用Select方法选择特定字段")
	selectedRecords := records.Select("id", "name", "price")
	defer selectedRecords.Release()
	fmt.Println("   选择后的记录:")
	for i, r := range selectedRecords {
		fmt.Printf("   记录%d: id=%v, name=%v, price=%v\n", i+1, r["id"], r["name"], r["price"])
	}
	fmt.Println()

	fmt.Println("6. 使用Operation进行字段运算")
	addOp := record.NewAddOperation([]string{"price", "quantity"}, "total_value")
	operationRecords := records.Operation(addOp)
	defer operationRecords.Release()
	fmt.Println("   运算后的记录（添加total_value字段）:")
	for i, r := range operationRecords {
		fmt.Printf("   记录%d: name=%v, price=%v, quantity=%v, total_value=%v\n",
			i+1, r["name"], r["price"], r["quantity"], r["total_value"])
	}
	fmt.Println()

	fmt.Println("7. 演示对象池的使用")
	fmt.Println("   ✓ Record和Records对象池可以显著减少内存分配")
	fmt.Println("   ✓ 使用GetRecord()/GetRecords()从池获取对象")
	fmt.Println("   ✓ 使用Release()将对象放回池中复用")
	fmt.Println()

	fmt.Println("=== 示例运行完成 ===")
}
```

---

## 4.2 Record 数据结构

Record 是 sfsDb 中表示单条记录的数据结构，它本质上是一个 `map[string]any` 类型。

```go
type Record map[string]any
```

### 主要方法：

#### Select 方法

从记录中选择指定的字段，返回一个新的记录。

```go
func (r Record) Select(keys ...string) Record
```

**示例：**
```go
r := record.GetRecord()
r["id"] = 1
r["name"] = "Alice"
r["age"] = 25

selected := r.Select("id", "name")
// selected 包含 {"id": 1, "name": "Alice"}
```

#### Release 方法

将 Record 对象放回对象池，实现对象复用。

```go
func (r Record) Release()
```

**示例：**
```go
r := record.GetRecord()
defer r.Release()
// 使用 r...
```

---

## 4.3 Records 记录集合

Records 是 Record 的切片类型，表示多条记录的集合。

```go
type Records []Record
```

### 主要方法：

#### Select 方法

从记录集合中选择指定的字段，返回新的记录集合。

```go
func (rs Records) Select(fields ...string) Records
```

**示例：**
```go
records := record.GetRecords()
// 添加记录...

selected := records.Select("id", "name")
// selected 只包含 id 和 name 字段
```

#### Operation 方法

对记录集合执行字段运算，生成新的字段。

```go
func (rs Records) Operation(op ...Operation) Records
```

**示例：**
```go
addOp := record.NewAddOperation([]string{"price", "quantity"}, "total")
result := records.Operation(addOp)
// result 中的每条记录都包含 total 字段
```

#### OperationVertical 方法

对记录集合执行垂直操作（基于整个集合计算）。

```go
func (rs Records) OperationVertical(op ...VerticalOperation) Records
```

#### 集合操作方法

- **Intersect**: 计算记录集合的交集
- **Union**: 计算记录集合的并集
- **Difference**: 计算记录集合的差集
- **Contains**: 检查记录是否包含在集合中

**示例：**
```go
result1 := records1.Intersect(records2)
result2 := records1.Union(records2)
result3 := records1.Difference(records2)
exists := records1.Contains(someRecord)
```

#### Release 方法

将 Records 对象放回对象池。

```go
func (rs Records) Release()
```

---

## 4.4 对象池优化

sfsDb 使用对象池来优化 Record 和 Records 的内存分配，显著提升性能。

### 核心函数：

#### GetRecord

从对象池获取一个 Record 对象。

```go
func GetRecord() Record
```

#### GetRecordWithCapacity

从对象池获取一个指定初始容量的 Record 对象。

```go
func GetRecordWithCapacity(capacity int) Record
```

#### PutRecord

将 Record 对象放回对象池。

```go
func PutRecord(r Record)
```

#### GetRecords

从对象池获取一个 Records 对象。

```go
func GetRecords() Records
```

#### GetRecordsWithCapacity

从对象池获取一个指定初始容量的 Records 对象。

```go
func GetRecordsWithCapacity(capacity int) Records
```

#### PutRecords

将 Records 对象放回对象池。

```go
func PutRecords(rs Records)
```

### 最佳实践：

```go
// 使用 defer 确保对象被放回池中
r := record.GetRecord()
defer r.Release()

records := record.GetRecords()
defer records.Release()

// 执行操作...
```

---

## 4.5 记录操作

Operation 接口定义了对记录进行各种运算的能力。

### Operation 接口：

```go
type Operation interface {
	Evaluate(record map[string]any) any
	NewField() string
}
```

### 常用运算实现：

sfsDb 提供了 CommonOperation 结构体，支持多种常用字段运算：

#### 加法运算

```go
func NewAddOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewAddOperation([]string{"price", "tax"}, "total")
```

#### 减法运算

```go
func NewSubOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewSubOperation([]string{"revenue", "cost"}, "profit")
```

#### 乘法运算

```go
func NewMulOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewMulOperation([]string{"price", "quantity"}, "revenue")
```

#### 除法运算

```go
func NewDivOperation(fields []string, newField string, defaultDivisor float64) *CommonOperation
```

**示例：**
```go
op := record.NewDivOperation([]string{"total", "count"}, "average", 1.0)
```

#### 平均值运算

```go
func NewAvgOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewAvgOperation([]string{"score1", "score2", "score3"}, "average")
```

#### 最大值运算

```go
func NewMaxOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewMaxOperation([]string{"price1", "price2", "price3"}, "max_price")
```

#### 最小值运算

```go
func NewMinOperation(fields []string, newField string) *CommonOperation
```

**示例：**
```go
op := record.NewMinOperation([]string{"price1", "price2", "price3"}, "min_price")
```

#### 字符串连接运算

```go
func NewConcatOperation(fields []string, newField string, separator string) *CommonOperation
```

**示例：**
```go
op := record.NewConcatOperation([]string{"first_name", "last_name"}, "full_name", " ")
```

---

## 4.6 垂直操作

垂直操作是基于整个记录集合进行计算的操作。

### VerticalOperation 接口：

```go
type VerticalOperation interface {
	Evaluate(records Records) any
	NewField() string
}
```

垂直操作会计算整个集合的值，然后将结果添加到每条记录中。

---

## 4.7 批量操作函数

### BatchSelect

批量选择多个记录的字段，减少多次调用 Select 方法的开销。

```go
func BatchSelect(records Records, fields ...string) Records
```

### BatchOperation

批量对多个记录执行操作，减少多次调用 Operation 方法的开销。

```go
func BatchOperation(records Records, op ...Operation) Records
```

### BatchOperationVertical

批量对多个记录执行垂直操作，减少多次调用 OperationVertical 方法的开销。

```go
func BatchOperationVertical(records Records, op ...VerticalOperation) Records
```

---

## 4.8 性能优化建议

1. **始终使用对象池**：通过 GetRecord/GetRecords 获取对象，使用 Release 放回池中
2. **合理使用预分配容量**：使用 GetRecordWithCapacity 和 GetRecordsWithCapacity
3. **及时释放对象**：使用 defer 确保对象被正确释放
4. **使用批量操作**：对于多条记录的操作，优先使用 Batch* 函数
5. **避免不必要的字段复制**：只选择需要的字段

---

## 4.9 小结

Record 记录包是 sfsDb 的核心数据结构之一，它提供了：

- **灵活的 Record 类型**：基于 map[string]any 的动态数据结构
- **强大的 Records 集合**：支持选择、运算、集合操作等
- **高效的对象池**：显著减少内存分配，提升性能
- **丰富的运算接口**：支持多种字段运算和垂直操作

通过合理使用 Record 包，可以高效地处理和操作 sfsDb 中的数据。

---

**本章版本**：5.0.0
**最后更新**：2026-03-12
