package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// parseSize 解析大小字符串，支持 "64MB" 或 "67108864" 格式
func parseSize(s string) (int, error) {
	s = strings.TrimSpace(s)

	// 检查是否包含单位
	if strings.HasSuffix(s, "MB") {
		val := strings.TrimSuffix(s, "MB")
		if size, err := strconv.Atoi(val); err == nil {
			return size * 1024 * 1024, nil
		}
	}
	if strings.HasSuffix(s, "KB") {
		val := strings.TrimSuffix(s, "KB")
		if size, err := strconv.Atoi(val); err == nil {
			return size * 1024, nil
		}
	}
	if strings.HasSuffix(s, "GB") {
		val := strings.TrimSuffix(s, "GB")
		if size, err := strconv.Atoi(val); err == nil {
			return size * 1024 * 1024 * 1024, nil
		}
	}

	// 尝试直接解析为整数
	if size, err := strconv.Atoi(s); err == nil {
		return size, nil
	}

	return 0, fmt.Errorf("invalid size format: %s", s)
}

// loadConfigFromStore 从存储中加载配置到 opts
func loadConfigFromStore(path string, opts *opt.Options) error {
	// 创建临时配置用于打开存储
	tempOpts := &opt.Options{
		// 使用最小配置打开存储，只用于读取配置
		WriteBuffer:            4 * 1024 * 1024, // 4MB write buffer，最小配置
		OpenFilesCacheCapacity: 10,              // 最小打开文件缓存
		BlockCacheCapacity:     8 * 1024 * 1024, // 8MB block cache，最小配置
	}

	// 尝试打开临时存储来读取配置
	tempDB, err := leveldb.OpenFile(path, tempOpts)
	if err != nil {
		return err
	}
	defer tempDB.Close()

	// 定义配置项映射
	configItems := []struct {
		key    string
		parser func(string) (interface{}, error)
		apply  func(interface{}) error
	}{
		{
			key: "config:write_buffer",
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok {
					opts.WriteBuffer = size
				}
				return nil
			},
		},
		{
			key: "config:max_open_files",
			parser: func(s string) (interface{}, error) {
				return strconv.Atoi(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok {
					opts.OpenFilesCacheCapacity = size
				}
				return nil
			},
		},
		{
			key: "config:block_cache",
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok {
					opts.BlockCacheCapacity = size
				}
				return nil
			},
		},
	}

	// 加载所有配置项
	for _, item := range configItems {
		if value, err := tempDB.Get([]byte(item.key), nil); err == nil {
			if parsedVal, err := item.parser(string(value)); err == nil {
				item.apply(parsedVal)
			}
		}
	}

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
		if err := loadConfigFromStore(Path, opts); err == nil {
			// 配置加载成功，opts 已经被更新
		}
	}
	ldb, err := leveldb.OpenFile(Path, opts)
	if err != nil {
		// 尝试修复损坏的数据库
		ldb, err = leveldb.RecoverFile(Path, opts)
		if err != nil {
			// 修复失败，返回更详细的错误信息
			return nil, NewError(fmt.Sprintf("数据库打开失败且修复失败: 打开错误: %v, 修复错误: %v", err, err))
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
