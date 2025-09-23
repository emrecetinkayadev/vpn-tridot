package billinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/billing"
)

// Handler wires billing endpoints to the service layer.
type Handler struct {
	service *billing.Service
	logger  *zap.Logger
}

func New(service *billing.Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// ListPlans returns active subscription plans.
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("list plans failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// CreateCheckoutSession generates a checkout session for the authenticated user.
func (h *Handler) CreateCheckoutSession(c *gin.Context) {
	type request struct {
		PlanCode string `json:"plan_code" binding:"required"`
		Provider string `json:"provider"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	session, err := h.service.CreateCheckoutSession(c.Request.Context(), userID, req.PlanCode, req.Provider)
	if err != nil {
		switch err {
		case billing.ErrPlanNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case billing.ErrUnsupportedProvider, billing.ErrProviderDisabled:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			h.logger.Error("create checkout session failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"provider":     session.Provider,
		"session_id":   session.SessionID,
		"checkout_url": session.URL,
	})
}

// StripeWebhook handles Stripe webhook callbacks.
func (h *Handler) StripeWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		h.logger.Warn("read stripe webhook payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature"})
		return
	}

	if err := h.service.HandleWebhook(c.Request.Context(), "stripe", payload, signature); err != nil {
		switch err {
		case billing.ErrInvalidWebhook:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			h.logger.Error("stripe webhook handling failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook handling failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListPayments exposes payment history for the authenticated user.
func (h *Handler) ListPayments(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	payments, err := h.service.ListPayments(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list payments failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.UUID{}, false
	}

	switch v := value.(type) {
	case uuid.UUID:
		return v, true
	case string:
		uid, err := uuid.Parse(v)
		if err != nil {
			return uuid.UUID{}, false
		}
		return uid, true
	default:
		return uuid.UUID{}, false
	}
}
