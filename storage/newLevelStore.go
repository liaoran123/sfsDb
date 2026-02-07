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

	// 检查是否包含单位（支持大小写）
	sLower := strings.ToLower(s)
	var multiplier int
	var valueStr string

	switch {
	case strings.HasSuffix(sLower, "mb"):
		multiplier = 1024 * 1024
		valueStr = strings.TrimSuffix(s, "MB")
		if valueStr == s {
			valueStr = strings.TrimSuffix(s, "mb")
		}
	case strings.HasSuffix(sLower, "kb"):
		multiplier = 1024
		valueStr = strings.TrimSuffix(s, "KB")
		if valueStr == s {
			valueStr = strings.TrimSuffix(s, "kb")
		}
	case strings.HasSuffix(sLower, "gb"):
		multiplier = 1024 * 1024 * 1024
		valueStr = strings.TrimSuffix(s, "GB")
		if valueStr == s {
			valueStr = strings.TrimSuffix(s, "gb")
		}
	default:
		// 尝试直接解析为整数
		size, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid size format: %s, expected number or number with unit (KB, MB, GB)", s)
		}
		return size, nil
	}

	// 解析数值部分
	size, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size value: %s, %v", valueStr, err)
	}

	// 计算最终大小
	return size * multiplier, nil
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
		return fmt.Errorf("failed to open temp DB for config loading: %v", err)
	}
	defer tempDB.Close()

	// 定义配置项映射
	configItems := []struct {
		key    string
		parser func(string) (interface{}, error)
		apply  func(interface{}) error
		desc   string
	}{
		{
			key: "config:write_buffer",
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					opts.WriteBuffer = size
				}
				return nil
			},
			desc: "write buffer size",
		},
		{
			key: "config:max_open_files",
			parser: func(s string) (interface{}, error) {
				return strconv.Atoi(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					opts.OpenFilesCacheCapacity = size
				}
				return nil
			},
			desc: "max open files",
		},
		{
			key: "config:block_cache",
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					opts.BlockCacheCapacity = size
				}
				return nil
			},
			desc: "block cache capacity",
		},
		{
			key: "config:compression",
			parser: func(s string) (interface{}, error) {
				return strconv.ParseBool(s)
			},
			apply: func(val interface{}) error {
				if enabled, ok := val.(bool); ok {
					if enabled {
						opts.Compression = opt.DefaultCompression
					} else {
						opts.Compression = opt.NoCompression
					}
				}
				return nil
			},
			desc: "compression enabled",
		},
	}

	// 加载所有配置项
	configLoaded := false
	for _, item := range configItems {
		value, err := tempDB.Get([]byte(item.key), nil)
		if err != nil {
			// 配置项不存在，跳过
			continue
		}

		parsedVal, err := item.parser(string(value))
		if err != nil {
			// 解析失败，跳过该配置项
			continue
		}

		if err := item.apply(parsedVal); err != nil {
			// 应用失败，跳过
			continue
		}

		configLoaded = true
	}

	if !configLoaded {
		// 没有加载到任何配置，使用默认配置
		// 这里不返回错误，因为配置加载失败不应该阻止数据库打开
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
