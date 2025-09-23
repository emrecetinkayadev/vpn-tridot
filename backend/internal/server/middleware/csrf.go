package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

// CSRFGuard enforces origin checks for state-changing requests.
func CSRFGuard(cfg config.CSRFConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	allowedOrigins := toOriginSet(cfg.AllowedOrigins)
	protectedMethods := toUpperSet(cfg.ProtectedMethods)
	allowNoOrigin := cfg.AllowNoOrigin

	return func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		if _, ok := protectedMethods[method]; !ok {
			c.Next()
			return
		}

		origin := strings.ToLower(strings.TrimSpace(c.Request.Header.Get("Origin")))
		if origin == "" {
			if allowNoOrigin {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin required"})
			return
		}

		if len(allowedOrigins) > 0 {
			if _, allowAll := allowedOrigins["*"]; !allowAll {
				if _, ok := allowedOrigins[origin]; !ok {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
					return
				}
			}
		}

		c.Next()
	}
}

func toOriginSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			set["*"] = struct{}{}
			continue
		}
		set[strings.ToLower(trimmed)] = struct{}{}
	}
	return set
}

func toUpperSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		upper := strings.ToUpper(strings.TrimSpace(value))
		if upper == "" {
			continue
		}
		set[upper] = struct{}{}
	}
	return set
}
