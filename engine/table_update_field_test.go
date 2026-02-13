package engine

import (
	"testing"
)

// TestUpdateFieldNameWithVersionField 测试修改版本号字段"v"的行为
func TestUpdateFieldNameWithVersionField(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_update_field_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 验证"v"字段是否存在
	vValue, exists := table.GetField("v")
	if !exists {
		t.Error("Version field 'v' should exist after SetFields")
	}
	t.Logf("Version field 'v' exists with value: %v", vValue)

	// 尝试修改一个普通字段，应该成功
	err = table.UpdateFieldName("name", "username")
	if err != nil {
		t.Errorf("Failed to update field name 'name' to 'username': %v", err)
	}
	t.Log("Successfully updated field name 'name' to 'username'")

	// 尝试修改"v"字段，应该失败或被忽略
	err = table.UpdateFieldName("v", "version")
	if err == nil {
		t.Log("Warning: No error when trying to update version field 'v' to 'version'")
	} else {
		t.Logf("Got expected error when trying to update version field: %v", err)
	}

	// 验证"v"字段是否仍然存在
	vValue, exists = table.GetField("v")
	if !exists {
		t.Error("Version field 'v' should still exist after attempting to rename it")
	}
	t.Logf("Version field 'v' still exists with value: %v", vValue)

	// 验证"version"字段是否不存在
	versionValue, exists := table.GetField("version")
	if exists {
		t.Error("Field 'version' should not exist after attempting to rename 'v' to 'version'")
	}
	t.Logf("Field 'version' correctly does not exist: %v, %v", versionValue, exists)

	// 验证其他字段是否正常
	usernameValue, exists := table.GetField("username")
	if !exists {
		t.Error("Field 'username' should exist after renaming from 'name'")
	}
	t.Logf("Field 'username' exists with value: %v", usernameValue)

	ageValue, exists := table.GetField("age")
	if !exists {
		t.Error("Field 'age' should still exist")
	}
	t.Logf("Field 'age' exists with value: %v", ageValue)
}

// TestUpdateFieldNameNormalOperation 测试正常的字段重命名操作
func TestUpdateFieldNameNormalOperation(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_normal_update_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":      0,
		"product": "",
		"price":   0,
		"stock":   0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 验证所有字段是否存在
	fieldsToCheck := []string{"id", "product", "price", "stock", "v"}
	for _, field := range fieldsToCheck {
		_, exists := table.GetField(field)
		if !exists {
			t.Errorf("Field '%s' should exist after SetFields", field)
		}
		t.Logf("Field '%s' exists", field)
	}

	// 执行多个字段重命名操作
	renameOperations := []struct {
		oldName string
		newName string
	}{
		{"product", "item"},
		{"price", "cost"},
		{"stock", "inventory"},
	}

	for _, op := range renameOperations {
		err = table.UpdateFieldName(op.oldName, op.newName)
		if err != nil {
			t.Errorf("Failed to update field name '%s' to '%s': %v", op.oldName, op.newName, err)
		}
		t.Logf("Successfully updated field name '%s' to '%s'", op.oldName, op.newName)

		// 验证新字段名存在
		_, exists := table.GetField(op.newName)
		if !exists {
			t.Errorf("Field '%s' should exist after rename", op.newName)
		}

		// 验证旧字段名不存在
		_, exists = table.GetField(op.oldName)
		if exists {
			t.Errorf("Field '%s' should not exist after rename", op.oldName)
		}
	}

	// 验证"v"字段仍然存在
	_, exists := table.GetField("v")
	if !exists {
		t.Error("Version field 'v' should still exist after multiple field renames")
	}
	t.Log("Version field 'v' still exists after multiple field renames")
}
