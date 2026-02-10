package storage

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// 常量定义
const (
	// 默认配置值
	DefaultWriteBuffer            = 64 * 1024 * 1024  // 64MB
	DefaultOpenFilesCacheCapacity = 200               // 200个文件
	DefaultBlockCacheCapacity     = 128 * 1024 * 1024 // 128MB

	// 最小配置值（用于临时存储）
	MinWriteBuffer            = 4 * 1024 * 1024 // 4MB
	MinOpenFilesCacheCapacity = 10              // 10个文件
	MinBlockCacheCapacity     = 8 * 1024 * 1024 // 8MB

	// 配置键前缀
	ConfigPrefix = "config:"
)

// 配置键常量
const (
	ConfigKeyWriteBuffer  = ConfigPrefix + "write_buffer"
	ConfigKeyMaxOpenFiles = ConfigPrefix + "max_open_files"
	ConfigKeyBlockCache   = ConfigPrefix + "block_cache"
	ConfigKeyCompression  = ConfigPrefix + "compression"
)

// 场景常量
const (
	ScenarioEmbedded = "embedded"
	ScenarioIoT      = "iot"
	ScenarioEdge     = "edge"
	ScenarioGame     = "game"
	ScenarioDefault  = "default"
)

// Config 存储数据库配置参数
type Config struct {
	WriteBuffer            int
	OpenFilesCacheCapacity int
	BlockCacheCapacity     int
	Compression            opt.Compression
}

// ConfigManager 管理配置的加载、保存和访问
type ConfigManager struct {
	config Config
	mutex  sync.RWMutex
}

// 全局配置管理器实例
var configManager = &ConfigManager{
	config: Config{
		WriteBuffer:            DefaultWriteBuffer,
		OpenFilesCacheCapacity: DefaultOpenFilesCacheCapacity,
		BlockCacheCapacity:     DefaultBlockCacheCapacity,
		Compression:            opt.DefaultCompression,
	},
}

// GetConfigManager 获取全局配置管理器实例
func GetConfigManager() *ConfigManager {
	return configManager
}

// SetConfig 设置配置
func (cm *ConfigManager) SetConfig(config Config) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 验证并设置配置
	if config.WriteBuffer <= 0 {
		config.WriteBuffer = cm.config.WriteBuffer
	}
	if config.OpenFilesCacheCapacity <= 0 {
		config.OpenFilesCacheCapacity = cm.config.OpenFilesCacheCapacity
	}
	if config.BlockCacheCapacity <= 0 {
		config.BlockCacheCapacity = cm.config.BlockCacheCapacity
	}

	cm.config = config
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.config
}

// GetOptions 获取opt.Options
func (cm *ConfigManager) GetOptions() *opt.Options {
	config := cm.GetConfig()
	return &opt.Options{
		WriteBuffer:            config.WriteBuffer,
		OpenFilesCacheCapacity: config.OpenFilesCacheCapacity,
		BlockCacheCapacity:     config.BlockCacheCapacity,
		Compression:            config.Compression,
	}
}

// GetScenarioOptions 根据场景获取配置
func (cm *ConfigManager) GetScenarioOptions(scenario string) *opt.Options {
	switch scenario {
	case ScenarioEmbedded:
		return &opt.Options{
			WriteBuffer:            2 * 1024 * 1024,
			OpenFilesCacheCapacity: 5,
			BlockCacheCapacity:     4 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioIoT:
		return &opt.Options{
			WriteBuffer:            4 * 1024 * 1024,
			OpenFilesCacheCapacity: 10,
			BlockCacheCapacity:     8 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioEdge:
		return &opt.Options{
			WriteBuffer:            16 * 1024 * 1024,
			OpenFilesCacheCapacity: 50,
			BlockCacheCapacity:     32 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioGame:
		return &opt.Options{
			WriteBuffer:            64 * 1024 * 1024,
			OpenFilesCacheCapacity: 200,
			BlockCacheCapacity:     128 * 1024 * 1024,
			Compression:            opt.NoCompression,
		}
	default:
		// 使用当前配置
		return cm.GetOptions()
	}
}

// GetCustomOptions 根据自定义配置获取opt.Options
func (cm *ConfigManager) GetCustomOptions(config Config) *opt.Options {
	return &opt.Options{
		WriteBuffer:            config.WriteBuffer,
		OpenFilesCacheCapacity: config.OpenFilesCacheCapacity,
		BlockCacheCapacity:     config.BlockCacheCapacity,
		Compression:            config.Compression,
	}
}

// LoadFromStore 从存储中加载配置
func (cm *ConfigManager) LoadFromStore(path string) error {
	// 创建临时配置用于打开存储
	tempOpts := cm.getMinimalOptions()

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
			key: ConfigKeyWriteBuffer,
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					cm.mutex.Lock()
					cm.config.WriteBuffer = size
					cm.mutex.Unlock()
				}
				return nil
			},
			desc: "write buffer size",
		},
		{
			key: ConfigKeyMaxOpenFiles,
			parser: func(s string) (interface{}, error) {
				return strconv.Atoi(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					cm.mutex.Lock()
					cm.config.OpenFilesCacheCapacity = size
					cm.mutex.Unlock()
				}
				return nil
			},
			desc: "max open files",
		},
		{
			key: ConfigKeyBlockCache,
			parser: func(s string) (interface{}, error) {
				return parseSize(s)
			},
			apply: func(val interface{}) error {
				if size, ok := val.(int); ok && size > 0 {
					cm.mutex.Lock()
					cm.config.BlockCacheCapacity = size
					cm.mutex.Unlock()
				}
				return nil
			},
			desc: "block cache capacity",
		},
		{
			key: ConfigKeyCompression,
			parser: func(s string) (interface{}, error) {
				return strconv.ParseBool(s)
			},
			apply: func(val interface{}) error {
				if enabled, ok := val.(bool); ok {
					cm.mutex.Lock()
					if enabled {
						cm.config.Compression = opt.DefaultCompression
					} else {
						cm.config.Compression = opt.NoCompression
					}
					cm.mutex.Unlock()
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

// SaveToStore 将配置保存到存储
func (cm *ConfigManager) SaveToStore(path string) error {
	// 创建临时配置用于打开存储
	tempOpts := cm.getMinimalOptions()

	// 尝试打开临时存储来写入配置
	tempDB, err := leveldb.OpenFile(path, tempOpts)
	if err != nil {
		return fmt.Errorf("failed to open temp DB for config saving: %v", err)
	}
	defer tempDB.Close()

	// 获取当前配置
	config := cm.GetConfig()

	// 定义配置项
	configItems := map[string]string{
		ConfigKeyWriteBuffer:  strconv.Itoa(config.WriteBuffer),
		ConfigKeyMaxOpenFiles: strconv.Itoa(config.OpenFilesCacheCapacity),
		ConfigKeyBlockCache:   strconv.Itoa(config.BlockCacheCapacity),
		ConfigKeyCompression:  strconv.FormatBool(config.Compression == opt.DefaultCompression),
	}

	// 写入所有配置项
	batch := new(leveldb.Batch)
	for key, value := range configItems {
		batch.Put([]byte(key), []byte(value))
	}

	// 提交批量操作
	if err := tempDB.Write(batch, nil); err != nil {
		return fmt.Errorf("failed to write config batch: %v", err)
	}

	return nil
}

// getMinimalOptions 获取最小配置选项，用于临时打开存储
func (cm *ConfigManager) getMinimalOptions() *opt.Options {
	return &opt.Options{
		WriteBuffer:            MinWriteBuffer,
		OpenFilesCacheCapacity: MinOpenFilesCacheCapacity,
		BlockCacheCapacity:     MinBlockCacheCapacity,
	}
}

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

// 全局函数，保持向后兼容
var (
	// SetConfig 设置全局配置
	SetConfig = configManager.SetConfig

	// GetConfig 获取当前全局配置
	GetConfig = configManager.GetConfig

	// GetScenarioOptions 根据场景获取配置
	GetScenarioOptions = configManager.GetScenarioOptions

	// GetCustomOptions 根据自定义配置获取opt.Options
	GetCustomOptions = configManager.GetCustomOptions

	// LoadConfigFromStore 从存储中加载配置到全局配置
	LoadConfigFromStore = configManager.LoadFromStore

	// SaveConfigToStore 将全局配置保存到存储
	SaveConfigToStore = configManager.SaveToStore
)
