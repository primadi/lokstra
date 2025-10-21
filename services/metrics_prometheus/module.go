package metrics_prometheus

import (
	"sync"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/old_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const SERVICE_TYPE = "metrics_prometheus"

// Config represents the configuration for Prometheus metrics service.
type Config struct {
	Namespace string `json:"namespace" yaml:"namespace"` // namespace for all metrics
	Subsystem string `json:"subsystem" yaml:"subsystem"` // subsystem for all metrics
}

type metricsPrometheus struct {
	cfg      *Config
	registry *prometheus.Registry
	counters map[string]*prometheus.CounterVec
	histos   map[string]*prometheus.HistogramVec
	gauges   map[string]*prometheus.GaugeVec
	mu       sync.RWMutex
}

var _ serviceapi.Metrics = (*metricsPrometheus)(nil)

func (m *metricsPrometheus) IncCounter(name string, labels serviceapi.Labels) {
	m.mu.RLock()
	counter, exists := m.counters[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double check after acquiring write lock
		counter, exists = m.counters[name]
		if !exists {
			counter = promauto.With(m.registry).NewCounterVec(
				prometheus.CounterOpts{
					Namespace: m.cfg.Namespace,
					Subsystem: m.cfg.Subsystem,
					Name:      name,
					Help:      name,
				},
				m.getLabelKeys(labels),
			)
			m.counters[name] = counter
		}
		m.mu.Unlock()
	}

	counter.With(prometheus.Labels(labels)).Inc()
}

func (m *metricsPrometheus) ObserveHistogram(name string, value float64, labels serviceapi.Labels) {
	m.mu.RLock()
	histo, exists := m.histos[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double check after acquiring write lock
		histo, exists = m.histos[name]
		if !exists {
			histo = promauto.With(m.registry).NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: m.cfg.Namespace,
					Subsystem: m.cfg.Subsystem,
					Name:      name,
					Help:      name,
					Buckets:   prometheus.DefBuckets,
				},
				m.getLabelKeys(labels),
			)
			m.histos[name] = histo
		}
		m.mu.Unlock()
	}

	histo.With(prometheus.Labels(labels)).Observe(value)
}

func (m *metricsPrometheus) SetGauge(name string, value float64, labels serviceapi.Labels) {
	m.mu.RLock()
	gauge, exists := m.gauges[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double check after acquiring write lock
		gauge, exists = m.gauges[name]
		if !exists {
			gauge = promauto.With(m.registry).NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: m.cfg.Namespace,
					Subsystem: m.cfg.Subsystem,
					Name:      name,
					Help:      name,
				},
				m.getLabelKeys(labels),
			)
			m.gauges[name] = gauge
		}
		m.mu.Unlock()
	}

	gauge.With(prometheus.Labels(labels)).Set(value)
}

func (m *metricsPrometheus) getLabelKeys(labels serviceapi.Labels) []string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	return keys
}

func (m *metricsPrometheus) Registry() *prometheus.Registry {
	return m.registry
}

func (m *metricsPrometheus) Shutdown() error {
	return nil
}

func Service(cfg *Config) *metricsPrometheus {
	registry := prometheus.NewRegistry()
	return &metricsPrometheus{
		cfg:      cfg,
		registry: registry,
		counters: make(map[string]*prometheus.CounterVec),
		histos:   make(map[string]*prometheus.HistogramVec),
		gauges:   make(map[string]*prometheus.GaugeVec),
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		Namespace: utils.GetValueFromMap(params, "namespace", "app"),
		Subsystem: utils.GetValueFromMap(params, "subsystem", ""),
	}
	return Service(cfg)
}

func Register() {
	old_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		old_registry.AllowOverride(true))
}
