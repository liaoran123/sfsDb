package storage

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

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

// loadConfigFromStore 从存储中加载配置到 opts
func loadConfigFromStore(path string, opts *opt.Options) error {
	// 首先加载配置到全局配置
	if err := LoadConfigFromStore(path); err != nil {
		return err
	}

	// 然后将全局配置应用到传入的opts
	config := GetConfig()
	opts.WriteBuffer = config.WriteBuffer
	opts.OpenFilesCacheCapacity = config.OpenFilesCacheCapacity
	opts.BlockCacheCapacity = config.BlockCacheCapacity
	opts.Compression = config.Compression

	return nil
}

// createLevelDBStore 创建底层LevelDB存储实例（内部函数）
func createLevelDBStore(Path string, opts *opt.Options) (Store, error) {
	ldb, openErr := leveldb.OpenFile(Path, opts)
	if openErr != nil {
		// 尝试修复损坏的数据库
		ldb, recoverErr := leveldb.RecoverFile(Path, opts)
		if recoverErr != nil {
			// 修复失败，返回更详细的错误信息
			return nil, NewError(fmt.Sprintf("数据库打开失败且修复失败: 打开错误: %v, 修复错误: %v", openErr, recoverErr))
		}
		// 修复成功，直接使用恢复后的数据库实例
		return &LevelDBStore{
			ldb:        ldb,
			originalDB: ldb,
			isSnapshot: false,
			opts:       opts,
		}, nil
	}
	return &LevelDBStore{
		ldb:        ldb,
		originalDB: ldb,
		isSnapshot: false,
		opts:       opts,
	}, nil
}

// NewLevelDBStore 创建新的LevelDB存储实例
func (dm *DBManager) NewLevelDBStore(Path string, opts *opt.Options, encryptConfig ...*EncryptionConfig) (Store, error) {
	if opts == nil {
		// 创建默认配置
		opts = &opt.Options{
			// 设置默认选项
			WriteBuffer:            64 * 1024 * 1024,  // 64MB write buffer
			OpenFilesCacheCapacity: 200,               // 打开文件缓存，增加以提高并发读取性能
			BlockCacheCapacity:     128 * 1024 * 1024, // 128MB block cache，增加以提高读取性能
		}

		// 尝试从存储中读取配置
		if err := loadConfigFromStore(Path, opts); err != nil {
			// 配置加载失败，使用默认配置继续
			// 这里不返回错误，因为配置加载失败不应该阻止数据库打开
		}
	}

	// 如果没有提供加密配置，直接返回普通存储
	if len(encryptConfig) == 0 || encryptConfig[0] == nil {
		return createLevelDBStore(Path, opts)
	}

	config := encryptConfig[0]

	// 创建底层LevelDB存储
	underlyingStore, err := createLevelDBStore(Path, opts)
	if err != nil {
		return nil, err
	}

	// 如果未启用加密，直接返回底层存储
	if !config.Enabled {
		return underlyingStore, nil
	}

	// 创建加密存储包装器
	encryptedStore, err := NewEncryptedStoreWrapper(underlyingStore, config)
	if err != nil {
		underlyingStore.Close()
		return nil, err
	}

	return encryptedStore, nil
}

// NewLevelDBStoreWithScenario 根据场景创建新的LevelDB存储实例
/*
- 创建 独立 的 Store 实例
- 不设置 到 dm.db
- 用于需要多个独立存储实例的场景
*/
func (dm *DBManager) NewLevelDBStoreWithScenario(Path string, scenario string, encryptConfig ...*EncryptionConfig) (Store, error) {
	opts := GetScenarioOptions(scenario)
	return dm.NewLevelDBStore(Path, opts, encryptConfig...)
}

// OpenDB 打开存储数据库，使用管理器内部的db变量，保证全局唯一实例
func (dm *DBManager) OpenDB(path string, encryptConfig ...*EncryptionConfig) (Store, error) {
	var err error
	dm.db, err = dm.NewLevelDBStore(path, nil, encryptConfig...)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return dm.db, nil
}

// OpenDBWithScenario 根据场景打开存储数据库，使用管理器内部的db变量，保证全局唯一实例
func (dm *DBManager) OpenDBWithScenario(path string, scenario string, encryptConfig ...*EncryptionConfig) (Store, error) {
	var err error
	opts := GetScenarioOptions(scenario)
	dm.db, err = dm.NewLevelDBStore(path, opts, encryptConfig...)
	if err != nil {
		return nil, fmt.Errorf("failed to open database with scenario %s: %v", scenario, err)
	}
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

// CloseDb 关闭数据库
func CloseDb() error {
	return dbManager.CloseDB()
}

/*
// ----------- 向后兼容层 -----------
//--------逐渐废弃------------------------

// 外接Store实例，只需传给KVDb即可。
var KVDb Store

// SetStore 设置外部存储实例
func SetStore(store Store) {
	KVDb = store
	dbManager.SetDB(store)
}

// NewLevelDBStore 创建新的LevelDB存储实例（向后兼容）
func NewLevelDBStore(Path string, opts *opt.Options, encryptConfig ...*EncryptionConfig) (Store, error) {
	return dbManager.NewLevelDBStore(Path, opts, encryptConfig...)
}

// NewLevelDBStoreWithScenario 根据场景创建新的LevelDB存储实例（向后兼容）
func NewLevelDBStoreWithScenario(Path string, scenario string, encryptConfig ...*EncryptionConfig) (Store, error) {
	return dbManager.NewLevelDBStoreWithScenario(Path, scenario, encryptConfig...)
}

// OpenDefaultDb 打开存储数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDb(Path string, encryptConfig ...*EncryptionConfig) (Store, error) {
	store, err := dbManager.OpenDB(Path, encryptConfig...)
	if err != nil {
		return nil, err
	}
	KVDb = store
	return store, nil
}

// OpenDefaultDbWithScenario 根据场景打开存储数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDbWithScenario(Path string, scenario string, encryptConfig ...*EncryptionConfig) (Store, error) {
	store, err := dbManager.OpenDBWithScenario(Path, scenario, encryptConfig...)
	if err != nil {
		return nil, err
	}
	KVDb = store
	return store, nil
}
*/
