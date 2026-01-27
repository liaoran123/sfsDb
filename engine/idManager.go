package engine

import (
	"fmt"
	"strings"
	"sync"

	"github.com/liaoran123/sfsDb/storage"
)

var TableIDManager *IDManager

// 常量定义
const (
	// IDMaxLimit ID的最大限制（1字节最多能创建255个ID）
	IDMaxLimit = 255
)

/*
IDManager 实现ID管理功能

功能描述：
1. 根据对象键（sys-table-name ,sys-tableid-idx-name，sys-tableid-field-name）读取已有ID
2. 如果对象键不存在，则从对应类型的自动增长计数器（sys-table,sys-tableid-idx，sys-tableid-field）获取新ID
3. 支持表、索引、字段等不同类型的ID管理
4. 保证线程安全
*/
type IDManager struct {
	kvStore     storage.Store // 用于持久化存储
	mutex       sync.Mutex    // 并发锁，保护并发操作
	ValueJoiner               // 嵌入值拼接器，用于生成系统ID键
}

// NewIDManager 创建一个新的ID管理器
func NewIDManager(kvStore storage.Store) *IDManager {
	return &IDManager{
		kvStore:     kvStore,
		ValueJoiner: ValueJoiner{},
	}
}

// UpdateName 更新对象的名称，保持ID不变
//
// 参数：
// - oldKey: 旧的对象键（如sys-table-oldName）
// - newKey: 新的对象键（如sys-table-newName）
//
// 返回：
// - error: 错误信息，如旧键不存在
//
// 功能流程：
// 1. 从旧键中获取ID
// 2. 删除旧键
// 3. 将ID存储到新键
// 4. 返回结果
func (m *IDManager) UpdateKey(oldKey string, newKey string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 1. 从旧键中获取ID
	idBytes, err := m.kvStore.Get([]byte(oldKey))
	if err != nil {
		return fmt.Errorf("旧键不存在: %s", oldKey)
	}

	// 2. 删除旧键
	if err := m.kvStore.Delete([]byte(oldKey)); err != nil {
		return fmt.Errorf("删除旧键失败: %v", err)
	}

	// 3. 将ID存储到新键
	if err := m.kvStore.Put([]byte(newKey), idBytes); err != nil {
		return fmt.Errorf("存储新键失败: %v", err)
	}

	return nil
}

// getNextIDUnsafe 内部方法：获取下一个ID，不进行锁保护
// 仅在已持有锁的情况下调用
func (m *IDManager) getNextIDUnsafe(key string) (uint8, error) {
	// 获取当前计数器值
	counterBytes, err := m.kvStore.Get([]byte(key))
	var counter uint8
	if err != nil {
		// 计数器不存在，初始化为0
		counter = 0
	} else {
		// 解析计数器值
		counter = uint8(counterBytes[0])
	}

	// 检查计数器是否已达上限
	if counter >= IDMaxLimit {
		return 0, fmt.Errorf("ID已达上限(%d)", IDMaxLimit)
	}

	// 生成新ID（使用当前计数器值）
	newID := counter

	// 递增计数器并存储
	m.kvStore.Put([]byte(key), []byte{counter + 1})

	return newID, nil
}

// incrementCounter 递增计数器并生成新ID
//
// 参数：
// - objTypeKey: 计数器类型键（如sys-table, sys-tableid-idx, sys-tableid-field）
// - originalKey: 原始对象键（如sys-table-name, sys-tableid-idx-name）
//
// 返回：
// - uint8: 生成的ID（0-255）
// - error: 错误信息，如ID已达上限
func (m *IDManager) incrementCounter(objTypeKey string, originalKey string) (uint8, error) {
	// 调用内部无锁方法生成新ID
	newID, err := m.getNextIDUnsafe(objTypeKey)
	if err != nil {
		return 0, err
	}

	// 将生成的ID存储到原始key
	m.kvStore.Put([]byte(originalKey), []byte{newID})

	return newID, nil
}

// GetNextID 通过key值获取下一个自动递增的ID
//
// 参数：
// - key: 用于自动递增的key（如sys-table, sys-index等）
//
// 返回：
// - uint8: 生成的ID（0-255）
// - error: 错误信息，如ID已达上限
//
// 功能流程：
// 1. 从key中获取当前计数器值
// 2. 检查ID是否已达上限
// 3. 递增并存储计数器
// 4. 返回生成的ID
func (m *IDManager) GetNextID(key string) (uint8, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 调用内部无锁方法生成新ID
	return m.getNextIDUnsafe(key)
}

// GetPreviousID 回退指定key的ID计数器，并返回回退前的ID
// 用于在错误时回退ID
//
// 参数：
// - key: 用于自动递增的key（如sys-table, sys-index等）
//
// 返回：
// - uint8: 回退前的ID值（0-255）
// - error: 错误信息，如ID已达下限（0）
func (m *IDManager) GetPreviousID(key string) (uint8, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取当前计数器值
	counterBytes, err := m.kvStore.Get([]byte(key))
	if err != nil {
		return 0, fmt.Errorf("计数器不存在，无法回退ID: %s", key)
	}

	// 检查值长度是否为1，因为我们存储的是uint8
	if len(counterBytes) != 1 {
		return 0, fmt.Errorf("计数器值无效，无法回退ID: %s", key)
	}

	// 解析计数器值
	counter := uint8(counterBytes[0])

	// 检查计数器是否已达下限
	if counter <= 0 {
		return 0, fmt.Errorf("ID已达下限(0)，无法回退")
	}

	// 回退前生成的ID（计数器值减1，因为计数器指向下一个要生成的ID）
	previousID := counter - 1

	// 存储回退后的计数器值，并处理可能的错误
	if err := m.kvStore.Put([]byte(key), []byte{previousID}); err != nil {
		return 0, fmt.Errorf("存储回退计数器失败: %v", err)
	}

	return previousID, nil
}

// GetCurrentID 获取指定key的当前ID值（未递增）
//
// 参数：
// - key: 用于自动递增的key（如sys-table, sys-index等）
//
// 返回：
// - uint8: 当前ID值（0-255）
// - error: 错误信息
func (m *IDManager) GetCurrentID(key string) (uint8, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取当前计数器值
	counterBytes, err := m.kvStore.Get([]byte(key))
	if err != nil {
		// 计数器不存在，返回0
		return 0, nil
	}

	// 检查值长度是否为1，因为我们存储的是uint8
	if len(counterBytes) != 1 {
		return 0, fmt.Errorf("计数器值无效: %s", key)
	}

	// 解析计数器值
	return uint8(counterBytes[0]), nil
}

// ResetID 重置指定key的ID计数器为0
//
// 参数：
// - key: 用于自动递增的key（如sys-table, sys-index等）
//
// 返回：
// - error: 错误信息
func (m *IDManager) ResetID(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 重置计数器为0，并处理可能的错误
	if err := m.kvStore.Put([]byte(key), []byte{0}); err != nil {
		return fmt.Errorf("重置计数器失败: %v", err)
	}
	return nil
}

// SetID 设置指定key的ID计数器值
//
// 参数：
// - key: 用于自动递增的key（如sys-table, sys-index等）
// - id: 要设置的ID值（0-254，因为下一次调用GetNextID会返回该值并递增）
//
// 返回：
// - error: 错误信息，如ID值超出范围
func (m *IDManager) SetID(key string, id uint8) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查ID值是否超出范围
	if id >= IDMaxLimit {
		return fmt.Errorf("ID值超出范围，最大允许值为%d", IDMaxLimit-1)
	}

	// 设置计数器值，并处理可能的错误
	if err := m.kvStore.Put([]byte(key), []byte{id}); err != nil {
		return fmt.Errorf("设置计数器失败: %v", err)
	}
	return nil
}

/*
GetOrCreateID 获取或创建基于名称和类型的ID

参数：
- name: 对象名称（如表名、索引名、字段名）
- objType: 对象类型前缀（table-表，idx-索引，field-字段）

返回：
- uint8: 生成或获取的ID（0-255）
- bool: 是否是新创建的ID
- error: 错误信息，如ID已达上限

功能流程：
1. 尝试从对象key获取已存在的ID
2. 如果ID存在，直接返回
3. 如果ID不存在，从类型计数器获取当前值
4. 检查ID是否已达上限
5. 递增并存储计数器
6. 将生成的ID存储到对象key
7. 返回生成的ID
bool: 是否是新创建的ID，true表示是，false表示不是
*/
func (m *IDManager) GetOrCreateID(key string) (uint8, bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 1. 生成对象key并尝试获取已存在的ID
	if id, err := m.kvStore.Get([]byte(key)); err == nil {
		// ID已存在，直接返回
		return uint8(id[0]), false, nil
	}

	// 2. ID不存在，生成新ID
	/*
		根据对象键（sys-table-name ,sys-tableid-idx-name，sys-tableid-field-name）读取已有ID
		如果对象键不存在，则从对应类型的自动增长计数器（sys-table,sys-tableid-idx，sys-tableid-field）获取新ID
	*/
	//根据key=sys-table-name ,sys-tableid-idx-name，sys-tableid-field-name 截取对象类型
	ks := strings.Split(key, "-")
	objTypeKeySlice := ks[:len(ks)-1]
	//objTypeKey=sys-table,sys-tableid-idx,sys-tableid-field

	// 将objTypeKey切片转换为字符串键
	objTypeKey := strings.Join(objTypeKeySlice, "-")

	// 调用递增计数器方法生成新ID
	newID, err := m.incrementCounter(objTypeKey, key)
	if err != nil {
		return 0, false, err
	}
	return newID, true, nil
}
