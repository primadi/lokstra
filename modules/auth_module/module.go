package auth_module

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const (
	FLOW_PASSWORD_NAME        = auth.FLOW_PREFIX + "password"
	SESSION_FACTORY_NAME      = auth.SESSION_PREFIX + "kvstore"
	TOKEN_ISSUER_FACTORY_NAME = auth.TOKEN_ISSUER_PREFIX + "jwt"
	SERVICE_FACTORY_NAME      = auth.SERVICE_PREFIX + "default"
)

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "Auth Service Module provides authentication and session management services."
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "auth_module"
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	regCtx.RegisterServiceFactory(FLOW_PASSWORD_NAME, createPasswordFlowFactory(regCtx))
	regCtx.RegisterServiceFactory(SESSION_FACTORY_NAME, createSessionFactory(regCtx))
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}

func createSessionFactory(regCtx registration.Context) service.ServiceFactory {
	return func(config any) (service.Service, error) {
		kvStore, err := registration.GetServiceFromConfig[serviceapi.KvStore](regCtx, config, "kvstore")
		if err != nil {
			return nil, err
		}
		return NewSession(kvStore), nil
	}
}

func createPasswordFlowFactory(regCtx registration.Context) service.ServiceFactory {
	return func(config any) (service.Service, error) {
		userRepo, err := registration.GetServiceFromConfig[auth.UserRepository](regCtx,
			config, "user_repo")
		if err != nil {
			return nil, err
		}
		return NewFlowPassword(userRepo)
	}
}
