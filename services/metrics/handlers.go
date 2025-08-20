package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GetHTTPHandler returns an HTTP handler for the metrics endpoint
func (m *MetricsService) GetHTTPHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry: m.registry,
	})
}

// RegisterCustomMetrics allows registering custom Prometheus collectors
func (m *MetricsService) RegisterCustomMetrics(collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		if err := m.registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}

// UnregisterCustomMetrics allows unregistering custom Prometheus collectors
func (m *MetricsService) UnregisterCustomMetrics(collectors ...prometheus.Collector) {
	for _, collector := range collectors {
		m.registry.Unregister(collector)
	}
}

// GetMetricsSummary returns a summary of registered metrics
func (m *MetricsService) GetMetricsSummary() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]int{
		"counters":   len(m.counters),
		"gauges":     len(m.gauges),
		"histograms": len(m.histograms),
	}
}
