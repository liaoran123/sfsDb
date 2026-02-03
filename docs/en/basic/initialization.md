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

## 1.4 Using External Storage Instances

In addition to using built-in storage engines, sfsDb also supports using externally implemented storage instances. Simply set it through the `SetStore` function:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// Custom storage implementation
type CustomStore struct {
    // Implement storage logic
}

// Implement Store interface methods
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // Implement get logic
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // Implement put logic
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // Implement delete logic
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // Implement batch operation logic
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // Implement iterator logic
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // Implement snapshot logic
    return nil, nil
}

func (s *CustomStore) Close() error {
    // Implement close logic
    return nil
}

func main() {
    // Create custom storage instance
    customStore := &CustomStore{}
    
    // Set external storage instance
    storage.SetStore(customStore)
    
    fmt.Println("External storage instance set successfully")
    
    // Now the entire sfsdb project will use this custom storage instance
    // For example, creating tables, inserting data, etc. will all be executed through this instance
}
```

**Usage scenarios**:
- Integrating third-party storage implementations
- Customizing storage logic for specific scenarios
- Using in-memory storage or mock storage in tests
- Implementing special storage features such as encryption, compression, etc.

**Notes**:
- When using external storage instances, you need to manage their lifecycle yourself
- Ensure to properly close the storage instance when it's no longer needed
- The external storage implementation must fully implement all methods of the `Store` interface