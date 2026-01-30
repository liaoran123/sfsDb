package engine

import (
	"testing"
)

// TestStrategyConfig 测试策略配置管理
func TestStrategyConfig(t *testing.T) {
	// 测试默认配置
	defaultConfig := GetStrategyConfig()
	if defaultConfig.StrategyType != StrategyStatic {
		t.Errorf("默认策略类型应为 StrategyStatic，实际为 %v", defaultConfig.StrategyType)
	}

	// 测试设置配置
	customConfig := StrategyConfig{
		StrategyType:       StrategyDynamic,
		ConcurrencyThreshold: 500,
		MemoryThreshold:    300,
		RequestThreshold:   500,
	}
	SetStrategyConfig(customConfig)

	// 测试获取配置
	retrievedConfig := GetStrategyConfig()
	if retrievedConfig.StrategyType != StrategyDynamic {
		t.Errorf("策略类型应为 StrategyDynamic，实际为 %v", retrievedConfig.StrategyType)
	}
	if retrievedConfig.ConcurrencyThreshold != 500 {
		t.Errorf("并发阈值应为 500，实际为 %v", retrievedConfig.ConcurrencyThreshold)
	}
	if retrievedConfig.MemoryThreshold != 300 {
		t.Errorf("内存阈值应为 300，实际为 %v", retrievedConfig.MemoryThreshold)
	}
	if retrievedConfig.RequestThreshold != 500 {
		t.Errorf("请求阈值应为 500，实际为 %v", retrievedConfig.RequestThreshold)
	}

	// 重置为默认配置
	SetStrategyConfig(defaultConfig)
}

// TestSystemStateMonitoring 测试系统状态监控
func TestSystemStateMonitoring(t *testing.T) {
	// 记录请求
	recordRequestStart()
	
	// 获取系统状态
	concurrency, memory, requests := getSystemState()
	
	// 验证状态值
	if concurrency < 0 {
		t.Errorf("并发数不应为负数，实际为 %v", concurrency)
	}
	if memory < 0 {
		t.Errorf("内存使用不应为负数，实际为 %v", memory)
	}
	if requests < 0 {
		t.Errorf("请求数不应为负数，实际为 %v", requests)
	}
	
	// 结束请求
	recordRequestEnd()
}

// TestDynamicStrategy 测试动态策略
func TestDynamicStrategy(t *testing.T) {
	// 保存原始配置
	originalConfig := GetStrategyConfig()
	defer SetStrategyConfig(originalConfig)

	// 设置低阈值，强制使用对象池
	SetStrategyConfig(StrategyConfig{
		StrategyType:       StrategyDynamic,
		ConcurrencyThreshold: 0, // 强制使用对象池
		MemoryThreshold:    0,
		RequestThreshold:   0,
	})

	// 测试获取切片
	s := GetStringSliceWithConfigStrategy()
	if s == nil {
		t.Errorf("获取的切片不应为 nil")
	}
	if len(s) != 0 {
		t.Errorf("获取的切片应为空，实际长度为 %v", len(s))
	}

	// 使用切片
	s = append(s, "test")
	if len(s) != 1 {
		t.Errorf("切片长度应为 1，实际为 %v", len(s))
	}

	// 归还切片
	PutStringSliceWithConfigStrategy(s)
}

// TestHybridStrategy 测试混合策略
func TestHybridStrategy(t *testing.T) {
	// 保存原始配置
	originalConfig := GetStrategyConfig()
	defer SetStrategyConfig(originalConfig)

	// 设置混合策略
	SetStrategyConfig(StrategyConfig{
		StrategyType:       StrategyHybrid,
		ConcurrencyThreshold: 1000, // 高阈值，默认使用静态策略
		MemoryThreshold:    500,
		RequestThreshold:   1000,
	})

	// 测试获取切片
	s := GetStringSliceWithConfigStrategy()
	if s == nil {
		t.Errorf("获取的切片不应为 nil")
	}
	if len(s) != 0 {
		t.Errorf("获取的切片应为空，实际长度为 %v", len(s))
	}

	// 使用切片
	s = append(s, "test")
	if len(s) != 1 {
		t.Errorf("切片长度应为 1，实际为 %v", len(s))
	}

	// 归还切片
	PutStringSliceWithConfigStrategy(s)
}

// TestStrategySwitching 测试策略切换
func TestStrategySwitching(t *testing.T) {
	// 保存原始配置
	originalConfig := GetStrategyConfig()
	defer SetStrategyConfig(originalConfig)

	// 测试切换到静态策略
	SwitchToStaticStrategy()
	config := GetStrategyConfig()
	if config.StrategyType != StrategyStatic {
		t.Errorf("策略类型应为 StrategyStatic，实际为 %v", config.StrategyType)
	}

	// 测试切换到动态策略
	SwitchToDynamicStrategy()
	config = GetStrategyConfig()
	if config.StrategyType != StrategyDynamic {
		t.Errorf("策略类型应为 StrategyDynamic，实际为 %v", config.StrategyType)
	}

	// 测试切换到混合策略
	SwitchToHybridStrategy()
	config = GetStrategyConfig()
	if config.StrategyType != StrategyHybrid {
		t.Errorf("策略类型应为 StrategyHybrid，实际为 %v", config.StrategyType)
	}

	// 测试更新阈值
	UpdateStrategyThresholds(2000, 1000, 2000)
	config = GetStrategyConfig()
	if config.ConcurrencyThreshold != 2000 {
		t.Errorf("并发阈值应为 2000，实际为 %v", config.ConcurrencyThreshold)
	}
	if config.MemoryThreshold != 1000 {
		t.Errorf("内存阈值应为 1000，实际为 %v", config.MemoryThreshold)
	}
	if config.RequestThreshold != 2000 {
		t.Errorf("请求阈值应为 2000，实际为 %v", config.RequestThreshold)
	}
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	// 保存原始配置
	originalConfig := GetStrategyConfig()
	defer SetStrategyConfig(originalConfig)

	// 确保默认使用静态策略
	SetStrategyConfig(StrategyConfig{
		StrategyType:       StrategyStatic,
		ConcurrencyThreshold: 1000,
		MemoryThreshold:    500,
		RequestThreshold:   1000,
	})

	// 测试使用静态策略
	s := GetStringSliceWithConfigStrategy()
	if s == nil {
		t.Errorf("获取的切片不应为 nil")
	}
	if len(s) != 0 {
		t.Errorf("获取的切片应为空，实际长度为 %v", len(s))
	}

	// 使用切片
	s = append(s, "test")
	if len(s) != 1 {
		t.Errorf("切片长度应为 1，实际为 %v", len(s))
	}

	// 归还切片
	PutStringSliceWithConfigStrategy(s)
}

// BenchmarkStrategyPerformance 基准测试不同策略的性能
func BenchmarkStrategyPerformance(b *testing.B) {
	// 保存原始配置
	originalConfig := GetStrategyConfig()
	defer SetStrategyConfig(originalConfig)

	// 基准测试静态策略
	b.Run("StaticStrategy", func(b *testing.B) {
		SetStrategyConfig(StrategyConfig{
			StrategyType:       StrategyStatic,
			ConcurrencyThreshold: 1000,
			MemoryThreshold:    500,
			RequestThreshold:   1000,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s := GetStringSliceWithConfigStrategy()
			s = append(s, "test")
			PutStringSliceWithConfigStrategy(s)
		}
	})

	// 基准测试动态策略
	b.Run("DynamicStrategy", func(b *testing.B) {
		SetStrategyConfig(StrategyConfig{
			StrategyType:       StrategyDynamic,
			ConcurrencyThreshold: 1000,
			MemoryThreshold:    500,
			RequestThreshold:   1000,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s := GetStringSliceWithConfigStrategy()
			s = append(s, "test")
			PutStringSliceWithConfigStrategy(s)
		}
	})

	// 基准测试混合策略
	b.Run("HybridStrategy", func(b *testing.B) {
		SetStrategyConfig(StrategyConfig{
			StrategyType:       StrategyHybrid,
			ConcurrencyThreshold: 1000,
			MemoryThreshold:    500,
			RequestThreshold:   1000,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s := GetStringSliceWithConfigStrategy()
			s = append(s, "test")
			PutStringSliceWithConfigStrategy(s)
		}
	})
}
