package authhandler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/auth"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/hcaptcha"
)

// Handler exposes authentication HTTP endpoints.
type Handler struct {
	service *auth.Service
	logger  *zap.Logger
	captcha CaptchaVerifier
}

// CaptchaVerifier abstracts captcha verification.
type CaptchaVerifier interface {
	Enabled() bool
	Verify(ctx context.Context, token, remoteIP string) error
}

func New(service *auth.Service, captcha CaptchaVerifier, logger *zap.Logger) *Handler {
	return &Handler{service: service, captcha: captcha, logger: logger}
}

func (h *Handler) SignUp(c *gin.Context) {
	type request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Captcha  string `json:"hcaptcha_token"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validateCaptcha(c.Request.Context(), req.Captcha, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	user, token, err := h.service.SignUp(c.Request.Context(), req.Email, req.Password, now)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case auth.ErrEmailAlreadyUsed:
			status = http.StatusConflict
		default:
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":                  user.ID,
		"email_verification_token": token,
	})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	type request struct {
		Token string `json:"token" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	if err := h.service.VerifyEmail(c.Request.Context(), req.Token, now); err != nil {
		status := http.StatusBadRequest
		if err == auth.ErrTokenNotFound {
			status = http.StatusGone
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "verified"})
}

func (h *Handler) Login(c *gin.Context) {
	type request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		OTP      string `json:"otp"`
		Captcha  string `json:"hcaptcha_token"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validateCaptcha(c.Request.Context(), req.Captcha, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta := auth.Metadata{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}
	now := time.Now().UTC()
	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password, req.OTP, meta, now)
	if err != nil {
		status := http.StatusUnauthorized
		switch err {
		case auth.ErrEmailNotVerified:
			status = http.StatusForbidden
		case auth.ErrInvalidCredentials:
			status = http.StatusUnauthorized
		case auth.ErrTOTPRequired:
			status = http.StatusUnauthorized
		case auth.ErrInvalidTOTP:
			status = http.StatusUnauthorized
		default:
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    int(result.ExpiresIn.Seconds()),
		"user_id":       result.UserID,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	type request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta := auth.Metadata{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}
	now := time.Now().UTC()
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, meta, now)
	if err != nil {
		status := http.StatusUnauthorized
		if err == auth.ErrTokenNotFound {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    int(result.ExpiresIn.Seconds()),
		"user_id":       result.UserID,
	})
}

func (h *Handler) RequestPasswordReset(c *gin.Context) {
	type request struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	token, err := h.service.RequestPasswordReset(c.Request.Context(), req.Email, now)
	if err != nil {
		h.logger.Error("password reset request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	response := gin.H{"status": "ok"}
	if token != "" {
		response["reset_token"] = token
	}

	c.JSON(http.StatusAccepted, response)
}

func (h *Handler) ConfirmPasswordReset(c *gin.Context) {
	type request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, now); err != nil {
		status := http.StatusBadRequest
		if err == auth.ErrTokenNotFound {
			status = http.StatusGone
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "password_reset"})
}

func (h *Handler) validateCaptcha(ctx context.Context, token, remoteIP string) error {
	if h.captcha == nil || !h.captcha.Enabled() {
		return nil
	}
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return hcaptcha.ErrTokenMissing
	}
	if err := h.captcha.Verify(ctx, trimmed, remoteIP); err != nil {
		if errors.Is(err, hcaptcha.ErrTokenMissing) || errors.Is(err, hcaptcha.ErrVerificationFailed) {
			return err
		}
		h.logger.Warn("captcha verification error", zap.Error(err))
		return hcaptcha.ErrVerificationFailed
	}
	return nil
}

func (h *Handler) SetupTOTP(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	now := time.Now().UTC()
	secret, uri, err := h.service.SetupTOTP(c.Request.Context(), userID, now)
	if err != nil {
		status := http.StatusBadRequest
		switch err {
		case auth.ErrTOTPAlreadyEnabled:
			status = http.StatusConflict
		case auth.ErrInvalidCredentials:
			status = http.StatusNotFound
		default:
			h.logger.Error("setup totp failed", zap.Error(err))
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":      secret,
		"otpauth_url": uri,
	})
}

func (h *Handler) ConfirmTOTP(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type request struct {
		Code string `json:"code" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	if err := h.service.ConfirmTOTP(c.Request.Context(), userID, req.Code, now); err != nil {
		status := http.StatusBadRequest
		switch err {
		case auth.ErrTOTPNotSetup:
			status = http.StatusPreconditionFailed
		case auth.ErrInvalidTOTP:
			status = http.StatusUnauthorized
		case auth.ErrInvalidCredentials:
			status = http.StatusNotFound
		default:
			h.logger.Error("confirm totp failed", zap.Error(err))
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "totp_enabled"})
}

func (h *Handler) DisableTOTP(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type request struct {
		Code string `json:"code" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	if err := h.service.DisableTOTP(c.Request.Context(), userID, req.Code, now); err != nil {
		status := http.StatusBadRequest
		switch err {
		case auth.ErrTOTPNotSetup:
			status = http.StatusPreconditionFailed
		case auth.ErrTOTPRequired:
			status = http.StatusUnauthorized
		case auth.ErrInvalidTOTP:
			status = http.StatusUnauthorized
		case auth.ErrInvalidCredentials:
			status = http.StatusNotFound
		default:
			h.logger.Error("disable totp failed", zap.Error(err))
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "totp_disabled"})
}

// RequireAuth extracts the authenticated user ID from context.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.UUID{}, false
	}

	uid, ok := value.(uuid.UUID)
	return uid, ok
}
