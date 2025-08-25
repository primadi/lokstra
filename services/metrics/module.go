package metrics

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/utils"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/prometheus/client_golang/prometheus"
)

const FACTORY_NAME = "metrics"

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return FACTORY_NAME
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		metricsConfig := &MetricsConfig{
			Enabled:               true,
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

		switch v := config.(type) {
		case map[string]any:
			// Parse enabled
			if enabled, ok := v["enabled"].(bool); ok {
				metricsConfig.Enabled = enabled
			}

			// Parse endpoint
			if endpoint := utils.GetValueFromMap(v, "endpoint", "/metrics"); endpoint != "" {
				metricsConfig.Endpoint = endpoint
			}

			// Parse namespace
			if namespace := utils.GetValueFromMap(v, "namespace", ""); namespace != "" {
				metricsConfig.Namespace = namespace
			}

			// Parse subsystem
			if subsystem := utils.GetValueFromMap(v, "subsystem", ""); subsystem != "" {
				metricsConfig.Subsystem = subsystem
			}

			// Parse buckets
			if bucketsInterface, ok := v["buckets"]; ok {
				if buckets, err := parseBuckets(bucketsInterface); err == nil {
					metricsConfig.Buckets = buckets
				} else {
					fmt.Printf("Invalid buckets configuration, using default: %v\n", err)
				}
			}

			// Parse labels
			if labelsInterface, ok := v["labels"]; ok {
				if labels, err := parseLabels(labelsInterface); err == nil {
					metricsConfig.Labels = labels
				} else {
					fmt.Printf("Invalid labels configuration, using default: %v\n", err)
				}
			}

			// Parse collect_interval
			if collectInterval := utils.GetValueFromMap(v, "collect_interval", "15s"); collectInterval != "" {
				if _, err := time.ParseDuration(collectInterval); err == nil {
					metricsConfig.CollectInterval = collectInterval
				} else {
					fmt.Printf("Invalid collect_interval '%s', using default '15s': %v\n", collectInterval, err)
					metricsConfig.CollectInterval = "15s"
				}
			}

			// Parse host
			if host := utils.GetValueFromMap(v, "host", "localhost"); host != "" {
				metricsConfig.Host = host
			}

			// Parse port
			if portInterface, ok := v["port"]; ok {
				if port, ok := portInterface.(int); ok {
					metricsConfig.Port = port
				} else if portFloat, ok := portInterface.(float64); ok {
					metricsConfig.Port = int(portFloat)
				}
			}

			// Parse timeout
			if timeout := utils.GetValueFromMap(v, "timeout", "10s"); timeout != "" {
				if _, err := time.ParseDuration(timeout); err == nil {
					metricsConfig.Timeout = timeout
				} else {
					fmt.Printf("Invalid timeout '%s', using default '10s': %v\n", timeout, err)
					metricsConfig.Timeout = "10s"
				}
			}

			// Parse include_go_metrics
			if includeGo, ok := v["include_go_metrics"].(bool); ok {
				metricsConfig.IncludeGoMetrics = includeGo
			}

			// Parse include_process_metrics
			if includeProcess, ok := v["include_process_metrics"].(bool); ok {
				metricsConfig.IncludeProcessMetrics = includeProcess
			}

		case nil:
			// Use default configuration
		default:
			return nil, fmt.Errorf("unsupported configuration type for metrics service: %T", config)
		}

		return NewServiceWithConfig(metricsConfig)
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "Metrics Service for Lokstra using Prometheus"
}

// parseBuckets converts interface{} to []float64 for histogram buckets
func parseBuckets(bucketsInterface interface{}) ([]float64, error) {
	switch v := bucketsInterface.(type) {
	case []interface{}:
		buckets := make([]float64, len(v))
		for i, bucket := range v {
			switch b := bucket.(type) {
			case float64:
				buckets[i] = b
			case int:
				buckets[i] = float64(b)
			case int64:
				buckets[i] = float64(b)
			default:
				return nil, fmt.Errorf("bucket at index %d is not a number: %T", i, bucket)
			}
		}
		return buckets, nil
	case []float64:
		return v, nil
	default:
		return nil, fmt.Errorf("buckets must be an array of numbers, got %T", bucketsInterface)
	}
}

// parseLabels converts interface{} to map[string]string for constant labels
func parseLabels(labelsInterface interface{}) (map[string]string, error) {
	switch v := labelsInterface.(type) {
	case map[string]interface{}:
		labels := make(map[string]string)
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				labels[key] = strValue
			} else {
				labels[key] = fmt.Sprintf("%v", value)
			}
		}
		return labels, nil
	case map[string]string:
		return v, nil
	default:
		return nil, fmt.Errorf("labels must be an object, got %T", labelsInterface)
	}
}

// GetModule returns the metrics service with serviceType "lokstra.metrics".
func GetModule() registration.Module {
	return &module{}
}

var _ registration.Module = (*module)(nil)
