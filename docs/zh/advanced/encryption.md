# 数据加密

sfsDb 提供了内置的数据加密功能，支持 AES-256-GCM 加密算法，为您的敏感数据提供安全保障。

## 概述

加密功能通过 `EncryptedStoreWrapper` 实现，它包装了底层的存储引擎，在数据写入时自动加密，读取时自动解密。

### 核心特性

- **AES-256-GCM 加密**：采用工业标准的加密算法，提供认证加密
- **密钥派生**：支持使用密码通过 PBKDF2 派生密钥
- **并发安全**：使用 `atomic.Value` 保证加密器的线程安全访问
- **解密缓存**：内置 LRU 缓存提高解密性能
- **密钥轮换**：支持运行时重新加密所有数据

## 快速开始

### 1. 使用主密钥加密

最简单的方式是直接提供 32 字节（256 位）的主密钥：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./encrypted_db"

    // 创建 32 字节（256 位）的主密钥
    masterKey := make([]byte, 32)
    // 注意：在生产环境中，请使用安全的密钥生成方式
    // 例如：crypto/rand.Read(masterKey)

    // 创建加密配置
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        MasterKey: masterKey,
    }

    // 创建带加密的存储
    dbManager := storage.GetDBManager()
    store, err := dbManager.NewLevelDBStore(dbPath, nil, encryptConfig)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()

    fmt.Println("加密存储初始化成功！")
}
```

### 2. 使用密码派生密钥

您也可以使用密码，系统会自动通过 PBKDF2 派生密钥：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./encrypted_db"

    // 创建加密配置，使用密码
    encryptConfig := &storage.EncryptionConfig{
        Enabled:    true,
        Password:   "your-strong-password",
        Salt:       []byte("your-random-salt"), // 可选，如果不提供会自动生成
        Iterations: 100000, // 可选，默认 100000
    }

    // 创建带加密的存储
    dbManager := storage.GetDBManager()
    store, err := dbManager.NewLevelDBStore(dbPath, nil, encryptConfig)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()

    fmt.Println("加密存储初始化成功！")

    // 获取保存的配置（包含自动生成的盐值）
    wrapper, ok := store.(*storage.EncryptedStoreWrapper)
    if ok {
        savedConfig := wrapper.GetEncryptionConfig()
        fmt.Printf("保存的盐值: %x\n", savedConfig.Salt)
    }
}
```

## EncryptionConfig 配置

`EncryptionConfig` 结构体包含以下配置项：

| 配置项 | 类型 | 说明 | 必填 |
|-------|------|------|------|
| `Enabled` | `bool` | 是否启用加密 | 是 |
| `Algorithm` | `string` | 加密算法，默认 "AES-256-GCM" | 否 |
| `MasterKey` | `[]byte` | 32 字节的主密钥 | 与 Password 二选一 |
| `Password` | `string` | 密码，用于派生密钥 | 与 MasterKey 二选一 |
| `Salt` | `[]byte` | 盐值，用于密码派生 | 否，自动生成 |
| `Iterations` | `int` | PBKDF2 迭代次数，默认 100000 | 否 |

### 密钥长度要求

- **MasterKey**：必须是 32 字节（256 位）
- **Password**：无长度限制，但建议使用强密码

## 使用场景配置与加密结合

您可以将场景配置与加密结合使用：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./edge_encrypted_db"

    // 创建加密配置
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: make([]byte, 32),
    }

    // 使用边缘计算场景配置 + 加密
    dbManager := storage.GetDBManager()
    store, err := dbManager.NewLevelDBStore(
        dbPath,
        storage.GetScenarioOptions(storage.ScenarioEdge),
        encryptConfig,
    )
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()

    fmt.Println("边缘计算场景 + 加密存储初始化成功！")
}
```

## 密钥轮换

sfsDb 支持运行时密钥轮换，即重新加密所有数据：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // ... 初始化 store ...

    // 转换为 EncryptedStoreWrapper
    wrapper, ok := store.(*storage.EncryptedStoreWrapper)
    if !ok {
        panic("store is not an EncryptedStoreWrapper")
    }

    // 生成新密钥
    newKey := make([]byte, 32)
    // crypto/rand.Read(newKey)

    // 执行密钥轮换（重新加密所有数据）
    err := wrapper.ReEncrypt(newKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("密钥轮换成功！")
}
```

## 场景加密建议

根据不同的边缘智能与 IoT 场景，建议如下：

| 场景 | 加密建议 | 说明 |
|------|----------|------|
| **边缘计算节点** | ⚠️ 可选但推荐 | 取决于数据敏感度 |
| **IoT 网关设备** | ⚠️ 可选但推荐 | 取决于设备部署环境 |
| **智能终端设备** | ✅ 强烈推荐 | 设备可能丢失或被盗 |

### 详细建议

1. **边缘计算节点**
   - 如果处理敏感工业数据 → 启用加密
   - 如果是普通监控数据 → 可选择不加密以提升性能

2. **IoT 网关设备**
   - 如果网关在物理安全可控的环境 → 可选
   - 如果网关在易接触或公共区域 → 建议启用

3. **智能终端设备** ⚠️ **高风险**
   - **必须启用加密**
   - 设备可能被盗
   - 本地数据包含敏感信息
   - 符合数据保护法规要求

## 性能考虑

加密会带来一定的性能开销，建议：

- **读多写少**：解密缓存会显著提升读性能
- **批量操作**：使用批量操作减少加密/解密次数
- **场景选择**：非敏感数据可选择不加密

## 获取加密配置

您可以获取当前的加密配置（返回副本以避免外部修改）：

```go
wrapper, ok := store.(*storage.EncryptedStoreWrapper)
if ok {
    config := wrapper.GetEncryptionConfig()
    fmt.Printf("算法: %s\n", config.Algorithm)
    fmt.Printf("是否启用: %v\n", config.Enabled)
}
```

## 安全最佳实践

1. **密钥管理**
   - 不要在代码中硬编码密钥
   - 使用安全的密钥管理系统
   - 定期轮换密钥

2. **密码安全**
   - 使用强密码
   - 配合随机盐值
   - 使用足够的迭代次数（建议 ≥ 100000）

3. **数据备份**
   - 备份时数据仍是加密的
   - 确保密钥也安全备份
   - 丢失密钥将导致数据永久无法访问

## 总结

sfsDb 的加密功能为边缘计算和 IoT 场景提供了灵活且安全的数据保护方案。您可以根据实际需求选择是否启用加密，以及使用主密钥还是密码派生的方式。
