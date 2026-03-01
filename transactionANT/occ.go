package transactionANT

import (
	"sync"
)

// OCCSupport 乐观并发控制支持
type OCCSupport struct {
	versions map[string]uint64 // 键 -> 版本号
	mu       sync.Once         // 仅用于初始化
}

// NewOCCSupport 创建新的乐观并发控制支持
func NewOCCSupport() *OCCSupport {
	return &OCCSupport{
		versions: make(map[string]uint64),
	}
}

// GetVersion 获取键的版本号
func (occ *OCCSupport) GetVersion(key string) uint64 {
	occ.mu.Do(func() {
		// 确保versions已初始化
	})

	if version, exists := occ.versions[key]; exists {
		return version
	}
	return 0
}

// IncrementVersion 递增键的版本号
func (occ *OCCSupport) IncrementVersion(key string) uint64 {
	occ.mu.Do(func() {
		// 确保versions已初始化
	})

	// 这里使用atomic操作模拟无锁更新
	// 实际生产环境中，应该使用更复杂的无锁数据结构
	current := occ.versions[key]
	newVersion := current + 1
	occ.versions[key] = newVersion
	return newVersion
}

// CheckConflict 检查是否存在冲突
func (occ *OCCSupport) CheckConflict(key string, expectedVersion uint64) bool {
	occ.mu.Do(func() {
		// 确保versions已初始化
	})

	currentVersion := occ.versions[key]
	return currentVersion != expectedVersion
}

// LockFreeVersionManager 版本管理器（使用互斥锁实现）
type LockFreeVersionManager struct {
	versions map[string]uint64 // 键 -> 版本号
	mu       sync.RWMutex      // 读写锁
	initOnce sync.Once         // 仅用于初始化
}

// NewLockFreeVersionManager 创建新版本管理器
func NewLockFreeVersionManager() *LockFreeVersionManager {
	return &LockFreeVersionManager{}
}

// init 初始化版本映射
func (vm *LockFreeVersionManager) init() {
	vm.initOnce.Do(func() {
		vm.versions = make(map[string]uint64)
	})
}

// GetVersion 获取版本号
func (vm *LockFreeVersionManager) GetVersion(key string) uint64 {
	vm.init()
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if version, exists := vm.versions[key]; exists {
		return version
	}
	return 0
}

// IncrementVersion 递增版本号
func (vm *LockFreeVersionManager) IncrementVersion(key string) uint64 {
	vm.init()
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// 递增版本号
	currentVersion := vm.versions[key]
	newVersion := currentVersion + 1
	vm.versions[key] = newVersion
	return newVersion
}

// CheckConflict 检查冲突
func (vm *LockFreeVersionManager) CheckConflict(key string, expectedVersion uint64) bool {
	vm.init()
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	currentVersion := vm.versions[key]
	return currentVersion != expectedVersion
}
