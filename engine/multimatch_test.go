package engine

import (
	"testing"
)

func TestMultiFieldMatch(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 创建匹配器
	idMatch := NewEqualMatch("id", 1)
	nameMatch := NewEqualMatch("name", "test")
	ageMatch := NewGreaterThanMatch("age", 20)
	emailMatch := NewLikeMatch("email", "%example%")
	activeMatch := NewEqualMatch("active", true)
	invalidMatch := NewEqualMatch("invalid", "value")

	// 测试 AND 逻辑：所有匹配器都必须匹配
	andMatch := NewAndMatch(idMatch, nameMatch, ageMatch, emailMatch, activeMatch)
	if !andMatch.Match(fields) {
		t.Error("AND match failed: all matchers should match")
	}

	// 测试 AND 逻辑：包含无效匹配器
	andMatchWithInvalid := NewAndMatch(idMatch, invalidMatch)
	if andMatchWithInvalid.Match(fields) {
		t.Error("AND match with invalid field failed: should return false")
	}

	// 测试 OR 逻辑：至少一个匹配器匹配
	orMatch := NewOrMatch(invalidMatch, idMatch)
	if !orMatch.Match(fields) {
		t.Error("OR match failed: at least one matcher should match")
	}

	// 测试 OR 逻辑：没有匹配器匹配
	orMatchNoMatch := NewOrMatch(invalidMatch, NewEqualMatch("id", 99))
	if orMatchNoMatch.Match(fields) {
		t.Error("OR match with no matches failed: should return false")
	}

	// 测试动态添加匹配器
	dynamicMatch := NewAndMatch(idMatch, nameMatch)
	dynamicMatch.AddMatcher(ageMatch)
	if !dynamicMatch.Match(fields) {
		t.Error("Dynamic AND match failed: added matcher should be included")
	}

	// 测试空匹配器列表
	emptyMatch := NewAndMatch()
	if emptyMatch.Match(fields) {
		t.Error("Empty AND match failed: should return false")
	}

	emptyOrMatch := NewOrMatch()
	if emptyOrMatch.Match(fields) {
		t.Error("Empty OR match failed: should return false")
	}
}

func TestComplexMatch(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 创建匹配器
	idMatch := NewEqualMatch("id", 1)
	nameMatch := NewEqualMatch("name", "test")
	ageMatch := NewGreaterThanMatch("age", 20)
	emailMatch := NewLikeMatch("email", "%example%")
	activeMatch := NewEqualMatch("active", true)
	invalidMatch := NewEqualMatch("invalid", "value")

	// 测试简单的 AND 复杂匹配
	complexAnd := NewAndComplexMatch(idMatch, nameMatch, ageMatch)
	if !complexAnd.Match(fields) {
		t.Error("Complex AND match failed: all matchers should match")
	}

	// 测试简单的 OR 复杂匹配
	complexOr := NewOrComplexMatch(invalidMatch, idMatch)
	if !complexOr.Match(fields) {
		t.Error("Complex OR match failed: at least one matcher should match")
	}

	// 测试嵌套的 AND-OR 逻辑
	// (id=1 AND name=test) OR (age>20 AND email like %example%)
	nestedMatch := NewOrComplexMatch(
		NewAndComplexMatch(idMatch, nameMatch),
		NewAndComplexMatch(ageMatch, emailMatch),
	)
	if !nestedMatch.Match(fields) {
		t.Error("Nested AND-OR match failed: should return true")
	}

	// 测试嵌套的 OR-AND 逻辑
	// (id=99 OR name=test) AND (age>20 AND active=true)
	nestedMatch2 := NewAndComplexMatch(
		NewOrComplexMatch(NewEqualMatch("id", 99), nameMatch),
		NewAndComplexMatch(ageMatch, activeMatch),
	)
	if !nestedMatch2.Match(fields) {
		t.Error("Nested OR-AND match failed: should return true")
	}

	// 测试复杂匹配器的短路逻辑
	// 应该短路，不执行所有匹配器
	shortCircuitMatch := NewAndComplexMatch(
		idMatch,      // 匹配
		invalidMatch, // 不匹配，应该短路
		nameMatch,    // 不应该执行
	)
	if shortCircuitMatch.Match(fields) {
		t.Error("Short circuit AND match failed: should return false")
	}

	// 测试 OR 短路逻辑
	shortCircuitOrMatch := NewOrComplexMatch(
		idMatch,      // 匹配，应该短路
		invalidMatch, // 不应该执行
		nameMatch,    // 不应该执行
	)
	if !shortCircuitOrMatch.Match(fields) {
		t.Error("Short circuit OR match failed: should return true")
	}

	// 测试动态添加匹配器到复杂匹配器
	dynamicComplexMatch := NewAndComplexMatch(idMatch, nameMatch)
	dynamicComplexMatch.AddMatcher(ageMatch)
	if !dynamicComplexMatch.Match(fields) {
		t.Error("Dynamic complex match failed: added matcher should be included")
	}
}

func TestComplexNestedMatch(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 创建匹配器
	idMatch := NewEqualMatch("id", 1)
	nameMatch := NewEqualMatch("name", "test")
	ageMatch := NewGreaterThanMatch("age", 20)
	emailMatch := NewLikeMatch("email", "%example%")
	activeMatch := NewEqualMatch("active", true)
	ageLessThanMatch := NewLessThanMatch("age", 30)

	// 测试深度嵌套的匹配器
	// ((id=1 AND name=test) OR (email like %example% AND active=true)) AND (age>20 AND age<30)
	depthNestedMatch := NewAndComplexMatch(
		NewOrComplexMatch(
			NewAndComplexMatch(idMatch, nameMatch),
			NewAndComplexMatch(emailMatch, activeMatch),
		),
		NewAndComplexMatch(ageMatch, ageLessThanMatch),
	)
	if !depthNestedMatch.Match(fields) {
		t.Error("Depth nested match failed: should return true")
	}

	// 测试无效类型的匹配器
	// 应该忽略无效类型，继续执行其他匹配器
	invalidTypeMatch := NewAndComplexMatch(
		idMatch,
		"invalid type", // 无效类型
		nameMatch,
	)
	if !invalidTypeMatch.Match(fields) {
		t.Error("Invalid type match failed: should ignore invalid type and return true")
	}
}
