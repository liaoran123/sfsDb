package engine

import (
	"testing"
)

// ---------------------- 非ACID事务基准测试 ----------------------

// BenchmarkTableNonACIDInsert 测试非ACID模式下的单次插入性能
func BenchmarkTableNonACIDInsert(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_nonacid_insert")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 创建测试数据
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}

		// 插入数据（非ACID模式）
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}
	}
}

// BenchmarkTableNonACIDBatchInsert 测试非ACID模式下的批量插入性能
func BenchmarkTableNonACIDBatchInsert(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_nonacid_batch_insert")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 创建批量操作
		batch := table.kvStore.GetBatch()

		// 插入多条数据
		for j := 0; j < 10; j++ {
			// 创建测试数据
			testData := map[string]any{
				"id":      i*10 + j + 1,
				"name":    "User" + string(rune('A'+(i*10+j)%26)),
				"age":     20 + (i*10+j)%30,
				"email":   "user" + string(rune('A'+(i*10+j)%26)) + "@example.com",
				"address": "Address " + string(rune('A'+(i*10+j)%26)),
			}

			// 插入数据（非ACID模式）
			_, err := table.Insert(&testData, batch)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}

		// 提交批量操作
		err = table.kvStore.WriteBatch(batch)
		if err != nil {
			b.Fatalf("Failed to write batch: %v", err)
		}
	}
}

// BenchmarkTableNonACIDSearch 测试非ACID模式下的搜索性能
func BenchmarkTableNonACIDSearch(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_nonacid_search")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 搜索条件
		searchData := map[string]any{"id": (i % 1000) + 1}

		// 搜索数据（非ACID模式）
		iter := table.Search(&searchData)
		if iter != nil {
			// 获取记录
			records := iter.GetRecords(true, 1)
			_ = records
			iter.Release()
		}
	}
}

// BenchmarkTableNonACIDUpdate 测试非ACID模式下的更新性能
func BenchmarkTableNonACIDUpdate(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_nonacid_update")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 更新条件和新值合并到一个map中
		updateData := map[string]any{"id": (i % 1000) + 1, "age": 30 + i%25}

		// 更新数据（非ACID模式）
		err := table.Update(&updateData)
		if err != nil {
			b.Fatalf("Failed to update data: %v", err)
		}
	}
}

// ---------------------- ACID事务基准测试 ----------------------

// BenchmarkTableACIDInsert 测试ACID模式下的单次插入性能
func BenchmarkTableACIDInsert(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_acid_insert")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 创建测试数据
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}

		// 在事务中插入数据
		_, err = tx.Insert(&testData)
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

// BenchmarkTableACIDBatchInsert 测试ACID模式下的批量插入性能
func BenchmarkTableACIDBatchInsert(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_acid_batch_insert")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入多条数据
		for j := 0; j < 10; j++ {
			// 创建测试数据
			testData := map[string]any{
				"id":      i*10 + j + 1,
				"name":    "User" + string(rune('A'+(i*10+j)%26)),
				"age":     20 + (i*10+j)%30,
				"email":   "user" + string(rune('A'+(i*10+j)%26)) + "@example.com",
				"address": "Address " + string(rune('A'+(i*10+j)%26)),
			}

			// 在事务中插入数据
			_, err := tx.Insert(&testData)
			if err != nil {
				tx.Rollback()
				b.Fatalf("Failed to insert data: %v", err)
			}
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTableACIDSearch 测试ACID模式下的搜索性能
func BenchmarkTableACIDSearch(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_acid_search")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_idx")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 搜索条件
		searchData := map[string]any{"id": (i % 1000) + 1}

		// 在事务中搜索数据
		iter := tx.Search(&searchData)
		if iter != nil {
			// 获取记录
			records := iter.GetRecords(true, 1)
			_ = records
			iter.Release()
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTableACIDUpdate 测试ACID模式下的更新性能
func BenchmarkTableACIDUpdate(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_acid_update")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 更新条件和新值合并到一个map中
		updateData := map[string]any{"id": (i % 1000) + 1, "age": 30 + i%25}

		// 在事务中更新数据
		err = tx.Update(&updateData)
		if err != nil {
			tx.Rollback()
			b.Fatalf("Failed to update data: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTableNonACIDDelete 测试非ACID模式下的删除性能
func BenchmarkTableNonACIDDelete(b *testing.B) {
	// 创建测试表
	table, err := TableNew("bench_nonacid_delete")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "address": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 创建测试数据
		testData := map[string]any{
			"id":      i + 1,
			"name":    "User" + string(rune('A'+i%26)),
			"age":     20 + i%30,
			"email":   "user" + string(rune('A'+i%26)) + "@example.com",
			"address": "Address " + string(rune('A'+i%26)),
		}

		// 插入数据（非ACID模式）
		id, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 删除数据（非ACID模式）
		deleteData := map[string]any{"id": id}
		err = table.Delete(&deleteData)
		if err != nil {
			b.Fatalf("Failed to delete data: %v", err)
		}
	}
}
