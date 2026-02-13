package engine

import (
	"sync"
	"testing"
)

// TestTableCacheConcurrentAccess 测试表缓存的并发访问
func TestTableCacheConcurrentAccess(t *testing.T) {
	// 创建一个测试表
	table := &Table{
		name: "test_table",
		fields: map[string]any{
			"id":   "int",
			"name": "string",
			"age":  "int",
		},
		id: 1,
	}

	// 初始化索引
	table.indexs = NewIndexs(&table.fields)



	// 并发测试参数
	const goroutineCount = 100
	const operationCount = 100

	// 等待组
	var wg sync.WaitGroup

	// 测试索引匹配缓存并发访问
	t.Run("MatchIndexCachedConcurrent", func(t *testing.T) {
		wg.Add(goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationCount; j++ {
					// 生成不同的字段组合
					fields := []string{"id"}
					if j%2 == 0 {
						fields = append(fields, "name")
					}
					
					// 调用缓存函数
					table.MatchIndex(fields...)
				}
			}(i)
		}
		wg.Wait()
	})

	// 测试字段转换缓存并发访问
	t.Run("FieldsToBytesNilCachedConcurrent", func(t *testing.T) {
		wg.Add(goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationCount; j++ {
					// 生成不同的字段值
					fields := &map[string]any{
						"id":   goroutineID*operationCount + j,
						"name": "test",
						"age":  j,
					}
					
					// 调用缓存函数
					table.FieldsToBytesNil(fields)
				}
			}(i)
		}
		wg.Wait()
	})

	// 测试混合并发访问
	t.Run("MixedConcurrentAccess", func(t *testing.T) {
		wg.Add(goroutineCount * 2)
		for i := 0; i < goroutineCount; i++ {
			// 测试索引匹配
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationCount; j++ {
					fields := []string{"id"}
					if j%2 == 0 {
						fields = append(fields, "name")
					}
					table.MatchIndex(fields...)
				}
			}(i)

			// 测试字段转换
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationCount; j++ {
					fields := &map[string]any{
						"id":   goroutineID*operationCount + j,
						"name": "test",
						"age":  j,
					}
					table.FieldsToBytesNil(fields)
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestCacheClearConcurrent 测试缓存清除时的并发访问
func TestCacheClearConcurrent(t *testing.T) {
	// 创建一个测试表
	table := &Table{
		name: "test_table",
		fields: map[string]any{
			"id":   "int",
			"name": "string",
		},
		id: 1,
	}

	// 初始化索引
	table.indexs = NewIndexs(&table.fields)



	// 并发测试参数
	const goroutineCount = 50
	const operationCount = 50

	// 等待组
	var wg sync.WaitGroup

	// 启动多个 goroutine 进行缓存操作
	wg.Add(goroutineCount + 1)
	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationCount; j++ {
				// 交替调用两个缓存函数
				if j%2 == 0 {
					// 测试索引匹配
					fields := []string{"id", "name"}
					table.MatchIndex(fields...)
				} else {
					// 测试字段转换
					fields := &map[string]any{
						"id":   goroutineID*operationCount + j,
						"name": "test",
					}
					table.FieldsToBytesNil(fields)
				}
			}
		}(i)
	}

	// 启动一个 goroutine 定期清除缓存
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			// 清除所有缓存
			ClearAllCaches()
		}
	}()

	wg.Wait()
}
