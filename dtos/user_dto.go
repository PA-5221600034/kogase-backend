package dtos

// UpdateUserRequest represents data for updating user information
type UpdateUserRequest struct {
	Name     string `json:"name" binding:"omitempty"`
	Password string `json:"password" binding:"omitempty"`
}
