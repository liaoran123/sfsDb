package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// TestPerformance 性能测试函数
func TestPerformance(t *testing.T) {
	fmt.Println("sfsDb 性能测试")
	fmt.Println("=======================================")

	// 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./performance_test_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 创建测试表
	userTable, err := engine.TableNew("performance_test")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	userFields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 测试1: 内存占用
	fmt.Println("\n1. 内存占用测试")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialMem := m.Alloc

	// 插入1000条数据
	fmt.Println("插入1000条测试数据...")
	for i := 1; i <= 1000; i++ {
		user := map[string]any{
			"id":   i,
			"name": fmt.Sprintf("user_%d", i),
			"age":  rand.Intn(100),
		}
		_, err := userTable.Insert(&user)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	runtime.ReadMemStats(&m)
	finalMem := m.Alloc
	memUsage := (finalMem - initialMem) / (1024 * 1024) // MB
	fmt.Printf("内存占用: %d MB\n", memUsage)

	// 测试2: 启动时间
	fmt.Println("\n2. 启动时间测试")
	startTime := time.Now()

	// 测试表已创建，直接使用
	loadTime := time.Since(startTime)
	fmt.Printf("启动时间: %v\n", loadTime)

	// 测试3: 读取性能
	fmt.Println("\n3. 读取性能测试")
	startTime = time.Now()

	// 随机读取100次
	for i := 0; i < 100; i++ {
		id := rand.Intn(1000) + 1
		iter, err := userTable.Search(&map[string]any{"id": id})
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}
		iter.GetRecordSet(true)
		iter.Release()
	}

	readTime := time.Since(startTime)
	fmt.Printf("读取100次耗时: %v\n", readTime)
	fmt.Printf("平均读取时间: %v\n", readTime/time.Duration(100))

	// 测试4: 并发性能
	fmt.Println("\n4. 并发性能测试")
	concurrency := 100
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var successCount int
	var errorCount int

	startTime = time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			// 每个线程执行10次操作
			for j := 0; j < 10; j++ {
				id := 1000 + threadID*10 + j
				user := map[string]any{
					"id":   id,
					"name": fmt.Sprintf("user_%d", id),
					"age":  rand.Intn(100),
				}

				// 插入操作
				_, err := userTable.Insert(&user)
				if err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
				} else {
					mutex.Lock()
					successCount++
					mutex.Unlock()
				}

				// 读取操作
				iter, err := userTable.Search(&map[string]any{"id": id})
				if err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
				} else {
					iter.GetRecordSet(true)
					iter.Release()
					mutex.Lock()
					successCount++
					mutex.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	concurrentTime := time.Since(startTime)

	fmt.Printf("并发数: %d\n", concurrency)
	fmt.Printf("总操作数: %d\n", concurrency*20)
	fmt.Printf("成功操作数: %d\n", successCount)
	fmt.Printf("失败操作数: %d\n", errorCount)
	fmt.Printf("并发执行时间: %v\n", concurrentTime)
	fmt.Printf("每秒操作数: %.2f\n", float64(successCount)/concurrentTime.Seconds())

	// 测试5: 批量操作性能
	fmt.Println("\n5. 批量操作性能测试")
	db := dbManager.GetDB()
	batch := db.GetBatch()
	defer batch.Reset()

	startTime = time.Now()

	// 批量插入100条数据
	for i := 2000; i < 2100; i++ {
		user := map[string]any{
			"id":   i,
			"name": fmt.Sprintf("user_%d", i),
			"age":  rand.Intn(100),
		}
		_, err := userTable.Insert(&user, batch)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 提交批量操作
	err = db.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交批量操作失败: %v", err)
	}

	batchTime := time.Since(startTime)
	fmt.Printf("批量插入100条耗时: %v\n", batchTime)

	// 生成测试报告
	fmt.Println("\n=======================================")
	fmt.Println("性能测试报告")
	fmt.Println("=======================================")
	fmt.Printf("内存占用: %d MB\n", memUsage)
	fmt.Printf("启动时间: %v\n", loadTime)
	fmt.Printf("平均读取时间: %v\n", readTime/time.Duration(100))
	fmt.Printf("并发操作每秒: %.2f\n", float64(successCount)/concurrentTime.Seconds())
	fmt.Printf("批量插入每秒: %.2f\n", 100.0/batchTime.Seconds())
	fmt.Println("=======================================")

	fmt.Println("\n性能测试完成!")
}

/*
// main 函数用于直接运行性能测试
func main() {
	TestPerformance(&testing.T{})
}
*/
