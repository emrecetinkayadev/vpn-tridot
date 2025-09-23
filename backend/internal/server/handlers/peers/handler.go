package peershandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/peers"
)

// Handler manages peer CRUD endpoints.
type Handler struct {
	service *peers.Service
	logger  *zap.Logger
}

func New(service *peers.Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	peersList, err := h.service.ListPeers(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list peers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list peers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"peers": peersList})
}

func (h *Handler) Usage(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	summary, err := h.service.UsageSummary(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("usage summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch usage"})
		return
	}

	response := gin.H{
		"total_bytes_tx":    summary.TotalBytesTX,
		"total_bytes_rx":    summary.TotalBytesRX,
		"total_bytes":       summary.TotalBytes(),
		"peer_count":        summary.PeerCount,
		"active_peer_count": summary.ActivePeerCount,
	}
	if summary.LastHandshakeAt != nil {
		response["last_handshake_at"] = summary.LastHandshakeAt
	} else {
		response["last_handshake_at"] = nil
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type request struct {
		NodeID       string   `json:"node_id" binding:"required"`
		RegionID     string   `json:"region_id" binding:"required"`
		DeviceName   string   `json:"device_name" binding:"required"`
		ClientPubKey string   `json:"client_public_key"`
		AllowedIPs   string   `json:"allowed_ips"`
		DNSServers   []string `json:"dns_servers"`
		Keepalive    *int     `json:"keepalive"`
		MTU          *int     `json:"mtu"`
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
	regionID, err := uuid.Parse(req.RegionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid region_id"})
		return
	}

	output, err := h.service.CreatePeer(c.Request.Context(), peers.CreatePeerInput{
		UserID:       userID,
		NodeID:       nodeID,
		RegionID:     regionID,
		DeviceName:   req.DeviceName,
		ClientPubKey: req.ClientPubKey,
		AllowedIPs:   req.AllowedIPs,
		DNSServers:   req.DNSServers,
		Keepalive:    req.Keepalive,
		MTU:          req.MTU,
	})
	if err != nil {
		switch err {
		case peers.ErrDeviceLimitReached:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			h.logger.Error("create peer", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"peer":               output.Peer,
		"client_private_key": output.ClientPrivateKey,
		"config":             output.Config,
		"config_token":       output.ConfigToken,
		"config_qr":          output.ConfigQR,
	})
}

func (h *Handler) Rename(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	peerID, err := uuid.Parse(c.Param("peerID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer id"})
		return
	}

	type request struct {
		DeviceName string `json:"device_name" binding:"required"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peer, err := h.service.RenamePeer(c.Request.Context(), userID, peerID, req.DeviceName)
	if err != nil {
		if errors.Is(err, peers.ErrPeerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			h.logger.Error("rename peer", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"peer": peer})
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	peerID, err := uuid.Parse(c.Param("peerID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer id"})
		return
	}

	if err := h.service.DeletePeer(c.Request.Context(), userID, peerID); err != nil {
		if errors.Is(err, peers.ErrPeerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			h.logger.Error("delete peer", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DownloadConfig(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	config, err := h.service.GetConfigByToken(c.Request.Context(), userID, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
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
