package domain

type Role string

const (
	RoleMarketer Role = "marketer"
	RoleManager  Role = "manager"
	RoleTenant   Role = "tenant"
	RoleAnalyst  Role = "analyst"
)

type Principal struct {
	UserID      string
	Role        Role
	TenantID    string
	AuthVersion int64
}
