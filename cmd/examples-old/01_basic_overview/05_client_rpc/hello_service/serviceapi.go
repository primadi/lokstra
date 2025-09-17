package hello_service

import "time"

// ============================
// INTERFACES
// ============================

// GreetingService interface with various return types
type GreetingService interface {
	// Return string, error
	Hello(name string) (string, error)

	// Return interface, error
	GetUser(id int) (UserIface, error)

	// Return slice of interface, error
	GetUsers(limit int) ([]UserIface, error)

	// Return map, error
	GetUserStats(id int) (map[string]interface{}, error)

	// Return struct, error
	GetSystemInfo() (SystemInfo, error)

	// Return primitive types, error
	GetUserCount() (int, error)
	GetUserActive(id int) (bool, error)
	GetServerTime() (time.Time, error)

	// Return interface{} (any), error
	GetDynamicData(dataType string) (interface{}, error)

	// Return only error (void operations)
	DeleteUser(id int) error
	ClearCache() error
	Ping() error
}

// UserIface represents user interface
type UserIface interface {
	GetID() int
	GetName() string
	GetEmail() string
	IsActive() bool
}

// SystemInfo struct for system information
type SystemInfo struct {
	Version   string  `json:"version"`
	Uptime    string  `json:"uptime"`
	Memory    string  `json:"memory"`
	CPUUsage  float64 `json:"cpu_usage"`
	Connected int     `json:"connected"`
}
