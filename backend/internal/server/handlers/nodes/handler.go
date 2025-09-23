package nodeshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/regions"
)

// Handler manages node registration and health endpoints.
type Handler struct {
	service        *regions.Service
	logger         *zap.Logger
	provisionToken string
}

func New(service *regions.Service, cfg config.NodeConfig, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger, provisionToken: cfg.ProvisionToken}
}

func (h *Handler) Register(c *gin.Context) {
	if !h.validateToken(c) {
		return
	}

	type request struct {
		RegionCode string  `json:"region_code" binding:"required"`
		Hostname   string  `json:"hostname" binding:"required"`
		PublicIPv4 *string `json:"public_ipv4"`
		PublicIPv6 *string `json:"public_ipv6"`
		PublicKey  string  `json:"public_key"`
		Endpoint   string  `json:"endpoint"`
		TunnelPort int     `json:"tunnel_port"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.service.RegisterNode(c.Request.Context(), regions.RegisterNodeInput{
		RegionCode: req.RegionCode,
		Hostname:   req.Hostname,
		PublicIPv4: req.PublicIPv4,
		PublicIPv6: req.PublicIPv6,
		PublicKey:  req.PublicKey,
		Endpoint:   req.Endpoint,
		TunnelPort: req.TunnelPort,
	})
	if err != nil {
		h.logger.Error("register node failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"node_id": node.ID})
}

func (h *Handler) ReportHealth(c *gin.Context) {
	if !h.validateToken(c) {
		return
	}

	type request struct {
		NodeID         string  `json:"node_id" binding:"required"`
		ActivePeers    int     `json:"active_peers"`
		CPUPercent     float64 `json:"cpu_percent"`
		ThroughputMbps float64 `json:"throughput_mbps"`
		PacketLoss     float64 `json:"packet_loss"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
		return
	}

	node, err := h.service.ReportHealth(c.Request.Context(), regions.HealthReportInput{
		NodeID:         nodeID,
		ActivePeers:    req.ActivePeers,
		CPUPercent:     req.CPUPercent,
		ThroughputMbps: req.ThroughputMbps,
		PacketLoss:     req.PacketLoss,
	})
	if err != nil {
		h.logger.Error("node health update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"capacity_score": node.CapacityScore})
}

func (h *Handler) validateToken(c *gin.Context) bool {
	if h.provisionToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "node provisioning disabled"})
		return false
	}

	token := c.GetHeader("X-Provision-Token")
	if token == "" || token != h.provisionToken {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	return true
}
