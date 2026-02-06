package monitor

import (
	"sync"
	"testing"
	"time"
)

// TestKeysMap_IncAsync 测试 KeysMap 的 IncAsync 方法
func TestKeysMap_IncAsync(t *testing.T) {
	// 创建 KeysMap 实例
	km := NewKeysMap()
	
	// 异步增加计数器
	km.IncAsync(1, 1, "test")
	
	// 等待一段时间，确保异步操作完成
	time.Sleep(100 * time.Millisecond)
	
	// 获取所有计数器数据
	all := km.GetAll()
	
	// 验证计数器值
	if len(all) != 1 {
		t.Errorf("Expected 1 counter, got %d", len(all))
	}
	
	if stats, ok := all[1]; ok {
		if stats.PutCount.Load() != 1 {
			t.Errorf("Expected PutCount=1, got %d", stats.PutCount.Load())
		}
		if stats.DeleteCount.Load() != 0 {
			t.Errorf("Expected DeleteCount=0, got %d", stats.DeleteCount.Load())
		}
	} else {
		t.Error("Counter for key 1 not found")
	}
}

// TestKeysMap_DecAsync 测试 KeysMap 的 DecAsync 方法
func TestKeysMap_DecAsync(t *testing.T) {
	// 创建 KeysMap 实例
	km := NewKeysMap()
	
	// 先同步增加计数器，确保键存在
	km.Inc(1, 1, "test")
	
	// 异步减少计数器
	km.DecAsync(1, 1, "test")
	
	// 等待一段时间，确保异步操作完成
	time.Sleep(100 * time.Millisecond)
	
	// 获取所有计数器数据
	all := km.GetAll()
	
	// 验证计数器值
	if stats, ok := all[1]; ok {
		if stats.DeleteCount.Load() != 1 {
			t.Errorf("Expected DeleteCount=1, got %d", stats.DeleteCount.Load())
		}
	} else {
		t.Error("Counter for key 1 not found")
	}
}

// TestKeysMap_ConcurrentIncAsync 测试 KeysMap 的 IncAsync 方法在并发场景下的表现
func TestKeysMap_ConcurrentIncAsync(t *testing.T) {
	// 创建 KeysMap 实例
	km := NewKeysMap()
	
	// 并发增加计数器
	var wg sync.WaitGroup
	count := 100
	
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			km.IncAsync(1, 1, "test")
		}()
	}
	
	// 等待所有协程完成
	wg.Wait()
	
	// 等待一段时间，确保所有异步操作完成
	time.Sleep(500 * time.Millisecond)
	
	// 获取所有计数器数据
	all := km.GetAll()
	
	// 验证计数器值
	if stats, ok := all[1]; ok {
		if stats.PutCount.Load() != int64(count) {
			t.Errorf("Expected PutCount=%d, got %d", count, stats.PutCount.Load())
		}
	} else {
		t.Error("Counter for key 1 not found")
	}
}

// TestIndexStatsMap_SettimeAsync 测试 IndexStatsMap 的 SettimeAsync 方法
func TestIndexStatsMap_SettimeAsync(t *testing.T) {
	// 创建 IndexStatsMap 实例
	ism := NewIndexStatsMap()
	
	// 异步记录索引耗时
	duration := 10 * time.Millisecond
	ism.SettimeAsync(1, duration, "test_table", "test_index")
	
	// 等待一段时间，确保异步操作完成
	time.Sleep(100 * time.Millisecond)
	
	// 获取所有统计数据
	all := ism.GetAll()
	
	// 验证统计数据
	if len(all) != 1 {
		t.Errorf("Expected 1 stats, got %d", len(all))
	}
	
	if stats, ok := all[1]; ok {
		if stats.Count.Load() != 1 {
			t.Errorf("Expected Count=1, got %d", stats.Count.Load())
		}
		if stats.TblName != "test_table" {
			t.Errorf("Expected TblName=test_table, got %s", stats.TblName)
		}
		if stats.IndxName != "test_index" {
			t.Errorf("Expected IndxName=test_index, got %s", stats.IndxName)
		}
	} else {
		t.Error("Stats for key 1 not found")
	}
}

// TestIndexStatsMap_ConcurrentSettimeAsync 测试 IndexStatsMap 的 SettimeAsync 方法在并发场景下的表现
func TestIndexStatsMap_ConcurrentSettimeAsync(t *testing.T) {
	// 创建 IndexStatsMap 实例
	ism := NewIndexStatsMap()
	
	// 并发记录索引耗时
	var wg sync.WaitGroup
	count := 100
	
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			duration := 10 * time.Millisecond
			ism.SettimeAsync(1, duration, "test_table", "test_index")
		}()
	}
	
	// 等待所有协程完成
	wg.Wait()
	
	// 等待一段时间，确保所有异步操作完成
	time.Sleep(500 * time.Millisecond)
	
	// 获取所有统计数据
	all := ism.GetAll()
	
	// 验证统计数据
	if stats, ok := all[1]; ok {
		if stats.Count.Load() != int64(count) {
			t.Errorf("Expected Count=%d, got %d", count, stats.Count.Load())
		}
	} else {
		t.Error("Stats for key 1 not found")
	}
}

// BenchmarkKeysMap_IncAsync 测试 KeysMap 的 IncAsync 方法的性能
func BenchmarkKeysMap_IncAsync(b *testing.B) {
	// 创建 KeysMap 实例
	km := NewKeysMap()
	
	// 重置计时器
	b.ResetTimer()
	
	// 执行性能测试
	for i := 0; i < b.N; i++ {
		km.IncAsync(i, 1, "test")
	}
	
	// 等待所有异步操作完成
	time.Sleep(1 * time.Second)
}

// BenchmarkIndexStatsMap_SettimeAsync 测试 IndexStatsMap 的 SettimeAsync 方法的性能
func BenchmarkIndexStatsMap_SettimeAsync(b *testing.B) {
	// 创建 IndexStatsMap 实例
	ism := NewIndexStatsMap()
	
	// 重置计时器
	b.ResetTimer()
	
	// 执行性能测试
	for i := 0; i < b.N; i++ {
		duration := 10 * time.Millisecond
		ism.SettimeAsync(i, duration, "test_table", "test_index")
	}
	
	// 等待所有异步操作完成
	time.Sleep(1 * time.Second)
}
