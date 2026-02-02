# 索引管理

## 4.1 索引接口概述

sfsDb 定义了一套完整的索引接口体系，为不同类型的索引提供了统一的操作方法。

### 4.1.1 基础索引接口

所有索引类型都实现了基础 `Index` 接口，定义了索引的核心方法：

```go
// 基础索引接口，定义所有索引类型共有的方法
type Index interface {
    // 添加索引字段
    AddFields(field ...string)
    // 获取索引字段列表
    GetFields() []string
    Len() int
    setId(id uint8)
    GetId() uint8
    Name() string
    SetName(name string) error
    //修改索引字段名称
    UpdateFields(oldfields string, newfields string)
    //删除索引字段
    DeleteFields(field ...string)
    //拼接前缀
    Prefix(tbid uint8) []byte
    //拼接值，不需要前缀
    Join(fieldsBytes *map[string][]byte) []byte
    //拼接前缀+值
    JoinPrefix(tbid uint8, val []byte) []byte
    // 拼接索引前缀+索引值，调用JoinPrefix，Join方法
    JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
    // 匹配索引字段
    MatchFields(fields ...string) bool
}
```

### 4.1.2 索引类型层次结构

sfsDb 支持三种主要索引类型，每种类型都有对应的接口和实现：

1. **主键索引 (PrimaryKey)** - 用于唯一标识记录
2. **普通索引 (NormalIndex)** - 用于加速查询
3. **全文索引 (FullTextIndex)** - 用于文本搜索

## 4.2 主键索引

### 4.2.1 主键索引接口

```go
// 主键接口，嵌入基础索引接口
type PrimaryKey interface {
    Index
    // 设置主键ID
    GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte
    GetfieldTypeLen(tablefields *map[string]any) *map[string]uint8
    Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
}
```

### 4.2.2 默认主键索引实现

```go
// 默认主键索引
// 组合主键时只支持固定长度的类型的组合。
// 字符串类型，必须指定长度。否则无法解析或存在转义问题导致bug。
type DefaultPrimaryKey struct {
    BaseIndex // 嵌入基础索引
}

func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error)
```

### 4.2.3 使用示例

```go
// 创建单主键表
userTable, err := engine.TableNew("user_table")
if err != nil {
    panic(err)
}

// 设置字段
userFields := map[string]any{
    "id":   0,     // 主键字段
    "name": "",    // 普通字段
    "age":  0,     // 普通字段
}
err = userTable.SetFields(userFields)
if err != nil {
    panic(err)
}

// 创建单主键索引
pk, err := engine.DefaultPrimaryKeyNew("pk_id")
err = pk.AddFields("id")
err = userTable.CreateIndex(pk)
if err != nil {
    panic(err)
}

// 创建复合主键索引
compositePk, err := engine.DefaultPrimaryKeyNew("pk_order_user")
if err != nil {
    panic(err)
}
// 添加多个主键字段
compositePk.AddFields("order_id", "user_id")
err = orderTable.CreateIndex(compositePk)
if err != nil {
    panic(err)
}
```

## 4.3 普通索引

### 4.3.1 普通索引接口

```go
// 普通索引接口，嵌入基础索引接口
type NormalIndex interface {
    Index
    // 将索引的value转换为主键map值

    // 由于NormalIndex完全匹配index接口，所以需要一个Tag方法来区别是否是二级索引。
    Tag() bool
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}
```

### 4.3.2 默认普通索引实现

```go
// 默认普通索引，二级索引
type DefaultNormalIndex struct {
    BaseIndex // 嵌入基础索引
}

func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error)

// Tag方法返回true，表示是二级索引。
func (dni *DefaultNormalIndex) Tag() bool
```

### 4.3.3 使用示例

```go
// 创建普通索引
normalIndex, err := engine.DefaultNormalIndexNew("idx_name")
if err != nil {
    panic(err)
}
// 添加索引字段
normalIndex.AddFields("name")
err = table.CreateIndex(normalIndex)
if err != nil {
    panic(err)
}

// 创建复合索引
compositeIndex, err := engine.DefaultNormalIndexNew("idx_name_age")
if err != nil {
    panic(err)
}
// 添加多个索引字段
compositeIndex.AddFields("name", "age")
err = table.CreateIndex(compositeIndex)
if err != nil {
    panic(err)
}
```

## 4.4 全文索引

### 4.4.1 全文索引接口

```go
// 全文索引接口，嵌入基础索引接口
type FullTextIndex interface {
    Index
    SetFullField(field string, len int) error
    //GetFtlen() int
    // 拼接全文索引值
    JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte
    // 分词方法
    Tokenize(nr string, ftlen int) (tokens []string)
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}
```

### 4.4.2 默认全文索引实现

```go
// 默认全文索引
type DefaultFullTextIndex struct {
    BaseIndex // 嵌入基础索引
    //全文索引切分字段
    ftsplit string
    //切分长度
    ftlen int
}

func DefaultFullTextIndexNew(name string) (*DefaultFullTextIndex, error)
```

### 4.4.3 使用示例

```go
// 创建全文索引
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// 添加字段，最后一个字段必须是主键
// 全文索引必须全量包括主键（单主键或组合主键），一般在后面。否则，会导致索引无唯一性。
fullTextIndex.AddFields("description", "id")
或
fullTextIndex.AddFields("description", "id", "did") //"id"，"did"是组合主键
// 设置全文索引字段和长度
fullTextIndex.SetFullField("description", 5) // 5表示全文索引的长度。中文一般选择5或7。主要是支持象形文字。非象形文字该全文索引不太合适。
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

## 4.5 示例：创建包含多种索引的表

```go
// 创建表
indexDemoTable, err := engine.TableNew("index_demo")
if err != nil {
    panic(err)
}

// 设置字段
indexFields := map[string]any{
    "id":    0,
    "name":  "",
    "age":   0,
    "email": "",
    "city":  "",
}
err = indexDemoTable.SetFields(indexFields)
if err != nil {
    panic(err)
}

// 1. 创建主键索引
primaryKey, err := engine.DefaultPrimaryKeyNew("pk_id")
primaryKey.AddFields("id")
err = indexDemoTable.CreateIndex(primaryKey)
if err != nil {
    panic(err)
}

// 2. 创建普通索引
emailIndex, err := engine.DefaultNormalIndexNew("idx_email")
emailIndex.AddFields("email")
err = indexDemoTable.CreateIndex(emailIndex)
if err != nil {
    panic(err)
}

// 3. 创建复合索引
cityAgeIndex, err := engine.DefaultNormalIndexNew("idx_city_age")
cityAgeIndex.AddFields("city", "age")
err = indexDemoTable.CreateIndex(cityAgeIndex)
if err != nil {
    panic(err)
}
```

## 4.6 索引实现细节

### 4.6.1 基础索引结构

所有索引类型都嵌入了 `BaseIndex` 结构体，提供了通用的索引功能：

```go
// 基础索引结构体，包含所有索引类型共有的字段和方法
type BaseIndex struct {
    fields []string
    id     uint8
    name   string
}
```

### 4.6.2 索引字段处理

- **添加字段**: `AddFields(field ...string)` - 向索引添加一个或多个字段
- **获取字段**: `GetFields() []string` - 获取索引的所有字段
- **更新字段名称**: `UpdateFields(oldfields string, newfields string)` - 修改索引字段名称
- **删除字段**: `DeleteFields(field ...string)` - 从索引中删除字段

### 4.6.3 索引值拼接

索引内部使用字节数组来存储索引值，提供了多种拼接方法：

- **Prefix**: 拼接索引前缀
- **Join**: 拼接字段值
- **JoinPrefix**: 拼接前缀和值
- **JoinValue**: 拼接完整的索引值

### 4.6.4 复合索引注意事项

1. **组合主键限制**: 组合主键只支持固定长度类型的组合，字符串类型必须指定长度
2. **字段顺序**: 索引字段的顺序会影响查询性能，应将最常用的字段放在前面
3. **索引大小**: 索引会增加存储开销，应只创建必要的索引

## 4.7 索引管理系统

### 4.7.1 Indexs 结构体

`Indexs` 是 sfsDb 的索引管理核心结构体，负责管理表的所有索引：

```go
// 索引管理结构
type Indexs struct {
    id     uint8
    indexs []Index
    fields *map[string]any //表字段
}
```

**主要功能**：
- 创建索引（检查字段存在性、索引名唯一性等）
- 删除索引
- 匹配索引（优先匹配主键索引，再匹配普通索引，最后匹配全文索引）
- 管理索引的生命周期

### 4.7.2 索引类型限制

#### 4.7.2.1 不支持布尔类型作为索引

sfsDb **不支持使用布尔类型（bool）作为索引字段**，原因如下：

- **索引键冲突**：布尔类型只有两个可能的值（true/false），会导致多个记录生成相同的索引键
- **记录覆盖**：在KV存储中，相同的索引键会被覆盖，只保留最后插入的记录
- **查询不准确**：由于索引键冲突，查询时只能找到最后插入的记录，而不是所有符合条件的记录
- **性能提升有限**：布尔字段的选择性通常不高，索引带来的性能提升有限

#### 4.7.2.2 替代方案

**方案1：使用主键索引过滤**

```go
// 1. 使用主键索引获取所有记录
searchData := map[string]any{"id": nil}
iter := table.Search(&searchData)
if iter != nil {
    defer GlobalTableIterPool.Put( iter)
}

// 2. 获取所有记录并过滤
allResults := iter.GetRecords(true)
defer record.PutRecords(allResults)
var activeRecords []map[string]any
for _, record := range allResults {
    if active, ok := record["active"].(bool); ok && active {
        activeRecords = append(activeRecords, record)
    }
}
```

**方案2：使用整数类型代替布尔类型**

- 将 `active` 字段定义为 `int` 类型
- 使用 `0` 表示 false，`1` 表示 true
- 这样可以利用现有的整数索引实现，避免索引键冲突

```go
// 定义字段时使用整数类型
fields := map[string]any{
    "id":     0,
    "name":   "",
    "active": 0, // 使用整数类型，0=false, 1=true
}

// 创建索引
activeIndex, err := engine.DefaultNormalIndexNew("idx_active")
activeIndex.AddFields("active")
err = table.CreateIndex(activeIndex)
```

### 4.7.3 表级索引管理方法

`sfsDb` 提供了丰富的表级索引管理方法，方便用户创建和管理索引：

#### 创建索引

```go
// 创建自定义索引
func (t *Table) CreateIndex(index Index) error {
	if t.indexIDManager == nil {
		t.indexIDManager = NewIDManager(t.kvStore)
	}
	fkey := t.indexIDManager.GenerateIndexKey(t.id, index.Name())
	idxID, isNew, err := t.indexIDManager.GetOrCreateID(fkey)
	if err != nil {
		return err
	}
	err = t.indexs.createIndex(index, idxID)
	if err != nil {
		// 回退ID
		_, err = t.indexIDManager.GetPreviousID(fkey)
		if err != nil {
			return err
		}
		return err
	}
	//对table现有数据创建对应的新索引数据
	if isNew {
		t.createIndexData(index)
	}
	return nil
}

// 创建普通复合索引
func (t *Table) CreateCompositeIndex(name string, fields ...string) error

// 创建主键复合索引
func (t *Table) CreateCompositePrimaryKey(name string, fields ...string) error

// 创建主键索引（支持单个或多个字段）
func (t *Table) CreatePrimaryKey(fields ...string) error

// 创建普通索引（简化版，直接指定名称和字段）
func (t *Table) CreateSimpleIndex(name string, fields ...string) error
```

#### 获取索引

```go
// 获取主键索引
func (t *Table) GetPrimaryKey() PrimaryKey

// 获取所有索引
func (t *Table) GetAllIndexes() []Index

// 根据名称获取索引
func (t *Table) GetIndexByName(name string) Index

// 根据字段名获取包含该字段的所有索引
func (t *Table) GetIndexesByField(field string) []Index

// 匹配索引
func (t *Table) MatchIndex(fields ...string) Index
```

#### 删除索引

**重要说明**：删除索引不会删除现存数据，不会对数据产生影响，只是存在冗余。

```go
// 删除指定名称的索引
//删除索引不会删除现存数据，不会对数据产生影响，只是存在冗余。
func (t *Table) DropIndex(name string) error

// 删除主键索引
func (t *Table) DropPrimaryKey() error
```

### 4.7.4 索引管理示例

```go
// 示例：表级索引管理

// 1. 创建各种类型的索引
// 创建主键索引
table.CreatePrimaryKey("id")

// 创建复合主键索引
table.CreateCompositePrimaryKey("pk_user_id", "user_id", "product_id")

// 创建普通索引
table.CreateSimpleIndex("idx_name", "name")

// 创建复合索引
table.CreateCompositeIndex("idx_name_age", "name", "age")

// 2. 获取索引信息
// 获取主键索引
pk := table.GetPrimaryKey()

// 获取所有索引
allIndexes := table.GetAllIndexes()

// 根据名称获取索引
nameIndex := table.GetIndexByName("idx_name")

// 根据字段获取索引
nameIndexes := table.GetIndexesByField("name")

// 3. 删除索引
// 删除普通索引
table.DropIndex("idx_name")

// 删除主键索引
table.DropPrimaryKey()
```

## 4.8 索引优化建议

1. **选择合适的索引类型**: 根据查询需求选择合适的索引类型
   - 唯一标识记录：使用主键索引
   - 加速普通查询：使用普通索引
   - 文本搜索：使用全文索引

2. **合理设计索引字段**: 
   - 主键字段应选择唯一、稳定的字段
   - 普通索引应选择经常用于查询条件的字段
   - 全文索引应选择需要文本搜索的字段

3. **控制索引数量**: 
   - 过多的索引会影响写入性能
   - 只为频繁查询的字段创建索引

4. **使用复合索引**: 
   - 对于多字段查询，复合索引比多个单列索引更高效
   - 遵循最左前缀原则设计复合索引

5. **定期维护索引**: 
   - 对于频繁更新的表，定期重建索引
   - 删除不再使用的索引

## 4.9 索引缓存设置

`SetIndexCacheSizeLimit` 函数用于设置索引缓存的大小限制，控制内存使用并优化索引操作性能。

```go
// 设置索引缓存大小限制
func (t *Table) SetIndexCacheSizeLimit(limit int) {
    //控制一个合理数值，防止缓存大小过大
    if limit <= 0 {
        limit = 1000
    }
    indexCacheSizeLimit = limit
}
```

**参数说明**：
- `limit`: 索引缓存的最大数量

**使用示例**：

```go
// 设置索引缓存大小为 2000
table.SetIndexCacheSizeLimit(2000)

// 设置索引缓存大小为默认值（1000）
table.SetIndexCacheSizeLimit(0) // 当值 ≤ 0 时，会使用默认值 1000
```

**注意事项**：
- 默认值为 1000，适用于大多数场景
- 较大的缓存大小可以提高频繁索引操作的性能，但会增加内存使用
- 较小的缓存大小可以减少内存使用，但可能会降低频繁索引操作的性能