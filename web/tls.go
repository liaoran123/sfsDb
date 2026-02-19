package web

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
)

// TLSConfig TLS配置
type TLSConfig struct {
	Enable   bool   `json:"enable"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// DefaultTLSConfig 默认TLS配置
func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		Enable:   true,
		CertFile: filepath.Join("certs", "server.crt"),
		KeyFile:  filepath.Join("certs", "server.key"),
	}
}

// LoadTLSConfig 加载TLS配置
func LoadTLSConfig(config TLSConfig) (*tls.Config, error) {
	// 检查证书文件是否存在
	if !fileExists(config.CertFile) {
		return nil, fmt.Errorf("certificate file not found: %s", config.CertFile)
	}

	// 检查密钥文件是否存在
	if !fileExists(config.KeyFile) {
		return nil, fmt.Errorf("key file not found: %s", config.KeyFile)
	}

	// 加载证书
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// 使用安全的密码套件
		CipherSuites: []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		// 使用安全的TLS版本
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	return tlsConfig, nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureCertDir 确保证书目录存在
func EnsureCertDir(certDir string) error {
	return os.MkdirAll(certDir, 0755)
}
