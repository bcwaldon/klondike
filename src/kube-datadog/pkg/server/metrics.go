package server

type MetricsBundle struct {
	metrics chan Metric
}

func (b *MetricsBundle) Metrics() <-chan Metric {
	return b.metrics
}

type Metric struct {
	Name  string
	Value int
	Tags  []string
}
