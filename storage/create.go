package storage

import "fmt"

// DBManager 管理数据库实例的创建、打开和关闭
type DBManager struct {
	db Store
}

// 全局数据库管理器实例
var dbManager = &DBManager{}

// GetDBManager 获取数据库管理器实例
func GetDBManager() *DBManager {
	return dbManager
}

// GetDB 获取当前数据库实例
func (dm *DBManager) GetDB() Store {
	return dm.db
}

// SetDB 设置数据库实例
func (dm *DBManager) SetDB(store Store) {
	dm.db = store
}

// OpenDB 打开存储数据库，使用管理器内部的db变量，保证全局唯一实例
func (dm *DBManager) OpenDB(path string) (Store, error) {
	var err error
	dm.db, err = NewLevelDBStore(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return dm.db, nil
}

// OpenDBWithEncryption 打开带加密的数据库，使用管理器内部的db变量，保证全局唯一实例
func (dm *DBManager) OpenDBWithEncryption(path string, config *EncryptionConfig) (Store, error) {
	// 创建底层LevelDB存储
	underlyingStore, err := NewLevelDBStore(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create underlying store: %v", err)
	}

	// 如果未启用加密，直接返回底层存储
	if !config.Enabled {
		dm.db = underlyingStore
		return dm.db, nil
	}

	// 创建加密存储包装器
	encryptedStore, err := NewEncryptedStoreWrapper(underlyingStore, config)
	if err != nil {
		underlyingStore.Close()
		return nil, fmt.Errorf("failed to create encrypted store: %v", err)
	}

	dm.db = encryptedStore
	return dm.db, nil
}

// CloseDB 关闭数据库
func (dm *DBManager) CloseDB() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}

// StoreConfig 存储配置
type StoreConfig struct {
	Path   string // 存储路径
	DBType string // 数据库类型
	// 其他配置项
}

// NewStoreWithEncryption 创建带加密的存储实例
func (dm *DBManager) NewStoreWithEncryption(path string, config *EncryptionConfig) (Store, error) {
	// 创建底层LevelDB存储
	underlyingStore, err := NewLevelDBStore(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create underlying store: %v", err)
	}

	// 如果未启用加密，直接返回底层存储
	if !config.Enabled {
		return underlyingStore, nil
	}

	// 创建加密存储包装器
	return NewEncryptedStoreWrapper(underlyingStore, config)
}

// CloseDb 关闭数据库
func CloseDb() error {
	return dbManager.CloseDB()
}

// ----------- 向后兼容层 -----------
//--------逐渐废弃------------------------

// 外接Store实例，只需传给KVDb即可。
var KVDb Store

// SetStore 设置外部存储实例
func SetStore(store Store) {
	KVDb = store
	dbManager.SetDB(store)
}

// OpenDefaultDb 打开存储数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDb(Path string) (Store, error) {
	store, err := dbManager.OpenDB(Path)
	if err != nil {
		return nil, err
	}
	KVDb = store
	return store, nil
}

// OpenDefaultDbWithEncryption 打开带加密的默认数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDbWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	store, err := dbManager.OpenDBWithEncryption(Path, config)
	if err != nil {
		return nil, err
	}
	KVDb = store
	return store, nil
}

// NewLevelDBStoreWithEncryption 创建带加密的LevelDB存储实例
func NewLevelDBStoreWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	return dbManager.NewStoreWithEncryption(Path, config)
}
