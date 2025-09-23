package middleware

import (
	"time"

	tollbooth "github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-gonic/gin"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

// RateLimit adds a token-bucket rate limiter for the API endpoints.
func RateLimit(rule config.RateLimitRuleConfig) gin.HandlerFunc {
	if !rule.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	lim := tollbooth.NewLimiter(rule.RequestsPerSecond, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lim.SetBurst(rule.Burst)
	lim.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})

	return tollbooth_gin.LimitHandler(lim)
}
