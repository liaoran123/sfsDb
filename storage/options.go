package storage

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	DefaultWriteBuffer             = 64 * 1024 * 1024  // 默认写入缓冲区大小，64MB
	DefaultOpenFilesCacheCapacity  = 200               // 默认打开文件缓存容量，200
	DefaultBlockCacheCapacity      = 128 * 1024 * 1024 // 默认块缓存容量，128MB
	EmbeddedWriteBuffer            = 2 * 1024 * 1024   // 嵌入式场景写入缓冲区大小，2MB
	EmbeddedOpenFilesCacheCapacity = 5                 // 嵌入式场景打开文件缓存容量，5
	EmbeddedBlockCacheCapacity     = 4 * 1024 * 1024   // 嵌入式场景块缓存容量，4MB
	IoTWriteBuffer                 = 4 * 1024 * 1024   // 物联网场景写入缓冲区大小，4MB
	IoTOpenFilesCacheCapacity      = 10                // 物联网场景打开文件缓存容量，10
	IoTBlockCacheCapacity          = 8 * 1024 * 1024   // 物联网场景块缓存容量，8MB
	EdgeWriteBuffer                = 16 * 1024 * 1024  // 边缘场景写入缓冲区大小，16MB
	EdgeOpenFilesCacheCapacity     = 50                // 边缘场景打开文件缓存容量，50
	EdgeBlockCacheCapacity         = 32 * 1024 * 1024  // 边缘场景块缓存容量，32MB
	GameWriteBuffer                = 64 * 1024 * 1024  // 游戏场景写入缓冲区大小，64MB
	GameOpenFilesCacheCapacity     = 200
	GameBlockCacheCapacity         = 128 * 1024 * 1024 // 游戏场景块缓存容量，128MB
)

const (
	ScenarioEmbedded = "embedded" // 嵌入式场景，默认配置
	ScenarioIoT      = "iot"      // 物联网场景
	ScenarioEdge     = "edge"     // 边缘场景
	ScenarioGame     = "game"     // 游戏场景
	ScenarioDefault  = "default"  // 默认场景
)

type Config struct {
	WriteBuffer            int             // 写入缓冲区大小，默认64MB
	OpenFilesCacheCapacity int             // 打开文件缓存容量，默认200
	BlockCacheCapacity     int             // 块缓存容量，默认128MB
	Compression            opt.Compression // 压缩算法，默认Snappy压缩
}

var embeddedConfig = Config{
	WriteBuffer:            EmbeddedWriteBuffer,
	OpenFilesCacheCapacity: EmbeddedOpenFilesCacheCapacity,
	BlockCacheCapacity:     EmbeddedBlockCacheCapacity,
	Compression:            opt.DefaultCompression,
}

var iotConfig = Config{
	WriteBuffer:            IoTWriteBuffer,
	OpenFilesCacheCapacity: IoTOpenFilesCacheCapacity,
	BlockCacheCapacity:     IoTBlockCacheCapacity,
	Compression:            opt.DefaultCompression,
}

var edgeConfig = Config{
	WriteBuffer:            EdgeWriteBuffer,
	OpenFilesCacheCapacity: EdgeOpenFilesCacheCapacity,
	BlockCacheCapacity:     EdgeBlockCacheCapacity,
	Compression:            opt.DefaultCompression,
}

var gameConfig = Config{
	WriteBuffer:            GameWriteBuffer,
	OpenFilesCacheCapacity: GameOpenFilesCacheCapacity,
	BlockCacheCapacity:     GameBlockCacheCapacity,
	Compression:            opt.NoCompression,
}

var defaultConfig = Config{
	WriteBuffer:            DefaultWriteBuffer,
	OpenFilesCacheCapacity: DefaultOpenFilesCacheCapacity,
	BlockCacheCapacity:     DefaultBlockCacheCapacity,
	Compression:            opt.DefaultCompression,
}

type ConfigManager struct {
	config Config
}

// 创建配置管理器实例
var configManager = &ConfigManager{
	config: defaultConfig,
}

// 获取配置管理器实例
func GetConfigManager() *ConfigManager {
	return configManager
}

// 设置配置
func (cm *ConfigManager) SetConfig(config Config) {
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

// 获取配置
func (cm *ConfigManager) GetConfig() Config {
	return cm.config
}

// 自定义配置
func (cm *ConfigManager) GetOptions() *opt.Options {
	config := cm.GetConfig()
	return &opt.Options{
		WriteBuffer:            config.WriteBuffer,
		OpenFilesCacheCapacity: config.OpenFilesCacheCapacity,
		BlockCacheCapacity:     config.BlockCacheCapacity,
		Compression:            config.Compression,
	}
}

// 获取指定场景的配置
func GetScenarioConfig(scenario string) Config {
	switch scenario {
	case ScenarioEmbedded:
		return embeddedConfig
	case ScenarioIoT:
		return iotConfig
	case ScenarioEdge:
		return edgeConfig
	case ScenarioGame:
		return gameConfig
	default:
		return defaultConfig
	}
}

// 获取指定场景的自定义配置，根据GetScenarioConfig返回的配置创建opt.Options
func GetScenarioOptions(scenario string) *opt.Options {
	config := GetScenarioConfig(scenario)
	return &opt.Options{
		WriteBuffer:            config.WriteBuffer,
		OpenFilesCacheCapacity: config.OpenFilesCacheCapacity,
		BlockCacheCapacity:     config.BlockCacheCapacity,
		Compression:            config.Compression,
	}
}

var (
	SetConfig = configManager.SetConfig // 设置配置
	GetConfig = configManager.GetConfig // 获取配置
)
