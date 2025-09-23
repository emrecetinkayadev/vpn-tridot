package regionshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/regions"
)

// Handler provides HTTP endpoints for regions.
type Handler struct {
	service *regions.Service
	logger  *zap.Logger
}

func New(service *regions.Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	regions, err := h.service.List(c.Request.Context())
	if err != nil {
		h.logger.Error("list regions failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list regions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions})
}
