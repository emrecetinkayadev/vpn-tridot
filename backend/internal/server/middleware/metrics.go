package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/metrics"
)

// Metrics records HTTP request metrics via the provided collector.
func Metrics(collector *metrics.Collector) gin.HandlerFunc {
	if collector == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		status := strconv.Itoa(c.Writer.Status())
		collector.ObserveHTTPRequest(c.Request.Method, path, status, time.Since(start))
	}
}
