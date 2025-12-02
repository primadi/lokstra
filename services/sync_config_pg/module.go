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
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

const SERVICE_TYPE = "sync_config_pg"

// Config represents the configuration for PostgreSQL-based SyncConfig service
type Config struct {
	DbPoolName         string        `json:"dbpool_name" yaml:"dbpool_name"`                 // Name of DbPoolManager to use
	TableName          string        `json:"table_name" yaml:"table_name"`                   // Table name for storing configs
	Channel            string        `json:"channel" yaml:"channel"`                         // PostgreSQL NOTIFY channel name
	HeartbeatInterval  time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`   // CRC heartbeat interval (default: 5 minutes)
	ReconnectInterval  time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`   // Reconnect attempt interval
	SyncOnMismatch     bool          `json:"sync_on_mismatch" yaml:"sync_on_mismatch"`       // Auto sync when CRC mismatch detected
	EnableNotification bool          `json:"enable_notification" yaml:"enable_notification"` // Enable LISTEN/NOTIFY (default: true)
}

func DefualtConfig() *Config {
	return &Config{
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  5 * time.Minute,
		ReconnectInterval:  10 * time.Second,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}
}

type subscriber struct {
	id       string
	callback serviceapi.ConfigChangeCallback
}

type syncConfigPG struct {
	cfg         *Config
	dbPool      serviceapi.DbPoolWithSchema
	listenerDB  *pgxpool.Pool
	mu          sync.RWMutex
	cache       map[string]any
	subscribers map[string]*subscriber
	crc         uint32
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

var _ serviceapi.SyncConfig = (*syncConfigPG)(nil)

func (s *syncConfigPG) Set(ctx context.Context, key string, value any) error {
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

func (s *syncConfigPG) GetString(ctx context.Context, key string, defaultValue string) string {
	value, err := s.Get(ctx, key)
	if err != nil {
		return defaultValue
	}

	if str, ok := value.(string); ok {
		return str
	}

	return defaultValue
}

func (s *syncConfigPG) GetInt(ctx context.Context, key string, defaultValue int) int {
	value, err := s.Get(ctx, key)
	if err != nil {
		return defaultValue
	}

	// Handle JSON number conversion
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		return utils.ParseInt(v, defaultValue)
	}

	return defaultValue
}

func (s *syncConfigPG) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	value, err := s.Get(ctx, key)
	if err != nil {
		return defaultValue
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return defaultValue
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

	ctx := context.Background()
	conn, err := s.listenerDB.Acquire(ctx)
	if err != nil {
		fmt.Printf("Failed to acquire listener connection: %v\n", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN "+s.cfg.Channel)
	if err != nil {
		fmt.Printf("Failed to LISTEN on channel %s: %v\n", s.cfg.Channel, err)
		return
	}

	for {
		select {
		case <-s.stopCh:
			return
		default:
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// Check if context is done
				if ctx.Err() != nil {
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
	close(s.stopCh)
	s.wg.Wait()

	if s.listenerDB != nil {
		s.listenerDB.Close()
	}

	if s.dbPool != nil {
		return s.dbPool.Shutdown()
	}

	return nil
}

// Service creates a new PostgreSQL-based SyncConfig service
func Service(cfg *Config, dbPoolManager serviceapi.DbPoolManager) (*syncConfigPG, error) {
	dbPool, err := dbPoolManager.GetNamedPool(cfg.DbPoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get db pool: %w", err)
	}

	// Get connection config for listener
	dsn, _, err := dbPoolManager.GetNamedDsn(cfg.DbPoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSN: %w", err)
	}

	var listenerDB *pgxpool.Pool
	if cfg.EnableNotification {
		listenerDB, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener pool: %w", err)
		}
	}

	service := &syncConfigPG{
		cfg:         cfg,
		dbPool:      dbPool,
		listenerDB:  listenerDB,
		cache:       make(map[string]any),
		subscribers: make(map[string]*subscriber),
		stopCh:      make(chan struct{}),
	}

	if err := service.init(context.Background()); err != nil {
		if listenerDB != nil {
			listenerDB.Close()
		}
		return nil, err
	}

	return service, nil
}

// ServiceFactory creates a SyncConfig service from configuration map
func ServiceFactory(deps map[string]any, params map[string]any) any {
	cfg := &Config{
		DbPoolName:         utils.GetValueFromMap(params, "dbpool_name", ""),
		TableName:          utils.GetValueFromMap(params, "table_name", "sync_config"),
		Channel:            utils.GetValueFromMap(params, "channel", "config_changes"),
		HeartbeatInterval:  time.Duration(utils.GetValueFromMap(params, "heartbeat_interval", 5)) * time.Minute,
		ReconnectInterval:  time.Duration(utils.GetValueFromMap(params, "reconnect_interval", 10)) * time.Second,
		SyncOnMismatch:     utils.GetValueFromMap(params, "sync_on_mismatch", true),
		EnableNotification: utils.GetValueFromMap(params, "enable_notification", true),
	}

	if cfg.DbPoolName == "" {
		panic("sync_config_pg requires 'dbpool_name' parameter")
	}

	// Get DbPoolManager from dependencies
	dbPoolManagerRaw, ok := deps["dbpool-manager"]
	if !ok {
		panic("sync_config_pg requires 'dbpool-manager' dependency")
	}

	dbPoolManager, ok := dbPoolManagerRaw.(serviceapi.DbPoolManager)
	if !ok {
		panic("dbpool-manager is not of type serviceapi.DbPoolManager")
	}

	svc, err := Service(cfg, dbPoolManager)
	if err != nil {
		panic(fmt.Sprintf("failed to create sync_config_pg service: %v", err))
	}

	return svc
}

// Register registers the SyncConfig service type
func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
