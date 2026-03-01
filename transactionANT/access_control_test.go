package transactionANT

import (
	"testing"
	"time"
)

// TestAccessControl 测试访问控制功能
func TestAccessControl(t *testing.T) {
	// 初始化访问控制管理器
	err := InitAccessControl()
	if err != nil {
		t.Fatalf("Failed to initialize access control: %v", err)
	}

	// 测试默认角色和权限
	acm := GetAccessControlManager()
	if acm == nil {
		t.Fatal("Access control manager is nil")
	}

	// 测试获取默认角色
	adminRole, err := acm.GetRole("role:admin")
	if err != nil {
		t.Errorf("Failed to get admin role: %v", err)
	}
	if adminRole == nil {
		t.Error("Admin role is nil")
	}

	userRole, err := acm.GetRole("role:user")
	if err != nil {
		t.Errorf("Failed to get user role: %v", err)
	}
	if userRole == nil {
		t.Error("User role is nil")
	}

	// 测试获取默认用户
	adminUser, err := acm.GetUser("user:admin")
	if err != nil {
		t.Errorf("Failed to get admin user: %v", err)
	}
	if adminUser == nil {
		t.Error("Admin user is nil")
	}

	// 测试权限检查
	hasPermission, err := acm.CheckPermission("user:admin", ResourceTypeTable, "*", PermissionRead)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("Admin should have read permission")
	}

	hasPermission, err = acm.CheckPermission("user:admin", ResourceTypeTable, "*", PermissionWrite)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("Admin should have write permission")
	}

	// 测试添加新用户
	newUser := &User{
		ID:        "user:test",
		Username:  "test",
		Password:  "test123",
		Roles:     []*Role{userRole},
		CreatedAt: adminUser.CreatedAt,
	}
	err = acm.AddUser(newUser)
	if err != nil {
		t.Errorf("Failed to add new user: %v", err)
	}

	// 测试新用户的权限
	hasPermission, err = acm.CheckPermission("user:test", ResourceTypeTable, "*", PermissionRead)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("Test user should have read permission")
	}

	hasPermission, err = acm.CheckPermission("user:test", ResourceTypeTable, "*", PermissionWrite)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if hasPermission {
		t.Error("Test user should not have write permission")
	}

	// 测试分配角色
	err = acm.AssignRoleToUser("user:test", "role:admin")
	if err != nil {
		t.Errorf("Failed to assign role: %v", err)
	}

	// 测试权限变更
	hasPermission, err = acm.CheckPermission("user:test", ResourceTypeTable, "*", PermissionWrite)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("Test user should have write permission after role assignment")
	}

	// 测试撤销角色
	err = acm.RevokeRoleFromUser("user:test", "role:admin")
	if err != nil {
		t.Errorf("Failed to revoke role: %v", err)
	}

	// 测试权限变更
	hasPermission, err = acm.CheckPermission("user:test", ResourceTypeTable, "*", PermissionWrite)
	if err != nil {
		t.Errorf("Failed to check permission: %v", err)
	}
	if hasPermission {
		t.Error("Test user should not have write permission after role revocation")
	}

	t.Log("Access control tests passed")
}

// TestSessionManagement 测试会话管理功能
func TestSessionManagement(t *testing.T) {
	// 初始化会话管理器
	err := InitSessionManager(30 * 60 * time.Second) // 30分钟超时
	if err != nil {
		t.Fatalf("Failed to initialize session manager: %v", err)
	}

	sm := GetSessionManager()
	if sm == nil {
		t.Fatal("Session manager is nil")
	}

	// 测试创建会话
	session, err := sm.CreateSession("user:admin")
	if err != nil {
		t.Errorf("Failed to create session: %v", err)
	}
	if session == nil {
		t.Error("Session is nil")
	}

	// 测试获取会话
	retrievedSession, err := sm.GetSession(session.ID)
	if err != nil {
		t.Errorf("Failed to get session: %v", err)
	}
	if retrievedSession == nil {
		t.Error("Retrieved session is nil")
	}
	if retrievedSession.UserID != "user:admin" {
		t.Error("Session user ID mismatch")
	}

	// 测试验证会话
	valid, err := sm.ValidateSession(session.ID)
	if err != nil {
		t.Errorf("Failed to validate session: %v", err)
	}
	if !valid {
		t.Error("Session should be valid")
	}

	// 测试从会话获取用户ID
	userID, err := sm.GetUserIDFromSession(session.ID)
	if err != nil {
		t.Errorf("Failed to get user ID from session: %v", err)
	}
	if userID != "user:admin" {
		t.Error("User ID mismatch")
	}

	// 测试刷新会话
	err = sm.RefreshSession(session.ID)
	if err != nil {
		t.Errorf("Failed to refresh session: %v", err)
	}

	// 测试删除会话
	err = sm.DeleteSession(session.ID)
	if err != nil {
		t.Errorf("Failed to delete session: %v", err)
	}

	// 测试删除后的会话
	_, err = sm.GetSession(session.ID)
	if err == nil {
		t.Error("Session should not exist after deletion")
	}

	t.Log("Session management tests passed")
}
