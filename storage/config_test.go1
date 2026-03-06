package storage

import (
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb/opt"
)

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	// 获取配置管理器实例
	configMgr := GetConfigManager()
	if configMgr == nil {
		t.Fatalf("GetConfigManager() returned nil")
	}

	// 测试获取默认配置
	defaultConfig := configMgr.GetConfig()
	if defaultConfig.WriteBuffer != DefaultWriteBuffer {
		t.Errorf("Expected WriteBuffer %d, got %d", DefaultWriteBuffer, defaultConfig.WriteBuffer)
	}

	// 测试设置配置
	customConfig := Config{
		WriteBuffer:            32 * 1024 * 1024,
		OpenFilesCacheCapacity: 100,
		BlockCacheCapacity:     64 * 1024 * 1024,
		Compression:            opt.DefaultCompression,
	}
	configMgr.SetConfig(customConfig)

	// 测试获取更新后的配置
	updatedConfig := configMgr.GetConfig()
	if updatedConfig.WriteBuffer != customConfig.WriteBuffer {
		t.Errorf("Expected WriteBuffer %d, got %d", customConfig.WriteBuffer, updatedConfig.WriteBuffer)
	}
	if updatedConfig.OpenFilesCacheCapacity != customConfig.OpenFilesCacheCapacity {
		t.Errorf("Expected OpenFilesCacheCapacity %d, got %d", customConfig.OpenFilesCacheCapacity, updatedConfig.OpenFilesCacheCapacity)
	}
	if updatedConfig.BlockCacheCapacity != customConfig.BlockCacheCapacity {
		t.Errorf("Expected BlockCacheCapacity %d, got %d", customConfig.BlockCacheCapacity, updatedConfig.BlockCacheCapacity)
	}
}

// TestScenarioOptions 测试场景配置
func TestScenarioOptions(t *testing.T) {
	configMgr := GetConfigManager()

	// 测试嵌入式场景配置
	embeddedOpts := configMgr.GetScenarioOptions(ScenarioEmbedded)
	if embeddedOpts.WriteBuffer != 2*1024*1024 {
		t.Errorf("Expected embedded WriteBuffer 2MB, got %d", embeddedOpts.WriteBuffer)
	}

	// 测试IoT场景配置
	iotOpts := configMgr.GetScenarioOptions(ScenarioIoT)
	if iotOpts.WriteBuffer != 4*1024*1024 {
		t.Errorf("Expected IoT WriteBuffer 4MB, got %d", iotOpts.WriteBuffer)
	}

	// 测试边缘计算场景配置
	edgeOpts := configMgr.GetScenarioOptions(ScenarioEdge)
	if edgeOpts.WriteBuffer != 16*1024*1024 {
		t.Errorf("Expected edge WriteBuffer 16MB, got %d", edgeOpts.WriteBuffer)
	}

	// 测试游戏场景配置
	gameOpts := configMgr.GetScenarioOptions(ScenarioGame)
	if gameOpts.WriteBuffer != 64*1024*1024 {
		t.Errorf("Expected game WriteBuffer 64MB, got %d", gameOpts.WriteBuffer)
	}

	// 测试默认场景配置
	defaultOpts := configMgr.GetScenarioOptions(ScenarioDefault)
	currentConfig := configMgr.GetConfig()
	if defaultOpts.WriteBuffer != currentConfig.WriteBuffer {
		t.Errorf("Expected default WriteBuffer %d, got %d", currentConfig.WriteBuffer, defaultOpts.WriteBuffer)
	}
}

// TestDBManager 测试数据库管理器
func TestDBManager(t *testing.T) {
	// 获取数据库管理器实例
	dbMgr := GetDBManager()
	if dbMgr == nil {
		t.Fatalf("GetDBManager() returned nil")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "sfsdb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试打开数据库
	db, err := dbMgr.OpenDB(tempDir)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	if db == nil {
		t.Fatalf("OpenDB returned nil DB")
	}

	// 测试获取数据库实例
	getDb := dbMgr.GetDB()
	if getDb == nil {
		t.Fatalf("GetDB returned nil")
	}

	// 测试关闭数据库
	if err := dbMgr.CloseDB(); err != nil {
		t.Fatalf("Failed to close DB: %v", err)
	}
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	// 测试全局配置函数
	customConfig := Config{
		WriteBuffer:            16 * 1024 * 1024,
		OpenFilesCacheCapacity: 50,
		BlockCacheCapacity:     32 * 1024 * 1024,
		Compression:            opt.DefaultCompression,
	}

	// 使用全局函数设置配置
	SetConfig(customConfig)

	// 使用全局函数获取配置
	globalConfig := GetConfig()
	if globalConfig.WriteBuffer != customConfig.WriteBuffer {
		t.Errorf("Expected WriteBuffer %d, got %d", customConfig.WriteBuffer, globalConfig.WriteBuffer)
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "sfsdb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试全局数据库函数
	db, err := OpenDefaultDb(tempDir)
	if err != nil {
		t.Fatalf("Failed to open default DB: %v", err)
	}
	if db == nil {
		t.Fatalf("OpenDefaultDb returned nil DB")
	}

	// 测试全局KVDb变量
	if KVDb == nil {
		t.Fatalf("KVDb is nil after OpenDefaultDb")
	}

	// 测试关闭数据库
	if err := CloseDb(); err != nil {
		t.Fatalf("Failed to close DB: %v", err)
	}
}
