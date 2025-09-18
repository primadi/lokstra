package hello_service

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/service"
)

// ============================
// USER IMPLEMENTATION
// ============================

// User implements UserIface
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
	CreateAt string `json:"create_at"`
}

func (u *User) GetID() int       { return u.ID }
func (u *User) GetName() string  { return u.Name }
func (u *User) GetEmail() string { return u.Email }
func (u *User) IsActive() bool   { return u.Active }

// ============================
// SERVICE IMPLEMENTATION
// ============================

type GreetingServiceImpl struct{}

// Return string, error
func (h *GreetingServiceImpl) Hello(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	return "Hello, " + name + "!", nil
}

// Return interface, error
func (h *GreetingServiceImpl) GetUser(id int) (UserIface, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

	user := &User{
		ID:       id,
		Name:     fmt.Sprintf("User-%d", id),
		Email:    fmt.Sprintf("user%d@example.com", id),
		Active:   true,
		CreateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return user, nil
}

// Return slice of interface, error
func (h *GreetingServiceImpl) GetUsers(limit int) ([]UserIface, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}

	var users []UserIface
	for i := 1; i <= limit; i++ {
		user := &User{
			ID:       i,
			Name:     fmt.Sprintf("User-%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Active:   i%2 == 1, // Alternate active status
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		users = append(users, user)
	}

	return users, nil
}

// Return map, error
func (h *GreetingServiceImpl) GetUserStats(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

	stats := map[string]any{
		"user_id":      id,
		"login_count":  42,
		"last_login":   time.Now().Unix(),
		"is_premium":   id%3 == 0,
		"balance":      123.45,
		"achievements": []string{"first_login", "complete_profile", "power_user"},
	}

	return stats, nil
}

// Return struct, error
func (h *GreetingServiceImpl) GetSystemInfo() (SystemInfo, error) {
	info := SystemInfo{
		Version:   "1.0.0",
		Uptime:    "5 days",
		Memory:    "512MB",
		CPUUsage:  25.5,
		Connected: 42,
	}
	return info, nil
}

// Return primitive types, error
func (h *GreetingServiceImpl) GetUserCount() (int, error) {
	return 1337, nil // Simulate user count
}

func (h *GreetingServiceImpl) GetUserActive(id int) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid user ID: %d", id)
	}

	// Simulate check - even IDs are active
	return id%2 == 0, nil
}

func (h *GreetingServiceImpl) GetServerTime() (time.Time, error) {
	return time.Now(), nil
}

// Return any (any), error
func (h *GreetingServiceImpl) GetDynamicData(dataType string) (any, error) {
	switch dataType {
	case "user":
		return &User{
			ID: 999, Name: "Dynamic User", Email: "dynamic@example.com",
			Active: true, CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	case "stats":
		return map[string]any{
			"total_users":   1337,
			"active_users":  892,
			"premium_users": 445,
			"last_updated":  time.Now().Unix(),
		}, nil
	case "message":
		return "This is a dynamic string message", nil
	case "number":
		return 42, nil
	case "list":
		return []string{"item1", "item2", "item3"}, nil
	default:
		return nil, fmt.Errorf("unknown data type: %s", dataType)
	}
}

// Return only error (void operations)
func (h *GreetingServiceImpl) DeleteUser(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user ID: %d", id)
	}

	// Simulate deletion logic
	fmt.Printf("User %d deleted successfully\n", id)
	return nil
}

func (h *GreetingServiceImpl) ClearCache() error {
	// Simulate cache clearing
	fmt.Println("Cache cleared successfully")
	return nil
}

func (h *GreetingServiceImpl) Ping() error {
	fmt.Println("Ping received at", time.Now())
	return nil
}

// Interface compliance checks
var _ GreetingService = (*GreetingServiceImpl)(nil)
var _ service.Service = (*GreetingServiceImpl)(nil)

func NewGreetingService() GreetingService {
	return &GreetingServiceImpl{}
}
