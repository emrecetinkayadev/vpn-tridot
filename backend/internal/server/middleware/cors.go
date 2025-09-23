package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

// CORS configures cross-origin resource sharing for the API.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	corsCfg := cors.Config{
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	allowAll := false
	allowOrigins := make([]string, 0, len(cfg.AllowOrigins))
	for _, origin := range cfg.AllowOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "*" {
			allowAll = true
			continue
		}
		if trimmed != "" {
			allowOrigins = append(allowOrigins, trimmed)
		}
	}

	if allowAll {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = allowOrigins
	}

	if corsCfg.MaxAge <= 0 {
		corsCfg.MaxAge = 10 * time.Minute
	}

	return cors.New(corsCfg)
}
