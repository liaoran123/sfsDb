# API 文档

## 核心类型

### Table
表结构，半结构泛型化支持，提供灵活的组合索引优化查询。

```go
type Table struct {
    id   uint8  // 表id
    name string // 表名
    fields         map[string]any   // 字段映射，string为字段名，any为字段值
    fieldsid       map[uint8]string // id到字段名的映射
    indexs         *Indexs          // 索引集合
    counter        AutoInt          // 自动增值计数器，使用自定义的AutoInt
    kvStore        storage.Store
    fieldIDManager *IDManager // 字段ID管理器
    indexIDManager *IDManager // 索引ID管理器
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
    Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
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
- **功能**：为表预设字段和类型，或更新现有字段映射
- **参数**：
  - `fields`：字段映射，string为字段名，any为字段值
- **返回值**：
  - `error`：错误信息

#### UpdateFieldName
```go
func (t *Table) UpdateFieldName(oldfield string, newfield string) error
```
- **功能**：修改字段名称，更新ID管理器和索引中的字段名
- **参数**：
  - `oldfield`：旧字段名
  - `newfield`：新字段名
- **返回值**：
  - `error`：错误信息
- **说明**：修改字段名时，必须先调用此方法，然后再调用SetFields更新字段映射，否则系统会将修改的字段视为新字段，导致ID管理器和索引字段名不匹配

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

#### SearchRange
```go
func (t *Table) SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (*TableIter, error)
```
- **功能**：根据字段值范围搜索记录，特别适用于时序数据库的时间范围查询
- **参数**：
  - `funIter`：迭代器函数，用于创建范围迭代器
  - `fieldname`：要搜索的字段名
  - `Start`：区间开始值。Start=nil表示从索引最小值开始
  - `Limit`：区间结束值。Limit=nil表示到索引最大值结束。同时为nil即表示遍历索引
- **返回值**：
  - `*TableIter`：表迭代器，用于遍历搜索结果
  - `error`：错误信息

#### RangeForAny
```go
func (t *Table) RangeForAny(funIter storage.FunIter, fieldname string, Start, Limit any) (storage.Iterator, Index, error)
```
- **功能**：创建区间迭代器，用于范围搜索和跳跃区间
- **参数**：
  - `funIter`：迭代器函数，用于创建范围迭代器。如果为nil，将使用表的默认迭代器
  - `fieldname`：要搜索的字段名
  - `Start`：区间开始值。Start=nil表示从索引最小值开始
  - `Limit`：区间结束值。Limit=nil表示到索引最大值结束。同时为nil即表示遍历索引
- **返回值**：
  - `storage.Iterator`：存储迭代器，用于遍历搜索结果
  - `Index`：使用的索引
  - `error`：错误信息

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

### NormalIndex
普通索引接口，嵌入基础索引接口。

```go
type NormalIndex interface {
    Index
    // 将索引的value转换为主键map值
    Tag() bool
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}

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

1. **及时释放资源**：使用完毕后调用 `engine.GlobalTableIterPool.Put(iter)` 释放迭代器资源，调用 `record.PutRecords(records)` 释放记录集合资源
2. **合理使用分页**：对于大量数据，使用分页功能减少内存占用
3. **使用跳跃区间**：对于需要跳过特定范围数据的场景，使用跳跃区间提高遍历效率
4. **使用匹配条件**：在遍历前设置匹配条件，减少不必要的记录处理
5. **批量操作**：对于删除和更新操作，使用迭代器的批量操作功能，减少数据库交互
6. **选择必要字段**：使用 `SetSelects()` 方法只选择必要的字段，减少数据传输和处理开销

## 时序数据处理 (Time Package)

### 核心类型

#### TimeGranularity
时间粒度类型，用于指定时间处理的精度。

```go
type TimeGranularity string

const (
    TimeGranularityMillisecond TimeGranularity = "millisecond"
    TimeGranularityMicrosecond TimeGranularity = "microsecond"
    TimeGranularitySecond      TimeGranularity = "second"
    TimeGranularityMinute      TimeGranularity = "minute"
    TimeGranularityHour        TimeGranularity = "hour"
    TimeGranularityDay         TimeGranularity = "day"
    TimeGranularityWeek        TimeGranularity = "week"
    TimeGranularityMonth       TimeGranularity = "month"
    TimeGranularityQuarter     TimeGranularity = "quarter"
    TimeGranularityYear        TimeGranularity = "year"
)
```

#### TimeAggregationResult
时间聚合结果结构，用于存储按时间粒度聚合的数据。

```go
type TimeAggregationResult struct {
    TimeKey string  `json:"time_key"`
    Value   float64 `json:"value"`
}
```

#### TimeWindow
时间窗口接口，用于时间窗口计算。

```go
type TimeWindow interface {
    // Next 移动到下一个窗口，返回是否还有下一个窗口
    Next() bool
    // Start 获取当前窗口的开始时间
    Start() time.Time
    // End 获取当前窗口的结束时间
    End() time.Time
    // Reset 重置窗口到初始状态
    Reset()
}
```

#### SlidingWindow
滑动窗口实现，支持固定大小的窗口以固定步长滑动。

```go
type SlidingWindow struct {
    startTime    time.Time
    endTime      time.Time
    windowSize   time.Duration
    stepSize     time.Duration
    currentStart time.Time
}
```

#### TumblingWindow
滚动窗口实现，窗口之间没有重叠。

```go
type TumblingWindow struct {
    startTime    time.Time
    endTime      time.Time
    windowSize   time.Duration
    currentStart time.Time
}
```

#### WindowAggregation
窗口聚合结果，用于存储按时间窗口聚合的数据。

```go
type WindowAggregation struct {
    WindowStart time.Time  `json:"window_start"`
    WindowEnd   time.Time  `json:"window_end"`
    Value       float64    `json:"value"`
}
```

#### TimeSeriesPoint
时间序列数据点，用于时间序列预测。

```go
type TimeSeriesPoint struct {
    Time  time.Time
    Value float64
}
```

#### MovingAveragePrediction
移动平均预测结果。

```go
type MovingAveragePrediction struct {
    PredictedPoints []TimeSeriesPoint
    WindowSize      int
}
```

#### LinearRegressionPrediction
线性回归预测结果。

```go
type LinearRegressionPrediction struct {
    PredictedPoints []TimeSeriesPoint
    Slope           float64
    Intercept       float64
    R2              float64 // 决定系数，衡量模型拟合度
}
```

#### CompressedTimeSeries
压缩后的时间序列数据。

```go
type CompressedTimeSeries struct {
    StartTime    time.Time
    Interval     time.Duration
    CompressedValues []byte
    CompressionType  string
}
```

#### TimeRangeQueryOptions
时间范围查询选项，用于优化时间范围查询。

```go
type TimeRangeQueryOptions struct {
    FieldName     string
    StartTime     time.Time
    EndTime       time.Time
    TimeGranularity TimeGranularity
    Inclusive     bool // 是否包含边界值
}
```

### 核心函数

#### 时间粒度格式化
```go
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string
```
- **功能**：将时间戳格式化为指定粒度的时间字符串
- **参数**：
  - `t`：时间对象
  - `granularity`：时间粒度
- **返回值**：
  - `string`：格式化后的时间字符串

#### 时间桶分配
```go
func TimeBucket(t time.Time, duration time.Duration) time.Time
```
- **功能**：将时间戳分配到指定粒度的时间桶
- **参数**：
  - `t`：时间对象
  - `duration`：时间桶大小
- **返回值**：
  - `time.Time`：时间桶的起始时间

#### 时间范围计算
```go
func TimeRange(start time.Time, granularity TimeGranularity) (time.Time, time.Time)
```
- **功能**：根据时间粒度计算时间范围
- **参数**：
  - `start`：起始时间
  - `granularity`：时间粒度
- **返回值**：
  - `time.Time`：范围起始时间
  - `time.Time`：范围结束时间

#### 时间粒度聚合
```go
func AggregateByTimeGranularity(records interface{}, timeField string, valueField string, 
    granularity TimeGranularity, aggregationType string) ([]TimeAggregationResult, error)
```
- **功能**：按时间粒度聚合数据
- **参数**：
  - `records`：记录集合，可以是 `[]map[string]any` 或 `record.Records`
  - `timeField`：时间字段名
  - `valueField`：值字段名
  - `granularity`：时间粒度
  - `aggregationType`：聚合类型（sum, avg, count, max, min）
- **返回值**：
  - `[]TimeAggregationResult`：聚合结果
  - `error`：错误信息

#### 时间戳转换

```go
// TimeToUnixTimestamp 将 time.Time 转换为秒级 Unix 时间戳
func TimeToUnixTimestamp(t time.Time) int64

// TimeToUnixTimestampMs 将 time.Time 转换为毫秒级 Unix 时间戳
func TimeToUnixTimestampMs(t time.Time) int64

// TimeToUnixTimestampNs 将 time.Time 转换为纳秒级 Unix 时间戳
func TimeToUnixTimestampNs(t time.Time) int64

// UnixTimestampToTime 将秒级 Unix 时间戳转换为 time.Time
func UnixTimestampToTime(timestamp int64) time.Time

// UnixTimestampMsToTime 将毫秒级 Unix 时间戳转换为 time.Time
func UnixTimestampMsToTime(timestampMs int64) time.Time

// UnixTimestampNsToTime 将纳秒级 Unix 时间戳转换为 time.Time
func UnixTimestampNsToTime(timestampNs int64) time.Time
```

#### 时间窗口操作

```go
// NewSlidingWindow 创建一个新的滑动窗口
func NewSlidingWindow(startTime, endTime time.Time, windowSize, stepSize time.Duration) *SlidingWindow

// NewTumblingWindow 创建一个新的滚动窗口
func NewTumblingWindow(startTime, endTime time.Time, windowSize time.Duration) *TumblingWindow

// AggregateByWindow 按时间窗口聚合数据
func AggregateByWindow(records []map[string]any, timeField string, valueField string, window TimeWindow, aggregationType string) ([]WindowAggregation, error)
```

#### 时间序列预测

```go
// NewMovingAveragePrediction 创建移动平均预测
func NewMovingAveragePrediction(points []TimeSeriesPoint, windowSize, predictCount int, interval time.Duration) *MovingAveragePrediction

// NewLinearRegressionPrediction 创建线性回归预测
func NewLinearRegressionPrediction(points []TimeSeriesPoint, predictCount int, interval time.Duration) *LinearRegressionPrediction

// PredictTimeSeries 预测时间序列数据
func PredictTimeSeries(points []TimeSeriesPoint, method string, params map[string]any) (any, error)
```

#### 时间序列数据压缩

```go
// CompressTimeSeries 压缩时间序列数据
func CompressTimeSeries(points []TimeSeriesPoint, compressionType string, interval time.Duration) (*CompressedTimeSeries, error)

// DecompressTimeSeries 解压缩时间序列数据
func DecompressTimeSeries(cts *CompressedTimeSeries, count int) ([]TimeSeriesPoint, error)

// GetCompressionRatio 计算压缩率
func GetCompressionRatio(originalSize, compressedSize int) float64
```

#### 增强的时间粒度支持

```go
// IsWeekday 检查给定时间是否为工作日
func IsWeekday(t time.Time) bool

// IsWeekend 检查给定时间是否为周末
func IsWeekend(t time.Time) bool

// GetQuarter 获取给定时间所在的季度
func GetQuarter(t time.Time) int

// GetWeekNumber 获取给定时间所在的周数（一年中的第几周）
func GetWeekNumber(t time.Time) int
```

#### 时间范围查询优化

```go
// NewTimeRangeQueryOptions 创建时间范围查询选项
func NewTimeRangeQueryOptions(fieldName string, startTime, endTime time.Time, granularity TimeGranularity) *TimeRangeQueryOptions

// SearchTimeRange 执行时间范围查询
func SearchTimeRange(table *engine.Table, options *TimeRangeQueryOptions) (*engine.TableIter, error)

// SearchTimeRangeWithGranularity 按时间粒度执行时间范围查询
func SearchTimeRangeWithGranularity(table *engine.Table, fieldName string, startTime, endTime time.Time, granularity TimeGranularity) (*engine.TableIter, error)

// AdjustTimeToGranularity 将时间调整到指定粒度的边界
func AdjustTimeToGranularity(t time.Time, granularity TimeGranularity) time.Time

// AdjustTimeRangeByGranularity 根据时间粒度调整时间范围
func AdjustTimeRangeByGranularity(startTime, endTime time.Time, granularity TimeGranularity) (time.Time, time.Time)

// TimeRangeQueryWithAggregation 带聚合的时间范围查询
func TimeRangeQueryWithAggregation(table *engine.Table, options *TimeRangeQueryOptions, valueField string, aggregationType string) ([]TimeAggregationResult, error)
```

### 使用示例

#### 基本时间处理
```go
import (
    "time"
    sfsTime "github.com/liaoran123/sfsDb/time"
)

// 格式化时间
now := time.Now()
hourStr := sfsTime.FormatTimeByGranularity(now, sfsTime.TimeGranularityHour)
fmt.Println("Hour:", hourStr) // 输出: Hour: 2024-12-25 14:00:00

// 时间桶分配
bucket := sfsTime.TimeBucket(now, time.Hour)
fmt.Println("Bucket:", bucket) // 输出时间桶的起始时间

// 时间范围计算
start, end := sfsTime.TimeRange(now, sfsTime.TimeGranularityDay)
fmt.Println("Day range:", start, "to", end)
```

#### 时间戳转换
```go
// 时间对象转整数时间戳
now := time.Now()
unixTime := sfsTime.TimeToUnixTimestamp(now)         // 秒级
unixTimeMs := sfsTime.TimeToUnixTimestampMs(now)     // 毫秒级
unixTimeNs := sfsTime.TimeToUnixTimestampNs(now)     // 纳秒级

// 整数时间戳转时间对象
timeObj := sfsTime.UnixTimestampToTime(unixTime)     // 秒级
timeObjMs := sfsTime.UnixTimestampMsToTime(unixTimeMs) // 毫秒级
timeObjNs := sfsTime.UnixTimestampNsToTime(unixTimeNs) // 纳秒级
```

#### 数据聚合
```go
// 假设我们有一组传感器数据记录
records := record.Records{...}

// 按小时聚合
hourlyResults, err := sfsTime.AggregateByTimeGranularity(
    records,
    "timestamp",    // 时间字段
    "value",        // 值字段
    sfsTime.TimeGranularityHour,  // 时间粒度
    "sum",          // 聚合类型
)

// 输出结果
for _, result := range hourlyResults {
    fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
```

#### 时间窗口计算
```go
// 创建滑动窗口
startTime := time.Now().Add(-10 * time.Minute)
endTime := time.Now()
windowSize := 2 * time.Minute
stepSize := 1 * time.Minute

window := sfsTime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

// 遍历窗口
for window.Next() {
    start := window.Start()
    end := window.End()
    fmt.Printf("Window: %s to %s\n", start, end)
}

// 按窗口聚合
var records []map[string]any
// ... 填充数据 ...

results, err := sfsTime.AggregateByWindow(
    records,
    "timestamp",
    "value",
    window,
    "sum"
)
```

#### 时间序列预测
```go
// 创建测试数据点
var points []sfsTime.TimeSeriesPoint
now := time.Now()
for i := 0; i < 10; i++ {
    points = append(points, sfsTime.TimeSeriesPoint{
        Time:  now.Add(time.Duration(i) * time.Minute),
        Value: float64(i),
    })
}

// 移动平均预测
maPrediction := sfsTime.NewMovingAveragePrediction(points, 3, 5, time.Minute)
for _, point := range maPrediction.PredictedPoints {
    fmt.Printf("Predicted: %s, Value: %.2f\n", point.Time, point.Value)
}

// 线性回归预测
lrPrediction := sfsTime.NewLinearRegressionPrediction(points, 5, time.Minute)
for _, point := range lrPrediction.PredictedPoints {
    fmt.Printf("Predicted: %s, Value: %.2f\n", point.Time, point.Value)
}
```

#### 时间序列数据压缩
```go
// 压缩时间序列数据
compressed, err := sfsTime.CompressTimeSeries(points, "delta", time.Minute)
if err != nil {
    panic(err)
}

// 解压缩时间序列数据
decompressed, err := sfsTime.DecompressTimeSeries(compressed, 10)
if err != nil {
    panic(err)
}

// 计算压缩率
originalSize := len(points) * 16 // 假设每个点16字节
compressedSize := len(compressed.CompressedValues)
compressionRatio := sfsTime.GetCompressionRatio(originalSize, compressedSize)
fmt.Printf("Compression ratio: %.2f\n", compressionRatio)
```

#### 时间范围查询优化
```go
// 创建时间范围查询选项
options := sfsTime.NewTimeRangeQueryOptions(
    "timestamp",
    startTime,
    endTime,
    sfsTime.TimeGranularityHour
)

// 执行时间范围查询
iter, err := sfsTime.SearchTimeRange(table, options)
if err != nil {
    panic(err)
}
defer iter.Release()

// 带聚合的时间范围查询
aggregationResults, err := sfsTime.TimeRangeQueryWithAggregation(
    table,
    options,
    "value",
    "sum"
)
if err != nil {
    panic(err)
}

// 输出聚合结果
for _, result := range aggregationResults {
    fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
```

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

## 事务管理

### WithTransaction
```go
func WithTransaction(batch storage.Batch, tables []*Table, fn func(map[*Table]Transaction) error) error
```
- **功能**：执行事务操作，确保多个表的操作原子性
- **参数**：
  - `batch`：批量操作对象
  - `tables`：参与事务的表列表
  - `fn`：事务回调函数，包含具体的事务逻辑
- **返回值**：
  - `error`：错误信息

### Transaction 接口
```go
type Transaction interface {
    Get(fields *map[string]any) ([]byte, error)
    Set(fields *map[string]any) error
    Delete(fields *map[string]any) error
}
```
- **功能**：事务操作接口，提供事务内的读写删操作
- **方法**：
  - `Get`：获取记录
  - `Set`：设置记录
  - `Delete`：删除记录

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

6. **事务优化**：
   - 对于多表操作，使用事务确保原子性
   - 减少事务范围，只包含必要的操作
   - 使用批量操作提高事务性能

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

// 插入测试数据
// 准备多条测试记录用于全文搜索
contentRecords := []map[string]any{
    {"id": 1, "title": "软件工程师招聘", "content": "我们正在寻找优秀的软件工程师，熟悉Go语言和数据库设计"},
    {"id": 2, "title": "产品经理职位", "content": "产品经理需要有良好的沟通能力和项目管理经验"},
    {"id": 3, "title": "设计师需求", "content": "UI/UX设计师，精通Figma和用户体验设计原则"},
    {"id": 4, "title": "数据分析师", "content": "数据分析师需要熟悉SQL和数据可视化工具"},
}

for _, record := range contentRecords {
    _, err = table.Insert(&record)
    if err != nil {
        panic(err)
    }
}

// 全文搜索示例
fmt.Println("=== 全文搜索示例 ===")

// 搜索1: 查找包含"工程师"的记录
fmt.Println("\n1. 搜索 '工程师':")
search1 := map[string]any{"content": "工程师"}
iter1 := table.Search(&search1)
defer iter1.Release()
for iter1.First(); iter1.Valid(); iter1.Next() {
    record := table.ParseRecordValue(iter1.Value())
    fmt.Printf("   - %s: %s\n", record["title"], record["content"])
}

// 搜索2: 查找包含"设计"的记录
fmt.Println("\n2. 搜索 '设计':")
search2 := map[string]any{"content": "设计"}
iter2 := table.Search(&search2)
defer iter2.Release()
for iter2.First(); iter2.Valid(); iter2.Next() {
    record := table.ParseRecordValue(iter2.Value())
    fmt.Printf("   - %s: %s\n", record["title"], record["content"])
}

// 搜索3: 查找包含"数据"的记录
fmt.Println("\n3. 搜索 '数据':")
search3 := map[string]any{"content": "数据"}
iter3 := table.Search(&search3)
defer iter3.Release()
for iter3.First(); iter3.Valid(); iter3.Next() {
    record := table.ParseRecordValue(iter3.Value())
    fmt.Printf("   - %s: %s\n", record["title"], record["content"])
}
```

### 字段修改示例

```go
// 创建表
userTable, err := engine.TableNew("employees")
if err != nil {
    panic(err)
}

// 设置初始字段
initialFields := map[string]any{
    "id":       0,    // 自动增值主键
    "name":     "",   // 字符串
    "age":      0,    // 整数
    "email":    "",   // 字符串
    "position": "",   // 字符串
}
err = userTable.SetFields(initialFields)
if err != nil {
    panic(err)
}

// 插入测试数据
emp1 := map[string]any{
    "name":     "李四",
    "age":      28,
    "email":    "lisi@example.com",
    "position": "开发工程师",
}
emp2 := map[string]any{
    "name":     "王五",
    "age":      35,
    "email":    "wangwu@example.com",
    "position": "项目经理",
}
_, err = userTable.Insert(&emp1)
if err != nil {
    panic(err)
}
_, err = userTable.Insert(&emp2)
if err != nil {
    panic(err)
}

fmt.Println("=== 字段修改前 ===")
// 搜索修改前的记录
iterBefore := userTable.Search(&map[string]any{"name": "李四"})
defer iterBefore.Release()
for iterBefore.First(); iterBefore.Valid(); iterBefore.Next() {
    record := userTable.ParseRecordValue(iterBefore.Value())
    fmt.Printf("修改前记录: %v\n", record)
}

// 字段修改流程：将 "position" 字段重命名为 "job_title"
// 1. 首先调用 UpdateFieldName 更新字段名称映射
fmt.Println("\n=== 执行字段修改 ===")
err = userTable.UpdateFieldName("position", "job_title")
if err != nil {
    panic(err)
}
fmt.Println("已调用 UpdateFieldName 重命名字段")

// 2. 然后调用 SetFields 更新字段映射
updatedFields := map[string]any{
    "id":        0,
    "name":      "",
    "age":       0,
    "email":     "",
    "job_title": "", // 使用新的字段名
}
err = userTable.SetFields(updatedFields)
if err != nil {
    panic(err)
}
fmt.Println("已调用 SetFields 更新字段映射")

// 验证字段修改后的数据
fmt.Println("\n=== 字段修改后 ===")

// 插入新记录，使用新的字段名
emp3 := map[string]any{
    "name":      "赵六",
    "age":       32,
    "email":     "zhaoliu@example.com",
    "job_title": "产品经理", // 使用新的字段名
}
_, err = userTable.Insert(&emp3)
if err != nil {
    panic(err)
}

// 搜索所有记录，验证旧记录和新记录都能正确显示
allIter := userTable.Search(nil)
defer allIter.Release()
for allIter.First(); allIter.Valid(); allIter.Next() {
    record := userTable.ParseRecordValue(allIter.Value())
    fmt.Printf("修改后记录: %v\n", record)
}

// 使用新的字段名进行搜索
fmt.Println("\n=== 使用新字段名搜索 ===")
searchByJob := map[string]any{"job_title": "经理"}
jobIter := userTable.Search(&searchByJob)
defer jobIter.Release()
for jobIter.First(); jobIter.Valid(); jobIter.Next() {
    record := userTable.ParseRecordValue(jobIter.Value())
    fmt.Printf("职位包含'经理'的记录: %v\n", record)
}
```

### 时序数据区间搜索示例

```go
// 时序数据区间搜索示例：IoT设备数据查询
func querySensorDataByTimeRange(startTime, endTime int) error {
    // 创建表
    sensorTable, err := engine.TableNew("sensor_data")
    if err != nil {
        return err
    }
    
    // 设置字段，timestamp作为主键
    fields := map[string]any{
        "timestamp": 0,   // 时间戳，int类型
        "value":     0.0, // 传感器值
        "sensor_id": "",  // 传感器ID
    }
    err = sensorTable.SetFields(fields)
    if err != nil {
        return err
    }
    
    // 创建主键索引
    pk, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        return err
    }
    pk.AddFields("timestamp")
    err = sensorTable.CreateIndex(pk)
    if err != nil {
        return err
    }
    
    // 插入测试数据（模拟IoT设备产生的时序数据）
    now := int(time.Now().Unix())
    for i := 0; i < 100; i++ {
        data := map[string]any{
            "timestamp": now + i,
            "value":     float64(i * 10),
            "sensor_id": fmt.Sprintf("sensor_%d", i%10),
        }
        _, err := sensorTable.Insert(&data)
        if err != nil {
            return err
        }
    }
    
    // 使用SearchRange进行时间范围查询
    fmt.Println("=== 时序数据区间搜索示例 ===")
    
    // 定义迭代器函数
    funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
        return sensorTable.kvStore.Iterator(start, limit)
    })
    
    // 执行区间搜索
    iter, err := sensorTable.SearchRange(funIter, "timestamp", startTime, endTime)
    if err != nil {
        return err
    }
    defer iter.Release()
    
    // 获取结果
    records := iter.GetRecords(true)
    defer records.Release()
    
    // 打印结果
    fmt.Printf("时间范围 [%d, %d] 内的传感器数据：\n", startTime, endTime)
    for _, record := range records {
        fmt.Printf("时间戳: %d, 传感器ID: %s, 值: %f\n", 
            record["timestamp"], record["sensor_id"], record["value"])
    }
    
    return nil
}

// 使用示例
func main() {
    // 查询最近10秒的传感器数据
    now := int(time.Now().Unix())
    err := querySensorDataByTimeRange(now-10, now)
    if err != nil {
        panic(err)
    }
}
```

### 事务使用示例

#### 示例1：金融事务示例：银行转账

```go
// 金融事务示例：银行转账
func transferFunds(fromID, toID int, amount float64) error {
    batch := storage.KVDb.GetBatch()
    if batch == nil {
        return fmt.Errorf("failed to create batch")
    }
    
    return engine.WithTransaction(batch, []*engine.Table{accountTable}, func(transactions map[*engine.Table]engine.Transaction) error {
        tx := transactions[accountTable]
        
        // 获取转出账户
        fromFields := map[string]any{"id": fromID}
        fromIter, err := tx.Search(&fromFields)
        if err != nil {
            return err
        }
        defer engine.GlobalTableIterPool.Put(fromIter)
        
        fromRecords := fromIter.GetRecords(true)
        defer record.PutRecords(fromRecords)
        if len(fromRecords) == 0 {
            return fmt.Errorf("from account not found")
        }
        
        fromRecord := fromRecords[0]
        fromBalance, ok := fromRecord["balance"].(float64)
        if !ok {
            return fmt.Errorf("invalid balance type")
        }
        
        // 检查余额
        if fromBalance < amount {
            return fmt.Errorf("insufficient funds")
        }
        
        // 获取转入账户
        toFields := map[string]any{"id": toID}
        toIter, err := tx.Search(&toFields)
        if err != nil {
            return err
        }
        defer engine.GlobalTableIterPool.Put(toIter)
        
        toRecords := toIter.GetRecords(true)
        defer record.PutRecords(toRecords)
        if len(toRecords) == 0 {
            return fmt.Errorf("to account not found")
        }
        
        toRecord := toRecords[0]
        toBalance, ok := toRecord["balance"].(float64)
        if !ok {
            return fmt.Errorf("invalid balance type")
        }
        
        // 更新余额
        fromRecordMap := map[string]any(fromRecord)
        fromRecordMap["balance"] = fromBalance - amount
        
        toRecordMap := map[string]any(toRecord)
        toRecordMap["balance"] = toBalance + amount
        
        // 保存更新
        if err := tx.Update(&fromRecordMap); err != nil {
            return err
        }
        if err := tx.Update(&toRecordMap); err != nil {
            return err
        }
        
        return nil
    })
}

// 使用示例
func main() {
    // 创建账户表
    accountTable, err := engine.TableNew("accounts")
    if err != nil {
        panic(err)
    }
    
    // 设置字段
    fields := map[string]any{
        "id":      0,
        "name":    "",
        "balance": 0.0,
    }
    err = accountTable.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // 插入测试数据
    account1 := map[string]any{"id": 1, "name": "Alice", "balance": 1000.0}
    account2 := map[string]any{"id": 2, "name": "Bob", "balance": 500.0}
    _, err = accountTable.Insert(&account1)
    if err != nil {
        panic(err)
    }
    _, err = accountTable.Insert(&account2)
    if err != nil {
        panic(err)
    }
    
    // 执行转账
    fmt.Println("=== 执行转账 ===")
    err = transferFunds(1, 2, 200.0)
    if err != nil {
        fmt.Printf("转账失败: %v\n", err)
        return
    }
    fmt.Println("转账成功！")
    
    // 验证结果
    fmt.Println("\n=== 转账后余额 ===")
    iter, err := accountTable.Search(nil)
    if err != nil {
        panic(err)
    }
    defer engine.GlobalTableIterPool.Put(iter)
    
    records := iter.GetRecords(true)
    defer record.PutRecords(records)
    
    for _, r := range records {
        fmt.Printf("账户 %s: 余额 %.2f\n", r["name"], r["balance"])
    }
}
```

#### 示例2：多表事务

```go
// 多表事务示例：订单处理
func processOrder(orderID, userID int, amount float64, productIDs []int) error {
    batch := storage.KVDb.GetBatch()
    if batch == nil {
        return fmt.Errorf("failed to create batch")
    }
    
    return engine.WithTransaction(batch, []*engine.Table{orderTable, userTable, productTable}, func(transactions map[*engine.Table]engine.Transaction) error {
        // 1. 检查用户余额
        userTx := transactions[userTable]
        userFields := map[string]any{"id": userID}
        userIter, err := userTx.Search(&userFields)
        if err != nil {
            return err
        }
        defer engine.GlobalTableIterPool.Put(userIter)
        
        userRecords := userIter.GetRecords(true)
        defer record.PutRecords(userRecords)
        if len(userRecords) == 0 {
            return fmt.Errorf("user not found")
        }
        
        userRecord := userRecords[0]
        userBalance, ok := userRecord["balance"].(float64)
        if !ok {
            return fmt.Errorf("invalid balance type")
        }
        
        if userBalance < amount {
            return fmt.Errorf("insufficient funds")
        }
        
        // 2. 检查产品库存
        productTx := transactions[productTable]
        for _, productID := range productIDs {
            productFields := map[string]any{"id": productID}
            productIter, err := productTx.Search(&productFields)
            if err != nil {
                return err
            }
            defer engine.GlobalTableIterPool.Put(productIter)
            
            productRecords := productIter.GetRecords(true)
            defer record.PutRecords(productRecords)
            if len(productRecords) == 0 {
                return fmt.Errorf("product %d not found", productID)
            }
            
            productRecord := productRecords[0]
            stock, ok := productRecord["stock"].(int)
            if !ok {
                return fmt.Errorf("invalid stock type")
            }
            
            if stock <= 0 {
                return fmt.Errorf("product %d out of stock", productID)
            }
        }
        
        // 3. 创建订单
        orderTx := transactions[orderTable]
        order := map[string]any{
            "id":      orderID,
            "user_id": userID,
            "amount":  amount,
            "status":  "completed",
        }
        _, err = orderTx.Insert(&order)
        if err != nil {
            return err
        }
        
        // 4. 扣减用户余额
        userRecordMap := map[string]any(userRecord)
        userRecordMap["balance"] = userBalance - amount
        if err := userTx.Update(&userRecordMap); err != nil {
            return err
        }
        
        // 5. 扣减产品库存
        for _, productID := range productIDs {
            productFields := map[string]any{"id": productID}
            productIter, err := productTx.Search(&productFields)
            if err != nil {
                return err
            }
            defer engine.GlobalTableIterPool.Put(productIter)
            
            productRecords := productIter.GetRecords(true)
            defer record.PutRecords(productRecords)
            if len(productRecords) == 0 {
                return fmt.Errorf("product %d not found", productID)
            }
            
            productRecord := productRecords[0]
            productRecordMap := map[string]any(productRecord)
            stock, _ := productRecord["stock"].(int)
            productRecordMap["stock"] = stock - 1
            if err := productTx.Update(&productRecordMap); err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

#### 示例3：并发事务

```go
// 并发事务示例：多线程转账
func testConcurrentTransfers() {
    var wg sync.WaitGroup
    var mutex sync.Mutex
    var errors []error
    
    // 启动10个并发转账
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(transferID int) {
            defer wg.Done()
            
            // 随机选择转出和转入账户
            fromID := 1 + transferID%3
            toID := 1 + (transferID+1)%3
            amount := 10.0
            
            err := transferFunds(fromID, toID, amount)
            if err != nil {
                mutex.Lock()
                errors = append(errors, fmt.Errorf("transfer %d failed: %v", transferID, err))
                mutex.Unlock()
                return
            }
            
            fmt.Printf("Transfer %d: %d -> %d, amount: %.2f\n", transferID, fromID, toID, amount)
        }(i)
    }
    
    wg.Wait()
    
    // 打印错误
    if len(errors) > 0 {
        fmt.Println("\nErrors:")
        for _, err := range errors {
            fmt.Println(err)
        }
    } else {
        fmt.Println("\nAll transfers completed successfully!")
    }
    
    // 验证最终余额
    fmt.Println("\nFinal balances:")
    iter, err := accountTable.Search(nil)
    if err != nil {
        panic(err)
    }
    defer engine.GlobalTableIterPool.Put(iter)
    
    records := iter.GetRecords(true)
    defer record.PutRecords(records)
    
    var totalBalance float64
    for _, r := range records {
        balance := r["balance"].(float64)
        fmt.Printf("Account %s: %.2f\n", r["name"], balance)
        totalBalance += balance
    }
    
    fmt.Printf("Total balance: %.2f\n", totalBalance)
}
```