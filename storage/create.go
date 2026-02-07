package storage

// 外接Store实例，只需传给KVDb即可。
var KVDb Store

// SetStore 设置外部存储实例
func SetStore(store Store) {
	KVDb = store
}

// OpenDefaultDb 打开存储数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDb(Path string) (Store, error) {
	var err error
	KVDb, err = NewLevelDBStore(Path, nil)
	if err != nil {
		return nil, err
	}
	return KVDb, nil
}

// OpenDefaultDbWithEncryption 打开带加密的默认数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDbWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	// 创建底层LevelDB存储
	underlyingStore, err := NewLevelDBStore(Path, nil)
	if err != nil {
		return nil, err
	}

	// 如果未启用加密，直接返回底层存储
	if !config.Enabled {
		KVDb = underlyingStore
		return KVDb, nil
	}

	// 创建加密存储包装器
	encryptedStore, err := NewEncryptedStoreWrapper(underlyingStore, config)
	if err != nil {
		underlyingStore.Close()
		return nil, err
	}

	KVDb = encryptedStore
	return KVDb, nil
}

// NewLevelDBStoreWithEncryption 创建带加密的LevelDB存储实例
func NewLevelDBStoreWithEncryption(Path string, config *EncryptionConfig) (Store, error) {
	// 创建底层LevelDB存储
	underlyingStore, err := NewLevelDBStore(Path, nil)
	if err != nil {
		return nil, err
	}

	// 如果未启用加密，直接返回底层存储
	if !config.Enabled {
		return underlyingStore, nil
	}

	// 创建加密存储包装器
	return NewEncryptedStoreWrapper(underlyingStore, config)
}
func CloseDb() error {
	if KVDb != nil {
		return KVDb.Close()
	}
	return nil
}

// StoreConfig 存储配置
type StoreConfig struct {
	Path   string // 存储路径
	DBType string // 数据库类型
	// 其他配置项
}

// NewStore 创建新的存储实例
func NewStore(config StoreConfig) (Store, error) {
	// 默认使用LevelDB实现
	//将来可以支持rocksdb，ToplingDB等其他实现
	switch config.DBType {
	case "rocksdb":
		// 使用RocksDB实现
		//return NewRocksDBStore(config)
		return nil, NewError("rocksdb not supported/未安装和配置 RocksDB 的 C++ 库")
	default:
		return NewLevelDBStore(config.Path, nil)
	}
}
