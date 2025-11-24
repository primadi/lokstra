package domain

// User represents a user entity in the user bounded context
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"` // active, suspended, deleted
	RoleID int    `json:"role_id"`
}

// UserProfile represents extended user profile information
type UserProfile struct {
	UserID  int    `json:"user_id"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Avatar  string `json:"avatar"`
}
