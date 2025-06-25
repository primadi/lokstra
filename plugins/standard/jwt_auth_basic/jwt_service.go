package jwt_auth_basic

import "lokstra/iface"

type JWTAuthService struct {
	name      string
	secret    string
	expires   int
	validator AuthCredentialValidator
}

func NewJWTAuthService(name string, cfg map[string]any) *JWTAuthService {
	secret := cfg["secret"].(string)
	expires := int(cfg["expires"].(int))
	return &JWTAuthService{name: name, secret: secret, expires: expires}
}

var _ iface.Service = (*JWTAuthService)(nil)

// func (j *JWTAuthService) Name() string    { return j.name }
// func (j *JWTAuthService) Type() string    { return "jwt-auth-basic" }
// func (j *JWTAuthService) IsEnabled() bool { return true }
// func (j *JWTAuthService) GetConfig(key string) any {
// 	switch key {
// 	case "secret":
// 		return j.secret
// 	case "expires":
// 		return j.expires
// 	default:
// 		return nil
// 	}
// }

// func (j *JWTAuthService) SetValidator(v AuthCredentialValidator) {
// 	j.validator = v
// }

// func (j *JWTAuthService) GetValidator() AuthCredentialValidator {
// 	return j.validator
// }
