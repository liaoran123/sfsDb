package record

import (
	"testing"
)

// TestRecordSelect 测试单个 Record 的 Select 方法
func TestRecordSelect(t *testing.T) {
	// 创建测试记录
	testRecord := Record{
		"id":     1,
		"name":   "张三",
		"age":    30,
		"email":  "zhangsan@example.com",
		"active": true,
	}

	// 测试用例1：选择部分字段
	t.Run("SelectPartialFields", func(t *testing.T) {
		result := testRecord.Select("name", "age")
		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}
		if result["name"] != "张三" {
			t.Errorf("Expected name '张三', got '%v'", result["name"])
		}
		if result["age"] != 30 {
			t.Errorf("Expected age 30, got %v", result["age"])
		}
		if _, ok := result["email"]; ok {
			t.Error("Expected 'email' not to be in result")
		}
	})

	// 测试用例2：选择所有字段（不传参数）
	t.Run("SelectAllFields", func(t *testing.T) {
		result := testRecord.Select()
		if len(result) != 5 {
			t.Errorf("Expected all 5 fields, got %d", len(result))
		}
		if !t.Failed() {
			t.Log("Select() with no arguments returns all fields")
		}
	})

	// 测试用例3：选择单个字段
	t.Run("SelectSingleField", func(t *testing.T) {
		result := testRecord.Select("email")
		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}
		if result["email"] != "zhangsan@example.com" {
			t.Errorf("Expected email 'zhangsan@example.com', got '%v'", result["email"])
		}
	})

	// 测试用例4：选择不存在的字段
	t.Run("SelectNonExistentField", func(t *testing.T) {
		result := testRecord.Select("nonexistent")
		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}
		if result["nonexistent"] != nil {
			t.Errorf("Expected nonexistent field to be nil, got %v", result["nonexistent"])
		}
	})

	// 测试用例5：nil Record
	t.Run("SelectFromNilRecord", func(t *testing.T) {
		var nilRecord Record
		result := nilRecord.Select("name")
		if result != nil {
			t.Errorf("Expected nil result from nil record, got %v", result)
		}
	})
}

// TestRecordsSelect 测试 Records 的 Select 方法
func TestRecordsSelect(t *testing.T) {
	// 创建测试记录集合
	testRecords := Records{
		{
			"id":     1,
			"name":   "张三",
			"age":    30,
			"email":  "zhangsan@example.com",
			"active": true,
		},
		{
			"id":     2,
			"name":   "李四",
			"age":    25,
			"email":  "lisi@example.com",
			"active": false,
		},
	}

	// 测试用例1：选择部分字段
	t.Run("SelectPartialFields", func(t *testing.T) {
		result := testRecords.Select("name", "age")
		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}
		if len(result[0]) != 2 {
			t.Errorf("Expected 2 fields in first record, got %d", len(result[0]))
		}
		if result[0]["name"] != "张三" || result[1]["name"] != "李四" {
			t.Error("Expected correct names")
		}
		if result[0]["age"] != 30 || result[1]["age"] != 25 {
			t.Error("Expected correct ages")
		}
	})

	// 测试用例2：选择所有字段（不传参数）
	t.Run("SelectAllFields", func(t *testing.T) {
		result := testRecords.Select()
		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}
		if len(result[0]) != 5 {
			t.Errorf("Expected all 5 fields in first record, got %d", len(result[0]))
		}
	})

	// 测试用例3：空 Records
	t.Run("SelectFromEmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		result := emptyRecords.Select("name")
		if len(result) != 0 {
			t.Errorf("Expected empty result from empty records, got %d records", len(result))
		}
	})
}

// TestRecordsContains 测试 Records 的 Contains 方法
func TestRecordsContains(t *testing.T) {
	// 创建测试记录集合
	testRecords := Records{
		{
			"id":     1,
			"name":   "张三",
			"age":    30,
		},
		{
			"id":     2,
			"name":   "李四",
			"age":    25,
		},
	}

	// 测试用例1：包含指定记录
	t.Run("ContainsExistingRecord", func(t *testing.T) {
		targetRecord := Record{"id": 1, "name": "张三", "age": 30}
		if !testRecords.Contains(targetRecord) {
			t.Error("Expected records to contain target record")
		}
	})

	// 测试用例2：不包含指定记录
	t.Run("DoesNotContainRecord", func(t *testing.T) {
		targetRecord := Record{"id": 3, "name": "王五", "age": 35}
		if testRecords.Contains(targetRecord) {
			t.Error("Expected records not to contain target record")
		}
	})

	// 测试用例3：包含部分字段的记录（不匹配）
	t.Run("DoesNotContainPartialRecord", func(t *testing.T) {
		targetRecord := Record{"name": "张三"} // 缺少id和age
		if testRecords.Contains(targetRecord) {
			t.Error("Expected records not to contain partial record")
		}
	})

	// 测试用例4：空 Records
	t.Run("ContainsFromEmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		targetRecord := Record{"id": 1}
		if emptyRecords.Contains(targetRecord) {
			t.Error("Expected empty records not to contain any record")
		}
	})
}

// TestRecordsIntersect 测试 Records 的 Intersect 方法
func TestRecordsIntersect(t *testing.T) {
	// 创建测试记录集合
	records1 := Records{
		{"id": 1, "name": "张三"},
		{"id": 2, "name": "李四"},
		{"id": 3, "name": "王五"},
	}

	records2 := Records{
		{"id": 2, "name": "李四"},
		{"id": 3, "name": "王五"},
		{"id": 4, "name": "赵六"},
	}

	// 测试用例1：正常交集
	t.Run("NormalIntersection", func(t *testing.T) {
		result := records1.Intersect(records2)
		if len(result) != 2 {
			t.Errorf("Expected 2 records in intersection, got %d", len(result))
		}
		// 验证结果包含预期记录
		expectedRecords := Records{
			{"id": 2, "name": "李四"},
			{"id": 3, "name": "王五"},
		}
		for _, expected := range expectedRecords {
			if !result.Contains(expected) {
				t.Errorf("Expected intersection to contain %v, but it didn't", expected)
			}
		}
	})

	// 测试用例2：空交集
	t.Run("EmptyIntersection", func(t *testing.T) {
		records3 := Records{{"id": 5, "name": "孙七"}}
		result := records1.Intersect(records3)
		if len(result) != 0 {
			t.Errorf("Expected empty intersection, got %d records", len(result))
		}
	})

	// 测试用例3：空 Records
	t.Run("IntersectWithEmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		result1 := records1.Intersect(emptyRecords)
		result2 := emptyRecords.Intersect(records1)
		if len(result1) != 0 || len(result2) != 0 {
			t.Error("Expected empty result when intersecting with empty records")
		}
	})
}

// TestRecordsUnion 测试 Records 的 Union 方法
func TestRecordsUnion(t *testing.T) {
	// 创建测试记录集合
	records1 := Records{
		{"id": 1, "name": "张三"},
		{"id": 2, "name": "李四"},
	}

	records2 := Records{
		{"id": 2, "name": "李四"}, // 重复记录
		{"id": 3, "name": "王五"},
	}

	// 测试用例1：正常并集
	t.Run("NormalUnion", func(t *testing.T) {
		result := records1.Union(records2)
		if len(result) != 3 {
			t.Errorf("Expected 3 records in union, got %d", len(result))
		}
		// 验证结果包含所有唯一记录
		expectedRecords := Records{
			{"id": 1, "name": "张三"},
			{"id": 2, "name": "李四"},
			{"id": 3, "name": "王五"},
		}
		if len(result) != len(expectedRecords) {
			t.Errorf("Expected %d records in union, got %d", len(expectedRecords), len(result))
		}
		for _, expected := range expectedRecords {
			if !result.Contains(expected) {
				t.Errorf("Expected union to contain %v, but it didn't", expected)
			}
		}
	})

	// 测试用例2：空 Records
	t.Run("UnionWithEmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		result1 := records1.Union(emptyRecords)
		result2 := emptyRecords.Union(records1)
		if len(result1) != len(records1) || len(result2) != len(records1) {
			t.Error("Expected union with empty records to return original records")
		}
	})

	// 测试用例3：多个集合的并集
	t.Run("UnionMultipleSets", func(t *testing.T) {
		records3 := Records{{"id": 4, "name": "赵六"}}
		result := records1.Union(records2, records3)
		if len(result) != 4 {
			t.Errorf("Expected 4 records in union of 3 sets, got %d", len(result))
		}
	})
}

// TestRecordsDifference 测试 Records 的 Difference 方法
func TestRecordsDifference(t *testing.T) {
	// 创建测试记录集合
	records1 := Records{
		{"id": 1, "name": "张三"},
		{"id": 2, "name": "李四"},
		{"id": 3, "name": "王五"},
	}

	records2 := Records{
		{"id": 2, "name": "李四"},
		{"id": 4, "name": "赵六"},
	}

	// 测试用例1：正常差集
	t.Run("NormalDifference", func(t *testing.T) {
		result := records1.Difference(records2)
		if len(result) != 2 {
			t.Errorf("Expected 2 records in difference, got %d", len(result))
		}
		// 验证结果包含预期记录
		expectedRecords := Records{
			{"id": 1, "name": "张三"},
			{"id": 3, "name": "王五"},
		}
		for _, expected := range expectedRecords {
			if !result.Contains(expected) {
				t.Errorf("Expected difference to contain %v, but it didn't", expected)
			}
		}
		// 验证结果不包含被减去的记录
		if result.Contains(records2[0]) {
			t.Error("Expected difference not to contain record from the subtracted set")
		}
	})

	// 测试用例2：空差集
	t.Run("EmptyDifference", func(t *testing.T) {
		result := records1.Difference(records1) // 自己减自己
		if len(result) != 0 {
			t.Errorf("Expected empty difference when subtracting set from itself, got %d records", len(result))
		}
	})

	// 测试用例3：空 Records
	t.Run("DifferenceFromEmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		result1 := records1.Difference(emptyRecords)
		result2 := emptyRecords.Difference(records1)
		if len(result1) != len(records1) {
			t.Error("Expected difference from empty records to return original records")
		}
		if len(result2) != 0 {
			t.Error("Expected empty records difference to return empty")
		}
	})

	// 测试用例4：多个集合的差集
	t.Run("DifferenceMultipleSets", func(t *testing.T) {
		records3 := Records{{"id": 3, "name": "王五"}}
		result := records1.Difference(records2, records3)
		if len(result) != 1 {
			t.Errorf("Expected 1 record in difference of 3 sets, got %d", len(result))
		}
		if !result.Contains(Record{"id": 1, "name": "张三"}) {
			t.Error("Expected difference to contain only record 1")
		}
	})
}

// TestRecordsOperation 测试 Records 的 Operation 方法
func TestRecordsOperation(t *testing.T) {
	// 创建测试记录集合
	testRecords := Records{
		{
			"id":     1,
			"name":   "张三",
			"age":    30,
			"salary": 5000.0,
			"bonus":  1000.0,
			"str1":   "hello",
			"str2":   "world",
		},
		{
			"id":     2,
			"name":   "李四",
			"age":    25,
			"salary": 4000.0,
			"bonus":  800.0,
			"str1":   "good",
			"str2":   "morning",
		},
	}

	// 测试用例1：空 Records
	t.Run("EmptyRecords", func(t *testing.T) {
		emptyRecords := Records{}
		result := emptyRecords.Operation()
		if result != nil {
			t.Errorf("Expected nil result from empty records, got %v", result)
		}
	})

	// 测试用例2：没有提供 Operation
	t.Run("NoOperation", func(t *testing.T) {
		result := testRecords.Operation()
		if len(result) != len(testRecords) {
			t.Errorf("Expected same number of records, got %d vs %d", len(result), len(testRecords))
		}
		// 验证结果是原记录的副本
		for i, record := range result {
			if len(record) != len(testRecords[i]) {
				t.Errorf("Expected same number of fields in record %d, got %d vs %d", i, len(record), len(testRecords[i]))
			}
		}
	})

	// 测试用例3：单个加法 Operation
	t.Run("SingleAddOperation", func(t *testing.T) {
		// 创建加法运算：salary + bonus
		op := NewAddOperation([]string{"salary", "bonus"}, "total_income")
		result := testRecords.Operation(op)
		
		// 验证结果记录数相同
		if len(result) != len(testRecords) {
			t.Errorf("Expected same number of records, got %d vs %d", len(result), len(testRecords))
		}
		
		// 验证每个记录都添加了新字段
		for i, record := range result {
			if _, ok := record["total_income"]; !ok {
				t.Errorf("Expected record %d to have 'total_income' field", i)
			}
			
			// 验证计算结果正确
			expected := testRecords[i]["salary"].(float64) + testRecords[i]["bonus"].(float64)
			if record["total_income"].(float64) != expected {
				t.Errorf("Expected total_income %f, got %f for record %d", expected, record["total_income"].(float64), i)
			}
		}
	})

	// 测试用例4：单个字符串连接 Operation
	t.Run("SingleConcatOperation", func(t *testing.T) {
		// 创建字符串连接运算：str1 + str2
		op := NewConcatOperation([]string{"str1", "str2"}, "combined_str", " ")
		result := testRecords.Operation(op)
		
		// 验证结果记录数相同
		if len(result) != len(testRecords) {
			t.Errorf("Expected same number of records, got %d vs %d", len(result), len(testRecords))
		}
		
		// 验证每个记录都添加了新字段
		for i, record := range result {
			if _, ok := record["combined_str"]; !ok {
				t.Errorf("Expected record %d to have 'combined_str' field", i)
			}
			
			// 验证连接结果正确
			expected := testRecords[i]["str1"].(string) + " " + testRecords[i]["str2"].(string)
			if record["combined_str"].(string) != expected {
				t.Errorf("Expected combined_str '%s', got '%s' for record %d", expected, record["combined_str"].(string), i)
			}
		}
	})

	// 测试用例5：多个 Operation（只处理第一个）
	t.Run("MultipleOperations", func(t *testing.T) {
		// 创建多个运算
		op1 := NewAddOperation([]string{"salary", "bonus"}, "total_income")
		op2 := NewSubOperation([]string{"salary", "bonus"}, "net_salary")
		
		// 只处理第一个 Operation
		result := testRecords.Operation(op1, op2)
		
		// 验证结果记录数相同
		if len(result) != len(testRecords) {
			t.Errorf("Expected same number of records, got %d vs %d", len(result), len(testRecords))
		}
		
		// 验证只有第一个运算结果被添加
		for i, record := range result {
			if _, ok := record["total_income"]; !ok {
				t.Errorf("Expected record %d to have 'total_income' field", i)
			}
			if _, ok := record["net_salary"]; ok {
				t.Errorf("Expected record %d not to have 'net_salary' field (only first operation should be processed)", i)
			}
		}
	})

	// 测试用例6：减法 Operation
	t.Run("SingleSubOperation", func(t *testing.T) {
		// 创建减法运算：salary - bonus
		op := NewSubOperation([]string{"salary", "bonus"}, "net_salary")
		result := testRecords.Operation(op)
		
		// 验证结果记录数相同
		if len(result) != len(testRecords) {
			t.Errorf("Expected same number of records, got %d vs %d", len(result), len(testRecords))
		}
		
		// 验证每个记录都添加了新字段
		for i, record := range result {
			if _, ok := record["net_salary"]; !ok {
				t.Errorf("Expected record %d to have 'net_salary' field", i)
			}
			
			// 验证计算结果正确
			expected := testRecords[i]["salary"].(float64) - testRecords[i]["bonus"].(float64)
			if record["net_salary"].(float64) != expected {
				t.Errorf("Expected net_salary %f, got %f for record %d", expected, record["net_salary"].(float64), i)
			}
		}
	})

	// 测试用例7：重复字段名检测
	t.Run("DuplicateFieldDetection", func(t *testing.T) {
		// 创建运算，故意使用已存在的字段名
		op := NewAddOperation([]string{"salary", "bonus"}, "salary")
		
		// 验证会panic
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for duplicate field name, but no panic occurred")
			}
		}()
		
		// 执行运算，应该panic
		testRecords.Operation(op)
	})
}
