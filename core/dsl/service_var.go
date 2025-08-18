package dsl

import "github.com/primadi/lokstra/serviceapi"

type ServiceVar[TParam any] struct {
	DbPool       serviceapi.DbPool
	DbSchemaName string

	Logger  serviceapi.Logger
	Metrics serviceapi.Metrics
	I18n    serviceapi.I18n

	Param *TParam
	Vars  map[string]any
}

func NewServiceVar[TParam any](dbPool serviceapi.DbPool,
	dbSchemaName string, logger serviceapi.Logger,
	metrics serviceapi.Metrics, i18n serviceapi.I18n, param *TParam,
	vars map[string]any) *ServiceVar[TParam] {
	return &ServiceVar[TParam]{
		DbPool:       dbPool,
		DbSchemaName: dbSchemaName,

		Logger:  logger,
		Metrics: metrics,
		I18n:    i18n,
		Param:   param,
		Vars:    vars,
	}
}
