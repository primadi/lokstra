package auth_module

import (
	"github.com/primadi/lokstra/core/iface"
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
func (m *module) Register(regCtx iface.RegistrationContext) error {
	regCtx.RegisterServiceFactory(FLOW_PASSWORD_NAME, createPasswordFlowFactory(regCtx))
	regCtx.RegisterServiceFactory(SESSION_FACTORY_NAME, createSessionFactory(regCtx))
	return nil
}

var _ iface.Module = (*module)(nil)

func GetModule() iface.Module {
	return &module{}
}

func createSessionFactory(regCtx iface.RegistrationContext) service.ServiceFactory {
	return func(config any) (service.Service, error) {
		kvStore, err := iface.GetServiceFromConfig[serviceapi.KvStore](regCtx, config, "kvstore")
		if err != nil {
			return nil, err
		}
		return NewSession(kvStore), nil
	}
}

func createPasswordFlowFactory(regCtx iface.RegistrationContext) service.ServiceFactory {
	return func(config any) (service.Service, error) {
		userRepo, err := iface.GetServiceFromConfig[auth.UserRepository](regCtx,
			config, "user_repo")
		if err != nil {
			return nil, err
		}
		return NewFlowPassword(userRepo)
	}
}
