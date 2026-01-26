package record

import (
	"testing"
)

// TestApplyWithExistingOperations 测试使用现有操作函数
func TestApplyWithExistingOperations(t *testing.T) {
	rs1 := Records{
		Record{"id": 1, "name": "test1"},
		Record{"id": 2, "name": "test2"},
		Record{"id": 3, "name": "test3"},
	}

	rs2 := Records{
		Record{"id": 2, "name": "test2"},
		Record{"id": 3, "name": "test3"},
		Record{"id": 4, "name": "test4"},
	}

	// 测试使用 Intersect 函数
	intersectResult := rs1.Apply(func(rs Records, other ...Records) Records {
		return rs.Intersect(other...)
	}, rs2)

	if len(intersectResult) != 2 {
		t.Errorf("Expected intersect result length 2, got %d", len(intersectResult))
	}

	// 测试使用 Union 函数
	unionResult := rs1.Apply(func(rs Records, other ...Records) Records {
		return rs.Union(other...)
	}, rs2)

	if len(unionResult) != 4 {
		t.Errorf("Expected union result length 4, got %d", len(unionResult))
	}

	// 测试使用 Difference 函数
	diffResult := rs1.Apply(func(rs Records, other ...Records) Records {
		return rs.Difference(other...)
	}, rs2)

	if len(diffResult) != 1 {
		t.Errorf("Expected difference result length 1, got %d", len(diffResult))
	}
}

// TestApplyWithCustomOperation 测试使用自定义操作函数
func TestApplyWithCustomOperation(t *testing.T) {
	rs1 := Records{
		Record{"id": 1, "name": "test1"},
		Record{"id": 2, "name": "test2"},
		Record{"id": 3, "name": "test3"},
	}

	// 自定义操作：返回所有 id 大于 1 的记录
	customOp := func(rs Records, other ...Records) Records {
		result := make(Records, 0)
		for _, r := range rs {
			if id, ok := r["id"].(int); ok && id > 1 {
				result = append(result, r)
			}
		}
		return result
	}

	customResult := rs1.Apply(customOp)

	if len(customResult) != 2 {
		t.Errorf("Expected custom result length 2, got %d", len(customResult))
	}
}

// TestApplyWithComposedOperation 测试使用组合操作函数
func TestApplyWithComposedOperation(t *testing.T) {
	rs1 := Records{
		Record{"id": 1, "name": "test1"},
		Record{"id": 2, "name": "test2"},
		Record{"id": 3, "name": "test3"},
	}

	rs2 := Records{
		Record{"id": 2, "name": "test2"},
		Record{"id": 3, "name": "test3"},
		Record{"id": 4, "name": "test4"},
	}

	rs3 := Records{
		Record{"id": 3, "name": "test3"},
		Record{"id": 4, "name": "test4"},
		Record{"id": 5, "name": "test5"},
	}

	// 组合操作：先求 rs1 和 rs2 的并集，再与 rs3 求交集
	composedOp := func(rs Records, other ...Records) Records {
		unionResult := rs.Union(other[0])
		return unionResult.Intersect(other[1])
	}

	composedResult := rs1.Apply(composedOp, rs2, rs3)

	if len(composedResult) != 2 {
		t.Errorf("Expected composed result length 2, got %d", len(composedResult))
	}
}
