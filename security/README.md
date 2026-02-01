# Security Package

安全包提供了sfsDb的安全系统，包括认证、授权、加密和审计功能。

## 功能特性

- **认证系统**：基于JWT的用户认证
- **授权系统**：基于角色的访问控制（RBAC）
- **数据加密**：AES-GCM数据加密
- **审计日志**：详细的操作审计记录
- **自定义扩展**：支持自定义认证和授权策略

## 核心模块

### 1. Auth

认证模块，负责用户认证和令牌管理。

#### 主要接口

```go
// Authenticator 认证器接口
type Authenticator interface {
	// Authenticate 认证用户
	Authenticate(username, password string) (*User, error)

	// GenerateToken 生成认证令牌
	GenerateToken(user *User) (string, error)

	// ValidateToken 验证令牌
	ValidateToken(token string) (*Claims, error)

	// GetUserByID 根据ID获取用户
	GetUserByID(userID string) (*User, error)
}
```

### 2. Access

访问控制模块，负责权限管理和访问控制。

#### 主要接口

```go
// AccessControl 访问控制接口
type AccessControl interface {
	// CheckPermission 检查用户是否有指定权限
	CheckPermission(userID, permissionID string) (bool, error)

	// CheckRole 检查用户是否有指定角色
	CheckRole(userID, roleID string) (bool, error)

	// GetUserRoles 获取用户的所有角色
	GetUserRoles(userID string) ([]Role, error)

	// GetRolePermissions 获取角色的所有权限
	GetRolePermissions(roleID string) ([]Permission, error)

	// AssignRole 为用户分配角色
	AssignRole(userID, roleID string) error

	// RevokeRole 撤销用户的角色
	RevokeRole(userID, roleID string) error
}
```

### 3. Encryption

加密模块，负责数据加密和解密。

#### 主要接口

```go
// Encryptor 加密器接口
type Encryptor interface {
	// Encrypt 加密数据
	Encrypt(plaintext []byte) ([]byte, error)

	// Decrypt 解密数据
	Decrypt(ciphertext []byte) ([]byte, error)

	// Algorithm 获取加密算法名称
	Algorithm() string
}
```

### 4. Audit

审计模块，负责记录和管理审计日志。

#### 主要接口

```go
// AuditLogger 审计日志记录器接口
type AuditLogger interface {
	// LogEvent 记录审计事件
	LogEvent(event AuditEvent)

	// GetEvents 获取审计事件
	GetEvents(filter map[string]string, limit, offset int) ([]AuditEvent, error)

	// GetEventByID 根据ID获取审计事件
	GetEventByID(eventID string) (*AuditEvent, error)

	// SearchEvents 搜索审计事件
	SearchEvents(query string, limit, offset int) ([]AuditEvent, error)
}
```

## 使用示例

### 1. 初始化安全系统

```go
import (
	"github.com/liaoran123/sfsDb/security"
	"github.com/liaoran123/sfsDb/security/encryption"
)

// 创建安全配置
config := security.SecurityConfig{
	Auth: auth.AuthConfig{
		Enabled:     true,
		JWTSecret:   "your_jwt_secret",
		TokenExpiry: 24 * time.Hour,
		Users: []auth.User{
			{
				ID:       "1",
				Username: "admin",
				Password: "password",
				Roles:    []string{"admin"},
			},
		},
	},
	Access: access.AccessConfig{
		Enabled: true,
		Roles: []access.Role{
			{
				ID:          "admin",
				Name:        "管理员",
				Description: "系统管理员",
				Permissions: []access.Permission{
					{
						ID:          "all",
						Name:        "所有权限",
						Description: "所有操作的权限",
					},
				},
			},
		},
		UserRoles: map[string][]string{
			"1": {"admin"},
		},
	},
	Encryption: encryption.EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-GCM",
		MasterKey: []byte("your_encryption_key"), // 16, 24, or 32 bytes
	},
	Audit: audit.AuditConfig{
		Enabled:    true,
		MaxEvents:  10000,
		EventTypes: []string{"auth", "access", "data"},
	},
}

// 创建安全系统
secure, err := security.NewSecurity(config)
if err != nil {
	panic(err)
}

// 启动安全系统
secure.Start()
```

### 2. 用户认证

```go
// 认证用户
user, err := secure.Auth().Authenticate("admin", "password")
if err != nil {
	// 认证失败
	return
}

// 生成令牌
token, err := secure.Auth().GenerateToken(user)
if err != nil {
	// 生成令牌失败
	return
}

// 验证令牌
claims, err := secure.Auth().ValidateToken(token)
if err != nil {
	// 令牌无效
	return
}
```

### 3. 访问控制

```go
// 检查权限
hasPermission, err := secure.Access().CheckPermission(user.ID, "all")
if err != nil {
	// 检查权限失败
	return
}

if !hasPermission {
	// 没有权限
	return
}

// 为用户分配角色
err = secure.Access().AssignRole(user.ID, "admin")
if err != nil {
	// 分配角色失败
	return
}
```

### 4. 数据加密

```go
// 加密数据
plaintext := []byte("sensitive data")
ciphertext, err := secure.Encryption().Encrypt(plaintext)
if err != nil {
	// 加密失败
	return
}

// 解密数据
decrypted, err := secure.Encryption().Decrypt(ciphertext)
if err != nil {
	// 解密失败
	return
}
```

### 5. 审计日志

```go
// 记录审计事件
event := audit.NewAuditEvent(
	"auth",
	user.ID,
	user.Username,
	"login",
	"system",
	"用户登录",
	true,
	"127.0.0.1",
)
secure.Audit().LogEvent(event)

// 搜索审计事件	events, err := secure.Audit().SearchEvents("login", 10, 0)
if err != nil {
	// 搜索失败
	return
}
```

## 配置选项

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| auth.enabled | bool | false | 是否启用认证 |
| auth.jwt_secret | string | "default_secret" | JWT密钥 |
| auth.token_expiry | time.Duration | 24h | 令牌过期时间 |
| access.enabled | bool | false | 是否启用访问控制 |
| encryption.enabled | bool | false | 是否启用加密 |
| encryption.algorithm | string | "AES-GCM" | 加密算法 |
| encryption.master_key | []byte | []byte{} | 加密密钥 |
| audit.enabled | bool | false | 是否启用审计 |
| audit.max_events | int | 10000 | 最大审计事件数 |

## 最佳实践

1. **使用强密钥**：为JWT和加密使用强密钥
2. **定期轮换密钥**：定期轮换加密密钥
3. **最小权限原则**：为用户分配最小必要的权限
4. **启用审计**：在生产环境中启用审计日志
5. **安全存储配置**：安全存储包含敏感信息的配置

## 故障排查

### 常见问题

1. **认证失败**
   - 检查用户名和密码是否正确
   - 检查用户是否存在

2. **令牌无效**
   - 检查令牌是否过期
   - 检查JWT密钥是否正确

3. **权限不足**
   - 检查用户是否有正确的角色
   - 检查角色是否有正确的权限

4. **加密失败**
   - 检查密钥长度是否正确（16, 24, 或 32字节）
   - 检查数据是否有效

## 版本历史

- **v1.0.0**：初始版本，支持基本认证和授权
- **v1.1.0**：添加数据加密和审计功能
- **v1.2.0**：优化性能和安全性
