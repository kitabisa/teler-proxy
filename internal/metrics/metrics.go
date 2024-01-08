package metrics

import (
	"github.com/kitabisa/teler-waf/threat"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Threats  *prometheus.GaugeVec
	Events   *prometheus.CounterVec
	Duration *prometheus.HistogramVec
}

func New(registry prometheus.Registerer) *Metrics {
	m := new(Metrics)

	var threats []string
	for _, t := range threat.List() {
		threats = append(threats, t.String())
	}

	m.Threats = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "threat_datasets",
		Help:      "Number of avaiable threat datasets.",
	}, threats)

	m.Events = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "teler_event",
		Help:      "Number of incoming teler WAF events.",
	}, []string{"rule", "threat"})

	m.Duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "request_duration_seconds",
		Help:      "Request duration times in seconds.",
		Buckets:   []float64{0.1, 0.15, 0.2, 0.25, 0.3},
	}, []string{"status", "method", "path"})

	registry.MustRegister(m.Threats, m.Events, m.Duration)

	return m
}
