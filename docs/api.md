# API 文档

## 核心类型

### Table
表结构，半结构泛型化支持，提供灵活的组合索引优化查询。

```go
type Table struct {
    id   uint8  // 表id
    name string // 表名
    fields map[string]any // 字段映射，string为字段名，any为字段值
    indexs  *Indexs // 索引集合
    counter AutoInt // 自动增值计数器
    kvStore storage.Store
}
```

### Record
记录类型，本质是 `map[string]any` 的别名，提供便捷的记录操作方法。

```go
type Record map[string]any
```

### Records
记录集合类型，提供批量记录操作方法。

```go
type Records []Record
```

### Index
索引接口，定义所有索引类型共有的方法。

```go
type Index interface {
    AddFields(field ...string)
    GetFields() []string
    Len() int
    SetId(id uint8)
    GetId() uint8
    Name() string
    SetName(name string) error
    Prefix(tbid uint8) []byte
    Join(fieldsBytes *map[string][]byte) []byte
    JoinPrefix(tbid uint8, val []byte) []byte
    JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
    MatchFields(fields ...string) bool
    Parse(fields []string, value []byte) *map[string][]byte
}
```

### PrimaryKey
主键接口，嵌入基础索引接口。

```go
type PrimaryKey interface {
    Index
    GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte
}
```

### FullTextIndex
全文索引接口，嵌入基础索引接口。

```go
type FullTextIndex interface {
    Index
    SetFullField(field string, len int) error
    JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte
    Tokenize(nr string, ftlen int) (tokens []string)
}
```

## 核心 API

### 表操作

#### TableNew
```go
func TableNew(name string) (*Table, error)
```
- **功能**：新建一个表
- **参数**：
  - `name`：表名，不能包含分隔符SPLIT("-")
- **返回值**：
  - `*Table`：表实例
  - `error`：错误信息

#### SetFields
```go
func (t *Table) SetFields(fields map[string]any) error
```
- **功能**：为表预设字段和类型
- **参数**：
  - `fields`：字段映射，string为字段名，any为字段值
- **返回值**：
  - `error`：错误信息

#### GetAllFields
```go
func (t *Table) GetAllFields() map[string]any
```
- **功能**：获取所有字段值，用于添加记录时，直接复制，无需自行创建
- **返回值**：
  - `map[string]any`：字段的副本

### 记录操作

#### Insert
```go
func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (currentID int, err error)
```
- **功能**：插入记录
- **参数**：
  - `fields`：记录数据
  - `batchs`：可选的批量操作对象
- **返回值**：
  - `currentID`：插入的记录ID（如果使用自动增值主键）
  - `error`：错误信息

#### Delete
```go
func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error
```
- **功能**：删除记录
- **参数**：
  - `fields`：包含主键字段的记录数据
  - `batchs`：可选的批量操作对象
- **返回值**：
  - `error`：错误信息

#### Update
```go
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error
```
- **功能**：更新记录，不支持修改主键字段
- **参数**：
  - `fields`：包含主键字段和要更新字段的记录数据
  - `batchs`：可选的批量操作对象
- **返回值**：
  - `error`：错误信息

#### Read
```go
func (t *Table) Read(fields *map[string]any) ([]byte, error)
```
- **功能**：从按主键数据库读取记录
- **参数**：
  - `fields`：包含主键字段的记录数据
- **返回值**：
  - `[]byte`：记录的原始字节数据
  - `error`：错误信息

### 查询操作

#### Search
```go
func (t *Table) Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter
```
- **功能**：根据条件搜索记录
- **参数**：
  - `fields`：搜索条件
  - `ops`：比较操作符，默认为 Like
- **返回值**：
  - `*TableIter`：表迭代器，用于遍历搜索结果

#### ForData
```go
func (t *Table) ForData() *TableIter
```
- **功能**：遍历表所有数据
- **返回值**：
  - `*TableIter`：表迭代器，用于遍历所有数据

### 索引操作

#### CreateIndex
```go
func (t *Table) CreateIndex(index Index) error
```
- **功能**：为表创建索引
- **参数**：
  - `index`：索引对象
- **返回值**：
  - `error`：错误信息

#### GetPrimaryKey
```go
func (t *Table) GetPrimaryKey() PrimaryKey
```
- **功能**：获取主键索引
- **返回值**：
  - `PrimaryKey`：主键索引对象

### 记录操作（Record）

#### Select
```go
func (r Record) Select(keys ...string) (rs Record)
```
- **功能**：选择显示的字段，如SQL语句中的select
- **参数**：
  - `keys`：要选择的字段名
- **返回值**：
  - `Record`：包含指定字段的记录

### 记录集合操作（Records）

#### Select
```go
func (rs Records) Select(fields ...string) (rs2 Records)
```
- **功能**：批量选择显示的字段
- **参数**：
  - `fields`：要选择的字段名
- **返回值**：
  - `Records`：包含指定字段的记录集合

#### Intersect
```go
func (rs Records) Intersect(other ...Records) (rs2 Records)
```
- **功能**：计算记录集合的交集
- **参数**：
  - `other`：其他记录集合
- **返回值**：
  - `Records`：交集结果

#### Union
```go
func (rs Records) Union(other ...Records) (rs2 Records)
```
- **功能**：计算记录集合的并集
- **参数**：
  - `other`：其他记录集合
- **返回值**：
  - `Records`：并集结果

#### Difference
```go
func (rs Records) Difference(other ...Records) (rs2 Records)
```
- **功能**：计算记录集合的差集
- **参数**：
  - `other`：其他记录集合
- **返回值**：
  - `Records`：差集结果

## 索引类型

### DefaultPrimaryKey
默认主键索引，支持自动增值。

```go
func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error)
```

### DefaultNormalIndex
默认普通索引，二级索引。

```go
func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error)
```

### DefaultFullTextIndex
默认全文索引，支持考据级全文搜索。

```go
func DefaultFullTextIndexNew(name string) (*DefaultFullTextIndex, error)
```

#### SetFullField
```go
func (dfi *DefaultFullTextIndex) SetFullField(field string, len int) error
```
- **功能**：指定哪个字段是全文索引字段，以及索引长度
- **参数**：
  - `field`：全文索引字段名
  - `len`：索引长度，限制为3-11个字符
- **返回值**：
  - `error`：错误信息

## 表迭代器（TableIter）

表迭代器用于遍历搜索结果或表中所有数据，提供了丰富的记录处理功能。

### 结构定义

```go
type TableIter struct {
    iter       storage.Iterator
    jumpRanges []storage.Iterator
    table      *Table
    match      []Match
    selects    []string
    index      Index //搜索时使用的索引
    move       map[bool]func() bool
    top        map[bool]func() bool
    mu         sync.Mutex
}
```

### 分页结构

```go
type Page struct {
    Start int // 分页开始位置，默认0
    Count int // 分页数量，默认-1表示返回所有记录
}
```

### 导出函数类型

```go
// 导出记录 ，用于流式处理删除修改记录操作等
type ExportRecord func(rd *Record) bool

// 导出函数
type Export func(k, v []byte) bool
```

### 主要方法

#### 创建迭代器

```go
func TableIterNew(table *Table, iter storage.Iterator, index Index, selects ...string) *TableIter
```
- **功能**：创建新的表迭代器
- **参数**：
  - `table`：关联的表
  - `iter`：底层存储迭代器
  - `index`：搜索时使用的索引
  - `selects`：要选择的字段
- **返回值**：
  - `*TableIter`：表迭代器实例

#### 基本导航方法

- `First()`：移动到第一个记录
- `Last()`：移动到最后一个记录
- `Next()`：移动到下一个记录
- `Prev()`：移动到上一个记录
- `Seek(key []byte)`：移动到大于等于指定key的位置
- `Key()`：获取当前记录的键
- `Value()`：获取当前记录的值
- `Release()`：释放迭代器资源

#### 高级功能

##### SetJumpRanges
```go
func (t *TableIter) SetJumpRanges(jumpRanges ...storage.Iterator)
```
- **功能**：设置跳跃区间，用于跳过某些数据
- **参数**：
  - `jumpRanges`：跳跃区间迭代器列表

##### SetMatch
```go
func (t *TableIter) SetMatch(match ...Match)
```
- **功能**：设置匹配条件，用于过滤记录
- **参数**：
  - `match`：匹配条件列表

##### SetSelects
```go
func (t *TableIter) SetSelects(fields ...string)
```
- **功能**：设置要选择的字段，如SQL语句中的select
- **参数**：
  - `fields`：要选择的字段名列表

##### GetRecords
```go
func (t *TableIter) GetRecords(esc bool, limit ...int) (r Records)
```
- **功能**：遍历迭代器返回解析后的记录
- **参数**：
  - `esc`：是否顺序遍历，true为顺序，false为倒序
  - `limit`：分页参数，start和count
- **返回值**：
  - `Records`：记录集合

##### Delete
```go
func (t *TableIter) Delete(limit ...int)
```
- **功能**：删除迭代器中的记录
- **参数**：
  - `limit`：限制删除数量

##### Update
```go
func (t *TableIter) Update(fields *map[string]any, limit ...int)
```
- **功能**：更新迭代器中的记录
- **参数**：
  - `fields`：要更新的字段和值
  - `limit`：限制更新数量

##### ExportRecord
```go
func (t *TableIter) ExportRecord(export ExportRecord, esc bool, limit ...int)
```
- **功能**：导出记录，用于流式处理
- **参数**：
  - `export`：导出函数，返回false则停止导出
  - `esc`：是否顺序遍历
  - `limit`：分页参数

##### ForExport
```go
func (t *TableIter) ForExport(esc bool, export Export)
```
- **功能**：遍历迭代器导出原始键值对
- **参数**：
  - `esc`：是否顺序遍历
  - `export`：导出函数，返回false则停止导出

##### GetPrimaryKeys
```go
func (t *TableIter) GetPrimaryKeys(k, v []byte, fields ...string) (r any)
```
- **功能**：提取主键值
- **参数**：
  - `k`：记录键
  - `v`：记录值
  - `fields`：要提取的字段，默认提取所有主键字段
- **返回值**：
  - `any`：主键值，组合主键返回拼接字符串，单主键返回对应类型值

##### Map
```go
func (t *TableIter) Map(fields ...string) (data map[any]bool)
```
- **功能**：将字段值转换为映射，用于与其他迭代器进行匹配
- **参数**：
  - `fields`：要转换的字段，默认使用主键字段
- **返回值**：
  - `map[any]bool`：字段值映射

##### Count
```go
func (t *TableIter) Count() int
```
- **功能**：统计索引记录数
- **返回值**：
  - `int`：记录数量

##### Exist
```go
func (t *TableIter) Exist() bool
```
- **功能**：判断是否存在指定的主键记录
- **返回值**：
  - `bool`：是否存在记录

### 使用示例

#### 基本遍历

```go
// 获取迭代器
iter := table.ForData()
defer iter.Release()

// 顺序遍历
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("记录: %v\n", record)
}

// 倒序遍历
for iter.Last(); iter.Valid(); iter.Prev() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("记录: %v\n", record)
}
```

#### 搜索遍历

```go
// 搜索
searchFields := map[string]any{
    "age": 30,
}
iter := table.Search(&searchFields)
defer iter.Release()

// 遍历搜索结果
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("搜索结果: %v\n", record)
}
```

#### 分页获取记录

```go
// 获取迭代器
iter := table.ForData()
defer iter.Release()

// 获取第2页，每页10条记录
records := iter.GetRecords(true, 10, 10)
for _, record := range records {
    fmt.Printf("记录: %v\n", record)
}
```

#### 批量更新记录

```go
// 搜索条件
searchFields := map[string]any{
    "status": "inactive",
}
iter := table.Search(&searchFields)
defer iter.Release()

// 批量更新
updateData := map[string]any{
    "status": "active",
}
iter.Update(&updateData)
```

#### 批量删除记录

```go
// 搜索条件
searchFields := map[string]any{
    "age": 0,
}
iter := table.Search(&searchFields)
defer iter.Release()

// 批量删除
iter.Delete()
```

#### 使用跳跃区间

```go
// 创建跳跃区间迭代器
// 例如：跳过状态为"deleted"的记录
jumpRange := createJumpRangeIterator("status", "deleted")

// 设置跳跃区间
iter.SetJumpRanges(jumpRange)

// 遍历，会自动跳过指定区间的记录
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("记录: %v\n", record)
}
```

#### 使用匹配条件

```go
// 创建匹配条件
match := NewMatchFunc(func(record *Record) bool {
    // 自定义匹配逻辑，例如：年龄大于25且状态为active
    age, ok1 := (*record)["age"].(int)
    status, ok2 := (*record)["status"].(string)
    return ok1 && ok2 && age > 25 && status == "active"
})

// 设置匹配条件
iter.SetMatch(match)

// 遍历，只会返回符合条件的记录
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("匹配记录: %v\n", record)
}
```

### 性能优化建议

1. **及时释放资源**：使用完毕后调用 `Release()` 方法释放迭代器资源
2. **合理使用分页**：对于大量数据，使用分页功能减少内存占用
3. **使用跳跃区间**：对于需要跳过特定范围数据的场景，使用跳跃区间提高遍历效率
4. **使用匹配条件**：在遍历前设置匹配条件，减少不必要的记录处理
5. **批量操作**：对于删除和更新操作，使用迭代器的批量操作功能，减少数据库交互
6. **选择必要字段**：使用 `SetSelects()` 方法只选择必要的字段，减少数据传输和处理开销

## 错误处理

### 常见错误
- 字段名包含分隔符
- 字段不存在于表中
- 字段类型不匹配
- 主键值不存在
- 索引创建失败

### 错误检查
```go
if err != nil {
    // 处理错误
    fmt.Printf("错误: %v\n", err)
    return
}
```

## 性能优化建议

1. **合理设计表结构**：
   - 只为必要字段创建索引
   - 合理选择字段类型

2. **批量操作**：
   - 对于大量插入操作，使用批量操作
   - 减少单独的数据库写入操作

3. **查询优化**：
   - 使用索引字段进行查询
   - 避免全表扫描

4. **全文索引优化**：
   - 合理设置全文索引字段的长度
   - 只对需要全文搜索的字段创建全文索引

5. **资源管理**：
   - 及时释放迭代器资源
   - 合理使用对象池

## 使用示例

### 基本使用

```go
// 创建表
table, err := engine.TableNew("users")
if err != nil {
    panic(err)
}

// 设置字段
fields := map[string]any{
    "id":   0,     // 自动增值主键
    "name": "",    // 字符串
    "age":  0,     // 整数
    "email": "",   // 字符串
}
err = table.SetFields(fields)
if err != nil {
    panic(err)
}

// 插入记录
user := map[string]any{
    "name": "张三",
    "age":  30,
    "email": "zhangsan@example.com",
}
id, err := table.Insert(&user)
if err != nil {
    panic(err)
}
fmt.Printf("插入成功，ID: %d\n", id)

// 搜索记录
searchFields := map[string]any{
    "name": "张三",
}
iter := table.Search(&searchFields)
defer iter.Release()
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("找到记录: %v\n", record)
}

// 更新记录
updateFields := map[string]any{
    "id": id,      // 必须包含主键
    "age": 31,     // 要更新的字段
}
err = table.Update(&updateFields)
if err != nil {
    panic(err)
}

// 删除记录
deleteFields := map[string]any{
    "id": id,      // 必须包含主键
}
err = table.Delete(&deleteFields)
if err != nil {
    panic(err)
}
```

### 全文索引使用

```go
// 创建全文索引
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_idx")
if err != nil {
    panic(err)
}
fullTextIndex.AddFields("content", "id") // 添加字段，最后一个字段必须是主键
fullTextIndex.SetFullField("content", 5) // 设置全文索引字段和长度
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}

// 插入带全文索引的记录
contentRecord := map[string]any{
    "id": 1,
    "content": "这是一段测试文本，用于全文索引测试",
}
_, err = table.Insert(&contentRecord)
if err != nil {
    panic(err)
}

// 全文搜索
searchContent := map[string]any{
    "content": "测试",
}
iter := table.Search(&searchContent)
defer iter.Release()
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("全文搜索找到: %v\n", record)
}
```