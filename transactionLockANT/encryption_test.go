package transactionLockANT

import (
	"bytes"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestEncryptionManager 测试加密管理器
func TestEncryptionManager(t *testing.T) {
	// 创建加密配置
	config := &TransactionEncryptionConfig{
		BaseConfig: storage.EncryptionConfig{
			Enabled:      true,
			Algorithm:    "AES-256-GCM",
			MasterKey:    []byte("01234567890123456789012345678901"), // 32字节密钥
		},
	}

	// 初始化加密管理器
	err := InitEncryption(config)
	if err != nil {
		t.Fatalf("初始化加密管理器失败: %v", err)
	}

	// 获取加密管理器
	manager := GetEncryptionManager()
	if manager == nil {
		t.Fatal("获取加密管理器失败")
	}

	// 测试数据
	testData := []byte("Hello, World!")

	// 测试加密
	encrypted, err := manager.Encrypt(testData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 测试解密
	decrypted, err := manager.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 验证解密后的数据与原始数据一致
	if !bytes.Equal(testData, decrypted) {
		t.Fatal("解密后的数据与原始数据不一致")
	}

	t.Log("加密/解密测试通过")
}

// TestFieldEncryption 测试字段级加密
func TestFieldEncryption(t *testing.T) {
	// 创建加密配置
	config := &TransactionEncryptionConfig{
		BaseConfig: storage.EncryptionConfig{
			Enabled:      true,
			Algorithm:    "AES-256-GCM",
			MasterKey:    []byte("01234567890123456789012345678901"), // 32字节密钥
		},
		FieldEncryption: map[string]FieldEncryptionConfig{
			"password": {
				Enabled:   true,
				Algorithm: "AES-256-GCM",
				KeyID:     "key1",
			},
		},
	}

	// 初始化加密管理器
	err := InitEncryption(config)
	if err != nil {
		t.Fatalf("初始化加密管理器失败: %v", err)
	}

	// 获取加密管理器
	manager := GetEncryptionManager()
	if manager == nil {
		t.Fatal("获取加密管理器失败")
	}

	// 测试数据
	testData := []byte("secret password")

	// 测试字段加密
	encrypted, err := manager.EncryptField("password", testData)
	if err != nil {
		t.Fatalf("字段加密失败: %v", err)
	}

	// 测试字段解密
	decrypted, err := manager.DecryptField("password", encrypted)
	if err != nil {
		t.Fatalf("字段解密失败: %v", err)
	}

	// 验证解密后的数据与原始数据一致
	if !bytes.Equal(testData, decrypted) {
		t.Fatal("解密后的数据与原始数据不一致")
	}

	t.Log("字段级加密/解密测试通过")
}

// TestEncryptionAccessControl 测试加密模块的访问控制
func TestEncryptionAccessControl(t *testing.T) {
	// 创建加密配置
	config := &TransactionEncryptionConfig{
		BaseConfig: storage.EncryptionConfig{
			Enabled:      true,
			Algorithm:    "AES-256-GCM",
			MasterKey:    []byte("01234567890123456789012345678901"), // 32字节密钥
		},
		AccessControl: AccessControlConfig{
			Enabled: true,
			Roles: map[string]Role{
				"admin": {
					Name: "admin",
					Permissions: []*Permission{
						{
							ID:          "perm:admin:read:users",
							Name:        "Read Users",
							Description: "Allow read access to users table",
							ResourceType: "table",
							ResourceID:  "users",
							Action:      "read",
							CreatedAt:   time.Now(),
						},
						{
							ID:          "perm:admin:write:users",
							Name:        "Write Users",
							Description: "Allow write access to users table",
							ResourceType: "table",
							ResourceID:  "users",
							Action:      "write",
							CreatedAt:   time.Now(),
						},
					},
				},
				"user": {
					Name: "user",
					Permissions: []*Permission{
						{
							ID:          "perm:user:read:users",
							Name:        "Read Users",
							Description: "Allow read access to users table",
							ResourceType: "table",
							ResourceID:  "users",
							Action:      "read",
							CreatedAt:   time.Now(),
						},
					},
				},
			},
		},
	}

	// 初始化加密管理器
	err := InitEncryption(config)
	if err != nil {
		t.Fatalf("初始化加密管理器失败: %v", err)
	}

	// 获取加密管理器
	manager := GetEncryptionManager()
	if manager == nil {
		t.Fatal("获取加密管理器失败")
	}

	// 测试管理员权限
	if !manager.CheckPermission("admin", "table", "users", "read") {
		t.Error("管理员应该有读取权限")
	}

	if !manager.CheckPermission("admin", "table", "users", "write") {
		t.Error("管理员应该有写入权限")
	}

	// 测试普通用户权限
	if !manager.CheckPermission("user", "table", "users", "read") {
		t.Error("普通用户应该有读取权限")
	}

	if manager.CheckPermission("user", "table", "users", "write") {
		t.Error("普通用户不应该有写入权限")
	}

	// 测试不存在的角色
	if manager.CheckPermission("guest", "table", "users", "read") {
		t.Error("不存在的角色不应该有任何权限")
	}

	t.Log("访问控制测试通过")
}