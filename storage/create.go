package storage

//添加一个全局锁

var KVDb Store

// StoreFactory 存储工厂函数类型
type StoreFactory func(config StoreConfig) (Store, error)

// storeFactories 存储工厂注册表
var storeFactories = make(map[string]StoreFactory)

// RegisterStoreFactory 注册存储工厂
func RegisterStoreFactory(dbType string, factory StoreFactory) {
	storeFactories[dbType] = factory
}

// StoreConfig 存储配置
type StoreConfig struct {
	Path             string            // 存储路径
	DBType           string            // 数据库类型
	EncryptionConfig *EncryptionConfig // 加密配置
	// 其他配置项
}

// OpenStore 打开存储数据库，支持不同存储类型和加密配置
func OpenStore(config StoreConfig) (Store, error) {
	// 创建底层存储
	underlyingStore, err := createUnderlyingStore(config)
	if err != nil {
		return nil, err
	}

	// 应用加密包装器
	return applyEncryption(underlyingStore, config.EncryptionConfig)
}

// init 初始化存储工厂注册表
func init() {
	// 注册默认的LevelDB存储工厂
	RegisterStoreFactory("leveldb", func(config StoreConfig) (Store, error) {
		return NewLevelDBStore(config.Path, nil)
	})

	// 注册rocksdb存储工厂（占位）
	RegisterStoreFactory("rocksdb", func(config StoreConfig) (Store, error) {
		return nil, NewError("rocksdb not supported/未安装和配置 RocksDB 的 C++ 库")
	})
}

// OpenDefaultDb 打开默认数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDb(Path string) (Store, error) {
	config := StoreConfig{
		Path:   Path,
		DBType: "leveldb",
	}
	var err error
	KVDb, err = OpenStore(config)
	if err != nil {
		return nil, err
	}
	return KVDb, nil
}

// OpenDefaultDbWithEncryption 打开带加密的默认数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDbWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	storeConfig := StoreConfig{
		Path:             Path,
		DBType:           "leveldb",
		EncryptionConfig: config,
	}
	var err error
	KVDb, err = OpenStore(storeConfig)
	if err != nil {
		return nil, err
	}
	return KVDb, nil
}

// NewLevelDBStoreWithEncryption 创建带加密的LevelDB存储实例
func NewLevelDBStoreWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	storeConfig := StoreConfig{
		Path:             Path,
		DBType:           "leveldb",
		EncryptionConfig: config,
	}
	return OpenStore(storeConfig)
}

// createUnderlyingStore 创建底层存储实例
func createUnderlyingStore(config StoreConfig) (Store, error) {
	// 从注册表中获取存储工厂
	factory, exists := storeFactories[config.DBType]
	if !exists {
		// 如果未找到指定类型的工厂，尝试使用默认的leveldb
		factory, exists = storeFactories["leveldb"]
		if !exists {
			return nil, NewError("no store factory found for " + config.DBType)
		}
	}

	// 使用工厂创建存储实例
	return factory(config)
}

// applyEncryption 应用加密包装器
func applyEncryption(underlyingStore Store, config *EncryptionConfig) (Store, error) {
	// 如果未启用加密或配置为nil，直接返回底层存储
	if config == nil || !config.Enabled {
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

func CloseDb() error {
	if KVDb != nil {
		return KVDb.Close()
	}
	return nil
}

// NewStore 创建新的存储实例（保持向后兼容）
func NewStore(config StoreConfig) (Store, error) {
	return OpenStore(config)
}

// 备份数据库
func BackupDb(Path string) error {
	//保存当前的源数据库引用
	sourceDb := KVDb
	if sourceDb == nil {
		return NewError("源数据库未打开")
	}

	//打开备份目标数据库，不修改全局KVDb
	backupDb, err := NewLevelDBStore(Path, nil)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	//创建源数据库的全库遍历迭代器，使用Iterator方法传入nil作为start和limit
	iter := sourceDb.Iterator(nil, nil)
	defer iter.Release()

	//遍历所有记录
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		//将记录写入备份数据库
		if err := backupDb.Put(key, value); err != nil {
			return err
		}
	}
	return nil
}
