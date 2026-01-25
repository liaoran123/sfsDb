package record

import (
	"testing"
)

// TestOperation 测试各种运算接口功能
func TestOperation(t *testing.T) {
	// 测试数据
	record := map[string]any{
		"field1": 10.0,
		"field2": 5.0,
		"field3": 3.0,
		"field4": 20.0,
		"str1":   "hello",
		"str2":   "world",
	}

	// 测试用例：加法运算
	t.Run("AddOperation", func(t *testing.T) {
		op := NewAddOperation([]string{"field1", "field2", "field3"}, "sum_result")
		result := op.Evaluate(record)
		expected := 18.0
		if result.(float64) != expected {
			t.Errorf("AddOperation failed: expected %f, got %f", expected, result.(float64))
		}
		if op.NewField() != "sum_result" {
			t.Errorf("AddOperation NewField failed: expected 'sum_result', got '%s'", op.NewField())
		}
	})

	// 测试用例：减法运算
	t.Run("SubOperation", func(t *testing.T) {
		op := NewSubOperation([]string{"field1", "field2"}, "sub_result")
		result := op.Evaluate(record)
		expected := 5.0
		if result.(float64) != expected {
			t.Errorf("SubOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：乘法运算
	t.Run("MulOperation", func(t *testing.T) {
		op := NewMulOperation([]string{"field1", "field2"}, "mul_result")
		result := op.Evaluate(record)
		expected := 50.0
		if result.(float64) != expected {
			t.Errorf("MulOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：除法运算
	t.Run("DivOperation", func(t *testing.T) {
		op := NewDivOperation([]string{"field1", "field2"}, "div_result", 1.0)
		result := op.Evaluate(record)
		expected := 2.0
		if result.(float64) != expected {
			t.Errorf("DivOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：平均值运算
	t.Run("AvgOperation", func(t *testing.T) {
		op := NewAvgOperation([]string{"field1", "field2", "field3"}, "avg_result")
		result := op.Evaluate(record)
		expected := 6.0
		if result.(float64) != expected {
			t.Errorf("AvgOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：求和运算（sum）
	t.Run("SumOperation", func(t *testing.T) {
		op := NewSumOperation([]string{"field1", "field2", "field3"}, "sum_result")
		result := op.Evaluate(record)
		expected := 18.0
		if result.(float64) != expected {
			t.Errorf("SumOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：最大值运算
	t.Run("MaxOperation", func(t *testing.T) {
		op := NewMaxOperation([]string{"field1", "field2", "field3", "field4"}, "max_result")
		result := op.Evaluate(record)
		expected := 20.0
		if result.(float64) != expected {
			t.Errorf("MaxOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：最小值运算
	t.Run("MinOperation", func(t *testing.T) {
		op := NewMinOperation([]string{"field1", "field2", "field3", "field4"}, "min_result")
		result := op.Evaluate(record)
		expected := 3.0
		if result.(float64) != expected {
			t.Errorf("MinOperation failed: expected %f, got %f", expected, result.(float64))
		}
	})

	// 测试用例：字符串连接运算
	t.Run("ConcatOperation", func(t *testing.T) {
		op := NewConcatOperation([]string{"str1", "str2"}, "concat_result", " ")
		result := op.Evaluate(record)
		expected := "hello world"
		if result.(string) != expected {
			t.Errorf("ConcatOperation failed: expected '%s', got '%s'", expected, result.(string))
		}
	})
}

// TestOperationIntegration 测试运算接口在实际记录中的集成使用
func TestOperationIntegration(t *testing.T) {
	// 原始记录
	record := map[string]any{
		"price":  100.0,
		"tax":    10.0,
		"discount": 5.0,
		"quantity": 2.0,
	}

	// 创建运算操作列表
	operations := []Operation{
		// 计算总价：price * quantity
		NewMulOperation([]string{"price", "quantity"}, "total_price"),
		// 计算实际支付：(price * quantity) + tax - discount
		NewSubOperation([]string{"total_price", "discount"}, "final_price"),
		// 计算平均单价：total_price / quantity
		NewDivOperation([]string{"total_price", "quantity"}, "avg_price", 1.0),
	}

	// 执行所有运算并添加结果到记录
	for _, op := range operations {
		result := op.Evaluate(record)
		newField := op.NewField()
		record[newField] = result
	}

	// 验证结果
	expectedTotalPrice := 200.0
	if record["total_price"].(float64) != expectedTotalPrice {
		t.Errorf("Integration test failed: expected total_price %f, got %f", expectedTotalPrice, record["total_price"].(float64))
	}

	expectedFinalPrice := 195.0 // 200 - 5
	if record["final_price"].(float64) != expectedFinalPrice {
		t.Errorf("Integration test failed: expected final_price %f, got %f", expectedFinalPrice, record["final_price"].(float64))
	}

	expectedAvgPrice := 100.0 // 200 / 2
	if record["avg_price"].(float64) != expectedAvgPrice {
		t.Errorf("Integration test failed: expected avg_price %f, got %f", expectedAvgPrice, record["avg_price"].(float64))
	}
}
