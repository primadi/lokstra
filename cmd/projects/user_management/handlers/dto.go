package handlers

type CreateUserRequestDTO struct {
	Username string         `json:"username" form:"username"`
	Email    string         `json:"email" form:"email"`
	Password string         `json:"password" form:"password"`
	FullName string         `json:"full_name" form:"full_name"`
	IsActive *bool          `json:"is_active,omitempty" form:"is_active"`
	Metadata map[string]any `json:"metadata,omitempty" form:"metadata"`
}

type UpdateUserRequestDTO struct {
	ID       string         `path:"id"`
	Username string         `json:"username" form:"username"`
	Email    string         `json:"email" form:"email"`
	Password string         `json:"password" form:"password"`
	FullName string         `json:"full_name" form:"full_name"`
	IsActive *bool          `json:"is_active,omitempty" form:"is_active"`
	Metadata map[string]any `json:"metadata,omitempty" form:"metadata"`
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
