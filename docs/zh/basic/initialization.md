# 数据库初始化

## 1.1 自定义数据库路径

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

## 1.2 自动使用默认路径

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

## 1.3 数据库关闭

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

## 1.4 使用外部存储实例

除了使用内置的存储引擎外，sfsDb 还支持使用外部实现的存储实例。只需通过`SetStore`函数设置即可：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// 自定义存储实现
type CustomStore struct {
    // 实现存储逻辑
}

// 实现 Store 接口的方法
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // 实现获取逻辑
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // 实现存储逻辑
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // 实现删除逻辑
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // 实现批量操作逻辑
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // 实现迭代器逻辑
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // 实现快照逻辑
    return nil, nil
}

func (s *CustomStore) Close() error {
    // 实现关闭逻辑
    return nil
}

func main() {
    // 创建自定义存储实例
    customStore := &CustomStore{}
    
    // 设置外部存储实例
    storage.SetStore(customStore)
    
    fmt.Println("外部存储实例设置成功")
    
    // 现在整个 sfsdb 项目都会使用这个自定义存储实例
    // 例如，创建表、插入数据等操作都会通过这个实例执行
}
```

**使用场景**：
- 集成第三方存储实现
- 为特定场景定制存储逻辑
- 在测试中使用内存存储或模拟存储
- 实现特殊的存储功能，如加密、压缩等

**注意事项**：
- 使用外部存储实例时，需要自行管理其生命周期
- 确保在不再使用时正确关闭存储实例
- 外部存储实现必须完整实现`Store`接口的所有方法