package lock

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// LockType 锁类型
type LockType int

const (
	// ReadLock 读锁
	ReadLock LockType = iota
	// WriteLock 写锁
	WriteLock
)

// LockMode 锁模式
type LockMode int

const (
	// Blocking 阻塞模式
	Blocking LockMode = iota
	// NonBlocking 非阻塞模式
	NonBlocking
)

// LockRequest 锁请求
type LockRequest struct {
	Key       string        // 锁键
	Type      LockType      // 锁类型
	Mode      LockMode      // 锁模式
	Timeout   time.Duration // 超时时间
	Requestor string        // 请求者标识
}

// LockStatus 锁状态
type LockStatus struct {
	Key        string    // 锁键
	Holders    []string  // 持有者
	LockType   LockType  // 当前锁类型
	ReadCount  int       // 读锁数量
	WriteCount int       // 写锁数量
	CreatedAt  time.Time // 创建时间
	LastAccess time.Time // 最后访问时间
}

// Manager 锁管理器
type Manager struct {
	// 锁映射
	locks   map[string]*LockStatus // 锁状态映射
	locksMu sync.RWMutex           // 锁映射的读写锁

	// 配置参数
	defaultTimeout time.Duration // 默认超时时间
	maxRetries     int           // 最大重试次数
	retryInterval  time.Duration // 重试间隔

	// 统计信息
	lockAcquires  int64 // 锁获取次数
	lockReleases  int64 // 锁释放次数
	lockConflicts int64 // 锁冲突次数
	lockTimeouts  int64 // 锁超时次数

	// 并发控制
	mu sync.RWMutex // 管理器的读写锁
}

// NewManager 创建新的锁管理器
func NewManager() *Manager {
	return &Manager{
		locks:          make(map[string]*LockStatus),
		defaultTimeout: 5 * time.Second,
		maxRetries:     5,
		retryInterval:  10 * time.Millisecond,
	}
}

// AcquireLock 获取锁
func (m *Manager) AcquireLock(req LockRequest) error {
	// 设置默认值
	if req.Timeout == 0 {
		req.Timeout = m.defaultTimeout
	}

	// 锁排序，防止死锁
	sortedKeys := []string{req.Key}
	sort.Strings(sortedKeys)

	// 尝试获取锁
	startTime := time.Now()
	retries := 0

	for time.Since(startTime) < req.Timeout {
		success, err := m.tryAcquireLock(req)
		if err != nil {
			return err
		}

		if success {
			atomic.AddInt64(&m.lockAcquires, 1)
			return nil
		}

		// 非阻塞模式直接返回
		if req.Mode == NonBlocking {
			atomic.AddInt64(&m.lockConflicts, 1)
			return fmt.Errorf("锁获取失败：锁被其他事务持有")
		}

		// 阻塞模式重试
		retries++
		if retries > m.maxRetries {
			break
		}

		// 指数退避
		sleepTime := m.retryInterval * time.Duration(1<<uint(retries-1))
		time.Sleep(sleepTime)
	}

	atomic.AddInt64(&m.lockTimeouts, 1)

	return fmt.Errorf("锁获取超时：%s", req.Key)
}

// tryAcquireLock 尝试获取锁
func (m *Manager) tryAcquireLock(req LockRequest) (bool, error) {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	lock, exists := m.locks[req.Key]
	if !exists {
		// 锁不存在，创建新锁
		m.locks[req.Key] = &LockStatus{
			Key:        req.Key,
			Holders:    []string{req.Requestor},
			LockType:   req.Type,
			ReadCount:  0,
			WriteCount: 0,
			CreatedAt:  time.Now(),
			LastAccess: time.Now(),
		}

		if req.Type == ReadLock {
			m.locks[req.Key].ReadCount = 1
		} else {
			m.locks[req.Key].WriteCount = 1
		}

		return true, nil
	}

	// 锁已存在，检查是否可以获取
	if req.Type == ReadLock {
		// 读锁可以与其他读锁共存
		if lock.LockType == ReadLock {
			// 检查请求者是否已持有锁
			for _, holder := range lock.Holders {
				if holder == req.Requestor {
					// 已持有锁，增加计数
					lock.ReadCount++
					lock.LastAccess = time.Now()
					return true, nil
				}
			}

			// 新的读锁请求者
			lock.Holders = append(lock.Holders, req.Requestor)
			lock.ReadCount++
			lock.LastAccess = time.Now()
			return true, nil
		}
	} else {
		// 写锁需要独占
		if lock.LockType == ReadLock && lock.ReadCount == 0 {
			// 读锁已释放，获取写锁
			lock.Holders = []string{req.Requestor}
			lock.LockType = WriteLock
			lock.WriteCount = 1
			lock.LastAccess = time.Now()
			return true, nil
		} else if lock.LockType == WriteLock {
			// 检查请求者是否已持有锁
			for _, holder := range lock.Holders {
				if holder == req.Requestor {
					// 已持有锁，增加计数
					lock.WriteCount++
					lock.LastAccess = time.Now()
					return true, nil
				}
			}
		}
	}

	// 锁冲突
	return false, nil
}

// ReleaseLock 释放锁
func (m *Manager) ReleaseLock(key, requestor string) error {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	lock, exists := m.locks[key]
	if !exists {
		return fmt.Errorf("锁不存在：%s", key)
	}

	// 检查请求者是否持有锁
	holderIndex := -1
	for i, holder := range lock.Holders {
		if holder == requestor {
			holderIndex = i
			break
		}
	}

	if holderIndex == -1 {
		return fmt.Errorf("请求者未持有锁：%s", key)
	}

	// 减少锁计数
	if lock.LockType == ReadLock {
		lock.ReadCount--
	} else {
		lock.WriteCount--
	}

	// 移除持有者
	lock.Holders = append(lock.Holders[:holderIndex], lock.Holders[holderIndex+1:]...)

	// 检查锁是否可以释放
	if lock.ReadCount == 0 && lock.WriteCount == 0 {
		delete(m.locks, key)
	} else {
		lock.LastAccess = time.Now()
	}

	atomic.AddInt64(&m.lockReleases, 1)

	return nil
}

// ReleaseAllLocks 释放请求者的所有锁
func (m *Manager) ReleaseAllLocks(requestor string) error {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	for key, lock := range m.locks {
		// 检查请求者是否持有锁
		hasHolder := false
		for _, holder := range lock.Holders {
			if holder == requestor {
				hasHolder = true
				break
			}
		}

		if hasHolder {
			// 移除持有者
			newHolders := []string{}
			for _, holder := range lock.Holders {
				if holder != requestor {
					newHolders = append(newHolders, holder)
				}
			}
			lock.Holders = newHolders

			// 重置计数
			if lock.LockType == ReadLock {
				lock.ReadCount = len(newHolders)
			} else {
				lock.WriteCount = len(newHolders)
			}

			// 检查锁是否可以释放
			if len(newHolders) == 0 {
				delete(m.locks, key)
			} else {
				lock.LastAccess = time.Now()
			}
		}
	}

	atomic.AddInt64(&m.lockReleases, 1)

	return nil
}

// GetLockStatus 获取锁状态
func (m *Manager) GetLockStatus(key string) (*LockStatus, error) {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	lock, exists := m.locks[key]
	if !exists {
		return nil, fmt.Errorf("锁不存在：%s", key)
	}

	// 返回锁状态的副本
	status := *lock
	status.Holders = make([]string, len(lock.Holders))
	copy(status.Holders, lock.Holders)

	return &status, nil
}

// GetAllLocks 获取所有锁
func (m *Manager) GetAllLocks() map[string]*LockStatus {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	locks := make(map[string]*LockStatus)
	for key, lock := range m.locks {
		status := *lock
		status.Holders = make([]string, len(lock.Holders))
		copy(status.Holders, lock.Holders)
		locks[key] = &status
	}

	return locks
}

// GetStats 获取锁管理器统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.locksMu.RLock()
	lockCount := len(m.locks)
	m.locksMu.RUnlock()

	return map[string]interface{}{
		"lockCount":      lockCount,
		"lockAcquires":   atomic.LoadInt64(&m.lockAcquires),
		"lockReleases":   atomic.LoadInt64(&m.lockReleases),
		"lockConflicts":  atomic.LoadInt64(&m.lockConflicts),
		"lockTimeouts":   atomic.LoadInt64(&m.lockTimeouts),
		"defaultTimeout": m.defaultTimeout,
		"maxRetries":     m.maxRetries,
		"retryInterval":  m.retryInterval,
	}
}

// SetDefaultTimeout 设置默认超时时间
func (m *Manager) SetDefaultTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultTimeout = timeout
}

// SetRetryPolicy 设置重试策略
func (m *Manager) SetRetryPolicy(maxRetries int, retryInterval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxRetries = maxRetries
	m.retryInterval = retryInterval
}

// ClearExpiredLocks 清理过期锁
func (m *Manager) ClearExpiredLocks(expiration time.Duration) int {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	count := 0
	now := time.Now()

	for key, lock := range m.locks {
		if now.Sub(lock.LastAccess) > expiration {
			delete(m.locks, key)
			count++
		}
	}

	return count
}

// SortKeys 对锁键进行排序，防止死锁
func (m *Manager) SortKeys(keys []string) []string {
	sorted := make([]string, len(keys))
	copy(sorted, keys)
	sort.Strings(sorted)
	return sorted
}

// IsLocked 检查键是否被锁定
func (m *Manager) IsLocked(key string) bool {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	_, exists := m.locks[key]
	return exists
}

// GetLockHolders 获取锁的持有者
func (m *Manager) GetLockHolders(key string) []string {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	lock, exists := m.locks[key]
	if !exists {
		return []string{}
	}

	holders := make([]string, len(lock.Holders))
	copy(holders, lock.Holders)
	return holders
}
