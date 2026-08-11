package user

import "time"

const (
	RoleAdmin  = "admin"
	RoleViewer = "viewer"
)

type Permission string

const (
	PermissionVolunteerRead    Permission = "volunteer.read"
	PermissionVolunteerCreate  Permission = "volunteer.create"
	PermissionVolunteerUpdate  Permission = "volunteer.update"
	PermissionVolunteerDelete  Permission = "volunteer.delete"
	PermissionDepartmentRead   Permission = "department.read"
	PermissionDepartmentManage Permission = "department.manage"
	PermissionUserRead         Permission = "user.read"
	PermissionUserManage       Permission = "user.manage"
)

type User struct {
	ID           string       `json:"id"`
	Username     string       `json:"username"`
	DisplayName  string       `json:"display_name"`
	AvatarURL    string       `json:"avatar_url"`
	PasswordHash string       `json:"-"`
	Role         string       `json:"role"`
	Permissions  []Permission `json:"permissions"`
	Active       bool         `json:"active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Session struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type CreateInput struct {
	Username    string
	DisplayName string
	Password    string
	Role        string
}

type UpdateInput struct {
	Username    string
	DisplayName string
	Role        string
}

func PermissionsForRole(role string) []Permission {
	switch role {
	case RoleAdmin:
		return []Permission{
			PermissionVolunteerRead,
			PermissionVolunteerCreate,
			PermissionVolunteerUpdate,
			PermissionVolunteerDelete,
			PermissionDepartmentRead,
			PermissionDepartmentManage,
			PermissionUserRead,
			PermissionUserManage,
		}
	case RoleViewer:
		return []Permission{PermissionVolunteerRead, PermissionDepartmentRead, PermissionUserRead}
	default:
		return []Permission{}
	}
}

func HasPermission(role string, permission Permission) bool {
	for _, candidate := range PermissionsForRole(role) {
		if candidate == permission {
			return true
		}
	}
	return false
}

type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}
