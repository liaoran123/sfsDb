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