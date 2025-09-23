package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/peers"
)

type peerRepoStub struct {
	peers     map[uuid.UUID]entities.Peer
	count     int
	createErr error
}

func newPeerRepoStub() *peerRepoStub {
	return &peerRepoStub{peers: make(map[uuid.UUID]entities.Peer)}
}

func (r *peerRepoStub) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Peer, error) {
	var out []entities.Peer
	for _, peer := range r.peers {
		if peer.UserID == userID {
			out = append(out, peer)
		}
	}
	return out, nil
}

func (r *peerRepoStub) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.count, nil
}

func (r *peerRepoStub) Create(ctx context.Context, peer entities.Peer) (entities.Peer, error) {
	if r.createErr != nil {
		return entities.Peer{}, r.createErr
	}
	peer.ID = uuid.New()
	peer.CreatedAt = time.Now()
	peer.UpdatedAt = time.Now()
	r.peers[peer.ID] = peer
	r.count++
	return peer, nil
}

func (r *peerRepoStub) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entities.Peer, error) {
	peer, ok := r.peers[id]
	if !ok {
		return entities.Peer{}, errors.New("not found")
	}
	return peer, nil
}

func (r *peerRepoStub) Rename(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (entities.Peer, error) {
	peer, ok := r.peers[id]
	if !ok {
		return entities.Peer{}, errors.New("not found")
	}
	peer.DeviceName = name
	r.peers[id] = peer
	return peer, nil
}

func (r *peerRepoStub) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	delete(r.peers, id)
	r.count--
	return nil
}

func (r *peerRepoStub) UsageSummaryByUser(ctx context.Context, userID uuid.UUID) (entities.UsageSummary, error) {
	return entities.UsageSummary{PeerCount: r.count}, nil
}

type nodeStoreStub struct {
	node entities.Node
}

func (n *nodeStoreStub) GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error) {
	if n.node.ID == uuid.Nil {
		n.node.ID = id
	}
	return n.node, nil
}

type tokenStoreStub struct {
	tokens map[uuid.UUID]entities.UserToken
}

func newTokenStoreStub() *tokenStoreStub {
	return &tokenStoreStub{tokens: make(map[uuid.UUID]entities.UserToken)}
}

func (s *tokenStoreStub) CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error) {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	s.tokens[token.ID] = token
	return token, nil
}

func (s *tokenStoreStub) GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error) {
	for _, token := range s.tokens {
		if token.TokenType == tokenType && token.TokenHash == tokenHash {
			return token, nil
		}
	}
	return entities.UserToken{}, errors.New("token not found")
}

func (s *tokenStoreStub) ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error {
	delete(s.tokens, tokenID)
	return nil
}

func TestPeersServiceCreateGeneratesServerKeys(t *testing.T) {
	repo := newPeerRepoStub()
	node := nodeStoreStub{node: entities.Node{PublicKey: wgtypes.Key{}.String(), Endpoint: "vpn.example.com:51820"}}
	tokens := newTokenStoreStub()
	service := peers.NewService(repo, &node, tokens)

	input := peers.CreatePeerInput{
		UserID:     uuid.New(),
		NodeID:     uuid.New(),
		RegionID:   uuid.New(),
		DeviceName: "Laptop",
	}

	out, err := service.CreatePeer(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, out.ClientPrivateKey)
	require.NotEmpty(t, out.Config)
	require.NotEmpty(t, out.ConfigToken)
	require.Contains(t, out.Config, "[Interface]")
}

func TestPeersServiceDeviceLimit(t *testing.T) {
	repo := newPeerRepoStub()
	repo.count = 5
	node := nodeStoreStub{node: entities.Node{PublicKey: wgtypes.Key{}.String(), Endpoint: "vpn.example.com:51820"}}
	tokens := newTokenStoreStub()
	service := peers.NewService(repo, &node, tokens)

	input := peers.CreatePeerInput{
		UserID:     uuid.New(),
		NodeID:     uuid.New(),
		RegionID:   uuid.New(),
		DeviceName: "Tablet",
	}

	_, err := service.CreatePeer(context.Background(), input)
	require.ErrorIs(t, err, peers.ErrDeviceLimitReached)
}

func TestPeersServiceConfigTokenFlow(t *testing.T) {
	repo := newPeerRepoStub()
	node := nodeStoreStub{node: entities.Node{PublicKey: "serverpk", Endpoint: "vpn.example.com:51820"}}
	tokens := newTokenStoreStub()
	service := peers.NewService(repo, &node, tokens)

	userID := uuid.New()
	out, err := service.CreatePeer(context.Background(), peers.CreatePeerInput{
		UserID:     userID,
		NodeID:     uuid.New(),
		RegionID:   uuid.New(),
		DeviceName: "Phone",
	})
	require.NoError(t, err)

	config, err := service.GetConfigByToken(context.Background(), userID, out.ConfigToken)
	require.NoError(t, err)
	require.Equal(t, out.Config, config)

	_, err = service.GetConfigByToken(context.Background(), userID, out.ConfigToken)
	require.Error(t, err)
}
