package setup

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/middleware"
)

// Dependencies lists route handlers required by the router.
type Dependencies struct {
	AuthHandler interface {
		SignUp(*gin.Context)
		VerifyEmail(*gin.Context)
		Login(*gin.Context)
		Refresh(*gin.Context)
		RequestPasswordReset(*gin.Context)
		ConfirmPasswordReset(*gin.Context)
		SetupTOTP(*gin.Context)
		ConfirmTOTP(*gin.Context)
		DisableTOTP(*gin.Context)
	}
	AuthMiddleware gin.HandlerFunc
	BillingHandler interface {
		ListPlans(*gin.Context)
		CreateCheckoutSession(*gin.Context)
		StripeWebhook(*gin.Context)
		ListPayments(*gin.Context)
	}
	RegionsHandler interface {
		List(*gin.Context)
	}
	NodesHandler interface {
		Register(*gin.Context)
		ReportHealth(*gin.Context)
	}
}

// Register wires the public routes for the API server.
func Register(engine *gin.Engine, cfg config.Config, deps Dependencies, logger *zap.Logger) {
	health := engine.Group("/health")
	health.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})
	health.GET("/ready", func(c *gin.Context) {
		if cfg.Auth.JWTSigningKey == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "reason": "missing jwt secret"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")
	api.Use(middleware.RateLimit(cfg.RateLimit))

	if deps.BillingHandler != nil {
		api.GET("/plans", deps.BillingHandler.ListPlans)
	}
	if deps.RegionsHandler != nil {
		api.GET("/regions", deps.RegionsHandler.List)
	}

	authGroup := api.Group("/auth")
	authGroup.POST("/signup", deps.AuthHandler.SignUp)
	authGroup.POST("/verify-email", deps.AuthHandler.VerifyEmail)
	authGroup.POST("/login", deps.AuthHandler.Login)
	authGroup.POST("/refresh", deps.AuthHandler.Refresh)
	authGroup.POST("/password-reset/request", deps.AuthHandler.RequestPasswordReset)
	authGroup.POST("/password-reset/confirm", deps.AuthHandler.ConfirmPasswordReset)

	protected := api.Group("")
	if deps.AuthMiddleware != nil {
		protected.Use(deps.AuthMiddleware)
	}
	protected.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	if deps.BillingHandler != nil {
		protected.POST("/checkout", deps.BillingHandler.CreateCheckoutSession)
		protected.GET("/account/payments", deps.BillingHandler.ListPayments)
		engine.POST("/api/v1/webhooks/stripe", deps.BillingHandler.StripeWebhook)
	}
	if deps.NodesHandler != nil {
		engine.POST("/api/v1/nodes/register", deps.NodesHandler.Register)
		engine.POST("/api/v1/nodes/health", deps.NodesHandler.ReportHealth)
	}

	authProtected := protected.Group("/auth")
	authProtected.POST("/totp/setup", deps.AuthHandler.SetupTOTP)
	authProtected.POST("/totp/confirm", deps.AuthHandler.ConfirmTOTP)
	authProtected.POST("/totp/disable", deps.AuthHandler.DisableTOTP)
}
