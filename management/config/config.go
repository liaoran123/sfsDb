package config

import (
	"strconv"
	"strings"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// ConfigInfo 配置信息
type ConfigInfo struct {
	StoreType string            // 存储类型
	Options   map[string]string // 配置选项
	Scenarios []string          // 支持的场景
	// 其他配置信息可以根据需要扩展
}

// ConfigManager 配置管理器
type ConfigManager struct {
}

// NewConfigManager 创建配置管理器
// 返回:
//   *ConfigManager: 配置管理器实例

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// GetConfig 获取当前配置
// 返回:
//   ConfigInfo: 配置信息
//   error: 错误信息

func (cm *ConfigManager) GetConfig() (ConfigInfo, error) {
	// 获取存储的配置
	config := storage.GetConfig()

	// 转换为 ConfigInfo 格式
	configInfo := ConfigInfo{
		StoreType: "LevelDB",
		Options: map[string]string{
			"write_buffer":   strconv.Itoa(config.WriteBuffer),
			"max_open_files": strconv.Itoa(config.OpenFilesCacheCapacity),
			"block_cache":    strconv.Itoa(config.BlockCacheCapacity),
			"compression":    strconv.FormatBool(config.Compression != opt.NoCompression),
		},
		Scenarios: []string{
			storage.ScenarioEmbedded,
			storage.ScenarioIoT,
			storage.ScenarioEdge,
			storage.ScenarioGame,
			storage.ScenarioDefault,
		},
	}

	return configInfo, nil
}

// SetConfig 设置配置
// 参数:
//   key: 配置键
//   value: 配置值

func (cm *ConfigManager) SetConfig(key string, value string) {
	// 获取当前配置
	config := storage.GetConfig()

	// 根据键设置值
	switch key {
	case "write_buffer":
		if size, err := ParseSize(value); err == nil {
			config.WriteBuffer = size
		}
	case "max_open_files":
		if size, err := strconv.Atoi(value); err == nil {
			config.OpenFilesCacheCapacity = size
		}
	case "block_cache":
		if size, err := ParseSize(value); err == nil {
			config.BlockCacheCapacity = size
		}
	case "compression":
		if enabled, err := strconv.ParseBool(value); err == nil {
			if enabled {
				config.Compression = opt.DefaultCompression
			} else {
				config.Compression = opt.NoCompression
			}
		}
	}

	// 设置配置
	storage.SetConfig(config)
}

// SetScenarioConfig 设置场景配置
// 参数:
//   scenario: 场景名称

func (cm *ConfigManager) SetScenarioConfig(scenario string) {
	// 获取场景配置
	scenarioOpts := storage.GetScenarioOptions(scenario)

	// 转换为 Config 格式
	config := storage.GetConfig()
	config.WriteBuffer = scenarioOpts.WriteBuffer
	config.OpenFilesCacheCapacity = scenarioOpts.OpenFilesCacheCapacity
	config.BlockCacheCapacity = scenarioOpts.BlockCacheCapacity
	config.Compression = scenarioOpts.Compression

	// 设置配置
	storage.SetConfig(config)
}

// GetOptimizationSuggestions 获取优化建议
// 返回:
//   []string: 优化建议列表
//   error: 错误信息

func (cm *ConfigManager) GetOptimizationSuggestions() ([]string, error) {
	// 获取当前配置
	config := storage.GetConfig()

	suggestions := []string{}

	// 基于当前配置生成优化建议
	if config.WriteBuffer < 16*1024*1024 {
		suggestions = append(suggestions, "考虑增加 write_buffer 大小以提高写入性能，建议设置为 16MB 或更高")
	}

	if config.OpenFilesCacheCapacity < 50 {
		suggestions = append(suggestions, "根据内存情况调整 max_open_files 参数，内存充足时可设置为 100 或更高")
	}

	if config.BlockCacheCapacity < 32*1024*1024 {
		suggestions = append(suggestions, "考虑增加 block_cache 大小以提高读取性能，建议设置为 32MB 或更高")
	}

	// 场景特定建议
	suggestions = append(suggestions, "对于 IoT 场景，建议使用 '"+storage.ScenarioIoT+"' 配置")
	suggestions = append(suggestions, "对于边缘计算场景，建议使用 '"+storage.ScenarioEdge+"' 配置")
	suggestions = append(suggestions, "对于嵌入式场景，建议使用 '"+storage.ScenarioEmbedded+"' 配置")

	return suggestions, nil
}

// ValidateConfig 验证配置
// 参数:
//   config: 配置信息
// 返回:
//   bool: 是否有效
//   error: 错误信息

func (cm *ConfigManager) ValidateConfig(config ConfigInfo) (bool, error) {
	// 检查必要的配置选项是否存在
	requiredOptions := []string{"write_buffer", "max_open_files", "block_cache", "compression"}
	for _, option := range requiredOptions {
		if _, ok := config.Options[option]; !ok {
			return false, nil
		}
	}

	return true, nil
}

// ResetConfig 重置配置为默认值

func (cm *ConfigManager) ResetConfig() {
	// 重置为默认配置
	defaultConfig := storage.Config{
		WriteBuffer:            storage.DefaultWriteBuffer,
		OpenFilesCacheCapacity: storage.DefaultOpenFilesCacheCapacity,
		BlockCacheCapacity:     storage.DefaultBlockCacheCapacity,
		Compression:            opt.DefaultCompression,
	}

	storage.SetConfig(defaultConfig)
}

// ParseSize 解析大小字符串，支持 "64MB" 或 "67108864" 格式
func ParseSize(s string) (int, error) {
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
			return 0, err
		}
		return size, nil
	}

	// 解析数值部分
	size, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}

	// 计算最终大小
	return size * multiplier, nil
}
