package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// User 用户信息
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"-"` // 密码不序列化
	Roles    []string `json:"roles"`
}

// Claims 令牌声明
type Claims struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Subject   string    `json:"subject"`
}

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

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled      bool          `json:"enabled"`
	JWTSecret    string        `json:"jwt_secret"`
	TokenExpiry  time.Duration `json:"token_expiry"`
	Users        []User        `json:"users"`
}

// JWTAuthenticator JWT认证器
type JWTAuthenticator struct {
	config  AuthConfig
	users   map[string]*User
}

// NewJWTAuthenticator 创建JWT认证器
func NewJWTAuthenticator(config AuthConfig) *JWTAuthenticator {
	users := make(map[string]*User)
	for i := range config.Users {
		user := config.Users[i]
		users[user.Username] = &user
		users[user.ID] = &user
	}

	return &JWTAuthenticator{
		config: config,
		users:  users,
	}
}

// Authenticate 认证用户
func (a *JWTAuthenticator) Authenticate(username, password string) (*User, error) {
	user, exists := a.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	if user.Password != password {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// GenerateToken 生成认证令牌
func (a *JWTAuthenticator) GenerateToken(user *User) (string, error) {
	expirationTime := time.Now().Add(a.config.TokenExpiry)
	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Roles:     user.Roles,
		ExpiresAt: expirationTime,
		IssuedAt:  time.Now(),
		Subject:   user.ID,
	}

	// 序列化claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// 生成签名
	signature := a.generateSignature(claimsJSON)

	// 组合令牌：claims.base64.signature.base64
	claimsEncoded := base64.StdEncoding.EncodeToString(claimsJSON)
	signatureEncoded := base64.StdEncoding.EncodeToString(signature)
	tokenString := claimsEncoded + "." + signatureEncoded

	return tokenString, nil
}

// ValidateToken 验证令牌
func (a *JWTAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
	// 分割令牌
	parts := make([]string, 0, 2)
	for i := 0; i < len(tokenString); i++ {
		if tokenString[i] == '.' {
			parts = append(parts, tokenString[:i])
			parts = append(parts, tokenString[i+1:])
			break
		}
	}

	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}

	// 解码claims
	claimsJSON, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token claims")
	}

	// 解码签名
	signature, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	// 验证签名
	expectedSignature := a.generateSignature(claimsJSON)
	if !hmac.Equal(signature, expectedSignature) {
		return nil, errors.New("invalid token signature")
	}

	// 解析claims
	claims := &Claims{}
	if err := json.Unmarshal(claimsJSON, claims); err != nil {
		return nil, errors.New("invalid token claims")
	}

	// 验证过期时间
	if time.Now().After(claims.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// GetUserByID 根据ID获取用户
func (a *JWTAuthenticator) GetUserByID(userID string) (*User, error) {
	user, exists := a.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// generateSignature 生成签名
func (a *JWTAuthenticator) generateSignature(data []byte) []byte {
	h := hmac.New(sha256.New, []byte(a.config.JWTSecret))
	h.Write(data)
	return h.Sum(nil)
}
