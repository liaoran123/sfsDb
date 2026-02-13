package engine

import (
	"fmt"
	"testing"
)

// 基准测试：单操作事务性能
func BenchmarkTransactionSingleInsert(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_single_insert")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 准备插入数据
		insertData := map[string]any{
			"id":   i,
			"name": fmt.Sprintf("name_%d", i),
			"age":  i % 100,
		}

		// 执行插入
		_, err = tx.Insert(&insertData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：多操作事务性能
func BenchmarkTransactionMultiOperation(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_multi_operation")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 先插入一些测试数据
	for i := 0; i < 100; i++ {
		insertData := map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("name_%d", i),
			"age":   i % 100,
			"score": i * 10,
		}
		_, err = table.Insert(&insertData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 操作1: 插入新记录
		insertData := map[string]any{
			"id":    100 + i,
			"name":  fmt.Sprintf("new_name_%d", i),
			"age":   25,
			"score": 85,
		}
		_, err = tx.Insert(&insertData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 操作2: 更新现有记录
		updateData := map[string]any{
			"id":    i % 100,
			"score": (i%100)*10 + 5,
		}
		err = tx.Update(&updateData)
		if err != nil {
			b.Fatalf("Failed to update data: %v", err)
		}

		// 操作3: 读取记录
		readData := map[string]any{"id": i % 100}
		_, err = tx.Read(&readData)
		if err != nil {
			b.Fatalf("Failed to read data: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：嵌套事务性能
func BenchmarkTransactionNested(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_nested")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 准备事务选项，允许嵌套事务
	options := DefaultTransactionOptions()
	options.AllowNested = true

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 开始外层事务
		outerTx, err := table.BeginWithOptions(options)
		if err != nil {
			b.Fatalf("Failed to begin outer transaction: %v", err)
		}

		// 操作1: 外层事务插入
		outerInsert := map[string]any{
			"id":   i*10 + 1,
			"name": fmt.Sprintf("outer_%d", i),
		}
		_, err = outerTx.Insert(&outerInsert)
		if err != nil {
			b.Fatalf("Failed to insert in outer transaction: %v", err)
		}

		// 开始内层事务
		innerTx, err := outerTx.BeginNested()
		if err != nil {
			b.Fatalf("Failed to begin inner transaction: %v", err)
		}

		// 操作2: 内层事务插入
		innerInsert := map[string]any{
			"id":   i*10 + 2,
			"name": fmt.Sprintf("inner_%d", i),
		}
		_, err = innerTx.Insert(&innerInsert)
		if err != nil {
			b.Fatalf("Failed to insert in inner transaction: %v", err)
		}

		// 提交内层事务
		err = innerTx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit inner transaction: %v", err)
		}

		// 提交外层事务
		err = outerTx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit outer transaction: %v", err)
		}
	}
}

// 基准测试：不同隔离级别性能
func BenchmarkTransactionIsolationLevels(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_isolation")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "value": 0}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 测试不同隔离级别
	isolationLevels := []string{
		ReadUncommitted,
		ReadCommitted,
		RepeatableRead,
		Serializable,
	}

	for _, level := range isolationLevels {
		b.Run(level, func(b *testing.B) {
			// 准备事务选项
			options := DefaultTransactionOptions()
			options.IsolationLevel = level

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// 开始事务
				tx, err := table.BeginWithOptions(options)
				if err != nil {
					b.Fatalf("Failed to begin transaction: %v", err)
				}

				// 插入数据
				insertData := map[string]any{
					"id":    i,
					"name":  fmt.Sprintf("name_%d", i),
					"value": i * 10,
				}
				_, err = tx.Insert(&insertData)
				if err != nil {
					b.Fatalf("Failed to insert data: %v", err)
				}

				// 读取数据
				readData := map[string]any{"id": i}
				_, err = tx.Read(&readData)
				if err != nil {
					b.Fatalf("Failed to read data: %v", err)
				}

				// 提交事务
				err = tx.Commit()
				if err != nil {
					b.Fatalf("Failed to commit transaction: %v", err)
				}
			}
		})
	}
}

// 基准测试：并发事务性能
func BenchmarkTransactionConcurrent(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_concurrent")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "value": 0}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 测试不同并发度
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)

			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					// 生成唯一ID
					id := i
					i++

					// 开始事务
					tx, err := table.Begin()
					if err != nil {
						b.Fatalf("Failed to begin transaction: %v", err)
					}

					// 插入数据
					insertData := map[string]any{
						"id":    id,
						"name":  fmt.Sprintf("name_%d", id),
						"value": id * 10,
					}
					_, err = tx.Insert(&insertData)
					if err != nil {
						b.Fatalf("Failed to insert data: %v", err)
					}

					// 提交事务
					err = tx.Commit()
					if err != nil {
						b.Fatalf("Failed to commit transaction: %v", err)
					}
				}
			})
		})
	}
}

// 基准测试：事务性能与非事务性能对比
func BenchmarkTransactionVsNonTransaction(b *testing.B) {
	// 准备测试表
	table, err := TableNew("benchmark_tx_vs_non_tx")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 测试事务性能
	b.Run("WithTransaction", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				b.Fatalf("Failed to begin transaction: %v", err)
			}

			// 插入数据
			insertData := map[string]any{
				"id":   i,
				"name": fmt.Sprintf("name_%d", i),
				"age":  i % 100,
			}
			_, err = tx.Insert(&insertData)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				b.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	})

	// 测试非事务性能
	b.Run("WithoutTransaction", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// 直接插入数据
			insertData := map[string]any{
				"id":   i + 10000,
				"name": fmt.Sprintf("name_%d", i+10000),
				"age":  i % 100,
			}
			_, err := table.Insert(&insertData)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}
	})
}
