package metrics

import (
	"sync"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsService implements serviceapi.Metrics using Prometheus
type MetricsService struct {
	config     *MetricsConfig
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	registry   *prometheus.Registry
	mu         sync.RWMutex
}

// MetricsConfig holds configuration for metrics service
type MetricsConfig struct {
	Enabled               bool              `json:"enabled"`
	Endpoint              string            `json:"endpoint"`
	Namespace             string            `json:"namespace"`
	Subsystem             string            `json:"subsystem"`
	Buckets               []float64         `json:"buckets"`
	Labels                map[string]string `json:"labels"`
	CollectInterval       string            `json:"collect_interval"`
	Host                  string            `json:"host"`
	Port                  int               `json:"port"`
	Timeout               string            `json:"timeout"`
	IncludeGoMetrics      bool              `json:"include_go_metrics"`
	IncludeProcessMetrics bool              `json:"include_process_metrics"`
}

// NewService creates a new metrics service
func NewService() (*MetricsService, error) {
	return NewServiceWithConfig(&MetricsConfig{
		Enabled:               true,
		Endpoint:              "/metrics",
		Namespace:             "",
		Subsystem:             "",
		Buckets:               prometheus.DefBuckets,
		Labels:                make(map[string]string),
		CollectInterval:       "15s",
		Host:                  "localhost",
		Port:                  0, // 0 means no separate server
		Timeout:               "10s",
		IncludeGoMetrics:      true,
		IncludeProcessMetrics: true,
	})
}

// NewServiceWithConfig creates a new metrics service with custom configuration
func NewServiceWithConfig(config *MetricsConfig) (*MetricsService, error) {
	registry := prometheus.NewRegistry()

	// Add default collectors based on configuration
	if config.IncludeGoMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
	}
	if config.IncludeProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	service := &MetricsService{
		config:     config,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		registry:   registry,
	}

	return service, nil
}

// IncCounter implements serviceapi.Metrics
func (m *MetricsService) IncCounter(name string, labels serviceapi.Labels) {
	if !m.config.Enabled {
		return
	}

	counter := m.getOrCreateCounter(name, labels)
	if counter != nil {
		counter.With(prometheus.Labels(labels)).Inc()
	}
}

// ObserveHistogram implements serviceapi.Metrics
func (m *MetricsService) ObserveHistogram(name string, value float64, labels serviceapi.Labels) {
	if !m.config.Enabled {
		return
	}

	histogram := m.getOrCreateHistogram(name, labels)
	if histogram != nil {
		histogram.With(prometheus.Labels(labels)).Observe(value)
	}
}

// SetGauge implements serviceapi.Metrics
func (m *MetricsService) SetGauge(name string, value float64, labels serviceapi.Labels) {
	if !m.config.Enabled {
		return
	}

	gauge := m.getOrCreateGauge(name, labels)
	if gauge != nil {
		gauge.With(prometheus.Labels(labels)).Set(value)
	}
}

// GetRegistry returns the Prometheus registry for HTTP handler
func (m *MetricsService) GetRegistry() *prometheus.Registry {
	return m.registry
}

// GetConfig returns the metrics configuration
func (m *MetricsService) GetConfig() *MetricsConfig {
	return m.config
}

// getOrCreateCounter gets or creates a counter metric
func (m *MetricsService) getOrCreateCounter(name string, labels serviceapi.Labels) *prometheus.CounterVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.counters[name]; exists {
		return counter
	}

	labelKeys := m.getLabelKeys(labels)
	counter := promauto.With(m.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        "Counter metric: " + name,
			ConstLabels: prometheus.Labels(m.config.Labels),
		},
		labelKeys,
	)

	m.counters[name] = counter
	return counter
}

// getOrCreateGauge gets or creates a gauge metric
func (m *MetricsService) getOrCreateGauge(name string, labels serviceapi.Labels) *prometheus.GaugeVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	labelKeys := m.getLabelKeys(labels)
	gauge := promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        "Gauge metric: " + name,
			ConstLabels: prometheus.Labels(m.config.Labels),
		},
		labelKeys,
	)

	m.gauges[name] = gauge
	return gauge
}

// getOrCreateHistogram gets or creates a histogram metric
func (m *MetricsService) getOrCreateHistogram(name string, labels serviceapi.Labels) *prometheus.HistogramVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	labelKeys := m.getLabelKeys(labels)
	histogram := promauto.With(m.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   m.config.Namespace,
			Subsystem:   m.config.Subsystem,
			Name:        name,
			Help:        "Histogram metric: " + name,
			ConstLabels: prometheus.Labels(m.config.Labels),
			Buckets:     m.config.Buckets,
		},
		labelKeys,
	)

	m.histograms[name] = histogram
	return histogram
}

// getLabelKeys extracts label keys from labels map
func (m *MetricsService) getLabelKeys(labels serviceapi.Labels) []string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	return keys
}

// parseCollectInterval parses collect interval string to duration
// func (m *MetricsService) parseCollectInterval() time.Duration {
// 	if duration, err := time.ParseDuration(m.config.CollectInterval); err == nil {
// 		return duration
// 	}
// 	return 15 * time.Second // default
// }

// Verify interface compliance
var _ serviceapi.Metrics = (*MetricsService)(nil)
var _ service.Service = (*MetricsService)(nil)
