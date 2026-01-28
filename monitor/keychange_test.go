package monitor

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// 重置测试数据
func resetTestData() {
	AtomicInt = make(map[string]*atomic.Int64)
	AtomicDec = make(map[string]*atomic.Int64)
}

// 测试基本计数功能
func TestAtomicMap_Inc(t *testing.T) {
	// 重置测试数据
	resetTestData()

	// 测试AtomicInt计数
	tableID := byte(1)
	indexID := byte(2)

	// 执行计数操作
	AtomicMap(AtomicInt).Inc(tableID, indexID)
	AtomicMap(AtomicInt).Inc(tableID, indexID)

	// 验证计数结果
	count := AtomicMap(AtomicInt).Get(tableID, indexID)
	if count != 2 {
		t.Errorf("AtomicInt.Inc() failed: expected 2, got %d", count)
	}
	fmt.Printf("AtomicInt count: %d\n", count)

	// 测试AtomicDec计数
	AtomicMap(AtomicDec).Inc(tableID, indexID)
	AtomicMap(AtomicDec).Inc(tableID, indexID)
	AtomicMap(AtomicDec).Inc(tableID, indexID)

	// 验证计数结果
	decCount := AtomicMap(AtomicDec).Get(tableID, indexID)
	if decCount != 3 {
		t.Errorf("AtomicDec.Inc() failed: expected 3, got %d", decCount)
	}
	fmt.Printf("AtomicDec count: %d\n", decCount)
}

// 测试并发计数功能
func TestAtomicMap_ConcurrentInc(t *testing.T) {
	// 重置测试数据
	resetTestData()

	tableID := byte(3)
	indexID := byte(4)
	concurrency := 1000
	expectedCount := int64(concurrency)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 并发执行计数操作
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			AtomicMap(AtomicInt).Inc(tableID, indexID)
		}()
	}

	wg.Wait()

	// 验证计数结果
	count := AtomicMap(AtomicInt).Get(tableID, indexID)
	if count != expectedCount {
		t.Errorf("Concurrent AtomicInt.Inc() failed: expected %d, got %d", expectedCount, count)
	}
	fmt.Printf("Concurrent AtomicInt count: %d\n", count)
}

// 测试不同表和索引的计数隔离
func TestAtomicMap_SeparateCounters(t *testing.T) {
	// 重置测试数据
	resetTestData()

	// 测试不同表的计数隔离
	table1ID := byte(5)
	table2ID := byte(6)
	indexID := byte(7)

	// 对表1执行计数
	AtomicMap(AtomicInt).Inc(table1ID, indexID)
	AtomicMap(AtomicInt).Inc(table1ID, indexID)

	// 对表2执行计数
	AtomicMap(AtomicInt).Inc(table2ID, indexID)

	// 验证表1计数
	count1 := AtomicMap(AtomicInt).Get(table1ID, indexID)
	if count1 != 2 {
		t.Errorf("Table 1 count failed: expected 2, got %d", count1)
	}
	fmt.Printf("Table 1 count: %d\n", count1)

	// 验证表2计数
	count2 := AtomicMap(AtomicInt).Get(table2ID, indexID)
	if count2 != 1 {
		t.Errorf("Table 2 count failed: expected 1, got %d", count2)
	}
	fmt.Printf("Table 2 count: %d\n", count2)
}
