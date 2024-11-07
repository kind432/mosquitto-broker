package models

type Role string

const (
	RoleAnonymous  Role = "Anonymous"
	RoleUser       Role = "User"
	RoleSuperAdmin Role = "SuperAdmin"
)

func (e Role) String() string {
	return string(e)
}
