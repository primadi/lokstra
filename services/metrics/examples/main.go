package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// User represents a simple user model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Simple metrics interface for demonstration
type MetricsService interface {
	IncCounter(name string, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
}

// MockMetricsService implements MetricsService for demonstration
type MockMetricsService struct{}

func (m *MockMetricsService) IncCounter(name string, labels map[string]string) {
	log.Printf("Counter %s incremented with labels: %v", name, labels)
}

func (m *MockMetricsService) SetGauge(name string, value float64, labels map[string]string) {
	log.Printf("Gauge %s set to %.2f with labels: %v", name, value, labels)
}

func (m *MockMetricsService) ObserveHistogram(name string, value float64, labels map[string]string) {
	log.Printf("Histogram %s observed %.3f with labels: %v", name, value, labels)
}

// UserHandler handles user-related HTTP requests with metrics
type UserHandler struct {
	metrics MetricsService
	users   []User // Simple in-memory storage
}

func NewUserHandler(metricsService MetricsService) *UserHandler {
	return &UserHandler{
		metrics: metricsService,
		users: []User{
			{ID: 1, Name: "John Doe", Email: "john@example.com"},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		},
	}
}

// GetUsers returns all users with metrics tracking
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Track request
	labels := map[string]string{
		"method":   "GET",
		"endpoint": "/users",
	}
	h.metrics.IncCounter("http_requests_total", labels)

	// Set active connections gauge
	h.metrics.SetGauge("active_connections", 5, map[string]string{
		"handler": "users",
	})

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.users); err != nil {
		// Track error
		h.metrics.IncCounter("http_errors_total", map[string]string{
			"method":   "GET",
			"endpoint": "/users",
			"error":    "encode_failed",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Track response time
	duration := time.Since(start).Seconds()
	h.metrics.ObserveHistogram("request_duration_seconds", duration, labels)

	// Track successful response
	successLabels := map[string]string{
		"method":   "GET",
		"endpoint": "/users",
		"status":   "200",
	}
	h.metrics.IncCounter("http_responses_total", successLabels)
}

// CreateUser creates a new user with metrics tracking
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	labels := map[string]string{
		"method":   "POST",
		"endpoint": "/users",
	}
	h.metrics.IncCounter("http_requests_total", labels)

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		// Track validation error
		h.metrics.IncCounter("http_errors_total", map[string]string{
			"method":   "POST",
			"endpoint": "/users",
			"error":    "invalid_json",
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Assign new ID
	user.ID = len(h.users) + 1
	h.users = append(h.users, user)

	// Track user creation
	h.metrics.IncCounter("users_created_total", map[string]string{
		"source": "api",
	})

	// Update user count gauge
	h.metrics.SetGauge("total_users", float64(len(h.users)), nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

	// Track response time
	duration := time.Since(start).Seconds()
	h.metrics.ObserveHistogram("request_duration_seconds", duration, labels)

	// Track successful response
	successLabels := map[string]string{
		"method":   "POST",
		"endpoint": "/users",
		"status":   "201",
	}
	h.metrics.IncCounter("http_responses_total", successLabels)
}

// MetricsMiddleware adds request metrics to all HTTP handlers
func MetricsMiddleware(metrics MetricsService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Track concurrent requests
			metrics.IncCounter("concurrent_requests", map[string]string{
				"method": r.Method,
			})

			// Track request size
			if r.ContentLength > 0 {
				metrics.ObserveHistogram("request_size_bytes", float64(r.ContentLength), map[string]string{
					"method": r.Method,
				})
			}

			next.ServeHTTP(w, r)

			// Track total request duration
			duration := time.Since(start).Seconds()
			metrics.ObserveHistogram("http_request_duration_seconds", duration, map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		})
	}
}

func main() {
	// Create mock metrics service for demonstration
	metricsService := &MockMetricsService{}

	// Create user handler
	userHandler := NewUserHandler(metricsService)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Apply metrics middleware
	handler := MetricsMiddleware(metricsService)(mux)

	// Register routes
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Health check endpoint with metrics
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		metricsService.IncCounter("health_checks_total", nil)
		metricsService.SetGauge("service_up", 1, map[string]string{
			"service": "user-api",
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Background metrics collector
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Simulate system metrics
			metricsService.SetGauge("memory_usage_bytes", float64(1024*1024*128), map[string]string{
				"type": "heap",
			})
			metricsService.SetGauge("goroutines_count", float64(10), nil)
			metricsService.IncCounter("background_tasks_total", map[string]string{
				"task": "metrics_collection",
			})
		}
	}()

	log.Println("Starting user API with metrics...")
	log.Println("API available at: http://localhost:8080")
	log.Println("- GET /users - List all users")
	log.Println("- POST /users - Create a new user")
	log.Println("- GET /health - Health check")
	log.Println()
	log.Println("Example requests:")
	log.Println("curl http://localhost:8080/users")
	log.Println("curl -X POST -H 'Content-Type: application/json' -d '{\"name\":\"Alice\",\"email\":\"alice@example.com\"}' http://localhost:8080/users")

	// Start the server
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
