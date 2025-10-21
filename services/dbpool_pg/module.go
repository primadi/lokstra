package dbpool_pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/old_registry"
)

// for single service module, module name equals service name
const SERVICE_TYPE = "dbpool_pg"

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

	MinConnections int           `json:"min_cons" yaml:"min_cons"`
	MaxConnections int           `json:"max_cons" yaml:"max_cons"`
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

func Service(cfg *Config) *pgxPostgresPool {
	dsn := cfg.GetFinalDSN()

	svc, err := NewPgxPostgresPool(context.Background(), dsn)
	if err != nil {
		return nil
	}
	return svc
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		DSN:            utils.GetValueFromMap(params, "dsn", ""),
		Host:           utils.GetValueFromMap(params, "host", "localhost"),
		Port:           utils.GetValueFromMap(params, "port", 5432),
		Database:       utils.GetValueFromMap(params, "database", "postgres"),
		Username:       utils.GetValueFromMap(params, "username", "postgres"),
		Password:       utils.GetValueFromMap(params, "password", ""),
		MinConnections: utils.GetValueFromMap(params, "min_connections", 0),
		MaxConnections: utils.GetValueFromMap(params, "max_connections", 4),
		MaxIdleTime:    utils.GetValueFromMap(params, "max_idle_time", 30*time.Minute),
		MaxLifetime:    utils.GetValueFromMap(params, "max_lifetime", time.Hour),
		SSLMode:        utils.GetValueFromMap(params, "sslmode", "disable"),
	}
	return Service(cfg)
}

func Register() {
	old_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		old_registry.AllowOverride(true))
}
