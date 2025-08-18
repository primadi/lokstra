package serviceapi

type Metrics interface {
	IncCounter(name string, labels Labels)
	ObserveHistogram(name string, value float64, labels Labels)
	SetGauge(name string, value float64, labels Labels)
}

type Labels = map[string]string
