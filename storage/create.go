package storage

//添加一个全局锁

var KVDb Store

/*
func init() {
	config := config.Cfg
	KVDb, _ = OpenDefaultDb(config.DbPath)
}
*/
// OpenStoreDb 打开存储数据库，使用公共KVDb变量，保证全局唯一实例
func OpenDefaultDb(Path string) (Store, error) {
	var err error
	KVDb, err = NewLevelDBStore(Path, nil)
	if err != nil {
		return nil, err
	}
	return KVDb, nil
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
