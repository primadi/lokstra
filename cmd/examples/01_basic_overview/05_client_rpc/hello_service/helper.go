package hello_service

// User matches server's User struct
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
	CreateAt string `json:"create_at"`
}

// GetEmail implements UserIface.
func (u *User) GetEmail() string {
	return u.Email
}

// GetID implements UserIface.
func (u *User) GetID() int {
	return u.ID
}

// GetName implements UserIface.
func (u *User) GetName() string {
	return u.Name
}

// IsActive implements UserIface.
func (u *User) IsActive() bool {
	return u.Active
}

var _ UserIface = (*User)(nil)
