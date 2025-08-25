package handlers

type CreateUserRequestDTO struct {
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Password string         `json:"password"`
	IsActive *bool          `json:"is_active,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type UpdateUserRequestDTO struct {
	ID       string         `path:"id"`
	Email    string         `json:"email"`
	Password string         `json:"password"`
	IsActive *bool          `json:"is_active,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type DeleteUserRequestDTO struct {
	UserID string `path:"id"`
}

type GetUserByNameRequestDTO struct {
	Username string `path:"username"`
}

type GetUserByIDRequestDTO struct {
	ID string `path:"id"`
}

// ListUserRequestDTO bisa kosong karena menggunakan BindPaginationQuery
type ListUserRequestDTO struct {
}
