# Database Initialization

## 1.1 Custom Database Path

By default, sfsDb uses the `kvdb` folder in the current directory as the database storage path. If you need to customize the database path, you can call the `OpenDefaultDb` function when starting the program:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Custom database path
    dbPath := "./my_custom_db"
    _, err := storage.OpenDefaultDb(dbPath)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Database initialization successful")
}
```

## 1.2 Automatic Use of Default Path

If you don't call the `OpenDefaultDb` function, the system will automatically use the default path `./kvdb` when creating the table for the first time:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // When TableNew is called for the first time, the database will be automatically initialized with the default path "./kvdb"
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Table created successfully, database automatically initialized")
}
```

## 1.3 Database Closure

When the program ends, you can call the `CloseDb` function to close the database and release resources:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // Close database when program ends
    defer storage.CloseDb()
    
    fmt.Println("Table created successfully")
}
```