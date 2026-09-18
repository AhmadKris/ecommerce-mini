package model

// CreateAddressRequest is the payload for POST /api/profile/addresses.
type CreateAddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required,min=2,max=255"`
	Phone         string `json:"phone" binding:"required,min=5,max=30"`
	AddressLine   string `json:"address_line" binding:"required,min=10,max=500"`
	City          string `json:"city" binding:"required,max=255"`
	Province      string `json:"province" binding:"required,max=255"`
	PostalCode    string `json:"postal_code" binding:"required,max=20"`
	IsDefault     bool   `json:"is_default"`
}

// UpdateAddressRequest is the payload for PUT /api/profile/addresses/:id —
// a full replace (unlike UpdateProductRequest), since every field here is
// cheap for a client to resend and there's no "explicit zero" ambiguity
// worth pointer fields for (an address with a blank city is never valid).
type UpdateAddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required,min=2,max=255"`
	Phone         string `json:"phone" binding:"required,min=5,max=30"`
	AddressLine   string `json:"address_line" binding:"required,min=10,max=500"`
	City          string `json:"city" binding:"required,max=255"`
	Province      string `json:"province" binding:"required,max=255"`
	PostalCode    string `json:"postal_code" binding:"required,max=20"`
	IsDefault     bool   `json:"is_default"`
}
