# sfsDb 数据库详细使用文档

## 1. 数据库初始化

### 1.1 自定义数据库路径

默认情况下，sfsDb会使用当前目录下的`kvdb`文件夹作为数据库存储路径。如果需要自定义数据库路径，可以在程序启动时调用`OpenDefaultDb`函数：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 自定义数据库路径
    dbPath := "./my_custom_db"
    _, err := storage.OpenDefaultDb(dbPath)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("数据库初始化成功")
}
```

### 1.2 自动使用默认路径

如果不调用`OpenDefaultDb`函数，系统会在首次创建表时自动使用默认路径`./kvdb`：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 首次调用TableNew时，会自动初始化数据库，使用默认路径"./kvdb"
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("表创建成功，数据库已自动初始化")
}
```

### 1.3 数据库关闭

程序结束时，可以调用`CloseDb`函数关闭数据库，释放资源：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 程序结束时关闭数据库
    defer storage.CloseDb()
    
    fmt.Println("表创建成功")
}
```

## 2. 创建表与设置字段

### 2.1 基本表创建/打开

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 设置表字段
    fields := map[string]any{
        "id":   0,     // 自动增值主键（使用数字类型）
        "name": "",    // 字符串类型
        "age":  0,     // 整数类型
        "email": "",   // 字符串类型
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("表创建成功")
}
```

### 2.2 支持的字段类型

sfsDb 使用 `any` 类型（即 `interface{}`）支持所有 Go 语言类型。下表列出了常用的类型示例：

| 类型示例 | 描述 |
|---------|------|
| `0`     | 整数类型（int、int32、int64、uint、uint32、uint64 等） |
| `0.0`   | 浮点类型（float32、float64） |
| `""`    | 字符串类型 |
| `false` | 布尔类型 |
| `time.Now()` | 时间类型 |
| `complex(1, 2)` | 复数类型（complex64、complex128） |

#### 2.2.1 自定义类型处理

对于未在表中明确列出的类型（如自定义结构体、切片、映射等），sfsDb 会默认使用 JSON 序列化进行处理：

- **存储时**：通过 `json.Marshal()` 将自定义类型序列化为 JSON 字符串，然后存储为字节数组
- **读取时**：通过 `json.Unmarshal()` 将 JSON 字符串反序列化为相应类型

**示例：使用自定义结构体作为字段类型**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

// 自定义结构体类型
type Address struct {
    City    string `json:"city"`
    Street  string `json:"street"`
    ZipCode string `json:"zip_code"`
}

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./custom_type_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // 设置字段，包含自定义结构体类型
    fields := map[string]any{
        "id":      0,
        "name":    "",
        "age":     0,
        "address": Address{}, // 使用自定义结构体类型
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 插入带有自定义结构体的记录
    user := map[string]any{
        "name": "张三",
        "age":  30,
        "address": Address{
            City:    "北京",
            Street:  "朝阳区建国路",
            ZipCode: "100022",
        },
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("插入成功，ID: %d\n", id)

    // 读取记录
    readFields := map[string]any{"id": id}
    recordData, err := table.Read(&readFields)
    if err != nil {
        panic(err)
    }
    fmt.Printf("读取到的记录: %v\n", recordData)
}
```

#### 2.2.2 类型转换机制

sfsDb 在内部使用 `util` 包中的类型转换函数处理不同类型：

- `AnyToBytes()`：将任意类型转换为字节数组，用于存储
- `AnyToStr()`：将任意类型转换为字符串，用于显示或比较
- `StrToAny()`：将字符串转换为指定类型，用于读取

对于未明确支持的类型，这些函数都会使用 JSON 序列化/反序列化进行处理，确保所有 Go 语言类型都能被正确存储和读取。

#### 2.2.3 注意事项

1. **自定义类型必须可序列化**：用于字段的自定义类型必须能够被 JSON 序列化，否则会导致存储失败
2. **性能考虑**：复杂自定义类型的 JSON 序列化和反序列化会带来一定的性能开销
3. **类型安全**：读取自定义类型时，需要确保类型匹配，否则可能导致运行时错误
4. **字段扩展**：可以随时向表中添加新字段，无需修改表结构
5. **默认值**：字段的默认值用于类型推断，实际存储时会使用插入的数据

通过这种设计，sfsDb 实现了真正的半结构化数据存储，既支持常见的基本类型，也能灵活处理复杂的自定义类型。

### 2.3 示例：创建包含各种类型字段的表

```go
// 创建包含多种字段类型的表
complexTable, err := engine.TableNew("complex_table")
if err != nil {
    panic(err)
}

complexFields := map[string]any{
    "id":        0,          // 整数类型（主键）
    "name":      "",         // 字符串类型
    "age":       0,          // 整数类型
    "salary":    0.0,        // 浮点类型
    "active":    false,      // 布尔类型
    "created_at": time.Now(), // 时间类型
    "description": "",       // 字符串类型（用于全文搜索）
    "tags":      "",         // 字符串类型（用于标签）
}
err = complexTable.SetFields(complexFields)
if err != nil {
    panic(err)
}
```

## 3. 插入数据

### 3.1 基本插入

```go
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
```

### 2.2 自动增值

主键为id的字段，缺省时，sfsDb 会自动生成递增的值：

```go
// 不指定 ID 字段，系统自动生成
user2 := map[string]any{
    "name": "李四",
    "age":  25,
    "email": "lisi@example.com",
}
id2, err := table.Insert(&user2)
if err != nil {
    panic(err)
}
fmt.Printf("自动生成 ID: %d\n", id2) // 输出: 2

// 也可以显式设置 ID 为 nil，同样会自动生成
user3 := map[string]any{
    "id":   nil, // 显式设置为 nil，同样会自动生成
    "name": "王五",
    "age":  30,
    "email": "wangwu@example.com",
}
id3, err := table.Insert(&user3)
if err != nil {
    panic(err)
}
fmt.Printf("显式设置为 0 自动生成 ID: %d\n", id3) // 输出: 3
```

### 2.3 批量插入

```go
// 批量插入多条记录
users := []map[string]any{
    {"name": "王五", "age": 35, "email": "wangwu@example.com"},
    {"name": "赵六", "age": 28, "email": "zhaoliu@example.com"},
    {"name": "孙七", "age": 40, "email": "sunqi@example.com"},
}

for _, user := range users {
    _, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
}
fmt.Println("普通批量插入完成")

// 手动事务批量插入（更高效的批量操作）
fmt.Println("\n手动事务批量插入:")

// 1. 获取批量操作对象（通过全局storage.KVDb获取）
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 2. 准备要插入的数据
batchUsers := []map[string]any{
    {"name": "周八", "age": 26, "email": "zhouba@example.com"},
    {"name": "吴九", "age": 32, "email": "wujiu@example.com"},
    {"name": "郑十", "age": 38, "email": "zhengshi@example.com"},
}

// 3. 将多个插入操作添加到同一个批处理中
for i, user := range batchUsers {
    _, err := table.Insert(&user, batch) // 传递batch参数，手动控制事务
    if err != nil {
        panic(fmt.Sprintf("插入第%d条记录失败: %v", i+1, err))
    }
    fmt.Printf("已添加第%d条记录到批处理\n", i+1)
}

// 4. 手动提交批处理（所有操作一次性执行）
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("批量提交失败: %v", err))
}
fmt.Println("手动事务批量插入完成")

// 5. 验证插入结果
fmt.Println("\n验证插入结果:")
allIter := table.ForData()
defer allIter.Release()
records := allIter.GetRecords(true)
fmt.Printf("表中共有%d条记录\n", len(records))

// 手动事务批量操作支持多种组合
fmt.Println("\n组合操作示例（插入+删除）:")
batch2 := storage.KVDb.GetBatch()
if batch2 == nil {
    panic("无法获取批量操作对象")
}

// 添加一条新记录
newUser := map[string]any{
    "name": "测试用户",
    "age":  25,
    "email": "test@example.com",
}
_, err = table.Insert(&newUser, batch2)
if err != nil {
    panic(err)
}

// 删除一条现有记录
deleteUser := map[string]any{
    "id": 1, // 删除ID为1的记录
}
err = table.Delete(&deleteUser, batch2)
if err != nil {
    panic(err)
}

// 提交组合操作
err = storage.KVDb.WriteBatch(batch2)
if err != nil {
    panic(fmt.Sprintf("组合操作提交失败: %v", err))
}
fmt.Println("组合操作完成")
```

## 3. 主键管理

### 3.1 单主键

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
```

### 3.2 复合主键

```go
// 创建复合主键表
orderTable, err := engine.TableNew("order_table")
if err != nil {
    panic(err)
}

// 设置复合主键字段
orderFields := map[string]any{
    "order_id": 0,     // 复合主键字段1
    "user_id":  0,     // 复合主键字段2
    "product":  "",    // 普通字段
    "quantity": 0,     // 普通字段
}
err = orderTable.SetFields(orderFields)
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

// 插入复合主键记录
order := map[string]any{
    "order_id": 1,
    "user_id":  100,
    "product":  "商品A",
    "quantity": 2,
}
_, err = orderTable.Insert(&order)
if err != nil {
    panic(err)
}
```

### 3.3 主键自动创建机制

当表中没有显式创建主键时，sfsDb 会在首次需要主键时自动创建一个默认主键。这个机制由 `GetPrimaryKey()` 方法实现：

```go
// 获取主键索引
func (t *Table) GetPrimaryKey() PrimaryKey {
    pk := t.indexs.getPrimaryKey()
    if pk == nil {
        //没有主键，需要创建一个默认主键。
        //所以主键必须在表未有数据前创建。
        pk, _ = DefaultPrimaryKeyNew("id")
        pk.AddFields("id")
        t.CreateIndex(pk)
    }
    return pk
}
```

**工作原理**：
1. 当调用 `GetPrimaryKey()` 方法时，系统会首先检查表是否已有主键
2. 如果没有主键，系统会自动创建一个名为 "id" 的默认主键
3. 默认主键会添加 "id" 字段作为主键字段
4. 然后将这个默认主键索引添加到表中
5. 最后返回创建的主键索引

**注意事项**：
- 系统注释明确指出："所以主键必须在表未有数据前创建。"
- 这意味着如果表中已经有数据，再调用需要主键的操作时，可能会导致数据不一致
- 因此，建议在表创建后、插入数据前，显式设置主键

**使用场景**：
- 快速原型开发，暂时不需要自定义主键
- 简单应用，使用默认的自增 ID 作为主键即可满足需求

**索引创建顺序的重要性**：
在添加索引时，系统会执行 `createIndexData` 函数来为现有数据创建索引数据，该函数会调用 `GetPrimaryKey()` 方法：

```go
// 对table现有数据创建对应的新索引数据
func (t *Table) createIndexData(index Index) error {
    //排除主键
    if index == t.GetPrimaryKey() {
        return nil
    }
    //根据主键前缀遍历所有数据，创建索引数据
    //获取主键前缀
    pk := t.GetPrimaryKey() // 这里会触发自动创建主键
    pkPrefix := pk.Prefix(t.id)
    slice := util.NewRangeHelper(pkPrefix).FromComparison(util.Like, pkPrefix)
    //pkPrefix创建迭代器
    iter := t.kvStore.Iterator(slice.Start, slice.Limit)
    defer iter.Release()
    //遍历所有数据
    var value []byte
    for iter.Next() {
        _, value = iter.Key(), iter.Value()
        rval, err := pk.Parse(t.fieldsid, value)
        if err != nil {
            return err
        }
        if rval == nil {
            continue
        }
        indexValue := pk.GetID(rval)
        switch index := index.(type) {
        case NormalIndex:
            idxvalueofkey := index.Join(rval)
            idxkey := index.JoinPrefix(t.id, idxvalueofkey)
            t.kvStore.Put(idxkey, indexValue)
        case FullTextIndex:
            //func (c *batchContainer) Operation 函数基本一样
            fieldsBytes, err := pk.Parse(t.fieldsid, value)
            if err != nil {
                return err
            }
            if fieldsBytes == nil {
                continue
            }
            joinValues := index.JoinFullValues(fieldsBytes, t.id)
            defer util.PutBytesArray(joinValues)
            for _, joinValue := range joinValues {
                if joinValue == nil {
                    continue
                }
                // 为全文索引创建索引数据
                t.kvStore.Put(joinValue, indexValue)
            }
        }
    }
    return nil
}
```

**重要建议**：
1. **先添加主键，后添加其他索引**：如果在添加其他索引之前没有显式创建主键，系统会在执行 `createIndexData` 时自动创建一个默认主键
2. **避免自动创建主键**：自动创建的主键可能不符合你的业务需求，建议始终显式创建主键
3. **在数据插入前设置主键**：确保在插入任何数据之前设置好主键，以避免数据不一致

## 4. 索引管理

### 4.1 索引接口概述

sfsDb 定义了一套完整的索引接口体系，为不同类型的索引提供了统一的操作方法。

#### 4.1.1 基础索引接口

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

#### 4.1.2 索引类型层次结构

sfsDb 支持三种主要索引类型，每种类型都有对应的接口和实现：

1. **主键索引 (PrimaryKey)** - 用于唯一标识记录
2. **普通索引 (NormalIndex)** - 用于加速查询
3. **全文索引 (FullTextIndex)** - 用于文本搜索

### 4.2 主键索引

#### 4.2.1 主键索引接口

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

#### 4.2.2 默认主键索引实现

```go
// 默认主键索引
// 组合主键时只支持固定长度的类型的组合。
// 字符串类型，必须指定长度。否则无法解析或存在转义问题导致bug。
type DefaultPrimaryKey struct {
    BaseIndex // 嵌入基础索引
}

func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error)
```

#### 4.2.3 使用示例

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

### 4.3 普通索引

#### 4.3.1 普通索引接口

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

#### 4.3.2 默认普通索引实现

```go
// 默认普通索引，二级索引
type DefaultNormalIndex struct {
    BaseIndex // 嵌入基础索引
}

func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error)

// Tag方法返回true，表示是二级索引。
func (dni *DefaultNormalIndex) Tag() bool
```

#### 4.3.3 使用示例

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

### 4.4 全文索引

#### 4.4.1 全文索引接口

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

#### 4.4.2 默认全文索引实现

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

#### 4.4.3 使用示例

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
fullTextIndex.AddFields("description", "id", "did") //"id"，“did”是组合主键
// 设置全文索引字段和长度
fullTextIndex.SetFullField("description", 5) // 5表示全文索引的长度。中文一般选择5或7。主要是支持象形文字。非象形文字该全文索引不太合适。
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

### 4.5 示例：创建包含多种索引的表

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

### 4.6 索引实现细节

#### 4.6.1 基础索引结构

所有索引类型都嵌入了 `BaseIndex` 结构体，提供了通用的索引功能：

```go
// 基础索引结构体，包含所有索引类型共有的字段和方法
type BaseIndex struct {
    fields []string
    id     uint8
    name   string
}
```

#### 4.6.2 索引字段处理

- **添加字段**: `AddFields(field ...string)` - 向索引添加一个或多个字段
- **获取字段**: `GetFields() []string` - 获取索引的所有字段
- **更新字段名称**: `UpdateFields(oldfields string, newfields string)` - 修改索引字段名称
- **删除字段**: `DeleteFields(field ...string)` - 从索引中删除字段

#### 4.6.3 索引值拼接

索引内部使用字节数组来存储索引值，提供了多种拼接方法：

- **Prefix**: 拼接索引前缀
- **Join**: 拼接字段值
- **JoinPrefix**: 拼接前缀和值
- **JoinValue**: 拼接完整的索引值

#### 4.6.4 复合索引注意事项

1. **组合主键限制**: 组合主键只支持固定长度类型的组合，字符串类型必须指定长度
2. **字段顺序**: 索引字段的顺序会影响查询性能，应将最常用的字段放在前面
3. **索引大小**: 索引会增加存储开销，应只创建必要的索引

### 4.7 索引管理系统

#### 4.7.1 Indexs 结构体

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

#### 4.7.2 表级索引管理方法

`sfsDb` 提供了丰富的表级索引管理方法，方便用户创建和管理索引：

##### 创建索引

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

##### 获取索引

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

##### 删除索引

**重要说明**：删除索引不会删除现存数据，不会对数据产生影响，只是存在冗余。

```go
// 删除指定名称的索引
//删除索引不会删除现存数据，不会对数据产生影响，只是存在冗余。
func (t *Table) DropIndex(name string) error

// 删除主键索引
func (t *Table) DropPrimaryKey() error
```

#### 4.7.3 索引管理示例

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

### 4.8 索引优化建议

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

通过合理使用和优化索引，可以显著提高 sfsDb 的查询性能，特别是在处理大量数据时。

## 5. 全文搜索

### 5.1 创建全文索引

```go
// 创建全文索引
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// 添加字段，最后一个字段必须是主键
fullTextIndex.AddFields("description", "id")
// 设置全文索引字段和长度
fullTextIndex.SetFullField("description", 5) // 5表示全文索引的长度
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

### 5.2 插入带全文索引的记录

```go
// 插入带全文索引的记录
contentRecords := []map[string]any{
    {"id": 1, "name": "商品1", "description": "这是一款高性能的笔记本电脑，适合编程和游戏"},
    {"id": 2, "name": "商品2", "description": "智能手机，拥有强大的摄像头和长续航电池"},
    {"id": 3, "name": "商品3", "description": "无线耳机，提供沉浸式音频体验"},
    {"id": 4, "name": "商品4", "description": "智能手表，可监测健康数据和接收通知"},
}

for _, record := range contentRecords {
    _, err = table.Insert(&record)
    if err != nil {
        panic(err)
    }
}
```

### 5.3 执行全文搜索

```go
// 全文搜索示例
fmt.Println("=== 全文搜索示例 ===")

// 搜索包含"笔记本"的记录
fmt.Println("\n1. 搜索 '笔记本':")
search1 := map[string]any{"description": "笔记本"}
iter1 := table.Search(&search1)
defer iter1.Release()
records1 := iter1.GetRecords(true)
for _, record := range records1 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 搜索包含"智能"的记录
fmt.Println("\n2. 搜索 '智能':")
search2 := map[string]any{"description": "智能"}
iter2 := table.Search(&search2)
defer iter2.Release()
records2 := iter2.GetRecords(true)
for _, record := range records2 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 3. 搜索结果字段选择示例
fmt.Println("\n3. 搜索结果字段选择:")
search3 := map[string]any{"description": "智能"}
iter3 := table.Search(&search3)
defer iter3.Release()
records3 := iter3.GetRecords(true)

// 使用 Select 方法只选择 name 字段
selectedNames := records3.Select("name")
fmt.Println("只显示匹配记录的名称:")
for _, record := range selectedNames {
    fmt.Printf("   - %s\n", record["name"])
}
```

## 6. 查询数据

### 6.1 基本搜索

```go
// 基本搜索：精确匹配
searchFields := map[string]any{
    "name": "张三",
}
iter := table.Search(&searchFields)
defer iter.Release()

// 获取所有匹配记录
records := iter.GetRecords(true)
for _, record := range records {
    fmt.Printf("找到记录: %v\n", record)
}
```

### 6.2 使用比较操作符

sfsDb 支持多种比较操作符，位于 `util` 包中：

| 操作符 | 描述 | 示例 |
|-------|------|------|
| `Equal` | 等于 | `util.Equal` |
| `NotEqual` | 不等于 | `util.NotEqual` |
| `GreaterThan` | 大于 | `util.GreaterThan` |
| `GreaterThanOrEqual` | 大于等于 | `util.GreaterThanOrEqual` |
| `LessThan` | 小于 | `util.LessThan` |
| `LessThanOrEqual` | 小于等于 | `util.LessThanOrEqual` |
| `Like` | 前缀匹配（类似 SQL LIKE） | `util.Like` |

### 6.3 比较操作符使用示例

```go
import (
    "github.com/liaoran123/sfsDb/util"
)

// 示例：使用比较操作符搜索

// 1. 搜索年龄大于30的用户
fmt.Println("\n年龄大于30的用户:")
ageGt30 := map[string]any{
    "age": 30,
}
iterGt30 := table.Search(&ageGt30, util.GreaterThan) // 传递比较操作符作为第二个参数
defer iterGt30.Release()
recordsGt30 := iterGt30.GetRecords(true)
for _, record := range recordsGt30 {
    fmt.Printf("   - %s: %d岁\n", record["name"], record["age"])
}

// 2. 前缀搜索（Like操作符）- 默认就是Like操作
fmt.Println("\n邮箱以'user'开头的用户:")
emailPrefix := map[string]any{
    "email": "user",
}
iterPrefix := table.Search(&emailPrefix) // 默认使用util.Like操作符，这里的like实则是前缀匹配
defer iterPrefix.Release()
recordsPrefix := iterPrefix.GetRecords(true)
for _, record := range recordsPrefix {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 3. 显式使用Like操作符
fmt.Println("\n名字以'张'开头的用户:")
namePrefix := map[string]any{
    "name": "张",
}
iterName := table.Search(&namePrefix, util.Like) // 显式指定util.Like操作符
defer iterName.Release()
recordsName := iterName.GetRecords(true)
for _, record := range recordsName {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 4. 精确匹配搜索
fmt.Println("\n精确搜索名字为'张三'的用户:")
exactSearch := map[string]any{
    "name": "张三",
}
iterExact := table.Search(&exactSearch, util.Equal) // 显式指定util.Equal操作符
defer iterExact.Release()
recordsExact := iterExact.GetRecords(true)
for _, record := range recordsExact {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 5. 不等于搜索
fmt.Println("\nid不等于1的用户:")
notEqualSearch := map[string]any{
    "id": 1,
}
iterNotEqual := table.Search(&notEqualSearch, util.NotEqual) // 使用util.NotEqual操作符
defer iterNotEqual.Release()
recordsNotEqual := iterNotEqual.GetRecords(true)
for _, record := range recordsNotEqual {
    fmt.Printf("   - %s: ID=%d\n", record["name"], record["id"])
}
```

### 6.4 匹配器接口（Match Interface）

sfsDb 支持自定义匹配器接口，用于实现复杂的查询逻辑：

```go
// Match 接口定义
type Match interface {
    Match(record map[string]any) bool
}

// 自定义匹配器示例：年龄大于指定值
 type AgeGreaterThanMatcher struct {
    MinAge int
}

func (m *AgeGreaterThanMatcher) Match(record map[string]any) bool {
    if age, ok := record["age"].(int); ok {
        return age > m.MinAge
    }
    return false
}

// 使用自定义匹配器
matcher := &AgeGreaterThanMatcher{MinAge: 30}
records := table.MatchRecords(matcher)
for _, record := range records {
    fmt.Printf("匹配记录: %v\n", record)
}
```

### 6.5 AND 匹配器

sfsDb 提供了内置的 `AND` 匹配器，用于实现类似 SQL 中的 IN、NOT IN、AND、OR 等操作，特别适合多表连接查询。

#### 6.5.1 AND 匹配器概述

`AND` 匹配器用于判断记录的字段值是否在指定的数据集合中，支持正向匹配（IN）和反向匹配（NOT IN）。

**核心功能**：
- 实现类似 SQL 的 IN 和 NOT IN 操作
- 支持多字段组合匹配
- 适合多表连接查询
- 支持与其他匹配器组合使用

#### 6.5.2 AND 结构体定义

```go
type AND struct {
    fields []string // 需要匹配的字段名
    data   map[any]bool // 匹配数据集合
    rule   bool // 匹配规则：true=IN/AND，false=NOT IN/OR
}
```

**字段说明**：
- `fields`：需要匹配的字段名列表，与记录中的 key 对应
- `data`：匹配数据集合，由 `TableIter.Map()` 方法生成或自定义
- `rule`：匹配规则，`true` 表示 IN/AND，`false` 表示 NOT IN/OR

#### 6.5.3 构造函数

```go
func NewAND(fields []string, data map[any]bool, rule ...bool) *AND
```

**参数说明**：
- `fields`：需要匹配的字段名列表
- `data`：匹配数据集合
- `rule`：可选参数，匹配规则，默认为 `true`

#### 6.5.4 使用示例

**示例 1：基本 IN 操作**

```go
// 假设我们有一个用户表，需要查询 ID 在指定集合中的用户

// 1. 获取 ID 集合
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. 创建 AND 匹配器
andMatcher := match.NewAND([]string{"id"}, idMap)

// 3. 使用匹配器
iter := table.Search(&map[string]any{"id": nil})
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)

// 结果：返回 ID 为 1、3、5 的用户
```

**示例 2：NOT IN 操作**

```go
// 查询 ID 不在指定集合中的用户

// 1. 获取 ID 集合
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. 创建 AND 匹配器，设置 rule=false 表示 NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 3. 使用匹配器
iter := table.Search(&map[string]any{"id": nil})
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)

// 结果：返回 ID 不为 1、3、5 的用户
```

**示例 3：多表连接查询（重点示例）**

```go
// 实现类似 SQL 的连接查询：SELECT table1.* FROM table1, table2 WHERE table1.id = table2.id

// 1. 获取两个表的迭代器
iter1 := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2 := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

// 2. 获取 table2 的 ID 映射
// Map() 方法生成 map[any]bool，键为指定字段的值
idMap := iter2.Map()

// 3. 创建 AND 匹配器
// 匹配 table1 的 id 字段是否在 table2 的 id 集合中
andMatcher := match.NewAND([]string{"id"}, idMap)

// 4. 设置匹配器并获取结果
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)

// 结果：返回 table1 中 ID 与 table2 中 ID 匹配的记录
```

**示例 4：多表连接的 NOT 操作**

```go
// 实现类似 SQL 的连接查询：SELECT table1.* FROM table1, table2 WHERE table1.id != table2.id

// 1. 获取两个表的迭代器
iter1 := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2 := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

// 2. 获取 table2 的 ID 映射
idMap := iter2.Map()

// 3. 创建 AND 匹配器，设置 rule=false 表示 NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 4. 设置匹配器并获取结果
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)

// 结果：返回 table1 中 ID 与 table2 中 ID 不匹配的记录
```

#### 6.5.5 组合主键匹配

`AND` 匹配器主要用于主键匹配，特别是组合主键的情况。当 `fields` 参数包含多个字段名时，它会通过 `util.MergeFields()` 将这些字段值合并为一个值进行匹配，这正是处理组合主键的典型方式：

```go
// 示例：组合主键匹配，假设表有组合主键 (user_id, product_id)

// 1. 获取另一个表的组合主键映射
// 假设iter2是包含组合主键的表的迭代器
// Map("user_id", "product_id") 生成组合主键值的映射
combinedKeyMap := iter2.Map("user_id", "product_id")

// 2. 创建 AND 匹配器
// 匹配当前表的组合主键 (user_id, product_id) 是否在另一个表的组合主键集合中
andMatcher := match.NewAND([]string{"user_id", "product_id"}, combinedKeyMap)

// 3. 使用匹配器
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)

// 结果：返回组合主键与另一个表匹配的记录
```

**注意事项**：
- AND 结构体主要设计用于主键值的匹配，特别是多个迭代器之间的主键匹配
- 当 `fields` 参数包含多个字段时，它会将这些字段视为组合主键进行处理
- 合并后的字段值格式取决于 `util.MergeFields()` 函数的实现
- 对于非主键字段的组合匹配，建议使用其他匹配器（如 FieldComparison）或自定义匹配器

**AND 匹配器的设计初衷**：
```go
/*
//该结构作用是多个迭代器的主键值进行相同或不相同的匹配
//rule为true时，多个迭代器的主键值必须相同，才匹配成功
//rule为false时，多个迭代器的主键值必须不同，才匹配成功
//rule为false的作用主要是用在跳跃查询中，比如sql语句中 field not in (1,2,3)，则data=map[any]bool{1:true,2:true,3:true}
*/
```

#### 6.5.6 与其他匹配器组合使用

`AND` 匹配器可以与其他匹配器组合使用，实现更复杂的查询逻辑：

```go
// 实现：SELECT * FROM table WHERE id IN (1,3,5) AND age > 25

// 1. 创建 AND 匹配器（ID IN (1,3,5)）
idMap := map[any]bool{1: true, 3: true, 5: true}
idMatcher := match.NewAND([]string{"id"}, idMap)

// 2. 创建 AgeGreaterThanMatcher（age > 25）
ageMatcher := &AgeGreaterThanMatcher{MinAge: 25}

// 3. 使用组合匹配器
iter := table.Search(&map[string]any{"id": nil})
defer iter.Release()

// 设置多个匹配器，它们之间是 AND 关系
iter.SetMatch(idMatcher, ageMatcher)
records := iter.GetRecords(true)

// 结果：返回 ID 为 1、3、5 且年龄大于 25 的用户
```

#### 6.5.7 AND 匹配器的优势

1. **高效的多表连接**：通过预计算的映射表，避免了嵌套循环，提高了连接查询的效率
2. **灵活的匹配规则**：支持 IN、NOT IN、AND、OR 等多种匹配方式
3. **支持多字段组合**：可以根据多个字段的组合值进行匹配
4. **易于与其他匹配器组合**：可以与其他自定义或内置匹配器组合使用
5. **适合处理复杂查询场景**：特别适合需要关联多个表或集合的查询场景

通过 `AND` 匹配器，sfsDb 实现了高效灵活的多表连接查询功能，为用户提供了强大的数据查询能力。

### 6.5 FieldComparison 匹配器

sfsDb 提供了内置的 `FieldComparison` 匹配器，用于实现各种比较操作，支持将匹配器应用于迭代器，特别适合主键迭代器对无索引字段的匹配：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/match"
)

func main() {
    // 创建测试表
    table, err := engine.TableNew("test_field_comparison")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "score":  0.0,
        "active": false,
    }

    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 创建主键索引
    pkIndex, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        panic(err)
    }
    pkIndex.AddFields("id")
    err = table.CreateIndex(pkIndex)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    testData := []map[string]any{
        {"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
        {"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
        {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
        {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
        {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
    }

    for _, record := range testData {
        _, err = table.Insert(&record)
        if err != nil {
            panic(err)
        }
    }

    // FieldComparison 匹配器示例
    fmt.Println("=== FieldComparison 匹配器示例 ===")

    // 1. 使用 FieldComparison 进行比较
    // 获取迭代器
    iter := table.Search(&map[string]any{"id": nil})
    defer iter.Release()

    // 创建 FieldComparison 匹配器
    matcher := match.NewFieldComparison("age", match.GreaterThan, 25)

    // 将匹配器设置到迭代器上
    iter.SetMatch(matcher)

    // 获取过滤后的记录
    records := iter.GetRecords(true)
    fmt.Printf("年龄大于25的记录 (%d 条):\n", len(records))
    for _, record := range records {
        fmt.Printf("   - %v\n", record)
    }

    // 2. 使用便捷函数创建匹配器
    iter2 := table.Search(&map[string]any{"id": nil})
    defer iter2.Release()

    // 使用 GreaterThanMatch 便捷函数
    highScoreMatcher := match.NewGreaterThanMatch("score", 90.0)
    iter2.SetMatch(highScoreMatcher)

    highScoreRecords := iter2.GetRecords(true)
    fmt.Printf("\n分数大于90的记录 (%d 条):\n", len(highScoreRecords))
    for _, record := range highScoreRecords {
        fmt.Printf("   - %v\n", record)
    }

    // 3. 使用 EqualMatch 便捷函数
    iter3 := table.Search(&map[string]any{"id": nil})
    defer iter3.Release()

    inactiveMatcher := match.NewEqualMatch("active", false)
    iter3.SetMatch(inactiveMatcher)

    inactiveRecords := iter3.GetRecords(true)
    fmt.Printf("\n非活跃用户 (%d 条):\n", len(inactiveRecords))
    for _, record := range inactiveRecords {
        fmt.Printf("   - %v\n", record)
    }
}
```

### 6.6 FieldComparison 支持的比较操作

| 比较操作 | 描述 | 便捷函数 |
|---------|------|---------|
| `Equal` | 等于 | `NewEqualMatch` |
| `NotEqual` | 不等于 | `NewNotEqualMatch` |
| `GreaterThan` | 大于 | `NewGreaterThanMatch` |
| `GreaterThanOrEqual` | 大于等于 | `NewGreaterThanOrEqualMatch` |
| `LessThan` | 小于 | `NewLessThanMatch` |
| `LessThanOrEqual` | 小于等于 | `NewLessThanOrEqualMatch` |
| `Like` | 前缀匹配 | `NewLikeMatch` |
| `Prefix` | 前缀匹配 | `NewPrefixMatch` |
| `Suffix` | 后缀匹配 | `NewSuffixMatch` |
| `Contains` | 包含匹配 | `NewContainsMatch` |

## 7. 字段修改

### 7.1 单个字段修改

```go
// 字段修改工作流：将 "position" 字段重命名为 "job_title"

// 1. 首先调用 UpdateFieldName 更新字段名称映射
err = table.UpdateFieldName("position", "job_title")
if err != nil {
    panic(err)
}

// 2. 然后调用 SetFields 更新字段映射
updatedFields := map[string]any{
    "id":        0,
    "name":      "",
    "age":       0,
    "email":     "",
    "job_title": "", // 使用新的字段名
}
err = table.SetFields(updatedFields)
if err != nil {
    panic(err)
}
```

### 7.2 批量字段修改

```go
// 批量修改多个字段

// 1. 依次调用 UpdateFieldName 更新多个字段名称映射
err = table.UpdateFieldName("old_field1", "new_field1")
if err != nil {
    panic(err)
}
err = table.UpdateFieldName("old_field2", "new_field2")
if err != nil {
    panic(err)
}
err = table.UpdateFieldName("old_field3", "new_field3")
if err != nil {
    panic(err)
}

// 2. 调用 SetFields 更新完整的字段映射
newFields := map[string]any{
    "id":         0,
    "name":       "",
    "new_field1": "",
    "new_field2": 0,
    "new_field3": 0.0,
}
err = table.SetFields(newFields)
if err != nil {
    panic(err)
}
```

## 8. 删除记录

### 8.1 单个记录删除

```go
// 删除单个记录
recordToDelete := map[string]any{
    "id": 1, // 必须包含主键
}
err = table.Delete(&recordToDelete)
if err != nil {
    panic(err)
}
fmt.Println("记录删除成功")
```

### 8.2 批量删除

```go
// 批量删除记录：删除年龄小于25的用户

// 首先搜索符合条件的记录
searchCriteria := map[string]any{
    "age": map[util.ComparisonOperator]any{util.LessThan: 25},
}
iter := table.Search(&searchCriteria)
defer iter.Release()

// 使用迭代器的 Delete 方法批量删除符合条件的记录
// 直接删除所有符合条件的记录（无限制）
iter.Delete()

// 或者限制删除数量：删除前2条符合条件的记录
// iter.Delete(2)
```

### 8.3 删除表的所有数据

sfsDb 提供了 `DeleteAll()` 方法，用于删除表中的所有数据。这个方法会遍历表的所有键值对，并使用批量操作删除所有数据。

```go
// 删除表的所有数据
func (t *Table) DeleteAll() error {
	// 获取表的所有kv键值对迭代器
	iter := t.For()
	defer iter.Release()
	
	// 创建批量操作
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("failed to get batch")
	}
	
	// 定义批量操作的大小限制
	const batchSizeLimit = 1000
	
	// 遍历并删除所有键值对
	count := 0
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
		count++
		
		// 当批量操作的大小达到限制时，执行批量操作并重置批量操作对象
		if count >= batchSizeLimit {
			// 提交批量操作
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}
			
			// 重置计数器和批量操作对象
			count = 0
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return fmt.Errorf("failed to get batch")
			}
		}
	}
	
	// 执行剩余的批量操作
	if count > 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}
	
	return nil
}
```

**使用示例**：

```go
// 删除表的所有数据
err = table.DeleteAll()
if err != nil {
    panic(err)
}
fmt.Println("表数据删除成功")
```

**工作原理**：
1. `DeleteAll()` 方法首先调用 `For()` 方法获取表的所有键值对迭代器
2. 然后创建批量操作对象，用于批量删除键值对
3. 定义批量操作的大小限制（默认为 1000），当达到限制时执行批量操作并重置
4. 遍历迭代器，将每个键添加到批量操作中
5. 当批量操作达到限制时，执行批量操作并重置
6. 最后执行剩余的批量操作

**注意事项**：
- `DeleteAll()` 方法会删除表中的所有数据，包括索引数据
- 该方法使用分批处理机制，确保在处理大量数据时不会超过批量操作的大小限制
- 删除操作不会影响其他表的数据
- 删除后表仍然存在，只是数据被清空
- 可以在删除后重新向表中插入数据

## 9. 对象池与内存管理

### 9.1 核心概念

sfsDb 实现了高效的对象池机制，用于管理 Record 和 Records 对象的生命周期，减少内存分配和垃圾收集开销。核心特性包括：

- **自动对象复用**：通过 sync.Pool 实现对象的重用
- **自动内存释放**：使用 runtime.SetFinalizer 确保对象在不再使用时自动释放
- **内存优化**：预分配合理容量，减少扩容开销

### 9.2 核心 API

#### 9.2.1 Record 对象操作

```go
// 从对象池获取 Record 对象
record := record.GetRecord()

// 使用 Record 对象
record["id"] = 1
record["name"] = "张三"
record["age"] = 30

// 手动将 Record 对象放回对象池（可选）
record.PutRecord(record)
```

#### 9.2.2 Records 对象操作

```go
// 从对象池获取 Records 对象
records := record.GetRecords()

// 从对象池获取指定容量的 Records 对象
recordsWithCapacity := record.GetRecordsWithCapacity(100)

// 使用 Records 对象
records = append(records, record1)
records = append(records, record2)

// 手动将 Records 对象放回对象池（可选）
record.PutRecords(records)
```

#### 9.2.3 Map 对象操作

sfsDb 还实现了 `map[any]bool` 类型的对象池管理，用于优化 `TableIter.Map()` 方法返回的映射对象。

```go
// 从对象池获取 map[any]bool 对象
import "github.com/liaoran123/sfsDb/engine"

// 从对象池获取 map[any]bool 对象
m := engine.GetMap()

// 使用 map 对象
m[1] = true
m[2] = true
m[3] = true

// 手动将 map 对象放回对象池（可选）
engine.PutMap(m)
```

**核心实现**：

```go
// mapPool 用于管理 map[any]bool 类型的对象池
var mapPool = sync.Pool{
	New: func() any {
		return make(map[any]bool)
	},
}

// GetMap 从对象池获取一个 map[any]bool 对象
func GetMap() map[any]bool {
	m := mapPool.Get().(map[any]bool)
	// 清空 map 中的所有键值对，确保返回的数据干净
	for k := range m {
		delete(m, k)
	}
	// 创建一个指向 map 的指针，用于设置 finalizer
	mapPtr := &m
	// 确保在对象被垃圾回收时清除其中的键值对
	runtime.SetFinalizer(mapPtr, func(ptr *map[any]bool) {
		if *ptr != nil {
			// 清空 map 中的所有键值对，确保对象池中的对象始终是干净的
			for k := range *ptr {
				delete(*ptr, k)
			}
		}
	})
	return m
}

// PutMap 将 map[any]bool 对象归还到对象池
func PutMap(m map[any]bool) {
	// 清空 map 中的所有键值对，确保归还的对象干净
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}
```

**使用场景**：
- 主要用于 `TableIter.Map()` 方法返回的映射对象，用于与其他迭代器进行匹配
- 适合需要频繁创建和销毁大 `map` 的场景，如生成匹配数据集合
- 自动管理对象生命周期，减少内存分配和垃圾回收开销

#### 9.2.4 批量记录格式化

sfsDb 提供了批量记录格式化功能，用于高效处理多个记录的格式化操作，减少函数调用开销和内存分配。

##### 9.2.4.1 核心 API

```go
// 批量格式化多个记录
func (t *Table) BatchFormatRecords(records []*map[string][]byte) [][]byte
```

##### 9.2.4.2 使用示例

```go
// 准备批量数据
records := make([]*map[string][]byte, 100)
for i := 0; i < 100; i++ {
	recordData := map[string][]byte{
		"id":    []byte(string(rune('0' + i%10))),
		"name":  []byte("User " + string(rune('A' + i%26))),
		"age":   []byte(string(rune('0' + i%10)) + string(rune('0' + i%10))),
		"email": []byte("user" + string(rune('a' + i%26)) + "@example.com"),
	}
	records[i] = &recordData
}

// 批量格式化记录
formattedRecords := table.BatchFormatRecords(records)

// 使用格式化后的记录
for i, formatted := range formattedRecords {
    fmt.Printf("Record %d: %s\n", i, formatted)
}
```

**性能优势**：
- 减少函数调用开销，批量处理多个记录只需一次函数调用
- 内存分配优化，预分配结果切片容量
- 适合处理大量记录的场景，如批量导入数据

#### 9.2.5 批量插入功能

sfsDb 提供了批量插入功能，用于高效处理多条记录的插入操作，减少事务开销和网络往返，显著提高插入性能。

##### 9.2.5.1 核心 API

```go
// BatchInsert 批量插入多条记录
// records []*map[string]any 要插入的记录列表
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsert(records []*map[string]any, batchs ...storage.Batch) ([]int, error)

// BatchInsertWithSize 带批量大小控制的批量插入
// records []*map[string]any 要插入的记录列表
// batchSize int 每批处理的记录数量
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertWithSize(records []*map[string]any, batchSize int, batchs ...storage.Batch) ([]int, error)
```

##### 9.2.5.2 使用示例

**基本批量插入**：

```go
// 创建测试记录
records := []*map[string]any{
    &map[string]any{"name": "Alice", "age": 25, "description": "Software Engineer"},
    &map[string]any{"name": "Bob", "age": 30, "description": "Product Manager"},
    &map[string]any{"name": "Charlie", "age": 35, "description": "Designer"},
    &map[string]any{"name": "David", "age": 40, "description": "Developer"},
    &map[string]any{"name": "Eve", "age": 45, "description": "Manager"},
}

// 批量插入
table, err := engine.TableNew("test_batch")
if err != nil {
    panic(err)
}

// 设置字段
err = table.SetFields(map[string]any{"id": 0, "name": "", "age": 0, "description": ""})
if err != nil {
    panic(err)
}

// 创建主键
pk, _ := engine.DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
table.CreateIndex(pk)

// 批量插入
ids, err := table.BatchInsert(records)
if err != nil {
    panic(err)
}

fmt.Printf("批量插入成功，插入的记录ID: %v\n", ids)
```

**带批量大小控制的批量插入**：

```go
// 创建大量测试记录
recordCount := 1000
records := make([]*map[string]any, recordCount)
for i := 0; i < recordCount; i++ {
    records[i] = &map[string]any{
        "name": fmt.Sprintf("Record%d", i),
        "age":  20 + i%50,
    }
}

// 带批量大小控制的批量插入
batchSize := 100 // 每批处理100条记录
ids, err := table.BatchInsertWithSize(records, batchSize)
if err != nil {
    panic(err)
}

fmt.Printf("批量插入成功，插入了 %d 条记录\n", len(ids))
```

**性能优势**：
- 减少事务开销，批量操作只启动和提交一次事务
- 减少网络往返，特别适合远程数据库场景
- 内存分配优化，预分配ID列表和批量操作容器
- 自动增值ID批量处理，减少锁竞争
- 适合处理大量数据的插入，如数据迁移、批量导入等场景

**并发安全**：
- 批量插入操作在并发情况下是安全的
- 自动增值ID的分配是原子的，避免ID重复
- 支持多 goroutine 同时进行批量插入操作

**使用场景**：
- 数据迁移和导入
- 批量生成测试数据
- 高频写入场景，如日志记录
- 数据分析和处理后的数据存储
- 需要高效插入大量记录的任何场景
formattedRecords := table.BatchFormatRecords(records)

// 使用格式化后的记录
for i, record := range formattedRecords {
	fmt.Printf("Record %d: %s\n", i, string(record))
}
```

##### 9.2.4.3 性能优势

| 操作类型 | 单次格式化 | 批量格式化 | 性能提升 |
|---------|-----------|-----------|----------|
| 100条记录 | ~120μs | ~91μs | ~24% |
| 1000条记录 | ~1200μs | ~915μs | ~24% |

##### 9.2.4.4 适用场景

- **批量导入数据**：处理大量数据导入时，批量格式化可以显著提高性能
- **批量更新操作**：一次性处理多个记录的更新操作
- **数据转换**：需要将大量记录转换为特定格式时
- **高频处理**：对性能要求较高的场景

#### 9.2.5 动态分析与监控功能

sfsDb 提供了动态分析切换方案，可根据系统状态自动选择最佳的对象创建策略，平衡性能和系统稳定性。

##### 9.2.4.1 策略类型

| 策略类型 | 描述 | 适用场景 |
|---------|------|----------|
| 静态分类策略（默认） | 基于对象类型选择创建方式，不进行系统状态监控 | 大多数场景，追求最佳性能 |
| 动态分析策略 | 基于系统状态自动切换创建方式 | 高并发场景，负载波动大 |
| 混合策略 | 结合静态和动态策略的优点 | 生产环境，平衡性能和稳定性 |

##### 9.2.4.2 配置方法

```go
// 设置策略配置
import "github.com/liaoran123/sfsDb/engine"

// 配置静态策略（默认）
engine.SetStrategyConfig(engine.StrategyConfig{
	StrategyType:       engine.StrategyStatic,
	ConcurrencyThreshold: 1000,
	MemoryThreshold:    500, // MB
	RequestThreshold:   1000, // 每分钟请求数
})

// 配置动态策略
engine.SetStrategyConfig(engine.StrategyConfig{
	StrategyType:       engine.StrategyDynamic,
	ConcurrencyThreshold: 1000,
	MemoryThreshold:    500,
	RequestThreshold:   1000,
})

// 配置混合策略
engine.SetStrategyConfig(engine.StrategyConfig{
	StrategyType:       engine.StrategyHybrid,
	ConcurrencyThreshold: 1000,
	MemoryThreshold:    500,
	RequestThreshold:   1000,
})
```

##### 9.2.4.3 快速策略切换

```go
// 切换到静态策略
engine.SwitchToStaticStrategy()

// 切换到动态策略
engine.SwitchToDynamicStrategy()

// 切换到混合策略
engine.SwitchToHybridStrategy()

// 更新策略阈值
engine.UpdateStrategyThresholds(500, 300, 500)
```

##### 9.2.4.4 使用示例

```go
// 使用配置策略获取 []string 切片
import "github.com/liaoran123/sfsDb/engine"

// 确保使用静态策略（默认）
engine.SwitchToStaticStrategy()

// 获取切片
slice := engine.GetStringSliceWithConfigStrategy()

// 使用切片
slice = append(slice, "field1")
slice = append(slice, "field2")

// 归还切片
engine.PutStringSliceWithConfigStrategy(slice)

// 在高负载场景下切换到混合策略
engine.SwitchToHybridStrategy()
// 更新阈值以适应具体环境
engine.UpdateStrategyThresholds(500, 300, 500)

// 继续使用相同的接口
slice2 := engine.GetStringSliceWithConfigStrategy()
// 使用并归还...
```

##### 9.2.4.5 系统状态监控

动态策略和混合策略会监控以下系统状态指标：

- **并发数**：当前系统并发请求数
- **内存使用**：系统内存使用情况（MB）
- **请求频率**：单位时间内的请求数（每分钟）

当这些指标超过配置的阈值时，系统会自动调整对象创建策略，以优化性能和资源使用。

##### 9.2.4.6 性能分析

| 策略类型 | 性能表现 | 适用场景 |
|---------|---------|----------|
| 静态策略 | 最佳性能，无监控开销 | 低负载场景，性能优先 |
| 动态策略 | 监控开销较大，性能较差 | 需要详细系统状态数据的场景 |
| 混合策略 | 平衡性能和稳定性 | 生产环境，负载波动大 |

##### 9.2.4.7 最佳实践

1. **默认使用静态策略**：在大多数场景下，静态策略提供最佳性能
2. **高负载场景使用混合策略**：在高并发、负载波动大的场景下，混合策略可以自动调整
3. **调整阈值参数**：根据具体硬件环境和应用场景调整切换阈值
4. **监控系统状态**：在生产环境中，监控系统状态，了解策略切换的时机

通过灵活使用不同的策略类型，您可以根据具体场景优化 sfsDb 的性能和资源使用，获得最佳的系统表现。

### 9.3 自动释放机制

sfsDb 实现了智能的自动释放机制，即使忘记手动调用 PutRecord/PutRecords，对象也会在垃圾收集时自动释放回对象池：

```go
// 获取 Record 对象
record := record.GetRecord()

// 使用 Record 对象
record["id"] = 1
record["name"] = "张三"

// 无需手动调用 PutRecord，对象会自动释放
// 当 record 变量超出作用域或被设为 nil 时，垃圾收集器会自动回收
```



### 9.5 内存管理最佳实践

#### 9.5.1 推荐使用方式

```go
// 方式 1：手动管理（推荐用于性能敏感场景）
record := record.GetRecord()
defer record.PutRecord(record)

// 使用 record...

// 方式 2：自动管理（推荐用于一般场景）
record := record.GetRecord()
// 使用 record...
// 无需手动 Put，会自动释放

// 方式 3：批量操作管理
records := record.GetRecords()
defer record.PutRecords(records)

// 添加记录
for i := 0; i < 100; i++ {
    r := record.GetRecord()
    r["id"] = i
    r["name"] = fmt.Sprintf("用户%d", i)
    records = append(records, r)
}
```

#### 9.5.2 性能优化建议

1. **预分配容量**：对于已知大小的 Records，使用 GetRecordsWithCapacity 预分配容量
2. **及时释放**：在不再需要对象时，尽快调用 PutRecord/PutRecords
3. **避免循环引用**：确保 Record 中的值不会导致循环引用，影响垃圾收集
4. **批量操作**：使用批量操作（如 BatchSelect、BatchOperation）减少对象创建
5. **合理使用 defer**：对于复杂函数，使用 defer 确保对象最终会被释放
6. **避免频繁创建**：在循环中重用 Record 对象，而不是每次循环都创建新对象
7. **注意作用域**：控制 Record 对象的作用域，避免不必要的长生命周期

#### 9.5.3 生产环境最佳实践

1. **性能监控**：
   - 设置监控告警，及时发现内存异常

2. **内存优化**：
   - 根据实际业务场景调整对象池的预分配容量
   - 对于大数据量操作，使用合适的批量大小
   - 避免在高频调用的代码路径中创建过多临时对象

3. **错误处理**：
   - 在错误处理路径中确保对象正确释放
   - 使用 defer 语句处理对象释放，确保即使发生错误也能释放资源

4. **并发处理**：
   - 利用 sync.Pool 的线程安全特性，在并发场景中安全使用
   - 避免在热点路径中产生过多的对象竞争

5. **代码审查**：
   - 建立代码审查机制，确保遵循对象池使用最佳实践
   - 检查是否有未释放的 Record/Records 对象
   - 验证对象池使用是否符合性能优化建议

6. **测试策略**：
   - 编写性能测试，验证对象池的效果
   - 模拟高并发场景，测试对象池在压力下的表现
   - 监控内存使用和垃圾收集情况

#### 9.5.4 实际应用场景

**场景 1：高频查询**

```go
// 高并发查询场景
handler := func(w http.ResponseWriter, r *http.Request) {
    // 获取 Records 对象
    records := record.GetRecords()
    defer record.PutRecords(records)
    
    // 执行查询
    // ...
    
    // 处理结果
    for _, r := range records {
        // 处理记录
    }
}
```

**场景 2：批量数据处理**

```go
// 批量数据处理
func processBatch(data []map[string]any) {
    // 预分配容量
    records := record.GetRecordsWithCapacity(len(data))
    defer record.PutRecords(records)
    
    for _, item := range data {
        r := record.GetRecord()
        for k, v := range item {
            r[k] = v
        }
        records = append(records, r)
    }
    
    // 处理记录
    // ...
}
```

**场景 3：复杂业务逻辑**

```go
// 复杂业务逻辑
func businessLogic() error {
    // 获取对象
    r := record.GetRecord()
    defer func() {
        // 确保即使发生错误也能释放对象
        record.PutRecord(r)
    }()
    
    // 业务逻辑
    // ...
    
    if err != nil {
        return err // 对象会通过 defer 释放
    }
    
    // 继续处理
    // ...
    
    return nil
}
```

### 9.6 实际使用示例

#### 9.6.1 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // 获取 Record 对象
    r := record.GetRecord()
    defer record.PutRecord(r)
    
    // 设置字段
    r["id"] = 1
    r["name"] = "张三"
    r["age"] = 30
    
    fmt.Printf("Record: %v\n", r)
    
    // 获取 Records 对象
    rs := record.GetRecords()
    defer record.PutRecords(rs)
    
    // 添加记录
    rs = append(rs, r)
    
    fmt.Printf("Records 长度: %d\n", len(rs))
}
```



#### 9.6.3 批量操作示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // 创建测试数据
    records := make(record.Records, 0, 100)
    for i := 0; i < 100; i++ {
        r := record.GetRecord()
        r["id"] = i
        r["name"] = fmt.Sprintf("用户%d", i)
        r["age"] = 20 + i%30
        records = append(records, r)
    }
    
    // 使用批量选择操作
    selected := record.BatchSelect(records, "id", "name")
    fmt.Printf("批量选择后记录数: %d\n", len(selected))
    
    // 释放所有对象
    for _, r := range records {
        record.PutRecord(r)
    }
    record.PutRecords(selected)
    record.PutRecords(records)
}
```

### 9.7 常见问题与解决方案

#### 9.7.1 对象池复用率低

**问题**：对象池复用率低，创建数远大于获取数

**解决方案**：
- 确保在不再使用对象时及时调用 PutRecord/PutRecords
- 检查是否存在对象泄露（如循环引用）
- 调整批量操作策略，减少临时对象创建

#### 9.7.2 内存使用过高

**问题**：应用内存使用持续增长

**解决方案**：
- 检查是否有未释放的 Record/Records 对象
- 考虑增加垃圾收集频率（debug.SetGCPercent）

#### 9.7.3 性能问题

**问题**：对象池操作成为性能瓶颈

**解决方案**：
- 使用 GetRecordsWithCapacity 预分配容量
- 优化批量操作，减少对象创建
- 考虑使用对象池的手动管理模式

## 10. 其他功能

### 10.1 记录集合操作

 sfsDb 提供了强大的记录集合操作功能，支持交集（Intersect）、并集（Union）和差集（Difference）等操作。此外，还支持动态操作函数，允许用户根据需求自定义集合操作：

#### 10.1.1 动态操作函数

sfsDb 提供了 `RecordsOperation` 类型和 `Apply` 方法，允许用户动态定义和应用集合操作：

```go
// RecordsOperation 定义了记录操作函数类型
type RecordsOperation func(rs Records, other ...Records) Records

// Apply 方法用于动态应用操作函数到记录集合
func (rs Records) Apply(op RecordsOperation, other ...Records) Records {
    return op(rs, other...)
}
```

**核心功能**：
- 允许用户根据需求自定义集合操作
- 支持将多个操作组合在一起
- 提供了更灵活的集合操作方式
- 可以与现有集合操作（Intersect、Union、Difference）结合使用

**使用场景**：
- 需要实现自定义集合操作逻辑时
- 需要组合多个集合操作时
- 需要动态选择集合操作时

**基本使用示例**：
```go
// 定义自定义操作函数
customOp := record.RecordsOperation(func(rs record.Records, other ...record.Records) record.Records {
    // 自定义操作逻辑
    result := make(record.Records, 0)
    for _, r := range rs {
        // 实现自定义过滤逻辑
        if age, ok := r["age"].(int); ok && age > 30 {
            result = append(result, r)
        }
    }
    return result
})

// 应用操作函数
result := records1.Apply(customOp)
```

除了动态操作函数，sfsDb 还支持以下内置集合操作：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // 创建测试记录集合
    records1 := record.Records{
        {"id": 1, "name": "张三", "age": 30},
        {"id": 2, "name": "李四", "age": 25},
        {"id": 3, "name": "王五", "age": 35},
    }

    records2 := record.Records{
        {"id": 2, "name": "李四", "age": 25},
        {"id": 3, "name": "王五", "age": 35},
        {"id": 4, "name": "赵六", "age": 40},
    }

    records3 := record.Records{
        {"id": 3, "name": "王五", "age": 35},
        {"id": 5, "name": "孙七", "age": 28},
    }

    fmt.Println("=== 集合操作示例 ===")
    fmt.Printf("集合1: %v\n", records1)
    fmt.Printf("集合2: %v\n", records2)
    fmt.Printf("集合3: %v\n", records3)

    // 1. 交集操作（Intersect）
    // 获取两个集合中都存在的记录
    fmt.Println("\n1. 交集操作 (records1 ∩ records2):")
    intersectResult := records1.Intersect(records2)
    for _, record := range intersectResult {
        fmt.Printf("   - %v\n", record)
    }

    // 2. 并集操作（Union）
    // 获取两个集合的所有唯一记录
    fmt.Println("\n2. 并集操作 (records1 ∪ records2):")
    unionResult := records1.Union(records2)
    for _, record := range unionResult {
        fmt.Printf("   - %v\n", record)
    }

    // 3. 差集操作（Difference）
    // 获取存在于第一个集合但不存在于第二个集合的记录
    fmt.Println("\n3. 差集操作 (records1 - records2):")
    differenceResult := records1.Difference(records2)
    for _, record := range differenceResult {
        fmt.Printf("   - %v\n", record)
    }

    // 4. 多个集合的交集
    fmt.Println("\n4. 多个集合交集 (records1 ∩ records2 ∩ records3):")
    multiIntersect := records1.Intersect(records2, records3)
    for _, record := range multiIntersect {
        fmt.Printf("   - %v\n", record)
    }

    // 5. 多个集合的并集
    fmt.Println("\n5. 多个集合并集 (records1 ∪ records2 ∪ records3):")
    multiUnion := records1.Union(records2, records3)
    for _, record := range multiUnion {
        fmt.Printf("   - %v\n", record)
    }

    // 6. 多个集合的差集
    fmt.Println("\n6. 多个集合差集 (records1 - records2 - records3):")
    multiDifference := records1.Difference(records2, records3)
    for _, record := range multiDifference {
        fmt.Printf("   - %v\n", record)
    }

    // 7. 结合 Select 方法使用
    fmt.Println("\n7. 结合 Select 方法的集合操作:")
    // 先选择特定字段，再进行集合操作
    selected1 := records1.Select("id", "name")
    selected2 := records2.Select("id", "name")
    fmt.Printf("   集合1 只选择 id 和 name: %v\n", selected1)
    fmt.Printf("   集合2 只选择 id 和 name: %v\n", selected2)
    
    selectedUnion := selected1.Union(selected2)
    fmt.Printf("   选择字段后的并集: %v\n", selectedUnion)

    // 8. 动态操作函数（Apply方法）
    // 使用RecordsOperation类型和Apply方法实现动态集合操作
    fmt.Println("\n8. 动态操作函数示例:")
    
    // 8.1 自定义操作：返回所有年龄大于30的记录
    fmt.Println("   8.1 自定义操作：返回所有年龄大于30的记录:")
    customOp := record.RecordsOperation(func(rs record.Records, other ...record.Records) record.Records {
        result := make(record.Records, 0)
        for _, r := range rs {
            if age, ok := r["age"].(int); ok && age > 30 {
                result = append(result, r)
            }
        }
        return result
    })
    
    customResult := records1.Apply(customOp)
    for _, r := range customResult {
        fmt.Printf("   - %v\n", r)
    }
    
    // 8.2 组合操作：先求交集再求并集
    fmt.Println("\n   8.2 组合操作：先求交集再求并集:")
    composedOp := record.RecordsOperation(func(rs record.Records, other ...record.Records) record.Records {
        // 先求当前集合与other[0]的交集
        intersectResult := rs.Intersect(other[0])
        // 再求交集结果与other[1]的并集
        unionResult := intersectResult.Union(other[1])
        return unionResult
    })
    
    composedResult := records1.Apply(composedOp, records2, records3)
    for _, r := range composedResult {
        fmt.Printf("   - %v\n", r)
    }
}
```

### 9.2 迭代器使用

**TableIter 结构创建方式**：
- TableIter 结构是由表的 `ForData()` 方法或 `Search()` 方法产生的
- `Search()` 方法位于 `d:\MyGo\src\sfsDb\engine\tableCRUD.go#L283`，用于根据条件搜索记录并返回迭代器
- `ForData()` 方法用于遍历表中的所有记录并返回迭代器

**跳跃区间 (jumpRanges) 功能**：

#### 概述
TableIter 结构体包含一个 `jumpRanges` 字段（位于 `d:\MyGo\src\sfsDb\engine\tableiter.go#L16`），用于存储跳跃区间的迭代器。跳跃区间是一种优化机制，用于在搜索时排除某些范围的记录，提高查询效率。

#### 工作原理
- **定义**：`jumpRanges []storage.Iterator` - 存储多个跳跃区间的迭代器
- **设置方法**：`TableIter.SetJumpRanges(jumpRanges ...storage.Iterator)` - 设置跳跃区间
- **使用场景**：主要用于不等于（!=）操作，当需要排除某些特定值的记录时
- **资源管理**：迭代器的 `Release()` 方法会自动释放所有跳跃区间的迭代器资源

#### 代码示例

**在 `Search` 方法中的应用**（位于 `d:\MyGo\src\sfsDb\engine\tableCRUD.go#L372-378`）：

```go
} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
    slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
    iter = t.kvStore.Iterator(slice.Start, slice.Limit)
    tbiter = TableIterNew(t, iter, idx)
    //设置跳跃区间
    neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
    tbiter.SetJumpRanges(t.kvStore.Iterator(neslice.Start, neslice.Limit))
}
```

#### 应用场景
1. **不等于操作**：当执行 `field != value` 类型的查询时
2. **范围排除**：排除某个范围内的记录
3. **多值排除**：排除多个特定值的记录
4. **性能优化**：通过跳跃区间减少需要扫描的记录数量

#### 内部实现
当 TableIter 执行迭代时，会检查当前记录是否在跳跃区间内，如果是，则跳过该记录，继续下一条记录的处理。这样可以有效减少需要处理的记录数量，提高查询性能。

#### 使用建议
- **合理使用**：对于需要排除特定值的查询，跳跃区间可以显著提高性能
- **资源管理**：不需要手动管理跳跃区间的迭代器资源，`TableIter.Release()` 会自动处理
- **内存考虑**：过多的跳跃区间可能会增加内存使用，应根据实际需求合理设置

### 9.2.1 遍历表所有键值对

sfsDb 提供了 `For()` 方法，用于遍历表的所有键值对。这个方法返回一个存储引擎级别的迭代器，可以直接访问表的所有键值对，包括数据和索引数据。

```go
// 遍历表所有kv键值对，用于快速复制表用或删除表数据
func (t *Table) For() storage.Iterator {
	pfx := []byte{byte(t.id), SPLIT[0]}
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, pfx)
	return t.kvStore.Iterator(slice.Start, slice.Limit)
}
```

**工作原理**：
1. `For()` 方法使用表的 ID 和分隔符创建一个前缀
2. 使用 `util.NewRangeHelper` 创建一个范围辅助对象
3. 使用 `FromComparison` 方法创建一个前缀匹配的范围
4. 调用存储引擎的 `Iterator` 方法返回一个迭代器，用于遍历该范围内的所有键值对

**使用场景**：
- 快速复制表数据
- 删除表的所有数据（如 `DeleteAll()` 方法中使用）
- 底层数据操作和调试

**使用示例**：

```go
// 使用 For() 方法遍历表的所有键值对
iter := table.For()
defer iter.Release()

for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    fmt.Printf("Key: %s, Value: %s\n", string(key), string(value))
}
```

**注意事项**：
- `For()` 方法返回的是存储引擎级别的迭代器，需要手动调用 `Release()` 方法释放资源
- 这个方法会遍历表的所有键值对，包括数据和索引数据
- 对于大型表，遍历可能会比较耗时，建议在适当的场景中使用
- `For()` 方法是 `DeleteAll()` 方法的基础，用于实现表数据的全量删除

```go
// 遍历所有记录
iter := table.ForData()
defer iter.Release()

// 推荐方法：使用 GetRecords() // 获取所有记录（正序）
records := iter.GetRecords(true)
for _, record := range records {
    fmt.Printf("记录: %v\n", record)
}

## 10. 半结构化数据支持

### 10.1 概述

sfsDb 提供了强大的半结构化数据支持，允许存储和查询复杂的嵌套数据结构，如 JSON 对象和数组。这一特性使 sfsDb 能够适应各种复杂的业务场景，无需预定义表结构。

### 10.2 技术实现

#### 10.2.1 底层存储
- **自动序列化**：使用 JSON 格式自动序列化半结构化数据
- **类型安全**：支持将半结构化数据转换回原始类型
- **灵活存储**：可以存储任意嵌套深度的半结构化数据

#### 10.2.2 核心方法
- **FieldsToBytes**：将包含半结构化数据的字段转换为字节数组
- **RecordByteToAny**：将字节数组转换回包含半结构化数据的字段

### 10.3 测试结果

#### 10.3.1 测试场景

我们创建了一个完整的测试来验证 sfsDb 的半结构化数据支持能力，包括以下场景：

1. **创建包含半结构化字段的表**
2. **插入包含复杂半结构化数据的记录**
3. **读取和验证半结构化数据**
4. **更新半结构化数据**
5. **删除包含半结构化数据的记录**




### 10.4 使用示例

#### 10.4.1 创建包含半结构化字段的表

```go
// 创建一个表
tableName := "test_semi_structured"
table, err := engine.TableNew(tableName)
if err != nil {
    panic(err)
}

// 定义表结构，包含半结构化字段
fields := map[string]any{
    "id":    0,                  // 主键
    "name":  "",                 // 字符串
    "age":   0,                  // 整型
    "data":  map[string]any{},   // 半结构化数据（map）
    "tags":  []string{},          // 半结构化数据（数组）
    "nested": map[string]any{
        "level1": map[string]any{
            "level2": "value",
        },
    },
}

// 设置表字段
if err := table.SetFields(fields); err != nil {
    panic(err)
}

// 设置主键
if err := table.CreatePrimaryKey("id"); err != nil {
    panic(err)
}
```

#### 10.4.2 插入包含半结构化数据的记录

```go
// 插入包含半结构化数据的记录
record := map[string]any{
    "id":   1,
    "name": "张三",
    "age":  25,
    "data": map[string]any{
        "address": "北京市朝阳区",
        "phone":   "13800138000",
        "email":   "zhangsan@example.com",
    },
    "tags": []string{"user", "active", "vip"},
    "nested": map[string]any{
        "level1": map[string]any{
            "level2": "value1",
            "level3": map[string]any{
                "value":  123,
                "active": true,
            },
        },
    },
}

id, err := table.Insert(&record)
if err != nil {
    panic(err)
}
fmt.Printf("插入成功，ID: %d\n", id)
```

#### 10.4.3 更新半结构化数据

```go
// 更新半结构化数据
updateRecord := map[string]any{
    "id":   1,
    "name": "张三（更新）",
    "data": map[string]any{
        "address": "北京市海淀区",
        "phone":   "13800138001",
        "email":   "zhangsan@example.com",
        "social": map[string]any{
            "wechat": "zhangsan_wechat",
        },
    },
    "tags": []string{"user", "active", "vip", "updated"},
}

if err := table.Update(&updateRecord); err != nil {
    panic(err)
}
fmt.Println("更新成功")
```

### 10.5 应用场景

半结构化数据支持使 sfsDb 适合以下场景：

1. **用户配置管理**：存储和管理用户的复杂配置信息
2. **IoT 设备数据**：处理传感器产生的不规则数据结构
3. **内容管理系统**：存储和查询复杂的内容结构
4. **API 响应缓存**：缓存和管理 API 响应的复杂数据结构
5. **日志数据存储**：存储和查询包含不同字段的日志数据

### 10.6 优势

1. **灵活性**：无需预定义表结构，适应业务需求的动态变化
2. **开发效率**：简化复杂数据结构的处理，减少代码量
3. **兼容性**：支持与结构化数据的混合使用
4. **性能**：针对半结构化数据进行了存储优化
5. **易用性**：提供直观的 API 来处理半结构化数据

**使用建议**：
- 在高并发环境下，建议使用指定版本号的更新方式，避免丢失更新
- 在低并发环境下，可以使用不指定版本号的更新方式，简化代码
- 当遇到乐观锁冲突时，建议实现重试机制，或提示用户重新操作
- 乐观锁机制适用于读多写少的场景

### 9.3 批量更新记录

使用迭代器的 Update 方法可以便捷地批量更新符合条件的记录：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/util"
)

func main() {
    // 创建或获取表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // 搜索条件：年龄小于25岁的用户
    searchCriteria := map[string]any{
        "age": map[util.ComparisonOperator]any{util.LessThan: 25},
    }
    iter := table.Search(&searchCriteria)
    defer iter.Release()

    // 准备更新的字段
    updateFields := map[string]any{
        "active": false, // 将年轻用户设为非活跃状态
    }

    // 使用迭代器的 Update 方法批量更新所有符合条件的记录
    iter.Update(&updateFields)

    // 或者限制更新数量：只更新前2条符合条件的记录
    // iter.Update(&updateFields, 2)

    fmt.Println("批量更新完成")
}
```

### 9.4 数据库备份

sfsDb 提供了数据库备份功能，可以将当前数据库的所有记录备份到指定路径：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 1. 打开源数据库（如果尚未打开）
    sourceDb, err := storage.OpenDefaultDb("./source_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 2. 执行数据库备份
    backupPath := "./backup_db"
    err = storage.BackupDb(backupPath)
    if err != nil {
        panic(fmt.Sprintf("数据库备份失败: %v", err))
    }

    fmt.Printf("数据库备份成功，备份路径: %s\n", backupPath)
}
```

**函数说明：**

```go
//备份数据库
func BackupDb(Path string) error
// 参数：
//   Path: 备份数据库的保存路径
// 返回值：
//   error: 备份过程中发生的错误，如果成功则返回nil
```

**使用注意事项：**

1. 备份操作会将源数据库的所有记录复制到目标路径
2. 备份过程中不会修改源数据库的状态
3. 备份完成后，目标路径将包含一个完整的数据库副本
4. 可以使用 `storage.OpenDefaultDb(backupPath)` 打开备份的数据库进行恢复或查询
5. 备份数据库是一个完整的数据库实例，可以独立使用

### 9.5 数据导入导出

sfsDb 支持将表数据导出为 CSV、JSON 和 SQL 格式，并支持从 CSV 和 JSON 格式导入数据。

#### 9.5.1 导出数据示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 打开数据库
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 获取表
    table, err := engine.TableNew("test_table")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "active": false,
        "score":  0.0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    testData := []map[string]any{
        {"id": 1, "name": "Alice", "age": 25, "active": true, "score": 85.5},
        {"id": 2, "name": "Bob", "age": 30, "active": false, "score": 90.0},
        {"id": 3, "name": "Charlie", "age": 35, "active": true, "score": 75.5},
    }

    for _, data := range testData {
        _, err := table.Insert(&data)
        if err != nil {
            panic(err)
        }
    }

    // 导出为 CSV
    err = table.ExportToCSV("./test.csv")
    if err != nil {
        panic(err)
    }
    fmt.Println("CSV 导出成功")

    // 导出为 JSON
    err = table.ExportToJSON("./test.json")
    if err != nil {
        panic(err)
    }
    fmt.Println("JSON 导出成功")

    // 导出为 SQL
    err = table.ExportToSQL("./test.sql")
    if err != nil {
        panic(err)
    }
    fmt.Println("SQL 导出成功")
}
```

#### 9.5.2 导入数据示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 打开数据库
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建新表用于导入数据
    table2, err := engine.TableNew("test_table2")
    if err != nil {
        panic(err)
    }

    // 设置表字段（与导出表相同）
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "active": false,
        "score":  0.0,
    }
    err = table2.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 从 CSV 导入数据（批量大小为 100）
    err = table2.ImportFromCSV("./test.csv", 100)
    if err != nil {
        panic(err)
    }
    fmt.Println("从 CSV 导入成功")

    // 创建另一个表用于导入 JSON 数据
    table3, err := engine.TableNew("test_table3")
    if err != nil {
        panic(err)
    }

    // 设置表字段（与导出表相同）
    err = table3.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 从 JSON 导入数据（批量大小为 100）
    err = table3.ImportFromJSON("./test.json", 100)
    if err != nil {
        panic(err)
    }
    fmt.Println("从 JSON 导入成功")
}
```

#### 9.5.3 函数说明

**导出功能：**

```go
// 将表数据导出为 CSV 格式
func (t *Table) ExportToCSV(filePath string) error
// 参数：
//   filePath: 导出文件路径
// 返回值：
//   error: 导出过程中发生的错误，如果成功则返回 nil

// 将表数据导出为 JSON 格式
func (t *Table) ExportToJSON(filePath string) error
// 参数：
//   filePath: 导出文件路径
// 返回值：
//   error: 导出过程中发生的错误，如果成功则返回 nil

// 将表数据导出为 SQL 格式
func (t *Table) ExportToSQL(filePath string) error
// 参数：
//   filePath: 导出文件路径
// 返回值：
//   error: 导出过程中发生的错误，如果成功则返回 nil
```

**导入功能：**

```go
// 从 CSV 文件导入数据到表
func (t *Table) ImportFromCSV(filePath string, batchSize int) error
// 参数：
//   filePath: 导入文件路径
//   batchSize: 批量导入大小，用于优化性能
// 返回值：
//   error: 导入过程中发生的错误，如果成功则返回 nil

// 从 JSON 文件导入数据到表
func (t *Table) ImportFromJSON(filePath string, batchSize int) error
// 参数：
//   filePath: 导入文件路径
//   batchSize: 批量导入大小，用于优化性能
// 返回值：
//   error: 导入过程中发生的错误，如果成功则返回 nil
```

#### 9.5.4 使用注意事项

1. **数据格式要求**：
   - CSV 文件必须包含表头，表头字段名必须与表字段名一致
   - JSON 文件必须是包含对象的数组，对象的键名必须与表字段名一致
   - SQL 文件包含创建表语句和插入语句，可用于初始化新表

2. **类型转换**：
   - 导入时会自动进行类型转换，确保数据类型与表定义匹配
   - JSON 中的数字会根据表定义自动转换为 int、float32 等类型
   - CSV 中的空值会被转换为表字段的默认值

3. **性能优化**：
   - 导入大量数据时，建议调整 batchSize 参数以优化性能
   - 较大的 batchSize 可以提高导入速度，但会增加内存占用
   - 默认 batchSize 为 100

4. **数据一致性**：
   - 导入过程中会进行类型检查，确保数据符合表定义
   - 导入失败时会返回详细错误信息，包括记录索引和错误原因
   - 建议在导入前备份原有数据

### 9.6 数据加密

sfsDb 提供了强大的数据加密功能，可以保护存储在磁盘上的数据安全性。加密功能使用 AES-256-GCM 算法，提供了灵活的密钥管理方式和良好的性能表现。

#### 9.6.1 加密概述

**核心特性**：
- **强加密算法**：使用 AES-256-GCM 算法，提供 256 位密钥保护
- **透明加密**：对上层应用透明，无需修改现有代码
- **灵活密钥管理**：支持直接密钥、密码派生密钥和密钥轮换
- **高性能**：通过解密缓存和批量操作优化性能
- **可选加密**：可以选择是否启用加密，或加密特定表

**加密范围**：
- 所有存储在磁盘上的数据
- 包括表数据、索引和元数据
- 数据库结构信息（表名、字段名）默认不加密

#### 9.6.2 加密配置

```go
// EncryptionConfig 加密配置
type EncryptionConfig struct {
    // 是否启用加密
    Enabled bool `json:"enabled"`
    
    // 加密算法，默认AES-256-GCM
    Algorithm string `json:"algorithm"`
    
    // 主密钥，直接提供的256位密钥
    MasterKey []byte `json:"master_key,omitempty"`
    
    // 密码，用于派生密钥
    Password string `json:"password,omitempty"`
    
    // 盐值，用于密码派生
    Salt []byte `json:"salt,omitempty"`
    
    // 迭代次数，用于密码派生，默认100000
    Iterations int `json:"iterations,omitempty"`
}
```

**配置说明**：
- `Enabled`：是否启用加密，默认 false
- `Algorithm`：加密算法，目前只支持 "AES-256-GCM"，默认值为空字符串（自动使用 AES-256-GCM）
- `MasterKey`：直接提供的 256 位密钥，与 Password 二选一
- `Password`：用于派生密钥的密码，与 MasterKey 二选一
- `Salt`：密码派生时使用的盐值，建议使用随机生成的 16 字节盐值
- `Iterations`：PBKDF2 算法的迭代次数，默认 100,000

#### 9.6.3 使用示例

##### 9.6.3.1 使用直接密钥加密

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 生成或获取256位主密钥
    masterKey := make([]byte, 32) // 256位密钥
    // 实际应用中，建议使用密码学安全的随机数生成器生成密钥
    for i := range masterKey {
        masterKey[i] = byte(i)
    }
    
    // 创建加密配置
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        MasterKey: masterKey,
    }
    
    // 打开加密数据库
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("加密数据库打开成功")
    
    // 后续操作与普通数据库相同
    // ...
}
```

##### 9.6.3.2 使用密码派生密钥

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建加密配置，使用密码派生密钥
    encryptConfig := &storage.EncryptionConfig{
        Enabled:    true,
        Algorithm:  "AES-256-GCM",
        Password:   "my_secure_password",
        Salt:       []byte("my_random_salt_123"), // 建议使用随机生成的盐值
        Iterations: 100000,
    }
    
    // 打开加密数据库
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("使用密码派生密钥的加密数据库打开成功")
    
    // 后续操作与普通数据库相同
    // ...
}
```

##### 9.6.3.3 多实例模式下的加密

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 生成测试密钥
    masterKey := make([]byte, 32)
    for i := range masterKey {
        masterKey[i] = byte(i)
    }
    
    // 创建加密配置
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: masterKey,
    }
    
    // 使用NewLevelDBStoreWithEncryption创建独立的加密数据库实例
    // 不使用全局KVDb变量
    db1, err := storage.NewLevelDBStoreWithEncryption("./encrypted_db_1", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer db1.Close()
    
    // 可以创建多个独立的加密数据库实例
    db2, err := storage.NewLevelDBStoreWithEncryption("./encrypted_db_2", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer db2.Close()
    
    fmt.Println("多个加密数据库实例创建成功")
    
    // 后续操作与普通数据库相同
    // ...
}
```

#### 9.6.4 API参考

**核心API**：

```go
// OpenDefaultDbWithEncryption 打开带加密的默认数据库
// 使用全局KVDb变量，保证全局唯一实例
func OpenDefaultDbWithEncryption(Path string, config *EncryptionConfig) (Store, error)

// NewLevelDBStoreWithEncryption 创建带加密的LevelDB存储实例
// 每次调用创建新实例，不使用全局变量
func NewLevelDBStoreWithEncryption(Path string, config *EncryptionConfig) (Store, error)

// ReEncrypt 重新加密所有数据（密钥轮换）
// 仅EncryptedStoreWrapper实现此方法
func (es *EncryptedStoreWrapper) ReEncrypt(newKey []byte) error
```

#### 9.6.5 密钥轮换

密钥轮换是一种重要的安全实践，可以定期更换加密密钥，降低密钥泄露的风险。

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始密钥
    oldKey := make([]byte, 32)
    for i := range oldKey {
        oldKey[i] = byte(i)
    }
    
    // 打开加密数据库
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: oldKey,
    }
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    // 存储一些数据
    // ...
    
    // 生成新密钥
    newKey := make([]byte, 32)
    for i := range newKey {
        newKey[i] = byte(255 - i)
    }
    
    // 执行密钥轮换
    encryptedWrapper, ok := db.(*storage.EncryptedStoreWrapper)
    if !ok {
        panic("Failed to convert to EncryptedStoreWrapper")
    }
    
    err = encryptedWrapper.ReEncrypt(newKey)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("密钥轮换成功")
    
    // 密钥轮换后，所有数据将使用新密钥加密
    // 后续操作与普通数据库相同
    // ...
}
```

#### 9.6.6 性能考虑

- **加密开销**：AES-256-GCM 加密解密速度很快，预计增加约 5-15% 的 CPU 开销
- **存储开销**：每条记录增加约 16 字节（认证标签）
- **缓存优化**：实现了解密缓存，减少重复解密操作
- **批量操作**：支持批量加密/解密，优化批量操作性能
- **配置建议**：
  - 对于高并发场景，建议使用直接密钥而非密码派生
  - 适当调整解密缓存大小（默认 1000 条记录）
  - 对批量操作进行优化，减少加密解密次数

#### 9.6.7 安全最佳实践

1. **密钥管理**：
   - 不要硬编码密钥在代码中
   - 考虑使用密钥管理服务（KMS）存储密钥
   - 定期进行密钥轮换
   - 确保密钥的安全性和完整性

2. **密码安全**：
   - 使用强密码（至少 16 个字符，包含大小写字母、数字和特殊字符）
   - 使用随机生成的盐值
   - 适当增加迭代次数（默认 100,000）

3. **部署安全**：
   - 确保数据库文件的访问权限正确设置
   - 考虑使用 TLS 加密网络传输
   - 定期备份加密数据库
   - 实施访问控制机制

通过合理配置和使用 sfsDb 的加密功能，可以有效保护数据安全，满足各种安全合规要求。

**TableIter 索引信息方法**：

```go
// 获取搜索时使用的索引的名称
func (t *TableIter) GetIndexName() string
// 返回值：
//   string: 索引名称

// 获取搜索时使用的索引的ID
func (t *TableIter) GetIndexId() uint8
// 返回值：
//   uint8: 索引ID
```

**使用示例**：

```go
// 创建查询迭代器
iter := table.Search(&map[string]any{"name": "Alice"})
if iter == nil {
    panic("查询失败")
}
defer iter.Release()

// 执行查询操作
records := iter.GetRecords(true)

// 获取索引详细信息
fmt.Printf("使用的索引名称: %s\n", iter.GetIndexName())
fmt.Printf("使用的索引ID (直接获取): %d\n", iter.GetIndexId())
```





### 9.7 键值变化跟踪

sfsDb 内置了键值变化跟踪功能，可以帮助开发者监控数据库中键值对的增减情况，包括记录插入（put）和删除（delete）操作的数量，便于监控数据库的运行状态和性能分析。

#### 9.7.1 跟踪数据结构

键值变化通过两个全局变量进行跟踪：

```go
// 递增变量：记录 put 操作的键值对数
// key 格式：table_id,index_id（由表ID和索引ID组成的字符串）
var AtomicInt map[string]*atomic.Int64
 
// 递减变量：记录 delete 操作的键值对数
// key 格式：table_id,index_id（由表ID和索引ID组成的字符串）
var AtomicDec map[string]*atomic.Int64
```

#### 9.7.2 跟踪数据访问

可以通过 `AtomicMap` 类型的方法来安全地访问和操作跟踪数据：

```go
// AtomicMap 是用于原子操作键值计数器的类型
// 使用时需要将 AtomicInt 或 AtomicDec 转换为 AtomicMap 类型

type AtomicMap map[string]*atomic.Int64

// Inc 递增指定键的计数器
func (m AtomicMap) Inc(tableID, indexID byte) {}

// Dec 递减指定键的计数器
func (m AtomicMap) Dec(tableID, indexID byte) {}

// Get 获取指定键的当前计数值
func (m AtomicMap) Get(tableID, indexID byte) int64 {}
```

#### 9.7.3 使用示例

**示例1：获取表的插入和删除操作数量**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/monitor"
)

func main() {
    // 假设表ID为0，索引ID为0
    tableID := byte(0)
    indexID := byte(0)
    
    // 获取插入操作数量
    putCount := monitor.AtomicMap(monitor.AtomicInt).Get(tableID, indexID)
    fmt.Printf("表 %d 索引 %d 的插入操作数量: %d\n", tableID, indexID, putCount)
    
    // 获取删除操作数量
    deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(tableID, indexID)
    fmt.Printf("表 %d 索引 %d 的删除操作数量: %d\n", tableID, indexID, deleteCount)
    
    // 计算净变化量
    netChange := putCount - deleteCount
    fmt.Printf("表 %d 索引 %d 的净变化量: %d\n", tableID, indexID, netChange)
}
```

**示例2：监控多个表和索引**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/monitor"
)

func main() {
    // 定义要监控的表和索引
    monitorConfigs := []struct {
        tableID  byte
        indexID  byte
        name     string
    }{
        {0, 0, "用户表-主键索引"},
        {0, 1, "用户表-姓名索引"},
        {1, 0, "订单表-主键索引"},
    }
    
    fmt.Println("数据库键值变化监控报告:")
    fmt.Println("==================================")
    
    for _, config := range monitorConfigs {
        putCount := monitor.AtomicMap(monitor.AtomicInt).Get(config.tableID, config.indexID)
        deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(config.tableID, config.indexID)
        netChange := putCount - deleteCount
        
        fmt.Printf("%s:\n", config.name)
        fmt.Printf("  - 插入操作: %d\n", putCount)
        fmt.Printf("  - 删除操作: %d\n", deleteCount)
        fmt.Printf("  - 净变化量: %d\n", netChange)
        fmt.Println()
    }
}
```

#### 9.7.4 自动跟踪的操作

sfsDb 会在以下操作中自动更新键值变化跟踪数据：

- **插入记录**：使用 `table.Insert()` 或 `table.BatchInsert()` 时，自动递增对应表和索引的插入计数器
- **删除记录**：使用 `table.Delete()` 或 `table.BatchDelete()` 时，自动递增对应表和索引的删除计数器
- **更新记录**：更新记录时，会先删除旧记录，再插入新记录，因此会同时更新插入和删除计数器

#### 9.7.5 使用场景

1. **数据库监控**：实时监控数据库的写入和删除操作，了解数据库的活跃程度
2. **性能分析**：分析不同表和索引的操作频率，识别热点数据
3. **容量规划**：根据键值变化趋势，预测数据库容量增长情况
4. **问题诊断**：当数据库性能异常时，检查键值变化情况，定位问题
5. **审计**：记录数据库的操作数量，用于审计和合规性要求

#### 9.7.6 注意事项

1. **并发安全性**：`AtomicInt` 和 `AtomicDec` 内部使用 `atomic.Int64` 保证并发安全，但直接访问 map 本身不是线程安全的，建议使用 `AtomicMap` 类型的方法进行操作
2. **数据持久化**：键值变化跟踪数据仅存在于内存中，重启应用后会丢失
3. **性能开销**：跟踪操作使用原子操作，性能开销较小，但在极高并发场景下可能会有影响
4. **键格式**：键格式为 `table_id,index_id`，其中 table_id 和 index_id 是 byte 类型的字符表示
5. **初始化**：跟踪数据在应用启动时自动初始化，无需手动操作
6. **适用范围**：仅跟踪通过 sfsDb API 进行的操作，直接修改底层存储不会被跟踪

### 9.8 记录运算

sfsDb 内置了记录运算功能，可以对记录集合中的字段进行各种运算，并将运算结果作为新字段添加到记录中。

#### 9.8.1 Records.Operation 方法

**方法签名**：
```go
func (rs Records) Operation(op ...Operation) (rs2 Records) {
    if len(rs) == 0 {
        return nil
    }
    if len(op) == 0 {
        return rs
    }
    // 直接创建结果切片，避免先复制再修改
    result := make(Records, 0, len(rs))
    for _, r := range rs {
        // 为每个记录创建新的副本，避免修改原记录
        newRecord := make(Record, len(r)+1)
        // 直接复制字段，避免使用 maps.Copy
        for k, v := range r {
            newRecord[k] = v
        }
        // 检测新字段名是否已经存在
        newField := op[0].NewField()
        if _, exists := newRecord[newField]; exists {
            panic(fmt.Sprintf("field '%s' already exists in record, cannot add duplicate field", newField))
        }
        // 添加新字段
        newRecord[newField] = op[0].Evaluate(r)
        // 将新记录添加到结果切片
        result = append(result, newRecord)
    }
    return result
}
```

**功能说明**：
- 对记录集合中的每个记录执行指定的运算操作
- 返回包含运算结果的新记录集合，不修改原记录
- 支持多个运算操作（当前版本只处理第一个）
- 自动检测新字段名是否已存在，避免重复

**参数说明**：
- `op`：运算操作对象，实现了 `Operation` 接口
- 通过`Operation` 接口，可以自定义运算操作，包括字段列表、新字段名和额外参数
- 自定义运算操作必须实现 `NewField` 和 `Evaluate` 方法，分别返回新字段名和运算结果
- 自定义运算操作可以使用 `args` 参数传递传递额外参数，用于运算时的计算



**返回值**：
- `rs2`：包含运算结果的新记录集合

#### 9.8.2 内置运算类型

sfsDb 内置了多种常用运算类型，通过 `CommonOperation` 结构体实现：

```go
type CommonOperation struct {
    opType   string // 运算类型：add, sub, mul, div, avg, sum, max, min, concat
    fields   []string // 参与运算的字段列表
    newField string // 运算后生成的新字段名
    args     map[string]any // 额外参数
}
```

**支持的运算类型**：

| 运算类型 | 说明 | 示例 |
|---------|------|------|
| `add`/`sum` | 加法运算 | `NewAddOperation([]string{"salary", "bonus"}, "total_income")` |
| `sub` | 减法运算 | `NewSubOperation([]string{"salary", "bonus"}, "net_salary")` |
| `mul` | 乘法运算 | `NewMulOperation([]string{"price", "quantity"}, "total_price")` |
| `div` | 除法运算 | `NewDivOperation([]string{"salary", "days"}, "daily_salary", 1.0)` |
| `avg` | 平均值运算 | `NewAvgOperation([]string{"score1", "score2"}, "avg_score")` |
| `max` | 最大值运算 | `NewMaxOperation([]string{"field1", "field2"}, "max_value")` |
| `min` | 最小值运算 | `NewMinOperation([]string{"field1", "field2"}, "min_value")` |
| `concat` | 字符串连接 | `NewConcatOperation([]string{"str1", "str2"}, "combined_str", " ")` |

#### 9.8.3 便捷创建函数

sfsDb 提供了便捷的函数来创建各种运算操作：

```go
// 创建加法运算实例
func NewAddOperation(fields []string, newField string) *CommonOperation

// 创建减法运算实例
func NewSubOperation(fields []string, newField string) *CommonOperation

// 创建乘法运算实例
func NewMulOperation(fields []string, newField string) *CommonOperation

// 创建除法运算实例
func NewDivOperation(fields []string, newField string, defaultDivisor float64) *CommonOperation

// 创建平均值运算实例
func NewAvgOperation(fields []string, newField string) *CommonOperation

// 创建求和运算实例
func NewSumOperation(fields []string, newField string) *CommonOperation

// 创建最大值运算实例
func NewMaxOperation(fields []string, newField string) *CommonOperation

// 创建最小值运算实例
func NewMinOperation(fields []string, newField string) *CommonOperation

// 创建字符串连接运算实例
func NewConcatOperation(fields []string, newField string, separator string) *CommonOperation
```

#### 9.8.4 使用示例

**示例1：计算员工总收入**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // 创建表
    table, _ := engine.TableNew("employees")
    
    // 添加一些记录
    record1 := map[string]any{"name": "张三", "salary": 5000.0, "bonus": 1000.0}
    record2 := map[string]any{"name": "李四", "salary": 4000.0, "bonus": 800.0}
    table.Insert(&record1)
    table.Insert(&record2)
    
    // 查询所有记录
    iter := table.Search(nil)
    records, _ := iter.GetRecords()
    
    // 创建加法运算：salary + bonus
    op := record.NewAddOperation([]string{"salary", "bonus"}, "total_income")
    
    // 应用运算
    resultRecords := records.Operation(op)
    
    // 打印结果
    for _, r := range resultRecords {
        fmt.Printf("姓名：%s，总收入：%.2f\n", r["name"], r["total_income"])
    }
}
```

**示例2：字符串连接**

```go
// 创建字符串连接运算：str1 + str2
op := record.NewConcatOperation([]string{"first_name", "last_name"}, "full_name", " ")
resultRecords := records.Operation(op)
```

**示例3：多种运算结合使用**

```go
// 先计算总收入
incomeOp := record.NewAddOperation([]string{"salary", "bonus"}, "total_income")
incomeRecords := records.Operation(incomeOp)

// 再计算平均工资
avgOp := record.NewAvgOperation([]string{"salary"}, "avg_salary")
avgRecords := incomeRecords.Operation(avgOp)
```

#### 9.8.6 垂直运算（OperationVertical）方法

**方法签名**：
```go
func (rs Records) OperationVertical(op ...VerticalOperation) (rs2 Records) {
    if len(rs) == 0 {
        return nil
    }
    if len(op) == 0 {
        return rs
    }
    // 实现省略
}
```

**功能说明**：
- 对记录集合执行垂直运算，即对多个记录的同一字段进行运算
- 返回包含运算结果的新记录集合，不修改原记录
- 支持多个垂直运算操作
- 自动检测新字段名是否已存在，避免重复

**参数说明**：
- `op`：垂直运算操作对象，实现了 `VerticalOperation` 接口
- 通过 `VerticalOperation` 接口，可以自定义垂直运算操作
- 自定义垂直运算操作必须实现 `NewField` 和 `Evaluate` 方法

**返回值**：
- `rs2`：包含垂直运算结果的新记录集合

#### 9.8.7 内置垂直运算类型

sfsDb 内置了多种常用垂直运算类型，通过 `CommonVerticalOperation` 结构体实现：

| 运算类型 | 说明 | 示例 |
|---------|------|------|
| `sum` | 求和运算 | `NewSumVerticalOperation("salary", "total_salary")` |
| `avg` | 平均值运算 | `NewAvgVerticalOperation("salary", "avg_salary")` |
| `count` | 计数运算 | `NewCountVerticalOperation("salary", "record_count")` |
| `max` | 最大值运算 | `NewMaxVerticalOperation("salary", "max_salary")` |
| `min` | 最小值运算 | `NewMinVerticalOperation("salary", "min_salary")` |
| `group` | 分组运算 | `NewGroupVerticalOperation("department", "group_result")` |

#### 9.8.8 垂直运算便捷创建函数

sfsDb 提供了便捷的函数来创建各种垂直运算操作：

```go
// 创建求和垂直运算实例
func NewSumVerticalOperation(field string, newField string) *CommonVerticalOperation

// 创建平均值垂直运算实例
func NewAvgVerticalOperation(field string, newField string) *CommonVerticalOperation

// 创建计数垂直运算实例
func NewCountVerticalOperation(field string, newField string) *CommonVerticalOperation

// 创建最大值垂直运算实例
func NewMaxVerticalOperation(field string, newField string) *CommonVerticalOperation

// 创建最小值垂直运算实例
func NewMinVerticalOperation(field string, newField string) *CommonVerticalOperation

// 创建分组垂直运算实例
func NewGroupVerticalOperation(field string, newField string) *CommonVerticalOperation
```

#### 9.8.9 垂直运算使用示例

**示例1：求和运算**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // 创建表
    table, err := engine.TableNew("employees")
    if err != nil {
        panic(err)
    }
    
    // 设置字段
    fields := map[string]any{
        "id":         0,
        "name":       "",
        "salary":     0.0,
        "department": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // 插入测试数据
    employees := []map[string]any{
        {"name": "张三", "salary": 5000.0, "department": "技术部"},
        {"name": "李四", "salary": 4000.0, "department": "技术部"},
        {"name": "王五", "salary": 6000.0, "department": "销售部"},
        {"name": "赵六", "salary": 5500.0, "department": "销售部"},
    }
    
    for _, emp := range employees {
        _, err = table.Insert(&emp)
        if err != nil {
            panic(err)
        }
    }
    
    // 查询所有记录
    iter := table.ForData()
    defer iter.Release()
    records := iter.GetRecords(true)
    
    // 创建求和垂直运算：计算总工资
    sumOp := record.NewSumVerticalOperation("salary", "total_salary")
    
    // 应用垂直运算
    result := records.OperationVertical(sumOp)
    
    // 打印结果
    fmt.Println("员工工资信息：")
    for _, r := range result {
        fmt.Printf("姓名：%s，部门：%s，工资：%.2f，总工资：%.2f\n", 
            r["name"], r["department"], r["salary"], r["total_salary"])
    }
}
```

**示例2：多种垂直运算结合使用**

```go
// 创建多个垂直运算
countOp := record.NewCountVerticalOperation("salary", "employee_count")
avgOp := record.NewAvgVerticalOperation("salary", "avg_salary")
maxOp := record.NewMaxVerticalOperation("salary", "max_salary")

// 应用多个垂直运算
result := records.OperationVertical(countOp, avgOp, maxOp)
```

**示例3：分组运算**

```go
// 创建分组垂直运算：按部门分组
groupOp := record.NewGroupVerticalOperation("department", "group_result")
result := records.OperationVertical(groupOp)

// 打印分组结果
for _, r := range result {
    fmt.Printf("姓名：%s\n", r["name"])
    fmt.Printf("分组结果：%v\n", r["group_result"])
    fmt.Println("---")
}
```

#### 9.8.10 垂直运算注意事项

1. **不可修改原记录**：`OperationVertical` 方法不会修改原记录，始终返回新的记录集合
2. **重复字段检测**：自动检测新字段名是否已存在，避免重复
3. **垂直运算特性**：垂直运算是对多条记录的同一字段进行运算，结果会添加到每条记录中
4. **性能考虑**：对于大量记录，垂直运算可能会消耗较多内存和 CPU
5. **数据类型**：垂直运算会自动处理不同数值类型之间的转换
6. **分组运算结果**：分组运算返回的是 map[any]Records 类型，包含分组后的记录集合
7. **空记录处理**：对空记录集合执行垂直运算会返回 nil

#### 9.8.11 记录运算注意事项

1. **不可修改原记录**：`Operation` 和 `OperationVertical` 方法不会修改原记录，始终返回新的记录集合
2. **重复字段检测**：自动检测新字段名是否已存在，避免重复
3. **运算顺序**：`Operation` 方法当前版本只处理第一个运算操作，`OperationVertical` 方法支持多个运算操作
4. **性能考虑**：对于大量记录，运算操作可能会消耗较多内存和 CPU
5. **数据类型**：运算操作会自动处理不同数值类型之间的转换
6. **错误处理**：当检测到重复字段时，会触发 panic，建议在开发阶段进行测试
7. **垂直与水平运算**：`Operation` 是水平运算（对单条记录的多个字段进行运算），`OperationVertical` 是垂直运算（对多条记录的同一字段进行运算）

### 9.9 事务管理

sfsDb 提供了完整的事务支持，通过 `Transaction` 接口可以方便地进行事务操作。事务支持 ACID 特性，确保数据的一致性和可靠性。

#### 9.9.1 事务接口介绍

```go
// Transaction 定义事务接口
type Transaction interface {
    // Insert 在事务中插入记录
    Insert(fields *map[string]any) (int, error)
    // Update 在事务中更新记录
    Update(fields *map[string]any) error
    // Delete 在事务中删除记录
    Delete(fields *map[string]any) error
    // Search 在事务中搜索记录（支持读一致性）
    Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter
    // Read 在事务中读取单条记录（支持读一致性）
    Read(fields *map[string]any) ([]byte, error)
    // Commit 提交事务
    Commit() error
    // Rollback 回滚事务
    Rollback() error
}
```

#### 9.9.2 事务基本使用流程

```go
// 1. 开始事务
tx, err := table.Begin()
if err != nil {
    panic(err)
}

try {
    // 2. 执行事务操作（插入、更新、删除、查询）
    _, err = tx.Insert(&insertRecord)
    if err != nil {
        panic(err)
    }
    
    err = tx.Update(&updateRecord)
    if err != nil {
        panic(err)
    }
    
    // 3. 提交事务
    err = tx.Commit()
    if err != nil {
        panic(err)
    }
    fmt.Println("事务提交成功")
} catch {
    // 4. 发生错误时回滚事务
    err = tx.Rollback()
    if err != nil {
        panic(err)
    }
    fmt.Println("事务回滚成功")
}
```

#### 9.9.3 事务方法详细说明

1. **`Insert(fields *map[string]any) (int, error)`**
   - 在事务中插入一条记录
   - 参数：包含字段名和值的映射，必须包含除自动增值主键外的所有必填字段
   - 返回值：插入记录的ID和错误信息
   - 事务内插入的记录可以立即被事务内的其他操作读取（读自己的写）

2. **`Update(fields *map[string]any) error`**
   - 在事务中更新一条记录
   - 参数：包含主键和要更新字段的映射，必须包含主键
   - 返回值：错误信息
   - 事务内更新的记录会立即反映在事务内的后续操作中

3. **`Delete(fields *map[string]any) error`**
   - 在事务中删除一条记录
   - 参数：包含主键的映射，必须包含主键
   - 返回值：错误信息
   - 事务内删除的记录在事务提交前仍可被事务内的操作读取

4. **`Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter`**
   - 在事务中搜索记录
   - 参数：搜索条件和比较操作符
   - 返回值：结果迭代器
   - 支持读一致性，基于快照读取数据，不受外部修改影响

5. **`Read(fields *map[string]any) ([]byte, error)`**
   - 在事务中读取单条记录
   - 参数：包含主键的映射，必须包含主键
   - 返回值：记录的字节数组和错误信息
   - 优先从事务缓存读取，支持读自己的写

6. **`Commit() error`**
   - 提交事务，将所有操作持久化到数据库
   - 返回值：错误信息
   - 提交后，事务内的所有操作都会生效
   - 提交后，事务对象不能再使用

7. **`Rollback() error`**
   - 回滚事务，取消所有操作
   - 返回值：错误信息
   - 回滚后，事务内的所有操作都不会生效
   - 回滚后，事务对象不能再使用

#### 9.9.4 事务使用示例

**示例1：基本事务操作**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 创建或获取表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 设置表字段
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
        "email": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // 开始事务
    tx, err := table.Begin()
    if err != nil {
        panic(err)
    }
    
    // 事务内插入记录
    insertRecord := map[string]any{"name": "张三", "age": 25, "email": "zhangsan@example.com"}
    id, err := tx.Insert(&insertRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Printf("事务内插入记录成功，ID: %d\n", id)
    
    // 事务内读取刚插入的记录
    readRecord := map[string]any{"id": id}
    recordData, err := tx.Read(&readRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Printf("事务内读取到刚插入的记录: %v\n", recordData)
    
    // 事务内更新记录
    updateRecord := map[string]any{"id": id, "age": 26}
    err = tx.Update(&updateRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Println("事务内更新记录成功")
    
    // 提交事务
    err = tx.Commit()
    if err != nil {
        panic(err)
    }
    fmt.Println("事务提交成功")
}
```

**示例2：事务回滚**

```go
// 开始事务
tx, err := table.Begin()
if err != nil {
    panic(err)
}

// 事务内插入记录
insertRecord := map[string]any{"name": "测试回滚", "age": 30, "email": "test@example.com"}
_, err = tx.Insert(&insertRecord)
if err != nil {
    tx.Rollback()
    panic(err)
}

// 事务内读取刚插入的记录
readRecord := map[string]any{"name": "测试回滚"}
iter := tx.Search(&readRecord)
defer iter.Release()
records := iter.GetRecords(true)
fmt.Printf("事务内读取到%d条记录\n", len(records))

// 回滚事务
err = tx.Rollback()
if err != nil {
    panic(err)
}
fmt.Println("事务回滚成功")

// 验证记录是否被回滚
readRecordAfterRollback := map[string]any{"name": "测试回滚"}
iterAfterRollback := table.Search(&readRecordAfterRollback)
defer iterAfterRollback.Release()
recordsAfterRollback := iterAfterRollback.GetRecords(true)
fmt.Printf("回滚后读取到%d条记录\n", len(recordsAfterRollback))
```

**示例3：事务的读一致性**

```go
// 开始事务
tx, err := table.Begin()
if err != nil {
    panic(err)
}

// 事务内读取初始数据
readRecord := map[string]any{"id": 1}
recordBeforeUpdate := readSingleRecord(tx, &readRecord)
fmt.Printf("事务开始时读取到的数据: %v\n", recordBeforeUpdate)

// 事务外更新数据
updateRecordOutside := map[string]any{"id": 1, "age": 100}
err = table.Update(&updateRecordOutside)
if err != nil {
    tx.Rollback()
    panic(err)
}
fmt.Println("事务外更新数据成功")

// 事务内再次读取数据，应该仍然是初始数据（读一致性）
recordInsideTx := readSingleRecord(tx, &readRecord)
fmt.Printf("事务内再次读取到的数据: %v\n", recordInsideTx)

// 提交事务
err = tx.Commit()
if err != nil {
    panic(err)
}

// 事务提交后读取数据，应该是更新后的数据
recordAfterCommit := readSingleRecord(table, &readRecord)
fmt.Printf("事务提交后读取到的数据: %v\n", recordAfterCommit)
```

#### 9.9.5 使用建议

1. **单表事务**：直接使用 `table.Begin()` 创建的事务，操作简洁方便
2. **跨表事务**：目前需要手动管理 batch，直接使用底层 API：
   ```go
   // 获取全局batch
   batch := storage.KVDb.GetBatch()
   
   // 跨表操作
   _, err := table1.Insert(&record1, batch)
   _, err := table2.Insert(&record2, batch)
   
   // 提交batch
   err = storage.KVDb.WriteBatch(batch)
   ```

#### 9.9.6 事务注意事项

1. **事务生命周期**：事务对象在 `Commit()` 或 `Rollback()` 后不能再使用
2. **读一致性**：事务内的读取操作基于快照，不受外部修改影响
3. **读自己的写**：事务内可以立即读取到自己刚刚写入的数据
4. **并发控制**：
   - 直接使用底层 `batch` API 支持并发写入操作
   - 通过 `table.Begin()` 方法创建的事务支持并发，每个事务有自己独立的快照和缓存，事务之间完全隔离
   - 可以在多线程中安全使用 `table.Begin()` 创建的事务，它们不会相互影响
   - 每个事务的写入操作最终会通过原子提交（`Commit()`）写入数据库，保证数据一致性
5. **性能考虑**：长时间持有事务会影响数据库性能，建议事务操作尽快完成
6. **错误处理**：必须处理事务操作中可能发生的错误，并在发生错误时回滚事务
7. **资源释放**：事务对象在使用完毕后会自动释放资源，无需手动处理

### 9.9 快照功能

sfsDb 支持 LevelDB 快照功能，可以创建数据库的一致性读取视图，用于读取一致性数据或实现读写分离。

#### 9.9.1 快照功能介绍

LevelDB 快照提供了数据库在某个时间点的一致性读取视图，即使在快照创建后数据库发生了修改，快照中的数据也不会变化。

**主要特性**：
- 一致性读取：快照创建后，无论数据库如何修改，快照中的数据始终保持一致
- 读写分离：可以在快照模式下进行读取操作，不影响写入操作
- 高效创建：快照创建操作非常高效，只需要很少的时间和内存
- 资源自动管理：快照资源会在切换回数据库模式时自动释放

#### 9.9.2 快照相关方法

```go
// SwitchToSnapshot 将当前实例切换到快照模式
// 如果当前已是快照模式，则返回错误
func (s *LevelDBStore) SwitchToSnapshot() error

// SwitchToDB 将当前实例切换回数据库模式
// 如果当前已是数据库模式，则返回错误
// 如果当前是快照模式，会先释放快照资源
func (s *LevelDBStore) SwitchToDB() error
```

#### 9.9.3 使用示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建表
    table, err := engine.TableNew("test_snapshot")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 初始插入一条记录
    record1 := map[string]any{"id": 1, "name": "初始数据", "age": 20}
    _, err = table.Insert(&record1)
    if err != nil {
        panic(err)
    }

    // 检查当前kvStore是否为LevelDBStore类型
    levelDBStore, ok := table.kvStore.(*storage.LevelDBStore)
    if !ok {
        fmt.Println("当前存储引擎不支持快照功能")
        return
    }

    // 创建快照
    err = levelDBStore.SwitchToSnapshot()
    if err != nil {
        panic(err)
    }
    fmt.Println("成功创建快照")

    // 在快照模式下读取数据
    readRecord1 := map[string]any{"id": 1}
    snapshotResult, err := table.Read(&readRecord1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("快照模式读取到的记录: %v\n", snapshotResult)

    // 切换回DB模式
    err = levelDBStore.SwitchToDB()
    if err != nil {
        panic(err)
    }
    fmt.Println("成功切换回DB模式")

    // 更新数据
    updateRecord := map[string]any{"id": 1, "name": "更新后数据", "age": 25}
    err = table.Update(&updateRecord)
    if err != nil {
        panic(err)
    }
    fmt.Println("成功更新记录")

    // 再次创建快照，此时快照会包含更新后的数据
    err = levelDBStore.SwitchToSnapshot()
    if err != nil {
        panic(err)
    }
    fmt.Println("成功再次创建快照")

    // 在新快照模式下读取数据，应该是更新后的数据
    newSnapshotResult, err := table.Read(&readRecord1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("新快照模式读取到的记录: %v\n", newSnapshotResult)

    // 切换回DB模式
    err = levelDBStore.SwitchToDB()
    if err != nil {
        panic(err)
    }
    fmt.Println("成功切换回DB模式")
}
```

#### 9.9.4 注意事项

1. **快照生命周期**：快照会在切换回数据库模式时自动释放资源
2. **嵌套快照**：不支持嵌套快照，即不能在快照模式下再次创建快照
3. **性能影响**：快照创建操作非常高效，但持有快照会影响 LevelDB 的垃圾回收，建议及时释放
4. **读取一致性**：快照提供的是创建时刻的一致性视图，之后的修改不会影响快照
5. **存储引擎支持**：只有 LevelDBStore 支持快照功能，其他存储引擎可能不支持
6. **事务交互**：快照只能读取到已提交事务的数据，未提交的事务对快照不可见

### 9.10 手动事务控制

sfsDb 支持手动事务控制，允许将多个操作（插入、更新、删除）组合到一个事务中，实现原子性操作。

#### 9.10.1 事务控制方法

```go
// 获取批量操作对象
func (s *LevelDBStore) GetBatch() Batch

// 提交批量操作
func (s *LevelDBStore) WriteBatch(batch Batch) error
```

#### 9.10.2 支持事务的操作

| 操作 | 方法签名 |
|------|---------|
| 插入 | `func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (int, error)` |
| 更新 | `func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error` |
| 删除 | `func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error` |

#### 9.10.3 使用示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./test_transaction")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建表
    table, err := engine.TableNew("test_manual_transaction")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 创建batch，手动控制事务
    batch := table.kvStore.GetBatch()
    if batch == nil {
        panic("获取batch失败")
    }

    fmt.Println("=== 在同一个事务中执行多个操作 ===")
    
    // 1. 插入第一条记录
    record1 := map[string]any{"id": 1, "name": "张三", "age": 20}
    _, err = table.Insert(&record1, batch)
    if err != nil {
        panic(fmt.Sprintf("插入第一条记录失败: %v", err))
    }
    fmt.Println("已添加第一条记录到事务")

    // 2. 插入第二条记录
    record2 := map[string]any{"id": 2, "name": "李四", "age": 25}
    _, err = table.Insert(&record2, batch)
    if err != nil {
        panic(fmt.Sprintf("插入第二条记录失败: %v", err))
    }
    fmt.Println("已添加第二条记录到事务")

    // 3. 更新第一条记录
    updateRecord1 := map[string]any{"id": 1, "name": "张三更新", "age": 21}
    err = table.Update(&updateRecord1, batch)
    if err != nil {
        panic(fmt.Sprintf("更新记录失败: %v", err))
    }
    fmt.Println("已添加更新操作到事务")

    // 4. 提交事务
    err = table.kvStore.WriteBatch(batch)
    if err != nil {
        panic(fmt.Sprintf("提交事务失败: %v", err))
    }
    fmt.Println("事务提交成功")

    // 验证所有操作是否生效
    fmt.Println("\n=== 验证事务结果 ===")
    
    // 读取第一条记录，应该是更新后的数据
    readRecord1 := map[string]any{"id": 1}
    result1, err := table.Read(&readRecord1)
    if err != nil {
        panic(fmt.Sprintf("读取第一条记录失败: %v", err))
    }
    fmt.Printf("第一条记录: %v\n", result1)

    // 读取第二条记录，应该存在
    readRecord2 := map[string]any{"id": 2}
    result2, err := table.Read(&readRecord2)
    if err != nil {
        panic(fmt.Sprintf("读取第二条记录失败: %v", err))
    }
    fmt.Printf("第二条记录: %v\n", result2)

    fmt.Println("\n事务执行成功，所有操作都已生效")
}
```

#### 9.10.4 事务回滚

在 sfsDb 中，事务回滚通过不提交事务来实现。如果在事务执行过程中发生错误，可以选择不调用 `WriteBatch` 方法，此时所有操作都不会生效，相当于回滚。

```go
// 创建batch
batch := table.kvStore.GetBatch()

// 执行一些操作
_, err = table.Insert(&record1, batch)
if err != nil {
    // 发生错误，不提交事务，相当于回滚
    fmt.Println("事务回滚")
    return
}

// 执行另一些操作
_, err = table.Insert(&record2, batch)
if err != nil {
    // 发生错误，不提交事务，相当于回滚
    fmt.Println("事务回滚")
    return
}

// 所有操作都成功，提交事务
err = table.kvStore.WriteBatch(batch)
if err != nil {
    panic(err)
}
```

#### 9.10.5 注意事项

1. **原子性**：事务中的所有操作要么全部成功，要么全部失败
2. **隔离性**：事务执行过程中，其他操作看不到中间结果
3. **一致性**：事务执行前后，数据库保持一致状态
4. **持久性**：事务提交后，数据持久化到存储
5. **批量操作**：手动事务适用于批量操作，可以提高性能
6. **资源管理**：未提交的事务会占用系统资源，建议及时处理
7. **嵌套事务**：不支持嵌套事务
8. **快照交互**：事务提交后，新创建的快照会包含事务中的修改

### 9.11 ACID事务

ACID 是数据库事务的四个核心特性，确保了数据库事务的可靠性和一致性。sfsDb 通过结合批量操作和快照功能，实现了完整的 ACID 事务支持。

#### 9.11.1 ACID 概念

| 特性 | 描述 |
|------|------|
| **原子性(Atomicity)** | 一个事务中的所有操作要么全部成功，要么全部失败，不会出现部分成功部分失败的情况 |
| **一致性(Consistency)** | 事务执行前后，数据库从一个一致性状态变换到另一个一致性状态，不会出现数据损坏或无效状态 |
| **隔离性(Isolation)** | 多个事务并发执行时，一个事务的执行不应影响其他事务的执行，每个事务都有独立的执行环境 |
| **持久性(Durability)** | 事务提交后，其结果应该永久保存在数据库中，即使系统崩溃或重启，数据也不会丢失 |

#### 9.11.2 sfsDb 的 ACID 实现

sfsDb 通过以下机制实现 ACID 事务：

1. **原子性**：
   - 使用 `GetBatch()` 和 `WriteBatch()` 方法，将多个操作组合到一个批处理中
   - `WriteBatch()` 方法确保所有操作要么全部成功，要么全部失败

2. **一致性**：
   - 事务执行前后进行数据验证
   - 快照功能提供一致性读取视图
   - 不允许无效的数据状态

3. **隔离性**：
   - 批量操作确保事务的原子执行
   - 快照功能提供读取一致性
   - 并发操作不同记录时，相互隔离

4. **持久性**：
   - LevelDB 的 WriteBatch 机制确保数据持久化
   - 事务提交后，数据立即写入磁盘

#### 9.11.3 完整 ACID 事务示例

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./test_acid")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建银行账户表
    accountTable, err := engine.TableNew("bank_accounts")
    if err != nil {
        panic(fmt.Sprintf("创建表失败: %v", err))
    }

    // 设置表字段
    accountFields := map[string]any{
        "id":      0,
        "name":    "",
        "balance": 0.0,
    }
    err = accountTable.SetFields(accountFields)
    if err != nil {
        panic(fmt.Sprintf("设置表字段失败: %v", err))
    }

    // 创建主键索引
    accountPK, err := engine.DefaultPrimaryKeyNew("pk_id")
    if err != nil {
        panic(fmt.Sprintf("创建主键失败: %v", err))
    }
    accountPK.AddFields("id")
    err = accountTable.CreateIndex(accountPK)
    if err != nil {
        panic(fmt.Sprintf("创建主键索引失败: %v", err))
    }

    // 初始化账户数据
    account1 := map[string]any{"id": 1, "name": "张三", "balance": 1000.0}
    account2 := map[string]any{"id": 2, "name": "李四", "balance": 2000.0}
    
    _, err = accountTable.Insert(&account1)
    if err != nil {
        panic(fmt.Sprintf("插入账户1失败: %v", err))
    }
    _, err = accountTable.Insert(&account2)
    if err != nil {
        panic(fmt.Sprintf("插入账户2失败: %v", err))
    }

    fmt.Println("初始账户余额:")
    printAccountBalance(accountTable, 1)
    printAccountBalance(accountTable, 2)

    // 执行转账事务
    fmt.Println("\n执行转账事务：张三向李四转账500元")
    
    // 1. 创建事务
    batch := accountTable.kvStore.GetBatch()
    if batch == nil {
        panic("获取batch失败")
    }

    // 2. 更新张三的账户（减少500元）
    updateAccount1 := map[string]any{"id": 1, "balance": 500.0}
    err = accountTable.Update(&updateAccount1, batch)
    if err != nil {
        panic(fmt.Sprintf("更新张三账户失败: %v", err))
    }

    // 3. 更新李四的账户（增加500元）
    updateAccount2 := map[string]any{"id": 2, "balance": 2500.0}
    err = accountTable.Update(&updateAccount2, batch)
    if err != nil {
        panic(fmt.Sprintf("更新李四账户失败: %v", err))
    }

    // 4. 提交事务
    err = accountTable.kvStore.WriteBatch(batch)
    if err != nil {
        panic(fmt.Sprintf("提交转账事务失败: %v", err))
    }

    fmt.Println("转账事务提交成功")

    // 验证转账结果
    fmt.Println("\n转账后账户余额:")
    printAccountBalance(accountTable, 1)
    printAccountBalance(accountTable, 2)

    // 验证总额一致性
    total := getAccountBalance(accountTable, 1) + getAccountBalance(accountTable, 2)
    fmt.Printf("\n账户总额: %.2f\n", total)
    if total != 3000.0 {
        panic(fmt.Sprintf("账户总额不一致，期望3000.0，实际%.2f", total))
    }
    fmt.Println("账户总额一致，事务执行成功")
}

// 辅助函数：打印账户余额
func printAccountBalance(table *engine.Table, id int) {
    readRecord := map[string]any{"id": id}
    iter := table.Search(&readRecord)
    defer iter.Release()
    records := iter.GetRecords(true)
    if len(records) == 1 {
        fmt.Printf("%s: %.2f\n", records[0]["name"], records[0]["balance"])
    }
}

// 辅助函数：获取账户余额
func getAccountBalance(table *engine.Table, id int) float64 {
    readRecord := map[string]any{"id": id}
    iter := table.Search(&readRecord)
    defer iter.Release()
    records := iter.GetRecords(true)
    if len(records) == 1 {
        return records[0]["balance"].(float64)
    }
    return 0.0
}
```

#### 9.11.4 并发情况下的事务处理

sfsDb 在并发情况下的事务表现：

1. **并发写入不同记录**：
   - 完全支持，不会出现冲突
   - 每个事务独立执行，互不影响
   - 适合高并发写入场景

2. **并发更新同一记录**：
   - 会出现更新丢失的情况
   - 这是因为 sfsDb 目前不支持行级锁或乐观锁
   - 建议在应用层实现额外的同步机制

3. **并发读写**：
   - 读取操作不会阻塞写入操作
   - 写入操作不会阻塞读取操作
   - 快照功能提供一致性读取视图

#### 9.11.5 ACID 事务的使用场景

1. **金融交易**：转账、支付等涉及资金变动的操作
2. **订单处理**：创建订单、扣减库存等组合操作
3. **数据迁移**：批量数据导入导出
4. **复杂更新**：需要修改多个相关记录的操作
5. **数据一致性要求高的场景**：确保数据完整性和一致性

#### 9.11.6 注意事项

1. **并发更新同一记录**：
   - sfsDb 目前不支持行级锁或乐观锁
   - 并发更新同一记录时，会出现更新丢失
   - 建议在应用层实现同步机制

2. **事务大小**：
   - 过大的事务会占用较多内存和磁盘空间
   - 建议将大事务拆分为多个小事务
   - 每个事务处理的数据量不宜过大

3. **错误处理**：
   - 始终检查并处理事务操作的错误
   - 及时回滚失败的事务
   - 避免长时间持有未提交的事务

4. **性能考虑**：
   - 事务操作比单条操作稍慢
   - 批量操作可以提高性能
   - 合理使用事务，避免不必要的事务

5. **快照交互**：
   - 事务提交后，新创建的快照会包含事务中的修改
   - 快照创建后，事务的修改不会影响快照中的数据

#### 9.11.7 最佳实践

1. **保持事务简短**：
   - 事务执行时间越短，并发冲突的可能性越小
   - 避免在事务中执行长时间的计算或IO操作

2. **批量操作**：
   - 将多个相关操作组合到一个事务中
   - 减少事务的数量，提高性能

3. **合理设计数据模型**：
   - 避免频繁更新同一记录
   - 设计合理的索引，提高查询性能

4. **错误处理**：
   - 实现完善的错误处理机制
   - 记录事务执行日志，便于调试和审计

5. **测试**：
   - 在开发环境中测试并发情况下的事务表现
   - 验证事务的ACID特性
   - 测试异常情况下的事务回滚

#### 9.11.8 使用建议

1. **单表事务**：
   - 直接使用 `table.Begin()` 创建的事务（当前设计）
   - 封装良好，使用简单，适合大多数单表操作场景
   - 示例：
     ```go
     // 开始事务
     tx, err := table.Begin()
     if err != nil {
         panic(err)
     }
     
     // 执行操作
     id, err := tx.Insert(&record1)
     if err != nil {
         tx.Rollback()
         panic(err)
     }
     
     // 提交事务
     err = tx.Commit()
     if err != nil {
         panic(err)
     }
     ```

2. **跨表事务**：
   - 目前需要手动管理 `batch`，直接使用底层 API
   - 适合需要在多个表之间保持原子性的操作
   - 示例：
     ```go
     // 获取batch
     batch := storage.KVDb.GetBatch()
     
     // 在表1中执行操作
     table1.Insert(&record1, batch)
     table1.Update(&record2, batch)
     
     // 在表2中执行操作
     table2.Insert(&record3, batch)
     table2.Delete(&record4, batch)
     
     // 提交事务
     storage.KVDb.WriteBatch(batch)
     ```

通过合理使用ACID事务，可以确保sfsDb在各种场景下的数据一致性和可靠性，特别是对于需要高可靠性的业务系统。

### 9.12 表结构序列化

#### 9.12.1 TableSchema 结构

`TableSchema` 是 `Table` 结构体的映射，用于序列化和反序列化表的元数据。它包含表的基本信息，但不包含运行时状态（如counter和kvStore）。

**TableSchema 结构定义**：

```go
type TableSchema struct {
    // 表ID
    ID uint8 `json:"id"`
    // 表名
    Name string `json:"name"`

    // 字段定义，string为字段名，any为字段值示例（用于类型推断）
    Fields map[string]any `json:"fields"`

    // 索引信息，包含所有索引的基本信息
    Indexes []IndexSchema `json:"indexes"`
}
```

**字段说明**：
- `ID`：表的唯一标识符
- `Name`：表名
- `Fields`：表的字段定义，包含字段名和字段值示例
- `Indexes`：表的索引信息，包含所有索引的基本信息

#### 9.12.2 TableSerialization 结构

`TableSerialization` 包含 `Table` 的完整序列化信息，包括运行时状态。用于完整的 `Table` 序列化和反序列化。

**TableSerialization 结构定义**：

```go
type TableSerialization struct {
    // TableSchema 元数据
    Schema *TableSchema `json:"schema"`

    // 自动增值的当前值（counter AutoInt）
    CurrentAutoID int `json:"current_auto_id"`
}
```

**字段说明**：
- `Schema`：表的元数据，类型为 `TableSchema`
- `CurrentAutoID`：自动增值的当前值

#### 9.12.3 IndexSchema 结构

`IndexSchema` 是 `Index` 的映射，用于序列化和反序列化索引的基本信息。

**IndexSchema 结构定义**：

```go
type IndexSchema struct {
    // 索引名称
    Name string `json:"name"`

    // 索引类型：primary, normal, fulltext
    Type string `json:"type"`

    // 索引字段，顺序表示索引顺序
    Fields []string `json:"fields"`

    // 是否唯一索引
    Unique bool `json:"unique"`
}
```

**字段说明**：
- `Name`：索引名称
- `Type`：索引类型，包括 primary（主键索引）、normal（普通索引）、fulltext（全文索引）
- `Fields`：索引字段列表，顺序表示索引顺序
- `Unique`：是否为唯一索引

#### 9.12.4 序列化和反序列化方法

**将 Table 转换为 TableSchema**：

```go
func (t *Table) ToSchema() *TableSchema
```

**将 Table 转换为 TableSerialization**：

```go
func (t *Table) ToSerialization() *TableSerialization
```

**从 TableSchema 创建 Table**：

```go
func FromSchema(schema *TableSchema) (*Table, error)
```

**从 TableSerialization 创建 Table**：

```go
func FromSerialization(serialization *TableSerialization) (*Table, error)
```

**将 Table 转换为 JSON 字符串**：

```go
func TableToJSON(t *Table) (string, error)
```

**从 JSON 字符串创建 Table**：

```go
func TableFromJSON(jsonStr string) (*Table, error)
```

#### 9.12.5 使用示例

**示例 1：将表结构序列化为 JSON**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 创建或获取表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 设置字段
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    if err := table.SetFields(fields); err != nil {
        panic(err)
    }
    
    // 创建索引
    if err := table.CreatePrimaryKey("id"); err != nil {
        panic(err)
    }
    if err := table.CreateSimpleIndex("idx_name_age", "name", "age"); err != nil {
        panic(err)
    }
    
    // 将表结构序列化为 JSON
    jsonStr, err := engine.TableToJSON(table)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("表结构 JSON:")
    fmt.Println(jsonStr)
}
```

**示例 2：从 JSON 字符串创建表**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 表结构 JSON 字符串
    jsonStr := `{
        "schema": {
            "id": 1,
            "name": "users",
            "fields": {
                "id": 0,
                "name": "",
                "age": 0
            },
            "indexes": [
                {
                    "name": "pk",
                    "type": "primary",
                    "fields": ["id"],
                    "unique": true
                },
                {
                    "name": "idx_name_age",
                    "type": "normal",
                    "fields": ["name", "age"],
                    "unique": false
                }
            ]
        },
        "current_auto_id": 0
    }`
    
    // 从 JSON 字符串创建表
    table, err := engine.TableFromJSON(jsonStr)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("表创建成功: %s (ID: %d)\n", table.GetName(), table.GetId())
    fmt.Println("表字段:")
    for field := range table.GetAllFields() {
        fmt.Printf("  - %s\n", field)
    }
    
    fmt.Println("表索引:")
    for _, index := range table.GetAllIndexes() {
        fmt.Printf("  - %s (类型: %T)\n", index.Name(), index)
    }
}
```

#### 9.12.6 使用场景

**TableSchema 和相关序列化功能的主要使用场景**：

1. **表结构备份和恢复**：将表结构序列化为 JSON 进行备份，需要时从 JSON 恢复
2. **表结构迁移**：在不同环境之间迁移表结构
3. **表结构分析**：分析表的结构和索引信息
4. **元数据存储**：存储表的元数据信息
5. **配置管理**：将表结构作为配置的一部分进行管理

#### 9.12.7 注意事项

1. **运行时状态**：TableSchema 不包含运行时状态（如 counter 和 kvStore），这些会在反序列化后重新初始化
2. **字段类型**：Fields 中的 any 类型是字段值示例，用于类型推断
3. **索引信息**：Indexes 包含所有索引的基本信息，但不包含索引的完整实现
4. **JSON 格式**：序列化和反序列化使用标准 JSON 格式，便于存储和传输
5. **错误处理**：序列化和反序列化过程中可能会出现错误，应始终检查并处理错误

## 11. 最佳实践

### 11.1 性能优化

1. **合理设计索引**：为频繁查询的字段创建索引
2. **避免全表扫描**：使用索引字段进行查询
3. **合理设置字段类型**：根据实际数据选择合适的字段类型
4. **及时释放资源**：使用 `defer iter.Release()` 确保迭代器资源被释放
5. **批量操作**：对于大量数据，使用批量插入和更新

#### 11.1.1 对象池和 Finalizer 机制

sfsDb 实现了对象池和 finalizer 机制，用于优化 Record 和 Records 对象的内存使用和性能：

**对象池工作原理**：
- 使用 `sync.Pool` 管理 Record 和 Records 对象的复用
- 通过 `GetRecord()` 和 `GetRecords()` 从池中获取对象
- 通过 `PutRecord()` 和 `PutRecords()` 将对象放回池

**Finalizer 机制**：
- 使用 `runtime.SetFinalizer` 为对象设置最终izer
- 当对象被垃圾回收时，自动将对象放回池中
- 确保即使忘记手动释放对象，系统也能正常工作

**使用示例**：

```go
// 正确使用 Record 对象
record := record.GetRecord()
defer record.PutRecord(record)

// 设置字段
record["id"] = 1
record["name"] = "Alice"

// 使用 record...

// 正确使用 Records 对象
records := record.GetRecords()
defer record.PutRecords(records)

// 添加记录
records = append(records, record)

// 使用 records...
```

**从 TableIter 获取记录的正确方式**：

```go
// 从迭代器获取记录
iter := table.Search(&searchFields)
defer iter.Release()

records := iter.GetRecords(true)
defer record.PutRecords(records) // 确保释放对象

// 使用 records...
for _, r := range records {
    fmt.Printf("Record: %v\n", r)
}
```

**性能优势**：
1. **减少内存分配**：通过对象复用，减少频繁创建和销毁对象的开销
2. **降低垃圾回收压力**：减少对象创建，降低垃圾回收的频率和开销
3. **双重保障**：既支持手动释放（性能更好），也支持自动释放（更安全）
4. **向后兼容**：不破坏现有代码，只是增强了安全性

**最佳实践**：
- 始终使用 `defer PutRecord()` 或 `defer PutRecords()` 确保对象被正确释放
- 对于性能敏感的场景，优先使用手动释放
- 对于复杂代码，依赖 finalizer 机制作为安全保障

通过对象池和 finalizer 机制，sfsDb 在处理大量记录时能够保持更好的性能和内存使用效率。

### 11.2 数据安全

1. **定期备份**：定期备份数据库文件
2. **错误处理**：始终检查并处理错误
3. **验证输入数据**：在插入前验证数据的合法性
4. **使用事务**：对于复杂操作，使用事务确保数据一致性

### 11.3 代码规范

1. **使用唯一表名**：在测试环境中使用时间戳生成唯一表名
2. **注释代码**：为复杂查询和操作添加注释
3. **遵循工作流**：修改字段时严格遵循先 `UpdateFieldName` 后 `SetFields` 的工作流
4. **使用常量**：为字段名和表名使用常量定义

## 12. 常见问题

### 12.1 字段修改后索引失效

**问题**：修改字段名称后，索引无法使用

**解决方案**：确保严格遵循字段修改工作流：先调用 `UpdateFieldName`，再调用 `SetFields`。`UpdateFieldName` 会自动更新所有索引中的字段名。

### 12.2 插入数据时主键冲突

**问题**：插入数据时出现主键冲突错误

**解决方案**：
1. 使用自动增值功能，不手动指定主键值
2. 确保手动指定的主键值唯一
3. 使用事务处理批量插入

### 12.3 全文搜索结果不符合预期

**问题**：全文搜索没有返回预期结果

**解决方案**：
1. 确保已正确创建全文索引
2. 检查全文索引字段设置是否正确
3. 验证搜索关键词是否与索引字段内容匹配
4. 检查全文索引的长度设置是否合适

### 12.4 查询性能慢

**问题**：查询操作响应时间长

**解决方案**：
1. 为查询字段创建索引
2. 优化查询条件，避免复杂的比较操作
3. 使用分页功能减少返回数据量
4. 考虑使用复合索引优化多字段查询

## 13. 总结

sfsDb 是一个灵活、高效的嵌入式数据库，支持多种数据类型、索引类型和查询方式。通过本文档的示例和说明，您应该能够掌握 sfsDb 的基本使用方法和最佳实践。

主要特点：

1. **灵活的数据模型**：支持多种字段类型和动态字段
2. **强大的索引系统**：支持单主键、复合主键、普通索引和全文索引
3. **丰富的查询功能**：支持比较操作符、自定义匹配器和全文搜索
4. **易于使用的 API**：简洁的 API 设计，易于集成到各种应用中
5. **高效的性能**：优化的存储结构和查询算法

建议在实际项目中，根据具体需求选择合适的功能和优化策略，以获得最佳的性能和可靠性。