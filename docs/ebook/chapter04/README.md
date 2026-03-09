# 第 4 章：索引与查询优化

索引是数据库性能的关键。本章将学习如何创建和使用各种索引，以及如何优化查询性能。

## 4.1 主键管理

### 4.1.1 单主键

**创建单主键索引：**

```go
package main

import (
    "log"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB("./index_example_db")
    if err != nil {
        log.Fatalf("打开数据库失败: %v", err)
    }
    defer dbManager.CloseAllDB()

    // 创建商品表
    productTable, err := db.CreateTable("products", map[string]string{
        "id":          "int",
        "name":        "string",
        "price":       "float64",
        "category":    "string",
        "stock":       "int",
        "created_at":  "int64",
    })
    if err != nil {
        log.Fatalf("创建商品表失败: %v", err)
    }

    // 添加主键索引（唯一索引）
    productTable.AddIndex("id", true)

    log.Println("商品表和主键索引创建成功")
}
```

**使用主键查询（最快）：**

```go
// 主键查询 - O(1) 复杂度
searchFields := map[string]any{
    "id": 1001,
}

data, err := productTable.Read(&searchFields)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}

log.Printf("查询结果: %s", string(data))
```

### 4.1.2 复合主键

**创建复合主键：**

```go
// 创建订单表，使用复合主键 (user_id, order_id)
orderTable, err := db.CreateTable("orders", map[string]string{
    "user_id":     "int",
    "order_id":    "int",
    "product_id":  "int",
    "quantity":    "int",
    "total_price": "float64",
    "status":      "string",
    "created_at":  "int64",
})
if err != nil {
    log.Fatalf("创建订单表失败: %v", err)
}

// 添加复合主键索引
orderTable.AddIndex("user_id,order_id", true)
```

**使用复合主键查询：**

```go
// 复合主键查询
searchFields := map[string]any{
    "user_id":  1,
    "order_id": 1001,
}

data, err := orderTable.Read(&searchFields)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
```

**复合索引的前缀匹配：**

```go
// 可以只使用复合索引的前缀进行查询
// 查询用户1的所有订单
searchFields := map[string]any{
    "user_id": 1,
}

// 这种查询也能利用复合索引 "user_id,order_id"
```

### 4.1.3 自动增值

**手动管理ID：**

```go
// 简单的ID管理器
type IDGenerator struct {
    nextID int
}

func NewIDGenerator() *IDGenerator {
    return &IDGenerator{nextID: 1}
}

func (g *IDGenerator) NextID() int {
    id := g.nextID
    g.nextID++
    return id
}

// 使用示例
generator := NewIDGenerator()

product := map[string]interface{}{
    "id":    generator.NextID(),
    "name":  "iPhone 15",
    "price": 7999.0,
}

productTable.Insert(&product)
```

## 4.2 普通索引

### 4.2.1 创建普通索引

**创建单字段普通索引：**

```go
// 为商品名称创建普通索引
productTable.AddIndex("name", false)

// 为商品分类创建普通索引
productTable.AddIndex("category", false)

// 为价格创建普通索引
productTable.AddIndex("price", false)
```

**创建复合普通索引：**

```go
// 创建复合索引 (category, price)
// 用于按分类和价格查询
productTable.AddIndex("category,price", false)

// 创建复合索引 (created_at, category)
// 用于按时间和分类查询
productTable.AddIndex("created_at,category", false)
```

**索引命名约定：**
- 单字段索引：字段名
- 复合索引：字段1_字段2_字段3

### 4.2.2 使用普通索引查询

**等值查询：**

```go
import "github.com/liaoran123/sfsDb/util"

// 使用普通索引查询分类为 "手机" 的商品
iter, err := productTable.Search(
    &map[string]any{"category": "手机"},
    util.Equal,
)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
defer iter.Close()

// 遍历结果
for iter.Next() {
    key, value := iter.Key(), iter.Value()
    log.Printf("商品: Key=%s, Value=%s", string(key), string(value))
}
```

**范围查询：**

```go
// 查询价格大于 5000 的商品
iter, err := productTable.Search(
    &map[string]any{"price": 5000.0},
    util.GreaterThan,
)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
defer iter.Close()
```

**使用复合索引：**

```go
// 查询分类为 "手机" 且价格小于 8000 的商品
// 使用复合索引 (category, price)
iter, err := productTable.Search(
    &map[string]any{"category": "手机", "price": 8000.0},
    util.LessOrEqual,
)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
defer iter.Close()
```

### 4.2.3 索引删除

```go
// 删除索引（注意：当前 sfsDb 可能不支持直接删除索引）
// 如果需要删除索引，可能需要重建表

// 重建表的方式：
// 1. 导出数据
// 2. 删除旧表
// 3. 创建新表（不包含要删除的索引）
// 4. 重新导入数据
```

## 4.3 全文索引

### 4.3.1 全文索引原理

**全文索引 vs 普通索引：**

| 特性 | 普通索引 | 全文索引 |
|------|---------|---------|
| 查询方式 | 精确匹配、范围匹配 | 关键词搜索 |
| 适用场景 | 分类、价格、日期等 | 文章标题、描述、评论 |
| 性能 | 高 | 中等 |
| 实现复杂度 | 低 | 高 |

**全文索引核心概念：**
- **分词**：将文本切分成词语
- **倒排索引**：词语 -> 文档列表
- **相关性评分**：计算匹配程度

### 4.3.2 创建全文索引

**注意：** sfsDb 的全文索引功能可能在特定版本中提供，请参考官方文档。

```go
// 假设的全文索引创建方式（示例）
articleTable, err := db.CreateTable("articles", map[string]string{
    "id":      "int",
    "title":   "string",
    "content": "string",
    "author":  "string",
})
if err != nil {
    log.Fatal(err)
}

// 添加主键
articleTable.AddIndex("id", true)

// （如果支持）添加全文索引
// articleTable.AddFullTextIndex("title,content")
```

### 4.3.3 全文搜索技巧

**替代方案：** 如果 sfsDb 不支持原生全文索引，可以考虑：

```go
// 方案1：使用关键词标签
// 在表中添加 keyword_tags 字段
// 存储分词后的关键词

// 方案2：结合外部搜索引擎
// 使用 Elasticsearch、Bleve 等
// sfsDb 存储原始数据，搜索引擎存储索引

// 方案3：简单的模糊查询
// 使用 Like 操作（如果支持）
```

## 4.4 查询优化

### 4.4.1 索引选择策略

**索引创建原则：**

1. **为 WHERE 条件字段创建索引**
   ```go
   // 经常按 category 查询 → 为 category 创建索引
   productTable.AddIndex("category", false)
   ```

2. **为 JOIN 字段创建索引**
   ```go
   // 经常按 user_id 关联查询 → 为 user_id 创建索引
   orderTable.AddIndex("user_id", false)
   ```

3. **为 ORDER BY 字段创建索引**
   ```go
   // 经常按 created_at 排序 → 为 created_at 创建索引
   productTable.AddIndex("created_at", false)
   ```

4. **复合索引的字段顺序很重要**
   ```go
   // 查询模式：WHERE category = ? AND price < ?
   // 索引顺序：category, price（正确）
   productTable.AddIndex("category,price", false)
   
   // 索引顺序：price, category（不能充分利用）
   ```

**索引 Cardinality（基数）：**

| 字段 | 基数 | 是否适合建索引 |
|------|------|---------------|
| id | 很高（唯一） | ✅ 非常适合 |
| category | 中等 | ✅ 适合 |
| status | 很低（只有几个值） | ❌ 不适合 |
| created_at | 很高 | ✅ 适合 |

### 4.4.2 查询条件优化

**避免全表扫描：**

```go
// ❌ 不好的做法：没有索引，全表扫描
// 如果没有为 price 创建索引
iter, err := productTable.Search(
    &map[string]any{"price": 100.0},
    util.GreaterThan,
)

// ✅ 好的做法：先创建索引，再查询
productTable.AddIndex("price", false)

iter, err := productTable.Search(
    &map[string]any{"price": 100.0},
    util.GreaterThan,
)
```

**优化查询范围：**

```go
// 查询 2024年1月的订单
// 方式1：分别查询年和月（如果有这两个字段）

// 方式2：使用时间戳范围查询（推荐）
startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local).Unix()
endTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local).Unix()

// 查询 createdAt >= startTime AND createdAt < endTime
```

**使用最左前缀原则：**

```go
// 复合索引 (a, b, c)
// 可以使用的查询模式：
// - a = ?
// - a = ? AND b = ?
// - a = ? AND b = ? AND c = ?

// 不能使用的查询模式：
// - b = ?
// - c = ?
// - b = ? AND c = ?
```

### 4.4.3 性能基准测试

**创建测试数据：**

```go
// 插入 10000 条测试数据
tx, err := transactionLockFree.NewTableTransaction(productTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

categories := []string{"手机", "电脑", "平板", "耳机", "充电器"}

for i := 0; i < 10000; i++ {
    product := map[string]interface{}{
        "id":         i + 1,
        "name":        fmt.Sprintf("商品%d", i+1),
        "price":       float64(100 + rand.Intn(10000)),
        "category":    categories[rand.Intn(len(categories))],
        "stock":       rand.Intn(1000),
        "created_at":  time.Now().Unix() - int64(rand.Intn(86400*30)),
    }
    _, err := tx.Insert(&product)
    if err != nil {
        log.Fatal(err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

**性能测试：**

```go
// 测试主键查询性能
start := time.Now()
for i := 0; i < 1000; i++ {
    searchFields := map[string]any{"id": rand.Intn(10000) + 1}
    _, err := productTable.Read(&searchFields)
    if err != nil {
        log.Printf("查询失败: %v", err)
    }
}
duration := time.Since(start)
log.Printf("主键查询 1000 次耗时: %v, 平均: %v", 
    duration, duration/1000)

// 测试普通索引查询性能
start = time.Now()
for i := 0; i < 1000; i++ {
    category := categories[rand.Intn(len(categories))]
    iter, err := productTable.Search(
        &map[string]any{"category": category},
        util.Equal,
    )
    if err != nil {
        log.Printf("查询失败: %v", err)
        continue
    }
    iter.Close()
}
duration = time.Since(start)
log.Printf("普通索引查询 1000 次耗时: %v, 平均: %v", 
    duration, duration/1000)
```

## 4.5 实战示例

### 4.5.1 电商商品查询优化

**场景：** 电商网站的商品列表页，支持：
- 按分类筛选
- 按价格范围筛选
- 按创建时间排序
- 分页查询

**优化方案：**

```go
package main

import (
    "fmt"
    "log"
    "math/rand"
    "time"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/util"
)

type ProductSearcher struct {
    productTable *engine.Table
}

func NewProductSearcher(db *storage.DB) (*ProductSearcher, error) {
    productTable, err := db.CreateTable("products", map[string]string{
        "id":          "int",
        "name":        "string",
        "price":       "float64",
        "category":    "string",
        "stock":       "int",
        "created_at":  "int64",
    })
    if err != nil {
        return nil, err
    }

    // 主键索引
    productTable.AddIndex("id", true)
    
    // 分类索引
    productTable.AddIndex("category", false)
    
    // 价格索引
    productTable.AddIndex("price", false)
    
    // 时间索引
    productTable.AddIndex("created_at", false)
    
    // 复合索引：分类+价格
    productTable.AddIndex("category,price", false)
    
    // 复合索引：分类+时间
    productTable.AddIndex("category,created_at", false)

    return &ProductSearcher{productTable: productTable}, nil
}

func (ps *ProductSearcher) SearchByCategory(category string, page, pageSize int) ([]map[string]interface{}, error) {
    iter, err := ps.productTable.Search(
        &map[string]any{"category": category},
        util.Equal,
    )
    if err != nil {
        return nil, err
    }
    defer iter.Close()

    var results []map[string]interface{}
    count := 0
    skip := (page - 1) * pageSize

    for iter.Next() {
        if count < skip {
            count++
            continue
        }
        if count >= skip+pageSize {
            break
        }
        
        // 解析数据...
        // results = append(results, parsedData)
        count++
    }

    return results, nil
}

func (ps *ProductSearcher) SearchByPriceRange(
    category string,
    minPrice, maxPrice float64,
    page, pageSize int,
) ([]map[string]interface{}, error) {
    // 使用复合索引 category,price
    // 先筛选分类，再筛选价格
    // 具体实现取决于 sfsDb 的查询能力
    
    return nil, nil
}

func (ps *ProductSearcher) SearchLatest(
    category string,
    limit int,
) ([]map[string]interface{}, error) {
    // 使用复合索引 category,created_at
    // 按时间倒序查询
    
    return nil, nil
}

func main() {
    rand.Seed(time.Now().UnixNano())

    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB("./ecommerce_db")
    if err != nil {
        log.Fatal(err)
    }
    defer dbManager.CloseAllDB()

    searcher, err := NewProductSearcher(db)
    if err != nil {
        log.Fatal(err)
    }

    // 查询手机分类的商品，第1页，每页10条
    products, err := searcher.SearchByCategory("手机", 1, 10)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("查询到 %d 个商品\n", len(products))
}
```

### 4.5.2 日志系统索引设计

**场景：** 系统日志存储和查询
- 按时间范围查询
- 按日志级别查询
- 按模块查询

**索引设计：**

```go
logTable, err := db.CreateTable("logs", map[string]string{
    "id":         "int",
    "timestamp":  "int64",
    "level":      "string",  // DEBUG, INFO, WARN, ERROR
    "module":     "string",  // user, order, payment
    "message":    "string",
    "metadata":   "string",  // JSON 格式的元数据
})
if err != nil {
    log.Fatal(err)
}

// 主键
logTable.AddIndex("id", true)

// 时间索引（最重要，按时间查询最多）
logTable.AddIndex("timestamp", false)

// 复合索引：时间 + 级别
logTable.AddIndex("timestamp,level", false)

// 复合索引：时间 + 模块
logTable.AddIndex("timestamp,module", false)

// 复合索引：时间 + 模块 + 级别
logTable.AddIndex("timestamp,module,level", false)
```

**查询示例：**

```go
// 查询最近1小时的ERROR日志
oneHourAgo := time.Now().Add(-time.Hour).Unix()

iter, err := logTable.Search(
    &map[string]any{
        "timestamp": oneHourAgo,
        "level":     "ERROR",
    },
    util.GreaterOrEqual,
)
```

## 4.6 本章小结

本章学习了 sfsDb 的索引与查询优化，包括：

✅ **主键管理**：单主键、复合主键、自动增值
✅ **普通索引**：创建索引、使用索引查询
✅ **查询优化**：索引选择策略、查询条件优化
✅ **性能基准测试**：如何测试和比较查询性能
✅ **实战示例**：电商商品查询、日志系统索引设计

**关键要点：**
1. 索引是提升查询性能的最有效手段
2. 合理选择索引字段（WHERE、JOIN、ORDER BY）
3. 复合索引的字段顺序很重要（最左前缀原则）
4. 定期进行性能测试，验证索引效果
5. 不要过度索引，索引会增加写入开销

下一章我们已经学习了事务处理，接下来将学习时序数据处理，这是工业物联网场景中的关键特性。
