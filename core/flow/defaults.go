package flow

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/serviceapi"
)

var defaultDbPool serviceapi.DbPool
var defaultLogger serviceapi.Logger
var defaultI18n serviceapi.I18n
var defaultMetrics serviceapi.Metrics
var defaultDbSchemaName string

func SetDefaultDbPoolService(pool serviceapi.DbPool) {
	defaultDbPool = pool
}

func SetDefaultLoggerService(logger serviceapi.Logger) {
	defaultLogger = logger
}

func SetDefaultI18nService(i18n serviceapi.I18n) {
	defaultI18n = i18n
}

func SetDefaultMetricsService(metrics serviceapi.Metrics) {
	defaultMetrics = metrics
}

func SetDefaultDbPool(regCtx registration.Context, name string) error {
	svc, err := registration.GetService[serviceapi.DbPool](regCtx, name)
	if err != nil {
		return err
	}

	SetDefaultDbPoolService(svc)
	return nil
}

func SetDefaultLogger(regCtx registration.Context, name string) error {
	svc, err := registration.GetService[serviceapi.Logger](regCtx, name)
	if err != nil {
		return err
	}

	SetDefaultLoggerService(svc)
	return nil
}

func SetDefaultI18n(regCtx registration.Context, name string) error {
	svc, err := registration.GetService[serviceapi.I18n](regCtx, name)
	if err != nil {
		return err
	}

	SetDefaultI18nService(svc)
	return nil
}

func SetDefaultMetrics(regCtx registration.Context, name string) error {
	svc, err := registration.GetService[serviceapi.Metrics](regCtx, name)
	if err != nil {
		return err
	}

	SetDefaultMetricsService(svc)
	return nil
}

func SetDefaultDbSchemaName(schemaName string) {
	defaultDbSchemaName = schemaName
}
