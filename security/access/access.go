package access

import (
	"errors"
)

// Permission 权限
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Role 角色
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

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

// AccessConfig 访问控制配置
type AccessConfig struct {
	Enabled      bool              `json:"enabled"`
	Roles        []Role            `json:"roles"`
	UserRoles    map[string][]string `json:"user_roles"` // userID -> []roleID
}

// RBACAccessControl 基于角色的访问控制
type RBACAccessControl struct {
	config        AccessConfig
	roles         map[string]Role
	userRoles     map[string][]string
	rolePermissions map[string]map[string]bool
}

// NewRBACAccessControl 创建RBAC访问控制
func NewRBACAccessControl(config AccessConfig) *RBACAccessControl {
	roles := make(map[string]Role)
	rolePermissions := make(map[string]map[string]bool)

	for _, role := range config.Roles {
		roles[role.ID] = role
		permissions := make(map[string]bool)
		for _, perm := range role.Permissions {
			permissions[perm.ID] = true
		}
		rolePermissions[role.ID] = permissions
	}

	return &RBACAccessControl{
		config:        config,
		roles:         roles,
		userRoles:     config.UserRoles,
		rolePermissions: rolePermissions,
	}
}

// CheckPermission 检查用户是否有指定权限
func (ac *RBACAccessControl) CheckPermission(userID, permissionID string) (bool, error) {
	roles, exists := ac.userRoles[userID]
	if !exists {
		return false, nil
	}

	for _, roleID := range roles {
		permissions, exists := ac.rolePermissions[roleID]
		if !exists {
			continue
		}

		if permissions[permissionID] {
			return true, nil
		}
	}

	return false, nil
}

// CheckRole 检查用户是否有指定角色
func (ac *RBACAccessControl) CheckRole(userID, roleID string) (bool, error) {
	roles, exists := ac.userRoles[userID]
	if !exists {
		return false, nil
	}

	for _, rID := range roles {
		if rID == roleID {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoles 获取用户的所有角色
func (ac *RBACAccessControl) GetUserRoles(userID string) ([]Role, error) {
	roles, exists := ac.userRoles[userID]
	if !exists {
		return []Role{}, nil
	}

	userRoles := make([]Role, 0, len(roles))
	for _, roleID := range roles {
		role, exists := ac.roles[roleID]
		if exists {
			userRoles = append(userRoles, role)
		}
	}

	return userRoles, nil
}

// GetRolePermissions 获取角色的所有权限
func (ac *RBACAccessControl) GetRolePermissions(roleID string) ([]Permission, error) {
	role, exists := ac.roles[roleID]
	if !exists {
		return []Permission{}, errors.New("role not found")
	}

	return role.Permissions, nil
}

// AssignRole 为用户分配角色
func (ac *RBACAccessControl) AssignRole(userID, roleID string) error {
	if _, exists := ac.roles[roleID]; !exists {
		return errors.New("role not found")
	}

	roles, exists := ac.userRoles[userID]
	if !exists {
		roles = []string{}
	}

	// 检查角色是否已分配
	for _, rID := range roles {
		if rID == roleID {
			return nil // 角色已分配
		}
	}

	roles = append(roles, roleID)
	ac.userRoles[userID] = roles

	return nil
}

// RevokeRole 撤销用户的角色
func (ac *RBACAccessControl) RevokeRole(userID, roleID string) error {
	roles, exists := ac.userRoles[userID]
	if !exists {
		return nil // 用户没有角色
	}

	newRoles := make([]string, 0, len(roles)-1)
	for _, rID := range roles {
		if rID != roleID {
			newRoles = append(newRoles, rID)
		}
	}

	ac.userRoles[userID] = newRoles

	return nil
}
