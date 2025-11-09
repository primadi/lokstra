package user

// UserService defines the interface for user-related operations
type UserService interface {
	GetByID(p *GetUserParams) (*User, error)
	List(p *ListUsersParams) ([]*User, error)
	Create(p *CreateUserParams) (*User, error)
	Update(p *UpdateUserParams) (*User, error)
	Delete(p *DeleteUserParams) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByID(id int) (*User, error)
	List() ([]*User, error)
	Create(user *User) (*User, error)
	Update(user *User) (*User, error)
	Delete(id int) error
}
