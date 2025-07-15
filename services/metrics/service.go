package metrics

import (
	"context"
	"fmt"
	"lokstra/common/iface"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsService struct {
	instanceName string
	registry     *prometheus.Registry
	config       map[string]any
	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
}

func (m *MetricsService) InstanceName() string {
	return m.instanceName
}

func (m *MetricsService) GetConfig(key string) any {
	return m.config[key]
}

func (m *MetricsService) GetRegistry() *prometheus.Registry {
	return m.registry
}

func (m *MetricsService) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *MetricsService) IncrementCounter(name string, labels prometheus.Labels) {
	if counter, ok := m.getCounter(name); ok {
		counter.With(labels).Inc()
	}
}

func (m *MetricsService) RecordDuration(name string, duration time.Duration, labels prometheus.Labels) {
	if histogram, ok := m.getHistogram(name); ok {
		histogram.With(labels).Observe(duration.Seconds())
	}
}

func (m *MetricsService) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	m.httpRequests.WithLabelValues(method, path, status).Inc()
	m.httpDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func (m *MetricsService) RegisterCounter(name, help string, labels []string) error {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	return m.registry.Register(counter)
}

func (m *MetricsService) RegisterHistogram(name, help string, labels []string, buckets []float64) error {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	return m.registry.Register(histogram)
}

func (m *MetricsService) getCounter(name string) (*prometheus.CounterVec, bool) {
	metric, err := m.registry.Gather()
	if err != nil {
		return nil, false
	}
	
	for _, mf := range metric {
		if mf.GetName() == name && mf.GetType() == prometheus.MetricType_COUNTER {
			return nil, true
		}
	}
	return nil, false
}

func (m *MetricsService) getHistogram(name string) (*prometheus.HistogramVec, bool) {
	metric, err := m.registry.Gather()
	if err != nil {
		return nil, false
	}
	
	for _, mf := range metric {
		if mf.GetName() == name && mf.GetType() == prometheus.MetricType_HISTOGRAM {
			return nil, true
		}
	}
	return nil, false
}

func newMetricsService(instanceName string, config map[string]any) (*MetricsService, error) {
	registry := prometheus.NewRegistry()

	httpRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	registry.MustRegister(httpRequests)
	registry.MustRegister(httpDuration)

	return &MetricsService{
		instanceName: instanceName,
		registry:     registry,
		config:       config,
		httpRequests: httpRequests,
		httpDuration: httpDuration,
	}, nil
}
