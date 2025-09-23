package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector wraps Prometheus metrics used by the API server.
type Collector struct {
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewCollector registers HTTP-level metrics with the default Prometheus registry.
func NewCollector(namespace, subsystem string) *Collector {
	if namespace == "" {
		namespace = "vpn_backend"
	}
	if subsystem == "" {
		subsystem = "http"
	}

	c := &Collector{}
	c.requestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Total number of processed HTTP requests.",
	}, []string{"method", "path", "status"})

	c.requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	return c
}

// ObserveHTTPRequest records an HTTP request metric sample.
func (c *Collector) ObserveHTTPRequest(method, path, status string, duration time.Duration) {
	if c == nil {
		return
	}
	c.requestTotal.WithLabelValues(method, path, status).Inc()
	c.requestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
}

// Handler returns an HTTP handler that exposes the registered metrics.
func (c *Collector) Handler() http.Handler {
	return promhttp.Handler()
}
