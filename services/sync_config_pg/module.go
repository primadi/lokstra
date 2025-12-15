package sync_config_pg

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

const SERVICE_TYPE = "sync_config_pg"

var (
	instanceMu sync.Mutex
	instances  = make(map[string]serviceapi.SyncConfig)
)

// Config represents the configuration for PostgreSQL-based SyncConfig service
type Config struct {
	DbPoolName         string        `json:"db_pool_name" yaml:"db_pool_name"`               // Named database pool               // Schema name
	TableName          string        `json:"table_name" yaml:"table_name"`                   // Table name for storing configs
	Channel            string        `json:"channel" yaml:"channel"`                         // PostgreSQL NOTIFY channel name
	HeartbeatInterval  time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`   // CRC heartbeat interval (default: 5 minutes)
	ReconnectInterval  time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`   // Reconnect attempt interval
	SyncOnMismatch     bool          `json:"sync_on_mismatch" yaml:"sync_on_mismatch"`       // Auto sync when CRC mismatch detected
	EnableNotification bool          `json:"enable_notification" yaml:"enable_notification"` // Enable LISTEN/NOTIFY (default: true)
}

type subscriber struct {
	id       string
	callback serviceapi.ConfigChangeCallback
}

type syncConfigPG struct {
	cfg         *Config
	dbPool      serviceapi.DbPool
	listenerDB  *pgxpool.Pool
	mu          sync.RWMutex
	cache       map[string]any
	subscribers map[string]*subscriber
	crc         uint32
	stopCh      chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

var _ serviceapi.SyncConfig = (*syncConfigPG)(nil)

func getDsnAndSchema(cfg *Config) (string, string) {
	deployConfig := deploy.Global().GetDeployConfig()
	if deployConfig == nil {
		panic("sync_config_pg: deploy config not found")
	}

	poolConfig, ok := deployConfig.NamedDbPools[cfg.DbPoolName]
	if !ok {
		panic(fmt.Sprintf("sync_config_pg: named pool '%s' not found in config", cfg.DbPoolName))
	}

	schema := poolConfig.Schema
	if schema == "" {
		schema = "public" // Default schema
	}

	return poolConfig.DSN, schema
}

// NewSyncConfigPG creates a new SyncConfig instance from config
// It will automatically get the database pool and create listener connection
// If an instance with the same configuration already exists, it will be reused (singleton per config)
func NewSyncConfigPG(cfg *Config) (serviceapi.SyncConfig, error) {
	// Check if instance already exists for this configuration
	instanceKey := fmt.Sprintf("%s:%s:%s", cfg.DbPoolName, cfg.TableName, cfg.Channel)

	instanceMu.Lock()
	if existing, ok := instances[instanceKey]; ok {
		instanceMu.Unlock()
		return existing, nil
	}
	instanceMu.Unlock()

	dsn, schema := getDsnAndSchema(cfg)

	// Get DSN for listener connection
	var listenerDB *pgxpool.Pool
	if cfg.EnableNotification {
		var err error
		listenerDB, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener pool: %w", err)
		}
	}

	// Create context with cancel for goroutine management
	ctx, cancel := context.WithCancel(context.Background())

	dbPool, err := dbpool_pg.NewPgxPostgresPool(dsn, schema, nil)
	if err != nil {
		if listenerDB != nil {
			listenerDB.Close()
		}
		cancel()
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	service := &syncConfigPG{
		cfg:         cfg,
		dbPool:      dbPool,
		listenerDB:  listenerDB,
		cache:       make(map[string]any),
		subscribers: make(map[string]*subscriber),
		stopCh:      make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize service (load data, start listener, start heartbeat)
	if err := service.init(context.Background()); err != nil {
		if listenerDB != nil {
			listenerDB.Close()
		}
		return nil, fmt.Errorf("failed to initialize sync config: %w", err)
	}

	// Register instance in singleton registry
	instanceMu.Lock()
	instances[instanceKey] = service
	instanceMu.Unlock()

	return service, nil
}

func (s *syncConfigPG) Set(ctx context.Context, key string, value any) error {
	// Check if value already exists and is unchanged (optimization)
	s.mu.RLock()
	existingValue, exists := s.cache[key]
	s.mu.RUnlock()

	if exists && equal(existingValue, value) {
		// Value unchanged - skip database write to reduce IO/network load
		return nil
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) 
		DO UPDATE SET value = $2, updated_at = NOW()
	`, s.cfg.TableName)

	conn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, query, key, valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	// Update local cache
	s.mu.Lock()
	s.cache[key] = value
	s.updateCRC()
	s.mu.Unlock()

	// Note: pg_notify is handled by database trigger automatically
	// No need to manually notify here

	// Notify local subscribers
	s.notifySubscribers(key, value)

	return nil
}

func (s *syncConfigPG) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	value, exists := s.cache[key]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("config key not found: %s", key)
	}

	return value, nil
}

func (s *syncConfigPG) Delete(ctx context.Context, key string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE key = $1`, s.cfg.TableName)

	conn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// Update local cache
	s.mu.Lock()
	delete(s.cache, key)
	s.updateCRC()
	s.mu.Unlock()

	// Note: pg_notify is handled by database trigger automatically
	// No need to manually notify here

	// Notify local subscribers
	s.notifySubscribers(key, nil)

	return nil
}

func (s *syncConfigPG) GetAll(ctx context.Context) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]any, len(s.cache))
	for k, v := range s.cache {
		result[k] = v
	}

	return result, nil
}

func (s *syncConfigPG) Subscribe(callback serviceapi.ConfigChangeCallback) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub := &subscriber{
		id:       uuid.New().String(),
		callback: callback,
	}

	s.subscribers[sub.id] = sub
	return sub.id
}

func (s *syncConfigPG) Unsubscribe(subscriptionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers, subscriptionID)
}

func (s *syncConfigPG) GetCRC() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.crc
}

func (s *syncConfigPG) Sync(ctx context.Context) error {
	return s.loadFromDB(ctx)
}

// Internal methods

func (s *syncConfigPG) init(ctx context.Context) error {
	// Load initial data
	if err := s.loadFromDB(ctx); err != nil {
		return err
	}

	// Start listener for notifications
	if s.cfg.EnableNotification {
		s.startListener()
	}

	// Start heartbeat
	s.startHeartbeat()

	return nil
}

func (s *syncConfigPG) loadFromDB(ctx context.Context) error {
	query := fmt.Sprintf(`SELECT key, value FROM %s`, s.cfg.TableName)

	conn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to load configs: %w", err)
	}
	defer rows.Close()

	// Collect new data first
	newCache := make(map[string]any)
	for rows.Next() {
		var key string
		var valueJSON []byte

		if err := rows.Scan(&key, &valueJSON); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		var value any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}

		newCache[key] = value
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Detect changes to notify subscribers
	s.mu.Lock()
	oldCache := s.cache
	s.cache = newCache
	s.updateCRC()
	s.mu.Unlock()

	// Notify subscribers about changes
	// 1. Notify updated/new keys
	for key, newValue := range newCache {
		if oldValue, exists := oldCache[key]; !exists || !equal(oldValue, newValue) {
			s.notifySubscribers(key, newValue)
		}
	}

	// 2. Notify deleted keys
	for key := range oldCache {
		if _, exists := newCache[key]; !exists {
			s.notifySubscribers(key, nil)
		}
	}

	return nil
}

// equal compares two values for equality (simple comparison)
func equal(a, b any) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func (s *syncConfigPG) updateCRC() {
	// Must be called with lock held
	keys := make([]string, 0, len(s.cache))
	for k := range s.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	crcData := ""
	for _, k := range keys {
		valueJSON, _ := json.Marshal(s.cache[k])
		crcData += k + ":" + string(valueJSON) + ";"
	}

	s.crc = crc32.ChecksumIEEE([]byte(crcData))
}

// sendHeartbeat sends CRC heartbeat to other instances via pg_notify
// This is separate from data change notifications which are handled by database triggers
func (s *syncConfigPG) sendHeartbeat(ctx context.Context) {
	s.mu.RLock()
	currentCRC := s.crc
	s.mu.RUnlock()

	notification := map[string]any{
		"action": "heartbeat",
		"crc":    currentCRC,
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		return
	}

	conn, _ := s.dbPool.Acquire(ctx)
	if conn != nil {
		defer conn.Release()
		query := `SELECT pg_notify($1, $2)`
		_, _ = conn.Exec(ctx, query, s.cfg.Channel, string(payload))
	}
}

func (s *syncConfigPG) notifySubscribers(key string, value any) {
	s.mu.RLock()
	subscribers := make([]*subscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mu.RUnlock()

	// Call callbacks without holding lock
	for _, sub := range subscribers {
		go sub.callback(key, value)
	}
}

func (s *syncConfigPG) startListener() {
	if s.listenerDB == nil {
		return
	}

	s.wg.Add(1)
	go s.listenForNotifications()
}

func (s *syncConfigPG) listenForNotifications() {
	defer s.wg.Done()

	conn, err := s.listenerDB.Acquire(s.ctx)
	if err != nil {
		fmt.Printf("Failed to acquire listener connection: %v\n", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(s.ctx, "LISTEN "+s.cfg.Channel)
	if err != nil {
		fmt.Printf("Failed to LISTEN on channel %s: %v\n", s.cfg.Channel, err)
		return
	}

	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ctx.Done():
			return
		default:
			notification, err := conn.Conn().WaitForNotification(s.ctx)
			if err != nil {
				// Check if context is done
				if s.ctx.Err() != nil {
					return
				}
				fmt.Printf("Notification error: %v\n", err)
				time.Sleep(s.cfg.ReconnectInterval)
				continue
			}

			if notification != nil {
				s.handleNotification(notification.Payload)
			}
		}
	}
}

func (s *syncConfigPG) handleNotification(payload string) {
	var notification struct {
		Action string `json:"action"`
		Key    string `json:"key"`
		Value  any    `json:"value"`
		CRC    uint32 `json:"crc"` // Only present in heartbeat messages
	}

	if err := json.Unmarshal([]byte(payload), &notification); err != nil {
		return
	}

	// Handle heartbeat - CRC validation only
	if notification.Action == "heartbeat" {
		s.mu.RLock()
		localCRC := s.crc
		s.mu.RUnlock()

		// Check CRC mismatch and trigger sync if needed
		if localCRC != notification.CRC && s.cfg.SyncOnMismatch {
			go s.Sync(context.Background())
		}
		return
	}

	// Handle data changes (insert/update/delete from trigger)
	s.mu.Lock()
	switch notification.Action {
	case "insert", "update":
		s.cache[notification.Key] = notification.Value
	case "delete":
		delete(s.cache, notification.Key)
	}
	s.updateCRC()
	s.mu.Unlock()

	// Notify local subscribers
	s.notifySubscribers(notification.Key, notification.Value)
}

func (s *syncConfigPG) startHeartbeat() {
	s.wg.Add(1)
	go s.heartbeatLoop()
}

func (s *syncConfigPG) heartbeatLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Send CRC heartbeat
			if s.cfg.EnableNotification {
				s.sendHeartbeat(context.Background())
			}
		}
	}
}

func (s *syncConfigPG) Shutdown() error {
	// Cancel context to unblock WaitForNotification
	if s.cancel != nil {
		s.cancel()
	}

	// Close stop channel
	close(s.stopCh)

	// Wait for all goroutines to finish
	s.wg.Wait()

	if s.listenerDB != nil {
		s.listenerDB.Close()
	}

	// Remove from singleton registry
	instanceKey := fmt.Sprintf("%s:%s:%s", s.cfg.DbPoolName, s.cfg.TableName, s.cfg.Channel)
	instanceMu.Lock()
	delete(instances, instanceKey)
	instanceMu.Unlock()

	return nil
}

// Service creates a new PostgreSQL-based SyncConfig service
// This is an alias for NewSyncConfigPG for backward compatibility
func Service(cfg *Config) (serviceapi.SyncConfig, error) {
	return NewSyncConfigPG(cfg)
}

// ServiceFactory creates a SyncConfig service from configuration map
func ServiceFactory(mapCfg map[string]any) any {
	cfg := &Config{
		DbPoolName:         utils.GetValueFromMap(mapCfg, "db_pool_name", "db_main"),
		TableName:          utils.GetValueFromMap(mapCfg, "table_name", "sync_config"),
		Channel:            utils.GetValueFromMap(mapCfg, "channel", "config_changes"),
		HeartbeatInterval:  utils.GetValueFromMap(mapCfg, "heartbeat_interval", 5*time.Minute),
		ReconnectInterval:  utils.GetValueFromMap(mapCfg, "reconnect_interval", 5*time.Second),
		SyncOnMismatch:     utils.GetValueFromMap(mapCfg, "sync_on_mismatch", true),
		EnableNotification: utils.GetValueFromMap(mapCfg, "enable_notification", true),
	}

	svc, err := Service(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create sync_config_pg service: %v", err))
	}

	return svc
}

// Register registers the SyncConfig service type
func Register(dbPoolName string, heartBeatInterval, reconnectInterval time.Duration) {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
	SetDefaultSyncConfigPG(dbPoolName, heartBeatInterval, reconnectInterval)
}

// registers the default SyncConfigPG service
func SetDefaultSyncConfigPG(syncDbPoolName string, heartBeatInterval, reconnectInterval time.Duration) {
	if lokstra_registry.HasService("sync-config") {
		return // Already registered
	}

	lokstra_registry.RegisterLazyService("sync-config", func() any {
		cfg := &Config{
			DbPoolName:         syncDbPoolName,
			TableName:          "sync_config",
			Channel:            "config_changes",
			SyncOnMismatch:     true,
			EnableNotification: true,

			HeartbeatInterval: heartBeatInterval,
			ReconnectInterval: reconnectInterval,
		}
		svc, err := Service(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to create default sync_config_pg service: %v", err))
		}
		return svc
	}, nil)
}
