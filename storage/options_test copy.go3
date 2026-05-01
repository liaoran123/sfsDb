package storage

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb/opt"
)

// TestConfigManager 测试 ConfigManager 的基本功能
func TestConfigManager(t *testing.T) {
	// 测试获取默认配置
	config := GetConfig()
	if config.WriteBuffer != DefaultWriteBuffer {
		t.Errorf("Expected default WriteBuffer %d, got %d", DefaultWriteBuffer, config.WriteBuffer)
	}
	if config.OpenFilesCacheCapacity != DefaultOpenFilesCacheCapacity {
		t.Errorf("Expected default OpenFilesCacheCapacity %d, got %d", DefaultOpenFilesCacheCapacity, config.OpenFilesCacheCapacity)
	}
	if config.BlockCacheCapacity != DefaultBlockCacheCapacity {
		t.Errorf("Expected default BlockCacheCapacity %d, got %d", DefaultBlockCacheCapacity, config.BlockCacheCapacity)
	}
	if config.Compression != opt.DefaultCompression {
		t.Errorf("Expected default Compression %v, got %v", opt.DefaultCompression, config.Compression)
	}
}

// TestSetConfig 测试 SetConfig 功能
func TestSetConfig(t *testing.T) {
	// 保存原始配置
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	// 创建自定义配置
	newConfig := Config{
		WriteBuffer:            128 * 1024 * 1024,
		OpenFilesCacheCapacity: 500,
		BlockCacheCapacity:     256 * 1024 * 1024,
		Compression:            opt.NoCompression,
	}

	// 设置新配置
	SetConfig(newConfig)

	// 验证配置已更新
	currentConfig := GetConfig()
	if currentConfig.WriteBuffer != newConfig.WriteBuffer {
		t.Errorf("Expected WriteBuffer %d, got %d", newConfig.WriteBuffer, currentConfig.WriteBuffer)
	}
	if currentConfig.OpenFilesCacheCapacity != newConfig.OpenFilesCacheCapacity {
		t.Errorf("Expected OpenFilesCacheCapacity %d, got %d", newConfig.OpenFilesCacheCapacity, currentConfig.OpenFilesCacheCapacity)
	}
	if currentConfig.BlockCacheCapacity != newConfig.BlockCacheCapacity {
		t.Errorf("Expected BlockCacheCapacity %d, got %d", newConfig.BlockCacheCapacity, currentConfig.BlockCacheCapacity)
	}
	if currentConfig.Compression != newConfig.Compression {
		t.Errorf("Expected Compression %v, got %v", newConfig.Compression, currentConfig.Compression)
	}
}

// TestSetConfigWithDefaults 测试 SetConfig 时的默认值处理
func TestSetConfigWithDefaults(t *testing.T) {
	// 保存原始配置
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	// 创建部分配置，某些值为 0
	partialConfig := Config{
		WriteBuffer:            0,
		OpenFilesCacheCapacity: 0,
		BlockCacheCapacity:     0,
		Compression:            opt.NoCompression,
	}

	// 设置配置
	SetConfig(partialConfig)

	// 验证 0 值被替换为原始值
	currentConfig := GetConfig()
	if currentConfig.WriteBuffer != originalConfig.WriteBuffer {
		t.Errorf("Expected WriteBuffer to remain %d, got %d", originalConfig.WriteBuffer, currentConfig.WriteBuffer)
	}
	if currentConfig.OpenFilesCacheCapacity != originalConfig.OpenFilesCacheCapacity {
		t.Errorf("Expected OpenFilesCacheCapacity to remain %d, got %d", originalConfig.OpenFilesCacheCapacity, currentConfig.OpenFilesCacheCapacity)
	}
	if currentConfig.BlockCacheCapacity != originalConfig.BlockCacheCapacity {
		t.Errorf("Expected BlockCacheCapacity to remain %d, got %d", originalConfig.BlockCacheCapacity, currentConfig.BlockCacheCapacity)
	}
	if currentConfig.Compression != opt.NoCompression {
		t.Errorf("Expected Compression to be %v, got %v", opt.NoCompression, currentConfig.Compression)
	}
}

// TestGetScenarioConfig 测试 GetScenarioConfig 函数
func TestGetScenarioConfig(t *testing.T) {
	testCases := []struct {
		name     string
		scenario string
		check    func(Config) bool
	}{
		{
			name:     "Embedded",
			scenario: ScenarioEmbedded,
			check: func(c Config) bool {
				return c.WriteBuffer == EmbeddedWriteBuffer &&
					c.OpenFilesCacheCapacity == EmbeddedOpenFilesCacheCapacity &&
					c.BlockCacheCapacity == EmbeddedBlockCacheCapacity
			},
		},
		{
			name:     "IoT",
			scenario: ScenarioIoT,
			check: func(c Config) bool {
				return c.WriteBuffer == IoTWriteBuffer &&
					c.OpenFilesCacheCapacity == IoTOpenFilesCacheCapacity &&
					c.BlockCacheCapacity == IoTBlockCacheCapacity
			},
		},
		{
			name:     "Edge",
			scenario: ScenarioEdge,
			check: func(c Config) bool {
				return c.WriteBuffer == EdgeWriteBuffer &&
					c.OpenFilesCacheCapacity == EdgeOpenFilesCacheCapacity &&
					c.BlockCacheCapacity == EdgeBlockCacheCapacity
			},
		},
		{
			name:     "Game",
			scenario: ScenarioGame,
			check: func(c Config) bool {
				return c.WriteBuffer == GameWriteBuffer &&
					c.OpenFilesCacheCapacity == GameOpenFilesCacheCapacity &&
					c.BlockCacheCapacity == GameBlockCacheCapacity &&
					c.Compression == opt.NoCompression
			},
		},
		{
			name:     "Default",
			scenario: ScenarioDefault,
			check: func(c Config) bool {
				return c.WriteBuffer == DefaultWriteBuffer &&
					c.OpenFilesCacheCapacity == DefaultOpenFilesCacheCapacity &&
					c.BlockCacheCapacity == DefaultBlockCacheCapacity
			},
		},
		{
			name:     "UnknownScenario",
			scenario: "unknown",
			check: func(c Config) bool {
				return c.WriteBuffer == DefaultWriteBuffer
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := GetScenarioConfig(tc.scenario)
			if !tc.check(config) {
				t.Errorf("Scenario %s config check failed", tc.name)
			}
		})
	}
}

// TestGetScenarioOptions 测试 GetScenarioOptions 函数
func TestGetScenarioOptions(t *testing.T) {
	testCases := []struct {
		name     string
		scenario string
	}{
		{"Embedded", ScenarioEmbedded},
		{"IoT", ScenarioIoT},
		{"Edge", ScenarioEdge},
		{"Game", ScenarioGame},
		{"Default", ScenarioDefault},
		{"Unknown", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := GetScenarioOptions(tc.scenario)
			if opts == nil {
				t.Errorf("Expected options for scenario %s, got nil", tc.name)
			}
		})
	}
}

// TestGetOptions 测试 GetOptions 方法
func TestGetOptions(t *testing.T) {
	// 保存原始配置
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	// 设置自定义配置
	customConfig := Config{
		WriteBuffer:            128 * 1024 * 1024,
		OpenFilesCacheCapacity: 500,
		BlockCacheCapacity:     256 * 1024 * 1024,
		Compression:            opt.NoCompression,
	}
	SetConfig(customConfig)

	// 获取 LevelDB 选项
	cm := GetConfigManager()
	opts := cm.GetOptions()

	if opts.WriteBuffer != customConfig.WriteBuffer {
		t.Errorf("Expected WriteBuffer %d, got %d", customConfig.WriteBuffer, opts.WriteBuffer)
	}
	if opts.OpenFilesCacheCapacity != customConfig.OpenFilesCacheCapacity {
		t.Errorf("Expected OpenFilesCacheCapacity %d, got %d", customConfig.OpenFilesCacheCapacity, opts.OpenFilesCacheCapacity)
	}
	if opts.BlockCacheCapacity != customConfig.BlockCacheCapacity {
		t.Errorf("Expected BlockCacheCapacity %d, got %d", customConfig.BlockCacheCapacity, opts.BlockCacheCapacity)
	}
	if opts.Compression != customConfig.Compression {
		t.Errorf("Expected Compression %v, got %v", customConfig.Compression, opts.Compression)
	}
}
