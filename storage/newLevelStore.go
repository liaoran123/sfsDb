package storage

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

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

// NewLevelDBStore 创建新的LevelDB存储实例
func NewLevelDBStore(Path string, opts *opt.Options) (Store, error) {
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
