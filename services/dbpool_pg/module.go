package dbpool_pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/primadi/lokstra/common/utils"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

// for single service module, module name equals service name
const MODULE_NAME = "lokstra.dbpool_pg"

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return MODULE_NAME
}

// Config represents the configuration for the PostgreSQL connection pool service.
// It can be provided in various formats, including a DSN string, a map, or a struct.
// If using DSN, it should be in the format:
// postgres://username:password@host:port/database?sslmode=disable&pool_min_conns=0&pool_max_conns=4&pool_max_conn_idle_time=30m&pool_max_conn_lifetime=1h
// Host, Port, Database, Username, and Password can be provided separately if DSN is not used.
// Other parameters like MinConnections, MaxConnections, MaxIdleTime, MaxLifetime, and SSLMode can also be set.
type Config struct {
	DSN      string `json:"dsn" yaml:"dsn"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Database string `json:"database" yaml:"database"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

	MinConnections int           `json:"min_connections" yaml:"min_connections"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections"`
	MaxIdleTime    time.Duration `json:"max_idle_time" yaml:"max_idle_time"`
	MaxLifetime    time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
	SSLMode        string        `json:"sslmode" yaml:"sslmode"`
}

func (cfg *Config) buildDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_min_conns=%d&pool_max_conns=%d&pool_max_conn_idle_time=%s&pool_max_conn_lifetime=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
		cfg.SSLMode, cfg.MinConnections, cfg.MaxConnections, cfg.MaxIdleTime, cfg.MaxLifetime)
}

func (cfg *Config) GetFinalDSN() string {
	if cfg.DSN == "" {
		return cfg.buildDSN()
	}
	dsnFinal := cfg.DSN

	if !strings.Contains(dsnFinal, "pool_min_conns=") {
		dsnFinal += fmt.Sprintf("&pool_min_conns=%d", cfg.MinConnections)
	}
	if !strings.Contains(dsnFinal, "pool_max_conns=") {
		dsnFinal += fmt.Sprintf("&pool_max_conns=%d", cfg.MaxConnections)
	}
	if !strings.Contains(dsnFinal, "pool_max_conn_idle_time=") {
		dsnFinal += fmt.Sprintf("&pool_max_conn_idle_time=%s", cfg.MaxIdleTime)
	}
	if !strings.Contains(dsnFinal, "pool_max_conn_lifetime=") {
		dsnFinal += fmt.Sprintf("&pool_max_conn_lifetime=%s", cfg.MaxLifetime)
	}

	if !strings.Contains(dsnFinal, "sslmode=") {
		dsnFinal += fmt.Sprintf("&sslmode=%s", cfg.SSLMode)
	}
	return dsnFinal
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		var dsn string

		switch t := config.(type) {
		case string:
			dsn = t
		case map[string]any:
			var cfg *Config
			if dk, ok := t["dsn"].(string); ok && dk != "" {
				cfg = &Config{DSN: dk}
			} else {
				cfg = &Config{
					Host:           utils.GetValueFromMap(t, "host", "localhost"),
					Port:           utils.GetValueFromMap(t, "port", 5432),
					Database:       utils.GetValueFromMap(t, "database", ""),
					Username:       utils.GetValueFromMap(t, "username", ""),
					Password:       utils.GetValueFromMap(t, "password", ""),
					MinConnections: utils.GetValueFromMap(t, "min_connections", 0),
					MaxConnections: utils.GetValueFromMap(t, "max_connections", 4),
					MaxIdleTime:    utils.GetDurationFromMap(t, "max_idle_time", "30m"),
					MaxLifetime:    utils.GetDurationFromMap(t, "max_lifetime", "1h"),
					SSLMode:        utils.GetValueFromMap(t, "sslmode", "disable"),
				}
			}
			dsn = cfg.GetFinalDSN()
		case []string:
			if len(t) == 1 {
				dsn = t[0]
			} else {
				return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration slice")
			}
		case *Config:
			if t != nil {
				dsn = t.GetFinalDSN()
			} else {
				return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration")
			}
		case Config:
			dsn = t.GetFinalDSN()
		default:
			return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration")
		}

		return NewPgxPostgresPool(context.Background(), dsn)
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "PostgreSQL Database Pool Service Module"
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}

func CreateService(regCtx registration.Context, serviceName string, config *Config) (*pgxPostgresPool, error) {
	dsn := config.DSN
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_min_conns=%d&pool_max_conns=%d&pool_max_conn_idle_time=%s&pool_max_conn_lifetime=%s",
			config.Username, config.Password, config.Host, config.Port, config.Database,
			config.SSLMode, config.MinConnections, config.MaxConnections, config.MaxIdleTime, config.MaxLifetime)
	}

	svc, err := NewPgxPostgresPool(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	regCtx.RegisterService(serviceName, svc, true)
	return svc, nil
}

func GetService(regCtx registration.Context, serviceName string) (*pgxPostgresPool, error) {
	svc, err := regCtx.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	pool, ok := svc.(*pgxPostgresPool)
	if !ok {
		return nil, fmt.Errorf("service %q is not a pgxPostgresPool", serviceName)
	}

	return pool, nil
}
