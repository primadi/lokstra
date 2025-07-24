package logger

import (
	"github.com/primadi/lokstra/core/registration"
)

const FACTORY_NAME = "logger"

type LoggerServiceModule struct{}

// Name implements registration.Module.
func (l *LoggerServiceModule) Name() string {
	return FACTORY_NAME
}

// Register implements registration.Module.
func (l *LoggerServiceModule) Register(regCtx registration.Context) error {
	regCtx.RegisterServiceFactory(l.Name(), ServiceFactory)
	return nil
}

// Description implements service.Module.
func (l *LoggerServiceModule) Description() string {
	return "Logger Service for Lokstra"
}

// GetModule returns the logger service with serviceType "lokstra.logger".
func GetModule() registration.Module {
	return &LoggerServiceModule{}
}

var _ registration.Module = (*LoggerServiceModule)(nil)
