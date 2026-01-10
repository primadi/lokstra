package pkga

// User represents a user from package A
type User struct {
	ID   string
	Name string
}

// Request represents a request from package A
type Request struct {
	Action string
	Data   string
}
