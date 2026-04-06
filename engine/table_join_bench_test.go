package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/match"
)

// TestTableJoinPerformance 测试多表连接查询性能并输出详细结果
func TestTableJoinPerformance(t *testing.T) {
	// 创建测试表
	table1, err := NewTable("test_table1")
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 设置表字段
	fields1 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	// 创建主键索引
	pk1, _ := NewDefaultPrimaryKey("pk1")
	pk1.AddFields("id")
	err = table1.CreateIndex(pk1)
	if err != nil {
		t.Fatalf("Failed to create primary key for table1: %v", err)
	}

	// 创建二级索引
	ageIdx1, _ := NewDefaultNormalIndex("age_index1")
	ageIdx1.AddFields("age")
	err = table1.CreateIndex(ageIdx1)
	if err != nil {
		t.Fatalf("Failed to create age index for table1: %v", err)
	}

	// 创建第二个测试表
	table2, err := TableNew("test_table2")
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 设置表字段
	fields2 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 创建主键索引
	pk2, _ := NewDefaultPrimaryKey("pk2")
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("Failed to create primary key for table2: %v", err)
	}

	// 创建二级索引
	ageIdx2, _ := NewDefaultNormalIndex("age_index2")
	ageIdx2.AddFields("age")
	err = table2.CreateIndex(ageIdx2)
	if err != nil {
		t.Fatalf("Failed to create age index for table2: %v", err)
	}

	// 测试不同规模的数据
	sizes := []int{100, 500, 1000}

	for _, size := range sizes {
		t.Logf("\n测试数据规模: %d 条记录", size)
		t.Logf("=====================================")

		// 清空表数据（简单起见，重新创建表）
		table1, _ = TableNew(fmt.Sprintf("test_table1_%d", size))
		err = table1.SetFields(fields1)
		if err != nil {
			t.Fatalf("Failed to set fields for table1: %v", err)
		}
		pk1, _ = NewDefaultPrimaryKey("pk1")
		pk1.AddFields("id")
		err = table1.CreateIndex(pk1)
		if err != nil {
			t.Fatalf("Failed to create primary key for table1: %v", err)
		}

		table2, _ = TableNew(fmt.Sprintf("test_table2_%d", size))
		err = table2.SetFields(fields2)
		if err != nil {
			t.Fatalf("Failed to set fields for table2: %v", err)
		}
		pk2, _ = NewDefaultPrimaryKey("pk2")
		pk2.AddFields("id")
		err = table2.CreateIndex(pk2)
		if err != nil {
			t.Fatalf("Failed to create primary key for table2: %v", err)
		}

		// 插入测试数据
		for i := 1; i <= size; i++ {
			// 插入 table1 数据
			record1 := map[string]any{
				"id":     i,
				"name":   "User" + string(rune('A'+i%26)),
				"age":    20 + i%30,
				"score":  60.0 + float64(i%40),
				"active": i%2 == 0,
			}
			_, err := table1.Insert(&record1)
			if err != nil {
				t.Fatalf("Failed to insert record1 %d: %v", i, err)
			}

			// 插入 table2 数据（与 table1 有部分重叠）
			if i%2 == 0 {
				record2 := map[string]any{
					"id":     i,
					"name":   "User" + string(rune('A'+i%26)),
					"age":    20 + i%30,
					"score":  60.0 + float64(i%40),
					"active": i%2 == 0,
				}
				_, err := table2.Insert(&record2)
				if err != nil {
					t.Fatalf("Failed to insert record2 %d: %v", i, err)
				}
			}
		}

		// 测试两表等值连接
		t.Logf("测试两表等值连接...")
		startTime := time.Now()

		// 获取表迭代器
		iter1, err := table1.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to create iter1: %v", err)
		}
		defer iter1.Release()

		iter2, err := table2.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to create iter2: %v", err)
		}
		defer iter2.Release()

		// 创建映射和匹配条件
		map2 := iter2.Map()
		defer iter2.ReleaseMap(map2)

		mach := match.NewAND([]string{"id"}, map2)
		iter1.SetMatch(mach)

		// 执行查询
		records := iter1.GetRecords(true)
		defer records.Release()

		// 计算执行时间
		duration := time.Since(startTime)
		t.Logf("查询完成，返回 %d 条记录，耗时: %v", len(records), duration)

		// 测试两表非等值连接
		t.Logf("测试两表非等值连接...")
		startTime = time.Now()

		// 创建非等值匹配条件
		machNotEqual := match.NewAND([]string{"id"}, map2, false)
		iter1.SetMatch(machNotEqual)

		// 执行查询
		recordsNotEqual := iter1.GetRecords(true)
		defer recordsNotEqual.Release()

		// 计算执行时间
		duration = time.Since(startTime)
		t.Logf("查询完成，返回 %d 条记录，耗时: %v", len(recordsNotEqual), duration)
	}

	t.Logf("\n性能测试完成！")
}

// BenchmarkTableJoinWithDifferentSizes 测试不同数据规模的连接性能
func BenchmarkTableJoinWithDifferentSizes(b *testing.B) {
	sizes := []int{100, 500, 1000, 2000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			// 创建测试表
			table1, err := TableNew("benchmark_table1_" + fmt.Sprintf("%d", size))
			if err != nil {
				b.Fatalf("Failed to create table1: %v", err)
			}

			// 设置表字段
			fields1 := map[string]any{"id": 0, "name": "", "age": 0}
			err = table1.SetFields(fields1)
			if err != nil {
				b.Fatalf("Failed to set fields for table1: %v", err)
			}

			// 创建主键索引
			pk1, _ := NewDefaultPrimaryKey("pk1")
			pk1.AddFields("id")
			err = table1.CreateIndex(pk1)
			if err != nil {
				b.Fatalf("Failed to create primary key for table1: %v", err)
			}

			// 创建第二个测试表
			table2, err := TableNew("benchmark_table2_" + fmt.Sprintf("%d", size))
			if err != nil {
				b.Fatalf("Failed to create table2: %v", err)
			}

			// 设置表字段
			fields2 := map[string]any{"id": 0, "name": "", "age": 0}
			err = table2.SetFields(fields2)
			if err != nil {
				b.Fatalf("Failed to set fields for table2: %v", err)
			}

			// 创建主键索引
			pk2, _ := NewDefaultPrimaryKey("pk2")
			pk2.AddFields("id")
			err = table2.CreateIndex(pk2)
			if err != nil {
				b.Fatalf("Failed to create primary key for table2: %v", err)
			}

			// 插入测试数据
			for i := 1; i <= size; i++ {
				// 插入 table1 数据
				record1 := map[string]any{
					"id":   i,
					"name": "User" + string(rune('A'+i%26)),
					"age":  20 + i%30,
				}
				_, err := table1.Insert(&record1)
				if err != nil {
					b.Fatalf("Failed to insert record1 %d: %v", i, err)
				}

				// 插入 table2 数据（与 table1 有部分重叠）
				if i%2 == 0 {
					record2 := map[string]any{
						"id":   i,
						"name": "User" + string(rune('A'+i%26)),
						"age":  20 + i%30,
					}
					_, err := table2.Insert(&record2)
					if err != nil {
						b.Fatalf("Failed to insert record2 %d: %v", i, err)
					}
				}
			}

			// 重置计时器
			b.ResetTimer()

			// 执行连接查询
			for i := 0; i < b.N; i++ {
				// 获取表迭代器
				iter1, err := table1.Search(&map[string]any{"id": nil})
				if err != nil {
					b.Fatalf("Failed to create iter1: %v", err)
				}

				iter2, err := table2.Search(&map[string]any{"id": nil})
				if err != nil {
					b.Fatalf("Failed to create iter2: %v", err)
				}

				// 创建映射和匹配条件
				map2 := iter2.Map()
				mach := match.NewAND([]string{"id"}, map2)
				iter1.SetMatch(mach)

				// 执行查询
				records := iter1.GetRecords(true)

				// 释放资源
				records.Release()
				iter1.Release()
				iter2.ReleaseMap(map2)
				iter2.Release()
			}
		})
	}
}
