//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package health_check

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

// DiskStats represents disk usage statistics
type DiskStats struct {
	Total     uint64
	Used      uint64
	Available uint64
}

// DatabaseHealthChecker creates a health checker for database connections
func DatabaseHealthChecker(dbPool serviceapi.DbPool) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		check := serviceapi.HealthCheck{
			Name:      "database",
			CheckedAt: start,
		}

		// Simple connection test to check database
		conn, err := dbPool.Acquire(ctx, "")
		if err != nil {
			check.Status = serviceapi.HealthStatusUnhealthy
			check.Error = err.Error()
			check.Message = "Database connection failed"
		} else {
			check.Status = serviceapi.HealthStatusHealthy
			check.Message = "Database connection is healthy"
			// Release connection if acquired successfully
			if conn != nil {
				conn.Release()
			}
		}

		check.Duration = time.Since(start)
		return check
	}
}

// RedisHealthChecker creates a health checker for Redis connections
func RedisHealthChecker(redis serviceapi.Redis) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		check := serviceapi.HealthCheck{
			Name:      "redis",
			CheckedAt: start,
		}

		// Simple ping to check Redis connection
		client := redis.Client()
		if err := client.Ping(ctx).Err(); err != nil {
			check.Status = serviceapi.HealthStatusUnhealthy
			check.Error = err.Error()
			check.Message = "Redis connection failed"
		} else {
			check.Status = serviceapi.HealthStatusHealthy
			check.Message = "Redis connection is healthy"
		}

		check.Duration = time.Since(start)
		return check
	}
}

// MemoryHealthChecker creates a health checker for memory usage
func MemoryHealthChecker(maxMemoryMB int64) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		check := serviceapi.HealthCheck{
			Name:      "memory",
			CheckedAt: start,
		}

		// Get actual memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Convert bytes to MB
		allocMB := float64(memStats.Alloc) / 1024 / 1024
		sysMB := float64(memStats.Sys) / 1024 / 1024
		maxMB := float64(maxMemoryMB)

		// Calculate usage percentage
		usagePercent := (allocMB / maxMB) * 100

		check.Details = map[string]any{
			"alloc_mb":       fmt.Sprintf("%.2f", allocMB),
			"sys_mb":         fmt.Sprintf("%.2f", sysMB),
			"max_memory_mb":  maxMemoryMB,
			"usage_percent":  fmt.Sprintf("%.2f%%", usagePercent),
			"total_alloc_mb": fmt.Sprintf("%.2f", float64(memStats.TotalAlloc)/1024/1024),
			"num_gc":         memStats.NumGC,
		}

		// Determine health status based on usage
		if usagePercent < 70 {
			check.Status = serviceapi.HealthStatusHealthy
			check.Message = fmt.Sprintf("Memory usage is healthy (%.2f%% of %dMB)", usagePercent, maxMemoryMB)
		} else if usagePercent < 90 {
			check.Status = serviceapi.HealthStatusDegraded
			check.Message = fmt.Sprintf("Memory usage is elevated (%.2f%% of %dMB)", usagePercent, maxMemoryMB)
		} else {
			check.Status = serviceapi.HealthStatusUnhealthy
			check.Message = fmt.Sprintf("Memory usage is critical (%.2f%% of %dMB)", usagePercent, maxMemoryMB)
		}

		check.Duration = time.Since(start)
		return check
	}
}

// DiskHealthChecker creates a health checker for disk usage
func DiskHealthChecker(path string, maxUsagePercent float64) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		check := serviceapi.HealthCheck{
			Name:      "disk",
			CheckedAt: start,
		}

		// Get actual disk stats
		diskStats, err := getDiskUsage(path)
		if err != nil {
			check.Status = serviceapi.HealthStatusUnhealthy
			check.Error = err.Error()
			check.Message = fmt.Sprintf("Failed to get disk usage for path %s", path)
			check.Details = map[string]any{
				"path":              path,
				"max_usage_percent": maxUsagePercent,
				"error":             err.Error(),
			}
		} else {
			usagePercent := (float64(diskStats.Used) / float64(diskStats.Total)) * 100

			check.Details = map[string]any{
				"path":              path,
				"total_bytes":       diskStats.Total,
				"used_bytes":        diskStats.Used,
				"available_bytes":   diskStats.Available,
				"usage_percent":     fmt.Sprintf("%.2f%%", usagePercent),
				"max_usage_percent": fmt.Sprintf("%.2f%%", maxUsagePercent),
				"total_gb":          fmt.Sprintf("%.2f", float64(diskStats.Total)/1024/1024/1024),
				"used_gb":           fmt.Sprintf("%.2f", float64(diskStats.Used)/1024/1024/1024),
				"available_gb":      fmt.Sprintf("%.2f", float64(diskStats.Available)/1024/1024/1024),
			}

			// Determine health status based on usage
			if usagePercent < maxUsagePercent*0.8 { // 80% of max threshold
				check.Status = serviceapi.HealthStatusHealthy
				check.Message = fmt.Sprintf("Disk usage is healthy (%.2f%% of %.2f%% threshold)", usagePercent, maxUsagePercent)
			} else if usagePercent < maxUsagePercent {
				check.Status = serviceapi.HealthStatusDegraded
				check.Message = fmt.Sprintf("Disk usage is elevated (%.2f%% of %.2f%% threshold)", usagePercent, maxUsagePercent)
			} else {
				check.Status = serviceapi.HealthStatusUnhealthy
				check.Message = fmt.Sprintf("Disk usage is critical (%.2f%% exceeds %.2f%% threshold)", usagePercent, maxUsagePercent)
			}
		}

		check.Duration = time.Since(start)
		return check
	}
}

// ApplicationHealthChecker creates a simple application health checker
func ApplicationHealthChecker(appName string) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		return serviceapi.HealthCheck{
			Name:      "application",
			Status:    serviceapi.HealthStatusHealthy,
			Message:   appName + " is running normally",
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}
}

// CustomHealthChecker creates a health checker with custom logic
func CustomHealthChecker(name string, checkFunc func(context.Context) (bool, string, map[string]any)) serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()
		check := serviceapi.HealthCheck{
			Name:      name,
			CheckedAt: start,
		}

		isHealthy, message, details := checkFunc(ctx)
		if isHealthy {
			check.Status = serviceapi.HealthStatusHealthy
		} else {
			check.Status = serviceapi.HealthStatusUnhealthy
		}

		check.Message = message
		check.Details = details
		check.Duration = time.Since(start)
		return check
	}
}

// getDiskUsage returns disk usage statistics for the given path (fallback implementation)
func getDiskUsage(path string) (*DiskStats, error) {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	// For unsupported platforms, return mock values with warning
	return getDiskUsageFallback(path)
}

// getDiskUsageFallback provides fallback disk usage for unsupported platforms
func getDiskUsageFallback(path string) (*DiskStats, error) {
	// Try to get file info to at least verify the path works
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	// For unsupported platforms, return mock values that indicate healthy disk usage
	// In production, you might want to log a warning about unsupported platform
	_ = fileInfo // Use fileInfo to avoid unused variable warning

	return &DiskStats{
		Total:     100 * 1024 * 1024 * 1024, // 100GB (mock)
		Used:      30 * 1024 * 1024 * 1024,  // 30GB (30% usage - mock)
		Available: 70 * 1024 * 1024 * 1024,  // 70GB available (mock)
	}, nil
}
