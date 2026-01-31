package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

// 定义加密相关错误
var (
	ErrInvalidKeyLength = NewError("invalid key length")
	ErrInvalidAlgorithm = NewError("invalid algorithm")
	ErrEncryptionFailed = NewError("encryption failed")
	ErrDecryptionFailed = NewError("decryption failed")
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	// 是否启用加密
	Enabled bool `json:"enabled"`

	// 加密算法，默认AES-256-GCM
	Algorithm string `json:"algorithm"`

	// 主密钥，可以是直接的密钥或密码派生
	MasterKey []byte `json:"master_key,omitempty"`

	// 密码，用于派生密钥
	Password string `json:"password,omitempty"`

	// 盐值，用于密码派生
	Salt []byte `json:"salt,omitempty"`

	// 迭代次数，用于密码派生
	Iterations int `json:"iterations,omitempty"`
}

// Encryptor 加密器接口
type Encryptor interface {
	// 加密数据
	Encrypt(plaintext []byte) ([]byte, error)

	// 解密数据
	Decrypt(ciphertext []byte) ([]byte, error)

	// 获取加密算法
	Algorithm() string
}

// AESGCMEncryptor AES-256-GCM加密器实现
type AESGCMEncryptor struct {
	key       []byte
	algorithm string
}

// NewAESGCMEncryptor 创建新的AES-GCM加密器
func NewAESGCMEncryptor(key []byte) (*AESGCMEncryptor, error) {
	// 验证密钥长度
	if len(key) != 32 { // 256位密钥
		return nil, ErrInvalidKeyLength
	}

	return &AESGCMEncryptor{
		key:       key,
		algorithm: "AES-256-GCM",
	}, nil
}

// Algorithm 返回加密算法
func (e *AESGCMEncryptor) Algorithm() string {
	return e.algorithm
}

// Encrypt 加密数据
func (e *AESGCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// 创建GCM模式
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// 生成随机nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}

	// 加密数据
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func (e *AESGCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// 创建GCM模式
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// 检查密文长度
	if len(ciphertext) < aesgcm.NonceSize() {
		return nil, ErrDecryptionFailed
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:aesgcm.NonceSize()], ciphertext[aesgcm.NonceSize():]

	// 解密数据
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// DeriveKey 使用PBKDF2派生密钥
func DeriveKey(password, salt []byte, iterations int) ([]byte, error) {
	// 设置默认值
	if iterations <= 0 {
		iterations = 100000
	}

	if len(salt) == 0 {
		// 生成随机盐值
		salt = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, err
		}
	}

	// 派生32字节（256位）密钥
	key := pbkdf2.Key(password, salt, iterations, 32, sha256.New)
	return key, nil
}

// EncryptedStoreWrapper 加密存储包装器
type EncryptedStoreWrapper struct {
	// 底层存储
	underlyingStore Store

	// 加密配置
	config *EncryptionConfig

	// 加密器
	encryptor Encryptor

	// 解密缓存
	decryptionCache map[string][]byte

	// 缓存互斥锁
	cacheMutex sync.RWMutex
}

// NewEncryptedStoreWrapper 创建新的加密存储包装器
func NewEncryptedStoreWrapper(underlyingStore Store, config *EncryptionConfig) (*EncryptedStoreWrapper, error) {
	// 验证配置
	if !config.Enabled {
		return nil, NewError("encryption is not enabled")
	}

	// 初始化密钥
	var key []byte
	var err error

	if len(config.MasterKey) > 0 {
		// 使用直接提供的密钥
		key = config.MasterKey
	} else if config.Password != "" {
		// 使用密码派生密钥
		key, err = DeriveKey([]byte(config.Password), config.Salt, config.Iterations)
		if err != nil {
			return nil, err
		}
		// 更新配置中的盐值和迭代次数
		if config.Iterations <= 0 {
			config.Iterations = 100000
		}
	} else {
		return nil, NewError("either master key or password must be provided")
	}

	// 创建加密器
	var encryptor Encryptor
	switch config.Algorithm {
	case "", "AES-256-GCM":
		encryptor, err = NewAESGCMEncryptor(key)
	default:
		return nil, ErrInvalidAlgorithm
	}

	if err != nil {
		return nil, err
	}

	return &EncryptedStoreWrapper{
		underlyingStore: underlyingStore,
		config:          config,
		encryptor:       encryptor,
		decryptionCache: make(map[string][]byte),
	}, nil
}

// GetEncryptionConfig 获取加密配置
func (es *EncryptedStoreWrapper) GetEncryptionConfig() *EncryptionConfig {
	return es.config
}

// ReEncrypt 重新加密所有数据
func (es *EncryptedStoreWrapper) ReEncrypt(newKey []byte) error {
	// 创建新的加密器
	newEncryptor, err := NewAESGCMEncryptor(newKey)
	if err != nil {
		return err
	}

	// 遍历所有数据
	iter := es.underlyingStore.Iterator(nil, nil)
	defer iter.Release()

	// 使用事务批量处理
	batch := es.underlyingStore.GetBatch()
	defer batch.Reset()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		encryptedValue := iter.Value()

		// 解密旧数据
		plaintext, err := es.encryptor.Decrypt(encryptedValue)
		if err != nil {
			return err
		}

		// 使用新密钥加密
		newEncryptedValue, err := newEncryptor.Encrypt(plaintext)
		if err != nil {
			return err
		}

		// 更新批次
		batch.Put(key, newEncryptedValue)
	}

	// 提交批次
	if err := es.underlyingStore.WriteBatch(batch); err != nil {
		return err
	}

	// 更新加密器
	es.encryptor = newEncryptor
	es.config.MasterKey = newKey

	// 清空缓存
	es.cacheMutex.Lock()
	es.decryptionCache = make(map[string][]byte)
	es.cacheMutex.Unlock()

	return nil
}

// Get 获取并解密数据
func (es *EncryptedStoreWrapper) Get(key []byte) ([]byte, error) {
	// 从缓存获取
	cacheKey := string(key)
	es.cacheMutex.RLock()
	if value, ok := es.decryptionCache[cacheKey]; ok {
		es.cacheMutex.RUnlock()
		return value, nil
	}
	es.cacheMutex.RUnlock()

	// 从底层存储获取加密数据
	encryptedValue, err := es.underlyingStore.Get(key)
	if err != nil {
		return nil, err
	}

	// 解密数据
	value, err := es.encryptor.Decrypt(encryptedValue)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	es.cacheMutex.Lock()
	es.decryptionCache[cacheKey] = value
	es.cacheMutex.Unlock()

	return value, nil
}

// Put 加密存储数据
func (es *EncryptedStoreWrapper) Put(key, value []byte) error {
	// 加密值
	encryptedValue, err := es.encryptor.Encrypt(value)
	if err != nil {
		return err
	}

	// 存储加密后的数据
	err = es.underlyingStore.Put(key, encryptedValue)
	if err != nil {
		return err
	}

	// 更新缓存
	cacheKey := string(key)
	es.cacheMutex.Lock()
	es.decryptionCache[cacheKey] = value
	es.cacheMutex.Unlock()

	return nil
}

// Delete 删除指定key
func (es *EncryptedStoreWrapper) Delete(key []byte) error {
	// 从底层存储删除
	err := es.underlyingStore.Delete(key)
	if err != nil {
		return err
	}

	// 从缓存删除
	cacheKey := string(key)
	es.cacheMutex.Lock()
	delete(es.decryptionCache, cacheKey)
	es.cacheMutex.Unlock()

	return nil
}

// GetBatch 创建批量操作对象
func (es *EncryptedStoreWrapper) GetBatch() Batch {
	// 返回加密批量操作对象
	return &encryptedBatch{
		underlyingBatch: es.underlyingStore.GetBatch(),
		encryptor:       es.encryptor,
	}
}

// WriteBatch 执行批量操作
func (es *EncryptedStoreWrapper) WriteBatch(batch Batch, put ...bool) error {
	// 提取底层batch
	var underlyingBatch Batch
	encryptedBatch, isEncrypted := batch.(*encryptedBatch)
	if isEncrypted {
		// 如果是加密批量操作，使用其底层batch
		underlyingBatch = encryptedBatch.underlyingBatch
	} else {
		// 否则直接使用传入的batch
		underlyingBatch = batch
	}

	// 执行批量操作
	err := es.underlyingStore.WriteBatch(underlyingBatch, put...)
	if err != nil {
		return err
	}

	// 如果是加密批量操作，更新缓存
	if isEncrypted {
		es.cacheMutex.Lock()
		for key, value := range encryptedBatch.decryptedValues {
			es.decryptionCache[key] = value
		}
		es.cacheMutex.Unlock()
	}

	return nil
}

// Iterator 创建迭代器
func (es *EncryptedStoreWrapper) Iterator(start, limit []byte) Iterator {
	// 返回加密迭代器
	return &encryptedIterator{
		underlyingIterator: es.underlyingStore.Iterator(start, limit),
		encryptor:          es.encryptor,
	}
}

// Snapshot 创建快照
func (es *EncryptedStoreWrapper) Snapshot() (Snapshot, error) {
	// 获取底层快照
	snapshot, err := es.underlyingStore.Snapshot()
	if err != nil {
		return nil, err
	}

	// 返回加密快照
	return &encryptedSnapshot{
		underlyingSnapshot: snapshot,
		encryptor:          es.encryptor,
	}, nil
}

// SwitchToSnapshot 切换到快照模式
func (es *EncryptedStoreWrapper) SwitchToSnapshot() error {
	return es.underlyingStore.SwitchToSnapshot()
}

// SwitchToDB 切换到数据库模式
func (es *EncryptedStoreWrapper) SwitchToDB() error {
	return es.underlyingStore.SwitchToDB()
}

// Close 关闭存储
func (es *EncryptedStoreWrapper) Close() error {
	// 清空缓存
	es.cacheMutex.Lock()
	es.decryptionCache = nil
	es.cacheMutex.Unlock()

	// 关闭底层存储
	return es.underlyingStore.Close()
}

// encryptedBatch 加密批量操作

type encryptedBatch struct {
	underlyingBatch Batch
	encryptor       Encryptor
	decryptedValues map[string][]byte
}

// Put 添加put操作
func (eb *encryptedBatch) Put(key []byte, value []byte) {
	// 加密值
	encryptedValue, err := eb.encryptor.Encrypt(value)
	if err != nil {
		return
	}

	// 添加到底层批量操作
	eb.underlyingBatch.Put(key, encryptedValue)

	// 保存解密后的值，用于更新缓存
	if eb.decryptedValues == nil {
		eb.decryptedValues = make(map[string][]byte)
	}
	eb.decryptedValues[string(key)] = value
}

// Delete 添加delete操作
func (eb *encryptedBatch) Delete(key []byte) {
	// 添加到底层批量操作
	eb.underlyingBatch.Delete(key)
}

// Len 获取批量操作数量
func (eb *encryptedBatch) Len() int {
	return eb.underlyingBatch.Len()
}

// Reset 重置批量操作
func (eb *encryptedBatch) Reset() {
	eb.underlyingBatch.Reset()
	eb.decryptedValues = nil
}

// encryptedIterator 加密迭代器

type encryptedIterator struct {
	underlyingIterator Iterator
	encryptor          Encryptor
	currentValue       []byte
}

// First 移动到第一个元素
func (ei *encryptedIterator) First() bool {
	if ei.underlyingIterator == nil {
		return false
	}
	if ei.underlyingIterator.First() {
		// 解密当前值
		value, err := ei.encryptor.Decrypt(ei.underlyingIterator.Value())
		if err != nil {
			ei.currentValue = nil
			return false
		}
		ei.currentValue = value
		return true
	}
	return false
}

// Last 移动到最后一个元素
func (ei *encryptedIterator) Last() bool {
	if ei.underlyingIterator == nil {
		return false
	}
	if ei.underlyingIterator.Last() {
		// 解密当前值
		value, err := ei.encryptor.Decrypt(ei.underlyingIterator.Value())
		if err != nil {
			ei.currentValue = nil
			return false
		}
		ei.currentValue = value
		return true
	}
	return false
}

// Seek 移动到大于等于指定key的位置
func (ei *encryptedIterator) Seek(key []byte) bool {
	if ei.underlyingIterator == nil {
		return false
	}
	if ei.underlyingIterator.Seek(key) {
		// 解密当前值
		value, err := ei.encryptor.Decrypt(ei.underlyingIterator.Value())
		if err != nil {
			ei.currentValue = nil
			return false
		}
		ei.currentValue = value
		return true
	}
	return false
}

// Next 移动到下一个元素
func (ei *encryptedIterator) Next() bool {
	if ei.underlyingIterator == nil {
		return false
	}
	if ei.underlyingIterator.Next() {
		// 解密当前值
		value, err := ei.encryptor.Decrypt(ei.underlyingIterator.Value())
		if err != nil {
			ei.currentValue = nil
			return false
		}
		ei.currentValue = value
		return true
	}
	return false
}

// Prev 移动到前一个元素
func (ei *encryptedIterator) Prev() bool {
	if ei.underlyingIterator == nil {
		return false
	}
	if ei.underlyingIterator.Prev() {
		// 解密当前值
		value, err := ei.encryptor.Decrypt(ei.underlyingIterator.Value())
		if err != nil {
			ei.currentValue = nil
			return false
		}
		ei.currentValue = value
		return true
	}
	return false
}

// Key 获取当前元素的key
func (ei *encryptedIterator) Key() []byte {
	if ei.underlyingIterator == nil {
		return nil
	}
	return ei.underlyingIterator.Key()
}

// Value 获取当前元素的value
func (ei *encryptedIterator) Value() []byte {
	return ei.currentValue
}

// Valid 检查迭代器是否有效
func (ei *encryptedIterator) Valid() bool {
	return ei.underlyingIterator != nil && ei.underlyingIterator.Valid() && ei.currentValue != nil
}

// Release 释放迭代器资源
func (ei *encryptedIterator) Release() {
	if ei.underlyingIterator != nil {
		ei.underlyingIterator.Release()
	}
	ei.currentValue = nil
}

// encryptedSnapshot 加密快照

type encryptedSnapshot struct {
	underlyingSnapshot Snapshot
	encryptor          Encryptor
}

// Get 从快照中获取指定key的值
func (es *encryptedSnapshot) Get(key []byte) ([]byte, error) {
	// 从底层快照获取加密数据
	encryptedValue, err := es.underlyingSnapshot.Get(key)
	if err != nil {
		return nil, err
	}

	// 解密数据
	return es.encryptor.Decrypt(encryptedValue)
}

// Iterator 从快照中创建迭代器
func (es *encryptedSnapshot) Iterator(start, limit []byte) Iterator {
	// 返回加密迭代器
	return &encryptedIterator{
		underlyingIterator: es.underlyingSnapshot.Iterator(start, limit),
		encryptor:          es.encryptor,
	}
}

// Release 释放快照资源
func (es *encryptedSnapshot) Release() error {
	return es.underlyingSnapshot.Release()
}
