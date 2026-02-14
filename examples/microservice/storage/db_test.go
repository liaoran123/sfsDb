package storage

import (
	"os"
	"testing"
	"time"
)

// TestInitLocalStorage 测试初始化本地存储
func TestInitLocalStorage(t *testing.T) {
	// 清理旧数据
	serviceName := "test-service"
	dbPath := "./" + serviceName + "_local_db"
	defer os.RemoveAll(dbPath)

	// 初始化本地存储
	err := InitLocalStorage(serviceName)
	if err != nil {
		t.Fatalf("初始化本地存储失败: %v", err)
	}
	defer CloseLocalStorage()

	// 验证数据库路径
	if GetDBPath() != dbPath {
		t.Errorf("数据库路径不匹配，期望: %s, 实际: %s", dbPath, GetDBPath())
	}
}

// TestSetAndGetState 测试设置和获取状态
func TestSetAndGetState(t *testing.T) {
	// 清理旧数据
	serviceName := "test-service"
	dbPath := "./" + serviceName + "_local_db"
	defer os.RemoveAll(dbPath)

	// 初始化本地存储
	err := InitLocalStorage(serviceName)
	if err != nil {
		t.Fatalf("初始化本地存储失败: %v", err)
	}
	defer CloseLocalStorage()

	// 测试数据
	testKey := "test-key"
	testValue := "test-value"

	// 设置状态
	err = SetState(testKey, testValue, 0)
	if err != nil {
		t.Fatalf("设置状态失败: %v", err)
	}

	// 获取状态
	value, err := GetState(testKey)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	// 验证值
	if value != testValue {
		t.Errorf("值不匹配，期望: %s, 实际: %s", testValue, value)
	}
}

// TestGetNonExistentState 测试获取不存在的状态
func TestGetNonExistentState(t *testing.T) {
	// 清理旧数据
	serviceName := "test-service"
	dbPath := "./" + serviceName + "_local_db"
	defer os.RemoveAll(dbPath)

	// 初始化本地存储
	err := InitLocalStorage(serviceName)
	if err != nil {
		t.Fatalf("初始化本地存储失败: %v", err)
	}
	defer CloseLocalStorage()

	// 获取不存在的状态
	value, err := GetState("non-existent-key")
	if err != nil {
		t.Fatalf("获取不存在的状态失败: %v", err)
	}

	// 验证值为空
	if value != "" {
		t.Errorf("期望空值，实际: %s", value)
	}
}

// TestSetStateWithExpire 测试设置带过期时间的状态
func TestSetStateWithExpire(t *testing.T) {
	// 清理旧数据
	serviceName := "test-service"
	dbPath := "./" + serviceName + "_local_db"
	defer os.RemoveAll(dbPath)

	// 初始化本地存储
	err := InitLocalStorage(serviceName)
	if err != nil {
		t.Fatalf("初始化本地存储失败: %v", err)
	}
	defer CloseLocalStorage()

	// 测试数据
	testKey := "expire-key"
	testValue := "expire-value"
	expireTime := time.Now().Unix() + 3600 // 1小时后过期

	// 设置带过期时间的状态
	err = SetState(testKey, testValue, expireTime)
	if err != nil {
		t.Fatalf("设置带过期时间的状态失败: %v", err)
	}

	// 获取状态
	value, err := GetState(testKey)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	// 验证值
	if value != testValue {
		t.Errorf("值不匹配，期望: %s, 实际: %s", testValue, value)
	}
}

// TestMultipleStates 测试多个状态
func TestMultipleStates(t *testing.T) {
	// 清理旧数据
	serviceName := "test-service"
	dbPath := "./" + serviceName + "_local_db"
	defer os.RemoveAll(dbPath)

	// 初始化本地存储
	err := InitLocalStorage(serviceName)
	if err != nil {
		t.Fatalf("初始化本地存储失败: %v", err)
	}
	defer CloseLocalStorage()

	// 测试数据
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// 设置多个状态
	for key, value := range testData {
		err = SetState(key, value, 0)
		if err != nil {
			t.Fatalf("设置状态失败 (key=%s): %v", key, err)
		}
	}

	// 获取并验证多个状态
	for key, expectedValue := range testData {
		value, err := GetState(key)
		if err != nil {
			t.Fatalf("获取状态失败 (key=%s): %v", key, err)
		}
		if value != expectedValue {
			t.Errorf("值不匹配 (key=%s)，期望: %s, 实际: %s", key, expectedValue, value)
		}
	}
}
