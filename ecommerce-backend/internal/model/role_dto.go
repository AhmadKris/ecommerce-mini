package model

// UpdateUserRolesRequest binds the body of PUT /api/admin/customers/:id/roles.
// RoleNames replaces the user's entire role set — omitting a role removes it,
// including a role adds it. The service validates that every name exists in the
// roles table before writing.
type UpdateUserRolesRequest struct {
	RoleNames []string `json:"role_names" binding:"required,min=1"`
}
