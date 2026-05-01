
### 优化总结

#### 主要改进

| 配置项 | 改进内容 |
|--------|----------|
| **布隆过滤器** | ✅ 为所有场景（除 Game）启用了布隆过滤器，避免 95% 无效磁盘读取 |
| **MaxOpenFiles** | ✅ 添加到 Config 结构体（但 opt.Options 不支持该字段） |
| **Embedded OpenFiles** | ✅ 从 5 调整到 10，更合理 |
| **极限模式** | ✅ 新增 `extremeConfig` 专门针对 128MB 内存设备 |

#### 配置对比

| 场景 | WriteBuffer | BlockCache | OpenFiles | BloomFilter |
|------|-------------|------------|-----------|-------------|
| **Embedded** | 2MB | 4MB | 10 | ✅ 10bits |
| **IoT** | 4MB | 8MB | 10 | ✅ 10bits |
| **Edge** | 16MB | 32MB | 50 | ✅ 10bits |
| **Extreme** | 2MB | 4MB | 5 | ✅ 10bits |
| **Game** | 64MB | 128MB | 200 | ❌ 禁用 |

#### 极限生存模式（128MB 内存设备）

```go
var extremeConfig = Config{
    WriteBuffer:            2MB,      // 2MB 写入缓冲区
    OpenFilesCacheCapacity: 5,        // 5 个打开文件缓存
    BlockCacheCapacity:     4MB,      // 4MB 块缓存
    Compression:            Snappy,   // 默认压缩
    FilterBitsPerKey:       10,       // 必须启用布隆过滤器
}
```

#### 资源占用预估

| 场景 | 预计内存占用 |
|------|-------------|
| Extreme | 15-25MB |
| Embedded | 20-30MB |
| IoT | 30-40MB |
| Edge | 50-70MB |

#### 为什么布隆过滤器不能省

边缘设备的磁盘（eMMC/SD卡）随机读取性能很差。布隆过滤器用极小的内存代价（约 10bits/key），可以避免 95% 的无效磁盘寻道，是边缘场景的"保命符"。