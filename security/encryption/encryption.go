package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encryptor 加密器接口
type Encryptor interface {
	// Encrypt 加密数据
	Encrypt(plaintext []byte) ([]byte, error)

	// Decrypt 解密数据
	Decrypt(ciphertext []byte) ([]byte, error)

	// Algorithm 获取加密算法名称
	Algorithm() string
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Enabled    bool   `json:"enabled"`
	Algorithm  string `json:"algorithm"`
	MasterKey  []byte `json:"master_key"`
	Password   string `json:"password"`
	Salt       []byte `json:"salt"`
	Iterations int    `json:"iterations"`
}

// AESGCMEncryptor AES-GCM加密器
type AESGCMEncryptor struct {
	key        []byte
	block      cipher.Block
	algorithm  string
}

// NewAESGCMEncryptor 创建AES-GCM加密器
func NewAESGCMEncryptor(key []byte) (*AESGCMEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid key length, must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &AESGCMEncryptor{
		key:       key,
		block:     block,
		algorithm: "AES-GCM",
	}, nil
}

// Encrypt 加密数据
func (e *AESGCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func (e *AESGCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Algorithm 获取加密算法名称
func (e *AESGCMEncryptor) Algorithm() string {
	return e.algorithm
}

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// KeyToBase64 将密钥转换为Base64字符串
func KeyToBase64(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// Base64ToKey 将Base64字符串转换为密钥
func Base64ToKey(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
