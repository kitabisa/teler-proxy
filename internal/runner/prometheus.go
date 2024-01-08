package runner

import (
	"strconv"
	"time"

	"net/http"

	"github.com/kitabisa/teler-proxy/internal/metrics"
	"github.com/kitabisa/teler-waf/threat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Prometheus struct {
	*metrics.Metrics
	*prometheus.Registry
}

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *Runner) initMetrics() error {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())

	m := metrics.New(reg)

	threatLabels := make(prometheus.Labels)
	for _, t := range threat.List() {
		count, err := t.Count()
		if err != nil {
			return err
		}

		threatLabels[t.String()] = strconv.Itoa(count)
	}

	m.Threats.With(threatLabels).Set(1)

	r.Prometheus.Metrics = m
	r.Prometheus.Registry = reg

	return nil
}

func (run *Runner) promMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()

		customResponseWriter := &customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(customResponseWriter, req)

		run.Prometheus.Metrics.Duration.With(prometheus.Labels{
			"method": req.Method,
			"status": customResponseWriter.StatusString(),
			"path":   req.URL.Path,
		}).Observe(time.Since(now).Seconds())
	})
}

func (w *customResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *customResponseWriter) Status() int {
	return w.statusCode
}

func (w *customResponseWriter) StatusString() string {
	return strconv.Itoa(w.Status())
}
