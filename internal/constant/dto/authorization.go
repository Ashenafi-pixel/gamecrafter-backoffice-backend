package dto

import (
	"github.com/google/uuid"
)

type Permissions struct {
	ID            uuid.UUID `gorm:"primary_key" json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	RequiresValue bool      `json:"requires_value"` // Whether this permission needs a value/limit
}

type Role struct {
	ID          uuid.UUID     `gorm:"primary_key" json:"id"`
	Name        string        `json:"name"`
	Permissions []Permissions `json:"permissions,omitempty"`
	// PermissionsWithValue is used for role create/edit screens so we can persist and re-load
	// per-permission limits (value/limit_type/limit_period).
	PermissionsWithValue []PermissionWithValue `json:"permissions_with_value,omitempty"`
}

type PermissionWithValue struct {
	PermissionID uuid.UUID  `json:"permission_id"`
	Value        *float64   `json:"value,omitempty"`        // NULL = unlimited, value = specific limit
	LimitType    *string    `json:"limit_type,omitempty"`   // "daily", "weekly", "monthly", or NULL
	LimitPeriod  *int       `json:"limit_period,omitempty"`  // Number of periods (e.g., 1 for "1 daily", 2 for "2 weekly")
}

type CreateRoleReq struct {
	Name        string                `json:"name"`
	Permissions []PermissionWithValue `json:"permissions"` // Updated to include values
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}
type UserRolesRes struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []Role    `json:"roles"`
}

type PermissionsToRoute struct {
	ID          uuid.UUID `json:"id"`
	EndPoint    string    `json:"endpoint"`
	Name        string    `json:"name"`
	Method      string    `json:"method"`
	Description string    `json:"description"`
}
type UpdatePermissionToRoleReq struct {
	RoleID      uuid.UUID             `json:"role_id"`
	Permissions []PermissionWithValue `json:"permissions"` // Updated to include values
}
type UpdatePermissionToRoleRes struct {
	Message string `json:"message"`
	Role    Role   `json:"role"`
}
type GetPermissionReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}
type GetPermissionData struct {
	Permissions Permissions `json:"permission"`
	Roles       []Role      `json:"roles"`
}

type GetPermissionRes struct {
	Message string            `json:"message"`
	Data    GetPermissionData `json:"data"`
}

type RolePermissions struct {
	RoleID      uuid.UUID   `json:"role_id"`
	Permissions []uuid.UUID `json:"permissions"`
}

type AssignPermissionToRoleData struct {
	ID           uuid.UUID  `json:"id"`
	RoleID       uuid.UUID  `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	Value        *float64   `json:"value,omitempty"`        // NULL = unlimited, value = specific limit
	LimitType    *string    `json:"limit_type,omitempty"`   // "daily", "weekly", "monthly", or NULL
	LimitPeriod  *int       `json:"limit_period,omitempty"`  // Number of periods
}

type AssignPermissionToRoleRes struct {
	Message string                     `json:"message"`
	Data    AssignPermissionToRoleData `json:"data"`
}

type GetRoleReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type RemoveRoleReq struct {
	RoleID uuid.UUID `json:"role_id"`
}

type AssignRoleToUserReq struct {
	RoleID uuid.UUID `json:"role_id"`
	UserID uuid.UUID `json:"user_id"`
}

type AssignRoleToUserRes struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []Role    `json:"roles"`
}

// PermissionsList is deprecated - all permissions are now managed via database migrations
// See: migrations/20250117000001_seed_page_permissions.up.sql
// Only keeping "super" permission for super admin role initialization
var PermissionsList = map[string]PermissionsToRoute{
	"super": {EndPoint: "*", Method: "*", Name: "super", Description: "super user has all permissions on the system"},
}
