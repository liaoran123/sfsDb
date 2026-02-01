package auth

import (
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	// 创建认证配置
	config := AuthConfig{
		Enabled:     true,
		JWTSecret:   "test_secret",
		TokenExpiry: 1 * time.Hour,
		Users: []User{
			{
				ID:       "1",
				Username: "admin",
				Password: "password",
				Roles:    []string{"admin"},
			},
			{
				ID:       "2",
				Username: "user",
				Password: "userpass",
				Roles:    []string{"user"},
			},
		},
	}

	// 创建认证器
	a := NewJWTAuthenticator(config)

	// 测试认证
	t.Run("Authenticate", func(t *testing.T) {
		// 测试成功认证
		user, err := a.Authenticate("admin", "password")
		if err != nil {
			t.Errorf("Expected successful authentication, got error: %v", err)
		}
		if user.Username != "admin" {
			t.Errorf("Expected username admin, got %s", user.Username)
		}

		// 测试失败认证（密码错误）
		_, err = a.Authenticate("admin", "wrongpassword")
		if err == nil {
			t.Errorf("Expected authentication error for wrong password, got nil")
		}

		// 测试失败认证（用户不存在）
		_, err = a.Authenticate("nonexistent", "password")
		if err == nil {
			t.Errorf("Expected authentication error for nonexistent user, got nil")
		}
	})

	// 测试令牌生成和验证
	t.Run("Token", func(t *testing.T) {
		// 获取用户
		user, err := a.Authenticate("admin", "password")
		if err != nil {
			t.Fatalf("Failed to authenticate user: %v", err)
		}

		// 生成令牌
		token, err := a.GenerateToken(user)
		if err != nil {
			t.Errorf("Expected token generation success, got error: %v", err)
		}

		// 验证令牌
		claims, err := a.ValidateToken(token)
		if err != nil {
			t.Errorf("Expected token validation success, got error: %v", err)
		}
		if claims.Username != "admin" {
			t.Errorf("Expected username admin in claims, got %s", claims.Username)
		}
		if claims.UserID != "1" {
			t.Errorf("Expected user ID 1 in claims, got %s", claims.UserID)
		}
		if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
			t.Errorf("Expected role admin in claims, got %v", claims.Roles)
		}
	})

	// 测试获取用户
	t.Run("GetUserByID", func(t *testing.T) {
		// 获取用户
		user, err := a.GetUserByID("1")
		if err != nil {
			t.Errorf("Expected user retrieval success, got error: %v", err)
		}
		if user.Username != "admin" {
			t.Errorf("Expected username admin, got %s", user.Username)
		}

		// 测试获取不存在的用户
		_, err = a.GetUserByID("999")
		if err == nil {
			t.Errorf("Expected error for nonexistent user, got nil")
		}
	})
}
