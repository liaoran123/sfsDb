package status

import (
	"runtime"

	"github.com/liaoran123/sfsDb/storage"
)

// MemoryInfo 内存使用信息
type MemoryInfo struct {
	Alloc      uint64 // 当前分配的内存大小（字节）
	TotalAlloc uint64 // 累计分配的内存大小（字节）
	Sys        uint64 // 从系统获取的内存大小（字节）
	NumGC      uint32 // GC 次数
}

// StorageInfo 存储使用信息
type StorageInfo struct {
	StoreType string // 存储类型
	// 其他存储相关信息可以根据具体存储实现扩展
}

// StatusInfo 数据库状态信息
type StatusInfo struct {
	Memory MemoryInfo   // 内存使用信息
	Storage StorageInfo // 存储使用信息
}

// StatusManager 状态管理器
type StatusManager struct {
	store storage.Store
}

// NewStatusManager 创建状态管理器
// 参数:
//   store: 存储实例
// 返回:
//   *StatusManager: 状态管理器实例

func NewStatusManager(store storage.Store) *StatusManager {
	return &StatusManager{
		store: store,
	}
}

// GetStatus 获取数据库状态
// 返回:
//   StatusInfo: 数据库状态信息
//   error: 错误信息

func (sm *StatusManager) GetStatus() (StatusInfo, error) {
	// 获取内存状态
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryInfo := MemoryInfo{
		Alloc:      memStats.Alloc,
		TotalAlloc: memStats.TotalAlloc,
		Sys:        memStats.Sys,
		NumGC:      memStats.NumGC,
	}

	// 获取存储状态
	storageInfo := StorageInfo{
		StoreType: getStoreType(sm.store),
	}

	return StatusInfo{
		Memory: memoryInfo,
		Storage: storageInfo,
	}, nil
}

// getStoreType 获取存储类型
// 参数:
//   store: 存储实例
// 返回:
//   string: 存储类型名称

func getStoreType(store storage.Store) string {
	storeType := "Unknown"
	
	// 根据存储实例的类型返回对应的类型名称
	switch store.(type) {
	case *storage.LevelDBStore:
		storeType = "LevelDB"
	case *storage.EncryptedStoreWrapper:
		storeType = "EncryptedLevelDB"
	// 可以根据需要添加其他存储类型
	}
	
	return storeType
}
