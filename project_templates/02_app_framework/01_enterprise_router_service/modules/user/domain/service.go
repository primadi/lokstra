package domain

// UserService defines business operations for users
type UserService interface {
	GetByID(p *GetUserRequest) (*User, error)
	List(p *ListUsersRequest) ([]*User, error)
	Create(p *CreateUserRequest) (*User, error)
	Update(p *UpdateUserRequest) (*User, error)
	Suspend(p *SuspendUserRequest) error
	Activate(p *ActivateUserRequest) error
	Delete(p *DeleteUserRequest) error
}

// UserRepository defines data access for users
type UserRepository interface {
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	List() ([]*User, error)
	Create(user *User) (*User, error)
	Update(user *User) (*User, error)
	Delete(id int) error
}
