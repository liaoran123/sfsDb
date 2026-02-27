package transactionLockANT

import (
	"fmt"
	"sync"
	"time"
)

// 权限类型常量
const (
	// 操作类型权限
	PermissionRead   = "READ"
	PermissionWrite  = "WRITE"
	PermissionDelete = "DELETE"
	PermissionCreate = "CREATE"

	// 资源类型
	ResourceTypeTable  = "TABLE"
	ResourceTypeField  = "FIELD"
	ResourceTypeSystem = "SYSTEM"
)

// Permission 权限结构体
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ResourceType string    `json:"resourceType"`
	ResourceID  string    `json:"resourceId"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Role 角色结构体
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []*Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"createdAt"`
}

// User 用户结构体
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // 实际应用中应该存储哈希值
	Roles     []*Role   `json:"roles"`
	CreatedAt time.Time `json:"createdAt"`
}

// AccessControlManager 访问控制管理器
type AccessControlManager struct {
	users      map[string]*User
	roles      map[string]*Role
	permissions map[string]*Permission
	mutex      sync.RWMutex
}

// NewAccessControlManager 创建访问控制管理器
func NewAccessControlManager() *AccessControlManager {
	return &AccessControlManager{
		users:       make(map[string]*User),
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
	}
}

// GlobalAccessControlManager 全局访问控制管理器
var GlobalAccessControlManager *AccessControlManager

// InitAccessControl 初始化访问控制管理器
func InitAccessControl() error {
	GlobalAccessControlManager = NewAccessControlManager()
	// 初始化默认角色和权限
	err := GlobalAccessControlManager.initDefaultRoles()
	if err == nil {
		// 通知事务模块访问控制已启用
		SetAccessControlEnabled(true)
	}
	return err
}

// GetAccessControlManager 获取访问控制管理器
func GetAccessControlManager() *AccessControlManager {
	return GlobalAccessControlManager
}

// 通知事务模块访问控制已启用
func SetAccessControlEnabled(enabled bool) {
	// 调用 transaction.go 中的函数
	setAccessControlEnabled(enabled)
}

// initDefaultRoles 初始化默认角色和权限
func (acm *AccessControlManager) initDefaultRoles() error {
	// 创建默认权限
	defaultPermissions := []*Permission{
		{
			ID:          "perm:read:table",
			Name:        "Read Table",
			Description: "允许读取表数据",
			ResourceType: ResourceTypeTable,
			ResourceID:  "*",
			Action:      PermissionRead,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "perm:write:table",
			Name:        "Write Table",
			Description: "允许写入表数据",
			ResourceType: ResourceTypeTable,
			ResourceID:  "*",
			Action:      PermissionWrite,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "perm:delete:table",
			Name:        "Delete Table",
			Description: "允许删除表数据",
			ResourceType: ResourceTypeTable,
			ResourceID:  "*",
			Action:      PermissionDelete,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "perm:create:table",
			Name:        "Create Table",
			Description: "允许创建表",
			ResourceType: ResourceTypeSystem,
			ResourceID:  "*",
			Action:      PermissionCreate,
			CreatedAt:   time.Now(),
		},
	}

	// 创建默认角色
	defaultRoles := []*Role{
		{
			ID:          "role:admin",
			Name:        "Administrator",
			Description: "系统管理员，拥有所有权限",
			Permissions: defaultPermissions,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "role:user",
			Name:        "User",
			Description: "普通用户，只拥有读取权限",
			Permissions: []*Permission{defaultPermissions[0]},
			CreatedAt:   time.Now(),
		},
	}

	// 保存默认权限和角色
	for _, perm := range defaultPermissions {
		acm.permissions[perm.ID] = perm
	}

	for _, role := range defaultRoles {
		acm.roles[role.ID] = role
	}

	// 创建默认管理员用户
	adminUser := &User{
		ID:        "user:admin",
		Username:  "admin",
		Password:  "admin123", // 实际应用中应该使用哈希
		Roles:     []*Role{defaultRoles[0]},
		CreatedAt: time.Now(),
	}
	acm.users[adminUser.ID] = adminUser

	return nil
}

// AddUser 添加用户
func (acm *AccessControlManager) AddUser(user *User) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	if _, exists := acm.users[user.ID]; exists {
		return fmt.Errorf("user already exists")
	}

	acm.users[user.ID] = user
	return nil
}

// GetUser 获取用户
func (acm *AccessControlManager) GetUser(userID string) (*User, error) {
	acm.mutex.RLock()
	defer acm.mutex.RUnlock()

	user, exists := acm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// AddRole 添加角色
func (acm *AccessControlManager) AddRole(role *Role) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	if _, exists := acm.roles[role.ID]; exists {
		return fmt.Errorf("role already exists")
	}

	acm.roles[role.ID] = role
	return nil
}

// GetRole 获取角色
func (acm *AccessControlManager) GetRole(roleID string) (*Role, error) {
	acm.mutex.RLock()
	defer acm.mutex.RUnlock()

	role, exists := acm.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role not found")
	}

	return role, nil
}

// AddPermission 添加权限
func (acm *AccessControlManager) AddPermission(permission *Permission) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	if _, exists := acm.permissions[permission.ID]; exists {
		return fmt.Errorf("permission already exists")
	}

	acm.permissions[permission.ID] = permission
	return nil
}

// GetPermission 获取权限
func (acm *AccessControlManager) GetPermission(permissionID string) (*Permission, error) {
	acm.mutex.RLock()
	defer acm.mutex.RUnlock()

	permission, exists := acm.permissions[permissionID]
	if !exists {
		return nil, fmt.Errorf("permission not found")
	}

	return permission, nil
}

// AssignRoleToUser 为用户分配角色
func (acm *AccessControlManager) AssignRoleToUser(userID, roleID string) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	user, exists := acm.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	role, exists := acm.roles[roleID]
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 检查角色是否已经分配
	for _, r := range user.Roles {
		if r.ID == roleID {
			return fmt.Errorf("role already assigned to user")
		}
	}

	user.Roles = append(user.Roles, role)
	return nil
}

// RevokeRoleFromUser 从用户撤销角色
func (acm *AccessControlManager) RevokeRoleFromUser(userID, roleID string) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	user, exists := acm.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// 查找并移除角色
	for i, r := range user.Roles {
		if r.ID == roleID {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("role not assigned to user")
}

// AssignPermissionToRole 为角色分配权限
func (acm *AccessControlManager) AssignPermissionToRole(roleID, permissionID string) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	role, exists := acm.roles[roleID]
	if !exists {
		return fmt.Errorf("role not found")
	}

	permission, exists := acm.permissions[permissionID]
	if !exists {
		return fmt.Errorf("permission not found")
	}

	// 检查权限是否已经分配
	for _, p := range role.Permissions {
		if p.ID == permissionID {
			return fmt.Errorf("permission already assigned to role")
		}
	}

	role.Permissions = append(role.Permissions, permission)
	return nil
}

// RevokePermissionFromRole 从角色撤销权限
func (acm *AccessControlManager) RevokePermissionFromRole(roleID, permissionID string) error {
	acm.mutex.Lock()
	defer acm.mutex.Unlock()

	role, exists := acm.roles[roleID]
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 查找并移除权限
	for i, p := range role.Permissions {
		if p.ID == permissionID {
			role.Permissions = append(role.Permissions[:i], role.Permissions[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("permission not assigned to role")
}

// CheckPermission 检查用户是否具有指定权限
func (acm *AccessControlManager) CheckPermission(userID, resourceType, resourceID, action string) (bool, error) {
	acm.mutex.RLock()
	defer acm.mutex.RUnlock()

	user, exists := acm.users[userID]
	if !exists {
		return false, fmt.Errorf("user not found")
	}

	// 检查用户的所有角色
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			// 检查权限是否匹配
			if permission.ResourceType == resourceType && 
			   (permission.ResourceID == "*" || permission.ResourceID == resourceID) && 
			   permission.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckTablePermission 检查用户对表的权限
func (acm *AccessControlManager) CheckTablePermission(userID, tableName, action string) (bool, error) {
	return acm.CheckPermission(userID, ResourceTypeTable, tableName, action)
}

// CheckFieldPermission 检查用户对字段的权限
func (acm *AccessControlManager) CheckFieldPermission(userID, tableName, fieldName, action string) (bool, error) {
	resourceID := fmt.Sprintf("%s.%s", tableName, fieldName)
	return acm.CheckPermission(userID, ResourceTypeField, resourceID, action)
}
