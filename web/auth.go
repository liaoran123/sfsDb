package web

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Role 定义用户角色
type Role string

const (
	// RoleAdmin 管理员角色，拥有所有权限
	RoleAdmin Role = "admin"
	// RoleOperator 运维人员角色，拥有操作权限
	RoleOperator Role = "operator"
	// RoleViewer 查看人员角色，只拥有查看权限
	RoleViewer Role = "viewer"
)

// Permission 定义权限
type Permission string

const (
	// PermissionRead 读取权限
	PermissionRead Permission = "read"
	// PermissionWrite 写入权限
	PermissionWrite Permission = "write"
	// PermissionAdmin 管理权限
	PermissionAdmin Permission = "admin"
)

// 角色权限映射
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionRead,
		PermissionWrite,
		PermissionAdmin,
	},
	RoleOperator: {
		PermissionRead,
		PermissionWrite,
	},
	RoleViewer: {
		PermissionRead,
	},
}

// APIKey 定义API密钥结构
type APIKey struct {
	Key  string
	Role Role
}

// AuthManager 认证管理器
type AuthManager struct {
	apiKeys map[string]Role
	mutex   sync.RWMutex
}

// NewAuthManager 创建认证管理器
func NewAuthManager() *AuthManager {
	return &AuthManager{
		apiKeys: make(map[string]Role),
	}
}

// GenerateAPIKey 生成新的API密钥
func (am *AuthManager) GenerateAPIKey(role Role) (string, error) {
	// 生成32字节的随机密钥
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return "", err
	}

	// 转换为十六进制字符串
	key := hex.EncodeToString(keyBytes)

	// 存储API密钥和角色
	am.mutex.Lock()
	am.apiKeys[key] = role
	am.mutex.Unlock()

	return key, nil
}

// ValidateAPIKey 验证API密钥
func (am *AuthManager) ValidateAPIKey(key string) (Role, bool) {
	am.mutex.RLock()
	role, exists := am.apiKeys[key]
	am.mutex.RUnlock()

	return role, exists
}

// RevokeAPIKey 撤销API密钥
func (am *AuthManager) RevokeAPIKey(key string) {
	am.mutex.Lock()
	delete(am.apiKeys, key)
	am.mutex.Unlock()
}

// AddAPIKey 添加API密钥
func (am *AuthManager) AddAPIKey(key string, role Role) {
	am.mutex.Lock()
	am.apiKeys[key] = role
	am.mutex.Unlock()
}

// HasPermission 检查角色是否有指定权限
func (am *AuthManager) HasPermission(role Role, permission Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// AuthMiddleware 认证中间件
func AuthMiddleware(authManager *AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取API密钥
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication required",
				"message": "API key is required in X-API-Key header",
			})
			c.Abort()
			return
		}

		// 验证API密钥
		role, valid := authManager.ValidateAPIKey(apiKey)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid api key",
				"message": "The provided API key is invalid or has been revoked",
			})
			c.Abort()
			return
		}

		// 将角色存储到上下文
		c.Set("role", role)
		c.Next()
	}
}

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(authManager *AuthManager, requiredPermission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取角色
		roleValue, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication required",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		role, ok := roleValue.(Role)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid role",
				"message": "Invalid role in context",
			})
			c.Abort()
			return
		}

		// 检查权限
		if !authManager.HasPermission(role, requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "permission denied",
				"message": "You don't have permission to perform this action",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
