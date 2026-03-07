# Data Encryption

sfsDb provides built-in data encryption functionality with AES-256-GCM algorithm, providing security for your sensitive data.

## Overview

Encryption is implemented through `EncryptedStoreWrapper`, which wraps the underlying storage engine, automatically encrypting data on write and decrypting on read.

### Core Features

- **AES-256-GCM Encryption**: Industry-standard encryption algorithm providing authenticated encryption
- **Key Derivation**: Supports key derivation from passwords using PBKDF2
- **Concurrency Safe**: Uses `atomic.Value` for thread-safe encryptor access
- **Decryption Cache**: Built-in LRU cache for improved decryption performance
- **Key Rotation**: Supports re-encrypting all data at runtime

## Quick Start

### 1. Using Master Key

The simplest way is to directly provide a 32-byte (256-bit) master key:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./encrypted_db"

    // Create a 32-byte (256-bit) master key
    masterKey := make([]byte, 32)
    // Note: In production, use secure key generation
    // Example: crypto/rand.Read(masterKey)

    // Create encryption config
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        MasterKey: masterKey,
    }

    // Create encrypted storage
    dbManager := storage.GetDBManager()
    store, err := dbManager.NewLevelDBStore(dbPath, nil, encryptConfig)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()

    fmt.Println("Encrypted storage initialized successfully!")
}
```

### 2. Using Password Derived Key

You can also use a password, and the system will automatically derive the key using PBKDF2:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./encrypted_db"

    // Create encryption config with password
    encryptConfig := &storage.EncryptionConfig{
        Enabled:    true,
        Password:   "your-strong-password",
        Salt:       []byte("your-random-salt"), // Optional, auto-generated if not provided
        Iterations: 100000, // Optional, default 100000
    }

    // Create encrypted storage
    dbManager := storage.GetDBManager()
    store, err := dbManager.NewLevelDBStore(dbPath, nil, encryptConfig)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()

    fmt.Println("Encrypted storage initialized successfully!")

    // Get saved config (includes auto-generated salt)
    wrapper, ok := store.(*storage.EncryptedStoreWrapper)
    if ok {
        savedConfig := wrapper.GetEncryptionConfig()
        fmt.Printf("Saved salt: %x\n", savedConfig.Salt)
    }
}
```

## EncryptionConfig Configuration

The `EncryptionConfig` struct contains the following configuration options:

| Configuration | Type | Description | Required |
|--------------|------|-------------|----------|
| `Enabled` | `bool` | Whether encryption is enabled | Yes |
| `Algorithm` | `string` | Encryption algorithm, default "AES-256-GCM" | No |
| `MasterKey` | `[]byte` | 32-byte master key | Either this or Password |
| `Password` | `string` | Password for key derivation | Either this or MasterKey |
| `Salt` | `[]byte` | Salt for password derivation | No, auto-generated |
| `Iterations` | `int` | PBKDF2 iterations, default 100000 | No |

### Key Length Requirements

- **MasterKey**: Must be 32 bytes (256 bits)
- **Password**: No length limit, but strong password recommended

## Using Scenario Configuration with Encryption

You can combine scenario configuration with encryption:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbPath := "./edge_encrypted_db"

    // Create encryption config
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: make([]byte, 32),
    }

    // Use edge computing scenario configuration + encryption
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

    fmt.Println("Edge computing scenario + encrypted storage initialized successfully!")
}
```

## Key Rotation

sfsDb supports runtime key rotation, which re-encrypts all data:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // ... initialize store ...

    // Convert to EncryptedStoreWrapper
    wrapper, ok := store.(*storage.EncryptedStoreWrapper)
    if !ok {
        panic("store is not an EncryptedStoreWrapper")
    }

    // Generate new key
    newKey := make([]byte, 32)
    // crypto/rand.Read(newKey)

    // Perform key rotation (re-encrypt all data)
    err := wrapper.ReEncrypt(newKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("Key rotation successful!")
}
```

## Scenario Encryption Recommendations

Based on different edge intelligence and IoT scenarios, recommendations are as follows:

| Scenario | Encryption Recommendation | Notes |
|----------|-------------------------|-------|
| **Edge Computing Nodes** | ⚠️ Optional but recommended | Depends on data sensitivity |
| **IoT Gateway Devices** | ⚠️ Optional but recommended | Depends on deployment environment |
| **Smart Terminal Devices** | ✅ Strongly recommended | Devices may be lost or stolen |

### Detailed Recommendations

1. **Edge Computing Nodes**
   - If processing sensitive industrial data → Enable encryption
   - If normal monitoring data → Optional, disable for better performance

2. **IoT Gateway Devices**
   - If gateway in physically secure environment → Optional
   - If gateway in accessible or public area → Recommended to enable

3. **Smart Terminal Devices** ⚠️ **High Risk**
   - **Must enable encryption**
   - Devices may be stolen
   - Local data contains sensitive information
   - Compliance with data protection regulations

## Performance Considerations

Encryption incurs some performance overhead, recommendations:

- **Read-heavy workloads**: Decryption cache significantly improves read performance
- **Batch operations**: Use batch operations to reduce encryption/decryption cycles
- **Scenario selection**: Non-sensitive data may choose to skip encryption

## Getting Encryption Configuration

You can get the current encryption configuration (returns a copy to prevent external modification):

```go
wrapper, ok := store.(*storage.EncryptedStoreWrapper)
if ok {
    config := wrapper.GetEncryptionConfig()
    fmt.Printf("Algorithm: %s\n", config.Algorithm)
    fmt.Printf("Enabled: %v\n", config.Enabled)
}
```

## Security Best Practices

1. **Key Management**
   - Don't hardcode keys in code
   - Use secure key management systems
   - Rotate keys regularly

2. **Password Security**
   - Use strong passwords
   - Use with random salt
   - Use sufficient iterations (recommended ≥ 100000)

3. **Data Backup**
   - Data remains encrypted in backups
   - Ensure keys are also securely backed up
   - Losing keys will result in permanent data inaccessibility

## Summary

sfsDb's encryption functionality provides a flexible and secure data protection scheme for edge computing and IoT scenarios. You can choose whether to enable encryption based on actual needs, and whether to use master key or password derivation.
