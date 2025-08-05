package standardservices

import (
	"errors"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const AUTH_FLOW_PASSWORD = auth.FLOW_PREFIX + "password"

func RegisterAllAuthFlow(regCtx iface.RegistrationContext) {
	AuthPasswordfactory := func(config any) (service.Service, error) {
		var ok bool
		var userRepo auth.UserRepository
		var r service.Service
		var err error

		switch cfg := config.(type) {
		case auth.UserRepository:
			return cfg, nil
		case string:
			r, err = regCtx.GetService(cfg)
			if err != nil {
				return nil, err
			}
		case map[string]any:
			var repoName string
			repoName, ok = cfg["user_repository"].(string)
			if !ok || repoName == "" {
				return nil, errors.New("user_repository must be provided")
			}
			r, err = regCtx.GetService(repoName)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unsupported config type for auth flow password")
		}

		userRepo, ok = r.(auth.UserRepository)
		if userRepo == nil || !ok {
			return nil, errors.New("invalid user repository type")
		}

		return userRepo, nil
	}

	regCtx.RegisterServiceFactory(AUTH_FLOW_PASSWORD, AuthPasswordfactory)
}
