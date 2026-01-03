package engine

import (
	"testing"
)

// TestTableSchemaSerialization 测试TableSchema的序列化和反序列化功能
func TestTableSchemaSerialization(t *testing.T) {
	// 创建一个Table实例
	table, err := TableNew("test_table")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   1,
		"name": "test",
		"age":  25,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建索引
	primaryKey, err := DefaultPrimaryKeyNew("id")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.indexs.CreateIndex(primaryKey); err != nil {
		t.Fatalf("添加主键索引失败: %v", err)
	}

	normalIndex, err := DefaultNormalIndexNew("name_age_index")
	if err != nil {
		t.Fatalf("创建普通索引失败: %v", err)
	}
	normalIndex.AddFields("name", "age")
	if err := table.indexs.CreateIndex(normalIndex); err != nil {
		t.Fatalf("添加普通索引失败: %v", err)
	}

	fulltextIndex, err := DefaultFullTextIndexNew("name_fulltext")
	if err != nil {
		t.Fatalf("创建全文索引失败: %v", err)
	}
	fulltextIndex.AddFields("name")
	if err := table.indexs.CreateIndex(fulltextIndex); err != nil {
		t.Fatalf("添加全文索引失败: %v", err)
	}

	// 测试ToSchema方法
	schema := table.ToSchema()
	if schema.Name != "test_table" {
		t.Errorf("Schema名称错误，预期: test_table, 实际: %s", schema.Name)
	}

	if len(schema.Fields) != 3 {
		t.Errorf("Schema字段数量错误，预期: 3, 实际: %d", len(schema.Fields))
	}

	if len(schema.Indexes) != 3 {
		t.Errorf("Schema索引数量错误，预期: 3, 实际: %d", len(schema.Indexes))
	}

	// 测试FromSchema方法
	newTable, err := FromSchema(schema)
	if err != nil {
		t.Fatalf("从Schema创建表失败: %v", err)
	}

	if newTable.GetName() != "test_table" {
		t.Errorf("从Schema创建的表名称错误，预期: test_table, 实际: %s", newTable.GetName())
	}

	// 测试ToSerialization方法
	table.counter.Set(100)
	serialization := table.ToSerialization()
	if serialization.CurrentAutoID != 100 {
		t.Errorf("序列化的CurrentAutoID错误，预期: 100, 实际: %d", serialization.CurrentAutoID)
	}

	// 测试FromSerialization方法
	newTableFromSerialization, err := FromSerialization(serialization)
	if err != nil {
		t.Fatalf("从Serialization创建表失败: %v", err)
	}

	if newTableFromSerialization.GetName() != "test_table" {
		t.Errorf("从Serialization创建的表名称错误，预期: test_table, 实际: %s", newTableFromSerialization.GetName())
	}

	// 测试TableToJSON和TableFromJSON方法
	jsonStr, err := TableToJSON(table)
	if err != nil {
		t.Fatalf("TableToJSON失败: %v", err)
	}

	newTableFromJSON, err := TableFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("TableFromJSON失败: %v", err)
	}

	if newTableFromJSON.GetName() != "test_table" {
		t.Errorf("从JSON创建的表名称错误，预期: test_table, 实际: %s", newTableFromJSON.GetName())
	}
}

// TestTableSerializationFull 测试完整的Table序列化和反序列化流程
func TestTableSerializationFull(t *testing.T) {
	// 创建表并添加数据
	table, err := TableNew("test_serialization_full")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   1,
		"name": "test",
		"age":  25,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 手动创建主键索引，确保索引数量的一致性
	primaryKey, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.indexs.CreateIndex(primaryKey); err != nil {
		t.Fatalf("添加主键索引失败: %v", err)
	}

	// 获取初始索引数量
	initialIndexes := table.indexs.GetAllIndexes()
	initialIndexCount := len(initialIndexes)
	t.Logf("初始索引数量: %d", initialIndexCount)

	// 手动设置counter初始值，确保测试的可重复性
	initialAutoID := 100
	table.counter.Set(initialAutoID)
	t.Logf("初始AutoID: %d", initialAutoID)

	// 添加数据，不指定id字段，让系统自动生成
	insertCount := 5
	for i := 1; i <= insertCount; i++ {
		row := map[string]any{
			"name": "test_" + string(rune('a'+i)),
			"age":  20 + i,
		}
		_, err := table.Insert(&row)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 序列化表
	serialization := table.ToSerialization()

	// 验证CurrentAutoID是否正确递增
	expectedAutoID := initialAutoID + insertCount
	if serialization.CurrentAutoID != expectedAutoID {
		t.Errorf("序列化的CurrentAutoID错误，预期: %d, 实际: %d", expectedAutoID, serialization.CurrentAutoID)
	}
	t.Logf("序列化后的AutoID: %d", serialization.CurrentAutoID)

	// 验证序列化后的索引数量与初始数量一致
	serializedIndexCount := len(serialization.Schema.Indexes)
	if serializedIndexCount != initialIndexCount {
		t.Errorf("序列化后的Schema.Indexes数量错误，预期: %d, 实际: %d", initialIndexCount, serializedIndexCount)
	}
	t.Logf("序列化后的索引数量: %d", serializedIndexCount)

	// 反序列化表
	newTable, err := FromSerialization(serialization)
	if err != nil {
		t.Fatalf("反序列化表失败: %v", err)
	}

	// 验证表的基本信息
	if newTable.GetName() != "test_serialization_full" {
		t.Errorf("反序列化后的表名称错误，预期: test_serialization_full, 实际: %s", newTable.GetName())
	}

	// 验证反序列化后的索引数量与序列化前一致
	allIndexes := newTable.indexs.GetAllIndexes()
	if len(allIndexes) != serializedIndexCount {
		t.Errorf("反序列化后的索引数量错误，预期: %d, 实际: %d", serializedIndexCount, len(allIndexes))
	}
	t.Logf("反序列化后的索引数量: %d", len(allIndexes))

	// 验证反序列化后的AutoID是否正确
	if newTable.counter.Get() != expectedAutoID {
		t.Errorf("反序列化后的AutoID错误，预期: %d, 实际: %d", expectedAutoID, newTable.counter.Get())
	}
	t.Logf("反序列化后的AutoID: %d", newTable.counter.Get())
}
