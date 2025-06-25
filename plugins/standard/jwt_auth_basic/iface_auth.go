package jwt_auth_basic

// AuthCredentialValidator is the hook interface your app must implement
// to validate a username and password and return userId and roleId.
type AuthCredentialValidator interface {
	ValidateCredentials(username, password string) (userId string, roleId string, err error)
}
