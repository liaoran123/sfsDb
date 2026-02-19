package engine

import (
	"fmt"
	"os"
	"testing"
)

// TestOpenTable_Comprehensive 测试OpenTable函数的全面功能
func TestOpenTable_Comprehensive(t *testing.T) {
	// 清理测试环境
	cleanupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	// 测试用例
	testCases := []struct {
		name        string
		tableName   string
		expectedErr bool
		setup       func() error
	}{{
		name:        "打开已存在的表",
		tableName:   "test_table",
		expectedErr: false,
		setup: func() error {
			// 创建一个测试表
			table, err := TableNew("test_table")
			if err != nil {
				return err
			}
			// 设置字段
			fields := map[string]any{
				"id":    0,
				"name":  "",
				"age":   0,
				"email": "",
			}
			return table.SetFields(fields)
		},
	}, {
		name:        "打开不存在的表",
		tableName:   "non_existent_table",
		expectedErr: false, // 目前实现中，不存在的表也会被打开
		setup:       func() error { return nil },
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行测试前的设置
			if tc.setup != nil {
				if err := tc.setup(); err != nil {
					t.Fatalf("测试设置失败: %v", err)
				}
			}

			// 创建表实例并打开表
			table := &Table{}
			err := table.OpenTable(tc.tableName)

			// 检查错误
			if tc.expectedErr && err == nil {
				t.Errorf("预期会出错，但没有错误")
			} else if !tc.expectedErr && err != nil {
				t.Errorf("预期不会出错，但出错了: %v", err)
			}

			// 验证表的属性
			if err == nil {
				if table.name != tc.tableName {
					t.Errorf("表名设置错误，预期: %s, 实际: %s", tc.tableName, table.name)
				}
				if table.fields == nil {
					t.Error("字段映射未初始化")
				}
				if table.fieldsid == nil {
					t.Error("字段ID映射未初始化")
				}
				if table.kvStore == nil {
					t.Error("存储实例未初始化")
				}
				if table.indexs == nil {
					t.Error("索引集合未初始化")
				}
				if table.timeFields == nil {
					t.Error("时间字段映射未初始化")
				}
				if table.deadlockDetector == nil {
					t.Error("死锁检测器未初始化")
				}
				t.Logf("成功打开表: %s, 表ID: %d", table.name, table.GetId())
			}
		})
	}
}

// TestOpenTable_Concurrency 测试并发打开同一个表
func TestOpenTable_Concurrency(t *testing.T) {
	// 清理测试环境
	cleanupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	// 创建一个测试表
	table, err := TableNew("concurrency_test")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 并发打开表
	const concurrency = 10
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			table := &Table{}
			errCh <- table.OpenTable("concurrency_test")
		}()
	}

	// 收集错误
	errors := []error{}
	for i := 0; i < concurrency; i++ {
		if err := <-errCh; err != nil {
			errors = append(errors, err)
		}
	}

	// 检查是否有错误
	if len(errors) > 0 {
		t.Errorf("并发打开表时出现错误: %v", errors)
	} else {
		t.Logf("并发打开表成功，共打开 %d 次", concurrency)
	}
}

// TestOpenTable_MultipleTables 测试打开多个不同的表
func TestOpenTable_MultipleTables(t *testing.T) {
	// 清理测试环境
	cleanupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	// 创建多个测试表
	tableNames := []string{"table1", "table2", "table3"}
	for _, tableName := range tableNames {
		table, err := TableNew(tableName)
		if err != nil {
			t.Fatalf("创建测试表 %s 失败: %v", tableName, err)
		}
		// 设置字段
		fields := map[string]any{
			"id":    0,
			"name":  "",
		}
		if err := table.SetFields(fields); err != nil {
			t.Fatalf("设置字段失败: %v", err)
		}
	}

	// 逐个打开表并验证
	for _, tableName := range tableNames {
		t.Run(fmt.Sprintf("打开表 %s", tableName), func(t *testing.T) {
			table := &Table{}
			err := table.OpenTable(tableName)
			if err != nil {
				t.Errorf("打开表 %s 失败: %v", tableName, err)
				return
			}
			if table.name != tableName {
				t.Errorf("表名不匹配，预期: %s, 实际: %s", tableName, table.name)
			}
			t.Logf("成功打开表: %s, 表ID: %d", table.name, table.GetId())
		})
	}
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment(t *testing.T) {
	t.Helper()

	// 删除测试数据库目录
	dbPath := "./kvdb"
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.RemoveAll(dbPath); err != nil {
			t.Logf("清理测试环境失败: %v", err)
		}
	}

	// 清理表ID管理器
	TableIDManager = nil
}
