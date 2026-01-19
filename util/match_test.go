package util

import (
	"testing"
	"time"
)

func TestMatchCompareWithNewOperators(t *testing.T) {
	// 测试字符串比较
	strMatch := NewMatch("hello world")
	
	// 测试Prefix操作符
	if !strMatch.Compare(Prefix, "hello") {
		t.Error("Prefix comparison failed: 'hello world' should start with 'hello'")
	}
	
	// 测试Suffix操作符
	if !strMatch.Compare(Suffix, "world") {
		t.Error("Suffix comparison failed: 'hello world' should end with 'world'")
	}
	
	// 测试Contains操作符
	if !strMatch.Compare(Contains, "lo w") {
		t.Error("Contains comparison failed: 'hello world' should contain 'lo w'")
	}
	
	// 测试Like操作符
	if !strMatch.Compare(Like, "hello%") {
		t.Error("Like comparison failed: 'hello world' should match 'hello%'")
	}
	
	if !strMatch.Compare(Like, "%world") {
		t.Error("Like comparison failed: 'hello world' should match '%world'")
	}
	
	if !strMatch.Compare(Like, "hello%world") {
		t.Error("Like comparison failed: 'hello world' should match 'hello%world'")
	}
	
	// 测试数值比较
	numMatch := NewMatch(42)
	
	if !numMatch.Compare(Equal, 42) {
		t.Error("Equal comparison failed: 42 should equal 42")
	}
	
	if !numMatch.Compare(GreaterThan, 41) {
		t.Error("GreaterThan comparison failed: 42 should be greater than 41")
	}
	
	if !numMatch.Compare(LessThan, 43) {
		t.Error("LessThan comparison failed: 42 should be less than 43")
	}
	
	// 测试时间比较
	timeMatch := NewMatch(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	
	if !timeMatch.Compare(Equal, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("Equal comparison failed for time")
	}
	
	if !timeMatch.Compare(LessThan, time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Error("LessThan comparison failed for time")
	}
	
	// 测试不匹配的情况
	if strMatch.Compare(Prefix, "world") {
		t.Error("Prefix comparison failed: 'hello world' should not start with 'world'")
	}
	
	if strMatch.Compare(Suffix, "hello") {
		t.Error("Suffix comparison failed: 'hello world' should not end with 'hello'")
	}
	
	if strMatch.Compare(Contains, "xyz") {
		t.Error("Contains comparison failed: 'hello world' should not contain 'xyz'")
	}
}

func TestMatchBasicOperations(t *testing.T) {
	// 测试基本比较操作
	match := NewMatch(100)
	
	// 测试Equal
	if !match.Equal(100) {
		t.Error("Equal failed: 100 should equal 100")
	}
	
	// 测试NotEqual
	if !match.NotEqual(99) {
		t.Error("NotEqual failed: 100 should not equal 99")
	}
	
	// 测试GreaterThan
	if !match.GreaterThan(99) {
		t.Error("GreaterThan failed: 100 should be greater than 99")
	}
	
	// 测试GreaterThanOrEqual
	if !match.GreaterThanOrEqual(100) {
		t.Error("GreaterThanOrEqual failed: 100 should be greater than or equal to 100")
	}
	
	if !match.GreaterThanOrEqual(99) {
		t.Error("GreaterThanOrEqual failed: 100 should be greater than or equal to 99")
	}
	
	// 测试LessThan
	if !match.LessThan(101) {
		t.Error("LessThan failed: 100 should be less than 101")
	}
	
	// 测试LessThanOrEqual
	if !match.LessThanOrEqual(100) {
		t.Error("LessThanOrEqual failed: 100 should be less than or equal to 100")
	}
	
	if !match.LessThanOrEqual(101) {
		t.Error("LessThanOrEqual failed: 100 should be less than or equal to 101")
	}
}

func TestMatchStringOperations(t *testing.T) {
	// 测试字符串操作
	match := NewMatch("test string")
	
	// 测试Prefix
	if !match.Prefix("test") {
		t.Error("Prefix failed: 'test string' should start with 'test'")
	}
	
	// 测试Suffix
	if !match.Suffix("string") {
		t.Error("Suffix failed: 'test string' should end with 'string'")
	}
	
	// 测试Contains
	if !match.Contains("st ") {
		t.Error("Contains failed: 'test string' should contain 'st '")
	}
	
	// 测试Like
	if !match.Like("test%") {
		t.Errorf("Like failed: 'test string' should match 'test%%'")
	}
	
	if !match.Like("%string") {
		t.Errorf("Like failed: 'test string' should match '%%string'")
	}
	
	if !match.Like("%st%") {
		t.Errorf("Like failed: 'test string' should match '%%st%%'")
	}
	
	if !match.Like("test string") {
		t.Error("Like failed: 'test string' should match 'test string'")
	}
	
	if match.Like("test%xyz") {
		t.Errorf("Like failed: 'test string' should not match 'test%%xyz'")
	}
}

func TestMatchExists(t *testing.T) {
	// 测试Exists方法
	match1 := NewMatch("test")
	if !match1.Exists() {
		t.Error("Exists failed: non-nil value should exist")
	}
	
	var nilValue any
	match2 := NewMatch(nilValue)
	if match2.Exists() {
		t.Error("Exists failed: nil value should not exist")
	}
}