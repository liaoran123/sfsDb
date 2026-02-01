package security

import (
	"time"

	"github.com/liaoran123/sfsDb/security/access"
	"github.com/liaoran123/sfsDb/security/audit"
	"github.com/liaoran123/sfsDb/security/auth"
	"github.com/liaoran123/sfsDb/security/encryption"
)

// Security 安全系统接口
type Security interface {
	// 认证相关
	Auth() auth.Authenticator

	// 访问控制相关
	Access() access.AccessControl

	// 加密相关
	Encryption() encryption.Encryptor

	// 审计相关
	Audit() audit.AuditLogger

	// 启动安全系统
	Start() error

	// 停止安全系统
	Stop() error
}

// SecurityConfig 安全系统配置
type SecurityConfig struct {
	Auth       auth.AuthConfig             `json:"auth"`
	Access     access.AccessConfig         `json:"access"`
	Encryption encryption.EncryptionConfig `json:"encryption"`
	Audit      audit.AuditConfig           `json:"audit"`
}

// Secure 安全系统实现
type Secure struct {
	config        SecurityConfig
	authenticator auth.Authenticator
	accessControl access.AccessControl
	encryptor     encryption.Encryptor
	auditLogger   audit.AuditLogger
}

// NewSecurity 创建安全系统
func NewSecurity(config SecurityConfig) (Security, error) {
	// 创建认证器
	authenticator := auth.NewJWTAuthenticator(config.Auth)

	// 创建访问控制器
	accessControl := access.NewRBACAccessControl(config.Access)

	// 创建加密器
	encryptor, err := encryption.NewAESGCMEncryptor(config.Encryption.MasterKey)
	if err != nil {
		return nil, err
	}

	// 创建审计日志记录器
	auditLogger := audit.NewConsoleAuditLogger(config.Audit)

	return &Secure{
		config:        config,
		authenticator: authenticator,
		accessControl: accessControl,
		encryptor:     encryptor,
		auditLogger:   auditLogger,
	}, nil
}

// Auth 获取认证器
func (s *Secure) Auth() auth.Authenticator {
	return s.authenticator
}

// Access 获取访问控制器
func (s *Secure) Access() access.AccessControl {
	return s.accessControl
}

// Encryption 获取加密器
func (s *Secure) Encryption() encryption.Encryptor {
	return s.encryptor
}

// Audit 获取审计日志记录器
func (s *Secure) Audit() audit.AuditLogger {
	return s.auditLogger
}

// Start 启动安全系统
func (s *Secure) Start() error {
	// 安全系统启动逻辑
	return nil
}

// Stop 停止安全系统
func (s *Secure) Stop() error {
	// 安全系统停止逻辑
	return nil
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Auth: auth.AuthConfig{
			Enabled:     false,
			JWTSecret:   "default_secret",
			TokenExpiry: 24 * time.Hour,
			Users:       []auth.User{},
		},
		Access: access.AccessConfig{
			Enabled:   false,
			Roles:     []access.Role{},
			UserRoles: make(map[string][]string),
		},
		Encryption: encryption.EncryptionConfig{
			Enabled:   false,
			Algorithm: "AES-GCM",
			MasterKey: []byte{},
		},
		Audit: audit.AuditConfig{
			Enabled:    false,
			LogFile:    "",
			MaxEvents:  10000,
			EventTypes: []string{},
		},
	}
}
