package pkgb

// User represents a user from package B (different from pkga.User)
type User struct {
	UserID   string
	FullName string
	Email    string
}

// Response represents a response from package B
type Response struct {
	Status  string
	Message string
}
