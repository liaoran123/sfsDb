# 创建表与设置字段

## 2.1 基本表创建/打开

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

## 2.2 支持的字段类型

sfsDb 使用 `any` 类型（即 `interface{}`）支持所有 Go 语言类型。下表列出了常用的类型示例：

| 类型示例 | 描述 |
|---------|------|
| `0`     | 整数类型（int、int32、int64、uint、uint32、uint64 等） |
| `0.0`   | 浮点类型（float32、float64） |
| `""`    | 字符串类型 |
| `false` | 布尔类型 |
| `time.Now()` | 时间类型 |
| `complex(1, 2)` | 复数类型（complex64、complex128） |

### 2.2.1 自定义类型处理

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

### 2.2.2 类型转换机制

sfsDb 在内部使用 `util` 包中的类型转换函数处理不同类型：

- `AnyToBytes()`：将任意类型转换为字节数组，用于存储
- `AnyToStr()`：将任意类型转换为字符串，用于显示或比较
- `StrToAny()`：将字符串转换为指定类型，用于读取

对于未明确支持的类型，这些函数都会使用 JSON 序列化/反序列化进行处理，确保所有 Go 语言类型都能被正确存储和读取。

### 2.2.3 注意事项

1. **自定义类型必须可序列化**：用于字段的自定义类型必须能够被 JSON 序列化，否则会导致存储失败
2. **性能考虑**：复杂自定义类型的 JSON 序列化和反序列化会带来一定的性能开销
3. **类型安全**：读取自定义类型时，需要确保类型匹配，否则可能导致运行时错误
4. **字段扩展**：可以随时向表中添加新字段，无需修改表结构
5. **默认值**：字段的默认值用于类型推断，实际存储时会使用插入的数据

通过这种设计，sfsDb 实现了真正的半结构化数据存储，既支持常见的基本类型，也能灵活处理复杂的自定义类型。

## 2.3 示例：创建包含各种类型字段的表

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