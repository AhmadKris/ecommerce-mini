package model

// CreateCategoryRequest is the payload for POST /api/admin/categories.
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}

// UpdateCategoryRequest is the payload for PUT /api/admin/categories/:id.
type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}
