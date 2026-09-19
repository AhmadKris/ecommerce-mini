package model

// CustomerListQuery binds GET /api/admin/customers query params.
type CustomerListQuery struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}
