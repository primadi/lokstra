package metrics

import (
	"testing"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetricsService_NewService(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	if service.config == nil {
		t.Fatal("Config should not be nil")
	}

	if !service.config.Enabled {
		t.Error("Service should be enabled by default")
	}
}

func TestMetricsService_IncCounter(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	// Increment counter
	labels := serviceapi.Labels{"method": "GET", "status": "200"}
	service.IncCounter("test_requests_total", labels)

	// Verify counter exists and has correct value
	if counter, exists := service.counters["test_requests_total"]; exists {
		metric := &dto.Metric{}
		counter.With(prometheus.Labels(labels)).Write(metric)

		if metric.GetCounter().GetValue() != 1.0 {
			t.Errorf("Expected counter value 1.0, got %f", metric.GetCounter().GetValue())
		}
	} else {
		t.Error("Counter should exist after increment")
	}
}

func TestMetricsService_SetGauge(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	// Set gauge value
	labels := serviceapi.Labels{"pool": "db"}
	service.SetGauge("test_connections", 42.0, labels)

	// Verify gauge exists and has correct value
	if gauge, exists := service.gauges["test_connections"]; exists {
		metric := &dto.Metric{}
		gauge.With(prometheus.Labels(labels)).Write(metric)

		if metric.GetGauge().GetValue() != 42.0 {
			t.Errorf("Expected gauge value 42.0, got %f", metric.GetGauge().GetValue())
		}
	} else {
		t.Error("Gauge should exist after set")
	}
}

func TestMetricsService_ObserveHistogram(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	// Observe histogram value
	labels := serviceapi.Labels{"endpoint": "/users"}
	service.ObserveHistogram("test_duration_seconds", 0.5, labels)

	// Verify histogram exists
	if histogram, exists := service.histograms["test_duration_seconds"]; exists {
		// For histogram, we need to get the metric family
		metricFamily, err := service.registry.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range metricFamily {
			if mf.GetName() == "test_duration_seconds" {
				found = true
				if len(mf.GetMetric()) > 0 {
					metric := mf.GetMetric()[0]
					if metric.GetHistogram().GetSampleCount() != 1 {
						t.Errorf("Expected histogram sample count 1, got %d", metric.GetHistogram().GetSampleCount())
					}
				}
				break
			}
		}

		if !found {
			t.Error("Histogram metric not found in registry")
		}

		_ = histogram // Suppress unused variable warning
	} else {
		t.Error("Histogram should exist after observe")
	}
}

func TestMetricsService_DisabledService(t *testing.T) {
	config := &MetricsConfig{
		Enabled:               false,
		Endpoint:              "/metrics",
		Namespace:             "",
		Subsystem:             "",
		Buckets:               prometheus.DefBuckets,
		Labels:                make(map[string]string),
		CollectInterval:       "15s",
		Host:                  "localhost",
		Port:                  0,
		Timeout:               "10s",
		IncludeGoMetrics:      true,
		IncludeProcessMetrics: true,
	}

	service, err := NewServiceWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	// Operations on disabled service should not create metrics
	labels := serviceapi.Labels{"test": "label"}
	service.IncCounter("disabled_counter", labels)
	service.SetGauge("disabled_gauge", 1.0, labels)
	service.ObserveHistogram("disabled_histogram", 1.0, labels)

	// Verify no metrics were created
	if len(service.counters) > 0 {
		t.Error("Disabled service should not create counters")
	}
	if len(service.gauges) > 0 {
		t.Error("Disabled service should not create gauges")
	}
	if len(service.histograms) > 0 {
		t.Error("Disabled service should not create histograms")
	}
}

func TestMetricsService_CustomConfig(t *testing.T) {
	config := &MetricsConfig{
		Enabled:               true,
		Endpoint:              "/custom-metrics",
		Namespace:             "test_app",
		Subsystem:             "api",
		Buckets:               []float64{0.1, 1.0, 10.0},
		Labels:                map[string]string{"service": "test"},
		CollectInterval:       "30s",
		Host:                  "0.0.0.0",
		Port:                  9090,
		Timeout:               "5s",
		IncludeGoMetrics:      false,
		IncludeProcessMetrics: false,
	}

	service, err := NewServiceWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	if service.config.Namespace != "test_app" {
		t.Errorf("Expected namespace 'test_app', got '%s'", service.config.Namespace)
	}

	if service.config.Subsystem != "api" {
		t.Errorf("Expected subsystem 'api', got '%s'", service.config.Subsystem)
	}

	if service.config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", service.config.Port)
	}
}

func TestMetricsService_GetMetricsSummary(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create metrics service: %v", err)
	}

	// Create some metrics
	labels := serviceapi.Labels{"test": "label"}
	service.IncCounter("test_counter", labels)
	service.SetGauge("test_gauge", 1.0, labels)
	service.ObserveHistogram("test_histogram", 1.0, labels)

	summary := service.GetMetricsSummary()

	if summary["counters"] != 1 {
		t.Errorf("Expected 1 counter, got %d", summary["counters"])
	}
	if summary["gauges"] != 1 {
		t.Errorf("Expected 1 gauge, got %d", summary["gauges"])
	}
	if summary["histograms"] != 1 {
		t.Errorf("Expected 1 histogram, got %d", summary["histograms"])
	}
}
