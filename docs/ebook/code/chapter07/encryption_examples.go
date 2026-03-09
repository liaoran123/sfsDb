package main

import (
	"fmt"
	"os"

	"github.com/liaoran123/sfsDb/storage"
)

func main3() {
	fmt.Println("=== sfsDb 加密存储示例 ===")

	example1BasicEncryption()
	example2KeyDerivation()
	example3EncryptedStore()
}

func example1BasicEncryption() {
	fmt.Println("\n--- 示例1: 基本加密解密 ---")

	// 创建 256 位密钥 (32 字节)
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i)
	}

	// 创建 AES-GCM 加密器
	encryptor, err := storage.NewAESGCMEncryptor(key)
	if err != nil {
		fmt.Printf("创建加密器失败: %v\n", err)
		return
	}

	fmt.Printf("加密算法: %s\n", encryptor.Algorithm())

	// 测试数据
	plaintext := []byte("Hello, 加密存储!")
	fmt.Printf("原始数据: %s\n", plaintext)

	// 加密
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密后: %x\n", ciphertext)

	// 解密
	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密后: %s\n", decrypted)

	if string(decrypted) == string(plaintext) {
		fmt.Println("✓ 加密解密验证成功!")
	}
}

func example2KeyDerivation() {
	fmt.Println("\n--- 示例2: 使用密码派生密钥 ---")

	password := []byte("my-secure-password-123")
	var salt []byte
	iterations := 100000

	// 派生密钥
	key, err := storage.DeriveKey(password, salt, iterations)
	if err != nil {
		fmt.Printf("派生密钥失败: %v\n", err)
		return
	}

	fmt.Printf("密码: %s\n", password)
	fmt.Printf("派生密钥长度: %d 字节\n", len(key))
	fmt.Printf("派生密钥(前16字节): %x\n", key[:16])

	// 使用派生密钥创建加密器
	encryptor, err := storage.NewAESGCMEncryptor(key)
	if err != nil {
		fmt.Printf("创建加密器失败: %v\n", err)
		return
	}

	fmt.Printf("使用派生密钥的加密器: %s\n", encryptor.Algorithm())
}

func example3EncryptedStore() {
	fmt.Println("\n--- 示例3: 加密存储包装器 ---")

	// 清理测试目录
	os.RemoveAll("./test_encrypted_db")
	defer os.RemoveAll("./test_encrypted_db")

	// 获取数据库管理器
	dbManager := storage.GetDBManager()

	// 创建加密配置
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i * 2)
	}

	config := &storage.EncryptionConfig{
		Enabled:    true,
		Algorithm:  "AES-256-GCM",
		MasterKey:  key,
		Password:   "",
		Salt:       nil,
		Iterations: 0,
	}

	// 使用 DBManager 创建加密存储
	store, err := dbManager.NewLevelDBStore("./test_encrypted_db", nil, config)
	if err != nil {
		fmt.Printf("创建加密存储失败: %v\n", err)
		return
	}
	defer store.Close()

	fmt.Println("加密存储创建成功")

	// 写入数据
	testData := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("user:1"), []byte(`{"name":"张三","age":25}`)},
		{[]byte("user:2"), []byte(`{"name":"李四","age":30}`)},
		{[]byte("config:theme"), []byte("dark")},
	}

	for _, data := range testData {
		err = store.Put(data.key, data.value)
		if err != nil {
			fmt.Printf("写入数据失败 %s: %v\n", data.key, err)
			return
		}
	}
	fmt.Printf("✓ 写入 %d 条加密数据\n", len(testData))

	// 读取数据
	for _, data := range testData {
		value, err := store.Get(data.key)
		if err != nil {
			fmt.Printf("读取数据失败 %s: %v\n", data.key, err)
			return
		}
		fmt.Printf("读取 %s: %s\n", data.key, value)
	}

	// 如果是加密存储包装器，可以获取加密配置
	if encryptedStore, ok := store.(*storage.EncryptedStoreWrapper); ok {
		retrievedConfig := encryptedStore.GetEncryptionConfig()
		fmt.Printf("加密配置 - 启用: %v, 算法: %s\n",
			retrievedConfig.Enabled,
			retrievedConfig.Algorithm)
	}
}
