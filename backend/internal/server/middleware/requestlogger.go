package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLoggerOptions controls request logging behaviour.
type RequestLoggerOptions struct {
	Enabled         bool
	Headers         []string
	MaskHeaders     []string
	QueryParams     []string
	MaskQueryParams []string
}

// RequestLogger sends structured request logs to zap.
func RequestLogger(logger *zap.Logger, opts RequestLoggerOptions) gin.HandlerFunc {
	headerAllow := toLookup(opts.Headers)
	headerMask := toMaskSet(opts.MaskHeaders)
	queryAllow := toLookup(opts.QueryParams)
	queryMask := toMaskSet(opts.MaskQueryParams)

	return func(c *gin.Context) {
		if !opts.Enabled {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		if headerAllow != nil && len(headerAllow) > 0 {
			headers := extractHeaders(c, headerAllow, headerMask)
			if len(headers) > 0 {
				fields = append(fields, zap.Any("headers", headers))
			}
		}

		if queryAllow != nil && len(queryAllow) > 0 {
			queries := extractQueries(c, queryAllow, queryMask)
			if len(queries) > 0 {
				fields = append(fields, zap.Any("query", queries))
			}
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}

func toLookup(values []string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]string, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		set[lower] = trimmed
	}
	return set
}

func toMaskSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "" {
			continue
		}
		set[lower] = struct{}{}
	}
	return set
}

func extractHeaders(c *gin.Context, allow map[string]string, mask map[string]struct{}) map[string]string {
	if len(allow) == 0 {
		return nil
	}
	result := make(map[string]string)
	for lower, original := range allow {
		actual := c.Request.Header.Get(original)
		if actual == "" {
			continue
		}
		if mask != nil {
			if _, ok := mask[lower]; ok {
				result[original] = "***"
				continue
			}
		}
		result[original] = actual
	}
	return result
}

func extractQueries(c *gin.Context, allow map[string]string, mask map[string]struct{}) map[string]string {
	if len(allow) == 0 {
		return nil
	}
	values := c.Request.URL.Query()
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string)
	for lower, original := range allow {
		vals, ok := values[original]
		if !ok || len(vals) == 0 {
			continue
		}
		joined := strings.Join(vals, ",")
		if mask != nil {
			if _, ok := mask[lower]; ok {
				result[original] = "***"
				continue
			}
		}
		result[original] = joined
	}
	return result
}
