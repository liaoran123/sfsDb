package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

func TestUtilMatch(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 测试 Equal 匹配器
	equalMatch := NewUtilMatch("id", util.Equal, 1)
	if !equalMatch.Match(fields) {
		t.Error("Equal match failed: id should equal 1")
	}

	// 测试 NotEqual 匹配器
	notEqualMatch := NewUtilMatch("id", util.NotEqual, 2)
	if !notEqualMatch.Match(fields) {
		t.Error("NotEqual match failed: id should not equal 2")
	}

	// 测试 GreaterThan 匹配器
	greaterThanMatch := NewUtilMatch("age", util.GreaterThan, 20)
	if !greaterThanMatch.Match(fields) {
		t.Error("GreaterThan match failed: age should be greater than 20")
	}

	// 测试 LessThan 匹配器
	lessThanMatch := NewUtilMatch("age", util.LessThan, 30)
	if !lessThanMatch.Match(fields) {
		t.Error("LessThan match failed: age should be less than 30")
	}

	// 测试 Like 匹配器
	likeMatch := NewUtilMatch("email", util.Like, "%example%")
	if !likeMatch.Match(fields) {
		t.Errorf("Like match failed: email should match '%%example%%'")
	}

	// 测试 Prefix 匹配器
	prefixMatch := NewUtilMatch("name", util.Prefix, "te")
	if !prefixMatch.Match(fields) {
		t.Error("Prefix match failed: name should start with 'te'")
	}

	// 测试 Suffix 匹配器
	suffixMatch := NewUtilMatch("name", util.Suffix, "st")
	if !suffixMatch.Match(fields) {
		t.Error("Suffix match failed: name should end with 'st'")
	}

	// 测试 Contains 匹配器
	containsMatch := NewUtilMatch("email", util.Contains, "example")
	if !containsMatch.Match(fields) {
		t.Error("Contains match failed: email should contain 'example'")
	}
}

func TestUtilMatchHelperFunctions(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 测试 NewEqualMatch
	equalMatch := NewEqualMatch("id", 1)
	if !equalMatch.Match(fields) {
		t.Error("NewEqualMatch failed: id should equal 1")
	}

	// 测试 NewNotEqualMatch
	notEqualMatch := NewNotEqualMatch("id", 2)
	if !notEqualMatch.Match(fields) {
		t.Error("NewNotEqualMatch failed: id should not equal 2")
	}

	// 测试 NewGreaterThanMatch
	greaterThanMatch := NewGreaterThanMatch("age", 20)
	if !greaterThanMatch.Match(fields) {
		t.Error("NewGreaterThanMatch failed: age should be greater than 20")
	}

	// 测试 NewLessThanMatch
	lessThanMatch := NewLessThanMatch("age", 30)
	if !lessThanMatch.Match(fields) {
		t.Error("NewLessThanMatch failed: age should be less than 30")
	}

	// 测试 NewLikeMatch
	likeMatch := NewLikeMatch("email", "%example%")
	if !likeMatch.Match(fields) {
		t.Errorf("NewLikeMatch failed: email should match '%%example%%'")
	}

	// 测试 NewPrefixMatch
	prefixMatch := NewPrefixMatch("name", "te")
	if !prefixMatch.Match(fields) {
		t.Error("NewPrefixMatch failed: name should start with 'te'")
	}

	// 测试 NewSuffixMatch
	suffixMatch := NewSuffixMatch("name", "st")
	if !suffixMatch.Match(fields) {
		t.Error("NewSuffixMatch failed: name should end with 'st'")
	}

	// 测试 NewContainsMatch
	containsMatch := NewContainsMatch("email", "example")
	if !containsMatch.Match(fields) {
		t.Error("NewContainsMatch failed: email should contain 'example'")
	}
}

func TestFieldMatch(t *testing.T) {
	// 创建测试数据
	fields := &map[string]any{
		"id":     1,
		"name":   "test",
		"age":    25,
		"email":  "test@example.com",
		"active": true,
	}

	// 测试 FieldMatch 与 util.Match 结合使用
	fieldMatch := NewFieldMatch("age", func(value any) bool {
		match := util.NewMatch(25)
		return match.Compare(util.Equal, value)
	})

	if !fieldMatch.Match(fields) {
		t.Error("FieldMatch failed: age should equal 25")
	}

	// 测试 FieldMatch 与自定义匹配函数
	customMatch := NewFieldMatch("active", func(value any) bool {
		// 检查是否为布尔值且为 true
		if b, ok := value.(bool); ok {
			return b
		}
		return false
	})

	if !customMatch.Match(fields) {
		t.Error("Custom FieldMatch failed: active should be true")
	}
}
