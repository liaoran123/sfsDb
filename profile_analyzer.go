package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main11111() {
	// 创建性能分析文件
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 开始 CPU 性能分析
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// 运行基准测试
	b := &testing.B{}
	BenchmarkTableOperations(b)

	fmt.Println("基准测试完成，性能分析数据已保存到 cpu.prof")
}

// BenchmarkTableOperations 基准测试表操作性能
func BenchmarkTableOperations(b *testing.B) {
	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer storage.GetDBManager().CloseDB()

	// 创建表
	table, err := engine.NewTable("benchmark_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := engine.NewDefaultPrimaryKey("pk_id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 基准测试插入操作
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user := map[string]any{
			"name": "Test",
			"age":  30,
		}
		_, err := table.Insert(&user)
		if err != nil {
			b.Fatalf("Failed to insert: %v", err)
		}
	}
}
