package peers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"image/png"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/random"
)

const (
	defaultDeviceLimit    = 5
	tokenTypePeerConfig   = "peer_config"
	tokenByteLength       = 32
	tokenTTL              = 24 * time.Hour
	defaultAllowedIPs     = "10.0.0.2/32"
	defaultDNSServers     = "1.1.1.1"
	defaultPersistentKeep = 25
)

var (
	ErrDeviceLimitReached = errors.New("device limit reached")
	ErrPeerNotFound       = errors.New("peer not found")
)

// Repository abstracts storage operations.
type Repository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Peer, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	Create(ctx context.Context, peer entities.Peer) (entities.Peer, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entities.Peer, error)
	Rename(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (entities.Peer, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UsageSummaryByUser(ctx context.Context, userID uuid.UUID) (entities.UsageSummary, error)
}

// NodeStore exposes node metadata required for config generation.
type NodeStore interface {
	GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error)
}

// TokenStore reuses the auth repository for managing single-use tokens.
type TokenStore interface {
	CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error)
	GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error)
	ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error
}

// Service handles peer lifecycle.
type Service struct {
	repo       Repository
	nodeStore  NodeStore
	tokenStore TokenStore
}

func NewService(repo Repository, nodeStore NodeStore, tokenStore TokenStore) *Service {
	return &Service{repo: repo, nodeStore: nodeStore, tokenStore: tokenStore}
}

// CreatePeerInput defines payload for peer creation.
type CreatePeerInput struct {
	UserID       uuid.UUID
	NodeID       uuid.UUID
	RegionID     uuid.UUID
	DeviceName   string
	ClientPubKey string
	AllowedIPs   string
	DNSServers   []string
	Keepalive    *int
	MTU          *int
}

// CreatePeerOutput returns created peer and configuration artifacts.
type CreatePeerOutput struct {
	Peer             entities.Peer
	ClientPrivateKey string
	Config           string
	ConfigToken      string
	ConfigQR         string
}

func (s *Service) ListPeers(ctx context.Context, userID uuid.UUID) ([]entities.Peer, error) {
	return s.repo.ListByUser(ctx, userID)
}

// UsageSummary aggregates traffic and activity information for the user.
func (s *Service) UsageSummary(ctx context.Context, userID uuid.UUID) (entities.UsageSummary, error) {
	return s.repo.UsageSummaryByUser(ctx, userID)
}

func (s *Service) CreatePeer(ctx context.Context, input CreatePeerInput) (CreatePeerOutput, error) {
	if input.UserID == uuid.Nil {
		return CreatePeerOutput{}, errors.New("user id required")
	}
	if input.NodeID == uuid.Nil || input.RegionID == uuid.Nil {
		return CreatePeerOutput{}, errors.New("node and region required")
	}
	if strings.TrimSpace(input.DeviceName) == "" {
		return CreatePeerOutput{}, errors.New("device name required")
	}

	count, err := s.repo.CountByUser(ctx, input.UserID)
	if err != nil {
		return CreatePeerOutput{}, err
	}
	if count >= defaultDeviceLimit {
		return CreatePeerOutput{}, ErrDeviceLimitReached
	}

	var clientPrivate wgtypes.Key
	var clientPublic wgtypes.Key
	if input.ClientPubKey == "" {
		clientPrivate, err = wgtypes.GeneratePrivateKey()
		if err != nil {
			return CreatePeerOutput{}, fmt.Errorf("generate client key: %w", err)
		}
		clientPublic = clientPrivate.PublicKey()
	} else {
		clientPublic, err = wgtypes.ParseKey(input.ClientPubKey)
		if err != nil {
			return CreatePeerOutput{}, fmt.Errorf("parse client public key: %w", err)
		}
	}

	preshared, err := wgtypes.GenerateKey()
	if err != nil {
		return CreatePeerOutput{}, fmt.Errorf("generate preshared key: %w", err)
	}

	allowed := input.AllowedIPs
	if allowed == "" {
		allowed = defaultAllowedIPs
	}
	var dns []string
	if len(input.DNSServers) == 0 {
		dns = []string{defaultDNSServers}
	} else {
		dns = input.DNSServers
	}

	peer := entities.Peer{
		UserID:       input.UserID,
		NodeID:       input.NodeID,
		RegionID:     input.RegionID,
		DeviceName:   input.DeviceName,
		PublicKey:    clientPublic.String(),
		PresharedKey: ptrString(preshared.String()),
		AllowedIPs:   allowed,
		DNSServers:   dns,
		Keepalive:    input.Keepalive,
		MTU:          input.MTU,
		Status:       "active",
	}

	peer, err = s.repo.Create(ctx, peer)
	if err != nil {
		return CreatePeerOutput{}, err
	}

	node, err := s.nodeStore.GetNodeByID(ctx, input.NodeID)
	if err != nil {
		return CreatePeerOutput{}, err
	}

	config := buildConfig(peer, node, clientPrivate.String())
	qrCode, err := generateQRCode(config)
	if err != nil {
		return CreatePeerOutput{}, err
	}
	token, err := s.issueConfigToken(ctx, peer.UserID, peer.ID, config)
	if err != nil {
		return CreatePeerOutput{}, err
	}

	return CreatePeerOutput{
		Peer:             peer,
		ClientPrivateKey: clientPrivate.String(),
		Config:           config,
		ConfigToken:      token,
		ConfigQR:         qrCode,
	}, nil
}

func (s *Service) RenamePeer(ctx context.Context, userID, peerID uuid.UUID, name string) (entities.Peer, error) {
	if strings.TrimSpace(name) == "" {
		return entities.Peer{}, errors.New("device name required")
	}
	peer, err := s.repo.Rename(ctx, peerID, userID, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Peer{}, ErrPeerNotFound
		}
		return entities.Peer{}, err
	}
	return peer, nil
}

func (s *Service) DeletePeer(ctx context.Context, userID, peerID uuid.UUID) error {
	if err := s.repo.Delete(ctx, peerID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPeerNotFound
		}
		return err
	}
	return nil
}

// GetConfigByToken returns config text for a single-use token and invalidates it.
func (s *Service) GetConfigByToken(ctx context.Context, userID uuid.UUID, token string) (string, error) {
	hash := hashToken(token)
	tokenRecord, err := s.tokenStore.GetUserToken(ctx, tokenTypePeerConfig, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("config token invalid")
		}
		return "", err
	}

	if tokenRecord.UserID != userID {
		return "", errors.New("config token does not belong to user")
	}

	if tokenRecord.ExpiresAt.Before(time.Now()) {
		return "", errors.New("config token expired")
	}

	if tokenRecord.Metadata == nil {
		return "", errors.New("config metadata missing")
	}

	var meta configTokenMetadata
	if err := json.Unmarshal([]byte(*tokenRecord.Metadata), &meta); err != nil {
		return "", errors.New("invalid config metadata")
	}

	if meta.PeerID == uuid.Nil {
		return "", errors.New("config metadata missing peer id")
	}

	if _, err := s.repo.GetByID(ctx, meta.PeerID, userID); err != nil {
		return "", err
	}

	if err := s.tokenStore.ConsumeUserToken(ctx, tokenRecord.ID); err != nil {
		return "", err
	}

	return meta.Config, nil
}

func (s *Service) issueConfigToken(ctx context.Context, userID, peerID uuid.UUID, config string) (string, error) {
	raw, err := random.String(tokenByteLength)
	if err != nil {
		return "", err
	}

	hash := hashToken(raw)
	metaBytes, err := json.Marshal(configTokenMetadata{PeerID: peerID, Config: config})
	if err != nil {
		return "", err
	}
	metaStr := string(metaBytes)

	token := entities.UserToken{
		UserID:    userID,
		TokenHash: hash,
		TokenType: tokenTypePeerConfig,
		ExpiresAt: time.Now().Add(tokenTTL),
		Metadata:  &metaStr,
	}
	if _, err := s.tokenStore.CreateUserToken(ctx, token); err != nil {
		return "", err
	}

	return raw, nil
}

func buildConfig(peer entities.Peer, node entities.Node, clientPrivate string) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	if clientPrivate != "" {
		sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", clientPrivate))
	}
	if peer.AllowedIPs != "" {
		sb.WriteString(fmt.Sprintf("Address = %s\n", peer.AllowedIPs))
	}
	if len(peer.DNSServers) > 0 {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(peer.DNSServers, ",")))
	}
	if peer.MTU != nil {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", *peer.MTU))
	}
	if peer.Keepalive != nil {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", *peer.Keepalive))
	} else {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", defaultPersistentKeep))
	}

	sb.WriteString("\n[Peer]\n")
	sb.WriteString(fmt.Sprintf("PublicKey = %s\n", node.PublicKey))
	if peer.PresharedKey != nil {
		sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", *peer.PresharedKey))
	}
	sb.WriteString(fmt.Sprintf("Endpoint = %s\n", node.Endpoint))
	sb.WriteString("AllowedIPs = 0.0.0.0/0, ::/0\n")

	return sb.String()
}

func generateQRCode(payload string) (string, error) {
	code, err := qr.Encode(payload, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("generate qr: %w", err)
	}
	scaled, err := barcode.Scale(code, 256, 256)
	if err != nil {
		return "", fmt.Errorf("scale qr: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("encode qr: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func ptrString(v string) *string {
	return &v
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type configTokenMetadata struct {
	PeerID uuid.UUID `json:"peer_id"`
	Config string    `json:"config"`
}
