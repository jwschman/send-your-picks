package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics for tracking HTTP request performance and volume
var (
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "syp_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status_code"},
	)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syp_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status_code"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "syp_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed.",
		},
	)
)

// init registers all Prometheus metrics collectors on package load
func init() {
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestsInFlight)
}

// MetricsMiddleware records request duration, total count, and in-flight gauge for each request
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		httpRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		httpRequestsInFlight.Dec()

		status := strconv.Itoa(c.Writer.Status())
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		httpRequestDuration.WithLabelValues(c.Request.Method, route, status).Observe(time.Since(start).Seconds())
		httpRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
	}
}
