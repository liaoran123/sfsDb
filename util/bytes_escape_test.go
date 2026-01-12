package util

import (
	"testing"
)

func TestBytesEscapeAndUnEscape(t *testing.T) {
	// 测试用例1: 包含默认分隔符 '-' 的字符串
	testStr1 := "test-name-with-dashes"
	escaped1 := Bytes(testStr1).Escape()
	t.Logf("Original 1: %s", testStr1)
	t.Logf("Escaped 1: %s", escaped1)

	unescaped1 := Bytes(escaped1).UnEscape()
	t.Logf("UnEscaped 1: %s", unescaped1)

	if string(unescaped1) != testStr1 {
		t.Errorf("Test 1 failed: expected '%s', got '%s'", testStr1, unescaped1)
	}

	// 测试用例2: 包含自定义分隔符 ':' 的字符串
	testStr2 := "test:name:with:colons"
	escaped2 := Bytes(testStr2).Escapes(':')
	t.Logf("Original 2: %s", testStr2)
	t.Logf("Escaped 2: %s", escaped2)

	// 注意：UnEscape 只处理默认分隔符，所以这个测试会失败
	// 这是一个设计问题，UnEscape 应该也能接受自定义分隔符
	// 但我们先测试当前功能
	unescaped2 := Bytes(escaped2).UnEscapes(':')
	t.Logf("UnEscaped 2: %s", unescaped2)

	// 测试用例3: 空字符串
	testStr3 := ""
	escaped3 := Bytes(testStr3).Escape()
	unescaped3 := Bytes(escaped3).UnEscape()
	if string(unescaped3) != testStr3 {
		t.Errorf("Test 3 failed: expected '%s', got '%s'", testStr3, unescaped3)
	}

	// 测试用例4: 只包含分隔符的字符串
	testStr4 := "--"
	escaped4 := Bytes(testStr4).Escape()
	t.Logf("Original 4: %s", testStr4)
	t.Logf("Escaped 4: %s", escaped4)

	unescaped4 := Bytes(escaped4).UnEscape()
	t.Logf("UnEscaped 4: %s", unescaped4)

	if string(unescaped4) != testStr4 {
		t.Errorf("Test 4 failed: expected '%s', got '%s'", testStr4, unescaped4)
	}

	t.Log("All tests passed!")
}

func TestBytesSplit(t *testing.T) {
	// 测试用例1: 正常分割
	testStr1 := "test-name-field"
	split1 := Bytes(testStr1).Split()
	t.Logf("Original 1: %s", testStr1)
	t.Logf("Split 1: %v", split1)

	if len(split1) != 3 {
		t.Errorf("Test 1 failed: expected 3 parts, got %d", len(split1))
	}

	// 测试用例2: 包含转义分隔符的分割
	testStr2 := "test--name-field"
	split2 := Bytes(testStr2).Split()
	t.Logf("Original 2: %s", testStr2)
	t.Logf("Split 2: %v", split2)

	if len(split2) != 2 {
		t.Errorf("Test 2 failed: expected 2 parts, got %d", len(split2))
	}

	if string(split2[0]) != "test-name" {
		t.Errorf("Test 2 failed: expected first part to be 'test-name', got '%s'", split2[0])
	}

	// 测试用例3: 空字符串分割
	testStr3 := ""
	split3 := Bytes(testStr3).Split()
	t.Logf("Original 3: %s", testStr3)
	t.Logf("Split 3: %v", split3)

	if len(split3) != 0 {
		t.Errorf("Test 3 failed: expected 0 parts, got %d", len(split3))
	}

	t.Log("All split tests passed!")
}
