package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/jwt"
)

// Auth ensures requests carry a placeholder bearer token until full JWT support lands.
func Auth(manager *jwt.Manager, security config.AdminSecurityConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		claims, err := manager.Verify(token)
		if err != nil {
			logger.Warn("jwt verification failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			logger.Warn("invalid user id in token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", uid)

		if len(security.IPAllowlist) > 0 {
			clientIP := clientIPFromContext(c)
			if !ipAllowed(clientIP, security.IPAllowlist) {
				logger.Warn("admin ip not allowed", zap.String("ip", clientIP))
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ip not allowed"})
				return
			}
		}
		c.Next()
	}
}

func clientIPFromContext(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(c.Request.RemoteAddr)
	}
	return ip
}

func ipAllowed(client string, allowlist []string) bool {
	if client == "" {
		return false
	}
	for _, entry := range allowlist {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if ipInCIDR(client, entry) {
				return true
			}
			continue
		}
		if client == entry {
			return true
		}
	}
	return false
}

func ipInCIDR(client, cidr string) bool {
	ip := net.ParseIP(client)
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}
