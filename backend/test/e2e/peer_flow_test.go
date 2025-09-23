package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/auth"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/billing"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/peers"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/hash"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/jwt"
)

type e2eAuthRepo struct {
	users  map[string]entities.User
	tokens map[uuid.UUID]entities.UserToken
}

func newE2EAuthRepo() *e2eAuthRepo {
	return &e2eAuthRepo{users: make(map[string]entities.User), tokens: make(map[uuid.UUID]entities.UserToken)}
}

func (r *e2eAuthRepo) CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error) {
	user := entities.User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, Status: "pending"}
	r.users[email] = user
	return user, nil
}

func (r *e2eAuthRepo) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	user, ok := r.users[email]
	if !ok {
		return entities.User{}, errors.New("not found")
	}
	return user, nil
}

func (r *e2eAuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return entities.User{}, errors.New("not found")
}

func (r *e2eAuthRepo) MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	for email, user := range r.users {
		if user.ID == userID {
			user.Status = "active"
			r.users[email] = user
			return nil
		}
	}
	return nil
}

func (r *e2eAuthRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return nil
}
func (r *e2eAuthRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}
func (r *e2eAuthRepo) UpsertTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return nil
}
func (r *e2eAuthRepo) EnableTOTP(ctx context.Context, userID uuid.UUID, enabledAt time.Time) error {
	return nil
}
func (r *e2eAuthRepo) DisableTOTP(ctx context.Context, userID uuid.UUID) error { return nil }
func (r *e2eAuthRepo) CreateSession(ctx context.Context, session entities.Session, userAgent, ipAddress string) (entities.Session, error) {
	return session, nil
}
func (r *e2eAuthRepo) GetSessionByTokenHash(ctx context.Context, hash string) (entities.Session, error) {
	return entities.Session{}, errors.New("not implemented")
}
func (r *e2eAuthRepo) UpdateSession(ctx context.Context, sessionID uuid.UUID, newHash string, expiresAt time.Time) (entities.Session, error) {
	return entities.Session{}, nil
}
func (r *e2eAuthRepo) RevokeSession(ctx context.Context, sessionID uuid.UUID) error   { return nil }
func (r *e2eAuthRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error { return nil }

func (r *e2eAuthRepo) CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error) {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	r.tokens[token.ID] = token
	return token, nil
}

func (r *e2eAuthRepo) GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error) {
	return entities.UserToken{}, errors.New("not implemented")
}

func (r *e2eAuthRepo) ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error {
	delete(r.tokens, tokenID)
	return nil
}

// Billing stubs reused from unit tests

type e2eBillingRepo struct {
	plans         map[string]entities.Plan
	subscriptions map[string]entities.Subscription
}

func newE2EBillingRepo() *e2eBillingRepo {
	return &e2eBillingRepo{plans: make(map[string]entities.Plan), subscriptions: make(map[string]entities.Subscription)}
}

func (r *e2eBillingRepo) UpsertPlan(ctx context.Context, plan entities.Plan) (entities.Plan, error) {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	r.plans[plan.Code] = plan
	return plan, nil
}
func (r *e2eBillingRepo) ListActivePlans(ctx context.Context) ([]entities.Plan, error) {
	return nil, nil
}
func (r *e2eBillingRepo) GetPlanByCode(ctx context.Context, code string) (entities.Plan, error) {
	plan, ok := r.plans[code]
	if !ok {
		return entities.Plan{}, errors.New("plan not found")
	}
	return plan, nil
}
func (r *e2eBillingRepo) UpsertSubscription(ctx context.Context, sub entities.Subscription) (entities.Subscription, error) {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	r.subscriptions[sub.ProviderSubscriptionID] = sub
	return sub, nil
}
func (r *e2eBillingRepo) GetSubscriptionByProviderID(ctx context.Context, provider, providerSubscriptionID string) (entities.Subscription, error) {
	sub, ok := r.subscriptions[providerSubscriptionID]
	if !ok {
		return entities.Subscription{}, errors.New("sub not found")
	}
	return sub, nil
}
func (r *e2eBillingRepo) RecordPayment(ctx context.Context, payment entities.Payment) (entities.Payment, error) {
	return payment, nil
}
func (r *e2eBillingRepo) ListPaymentsByUser(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	return nil, nil
}

type e2eBillingProvider struct {
	events []*billing.WebhookEvent
}

func (p *e2eBillingProvider) Name() string { return "stripe" }
func (p *e2eBillingProvider) CreateCheckoutSession(ctx context.Context, plan entities.Plan, user entities.User, cfg billing.ProviderCheckoutConfig) (billing.CheckoutSession, error) {
	return billing.CheckoutSession{Provider: p.Name(), SessionID: "sess_1", URL: "https://checkout"}, nil
}
func (p *e2eBillingProvider) ParseWebhook(payload []byte, signature string) (*billing.WebhookEvent, error) {
	if len(p.events) == 0 {
		return nil, nil
	}
	event := p.events[0]
	p.events = p.events[1:]
	return event, nil
}

// Peers stubs reused

type e2ePeerRepo struct {
	peers map[uuid.UUID]entities.Peer
	count int
}

func newE2EPeerRepo() *e2ePeerRepo {
	return &e2ePeerRepo{peers: make(map[uuid.UUID]entities.Peer)}
}

func (r *e2ePeerRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Peer, error) {
	return nil, nil
}
func (r *e2ePeerRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.count, nil
}
func (r *e2ePeerRepo) Create(ctx context.Context, peer entities.Peer) (entities.Peer, error) {
	peer.ID = uuid.New()
	r.peers[peer.ID] = peer
	r.count++
	return peer, nil
}
func (r *e2ePeerRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entities.Peer, error) {
	peer, ok := r.peers[id]
	if !ok {
		return entities.Peer{}, errors.New("not found")
	}
	return peer, nil
}
func (r *e2ePeerRepo) Rename(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (entities.Peer, error) {
	return entities.Peer{}, nil
}
func (r *e2ePeerRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error { return nil }
func (r *e2ePeerRepo) UsageSummaryByUser(ctx context.Context, userID uuid.UUID) (entities.UsageSummary, error) {
	return entities.UsageSummary{PeerCount: r.count}, nil
}

type e2eNodeStore struct {
	node entities.Node
}

func (n *e2eNodeStore) GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error) {
	if n.node.ID == uuid.Nil {
		n.node.ID = id
	}
	return n.node, nil
}

type e2eTokenStore struct {
	tokens map[string]entities.UserToken
}

func newE2ETokenStore() *e2eTokenStore {
	return &e2eTokenStore{tokens: make(map[string]entities.UserToken)}
}

func (t *e2eTokenStore) CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error) {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	t.tokens[token.TokenHash] = token
	return token, nil
}

func (t *e2eTokenStore) GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error) {
	token, ok := t.tokens[tokenHash]
	if !ok {
		return entities.UserToken{}, errors.New("not found")
	}
	return token, nil
}

func (t *e2eTokenStore) ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error {
	for hash, token := range t.tokens {
		if token.ID == tokenID {
			delete(t.tokens, hash)
			break
		}
	}
	return nil
}

func TestEndToEndPeerProvisioningFlow(t *testing.T) {
	ctx := context.Background()

	authRepo := newE2EAuthRepo()
	hasher, err := hash.NewArgon2Hasher(64*1024, 3, 16, 32, 2)
	require.NoError(t, err)
	jwtManager, err := jwt.NewManager("test-secret", "vpn", time.Minute)
	require.NoError(t, err)
	authService := auth.NewService(authRepo, hasher, jwtManager, config.AuthConfig{VerificationTokenTTL: time.Hour, RefreshTokenTTL: time.Hour})

	now := time.Now().UTC()
	user, verificationToken, err := authService.SignUp(ctx, "flow@example.com", "supersecurepw", now)
	require.NoError(t, err)
	require.NotEmpty(t, verificationToken)

	require.NoError(t, authRepo.MarkEmailVerified(ctx, user.ID, now))
	authRepo.users[user.Email] = entities.User{ID: user.ID, Email: user.Email, PasswordHash: authRepo.users[user.Email].PasswordHash, Status: "active"}

	billingRepo := newE2EBillingRepo()
	plan, _ := billingRepo.UpsertPlan(ctx, entities.Plan{Code: "vpn-monthly", Name: "Monthly", PriceCents: 1000, Currency: "TRY", BillingPeriod: "month", IntervalCount: 1, DeviceLimit: 5, IsActive: true})
	provider := &e2eBillingProvider{}
	providers := map[string]billing.PaymentProvider{"stripe": provider}
	billingCfg := config.BillingConfig{DefaultCurrency: "TRY"}
	billingService := billing.NewService(billingRepo, authRepo, providers, billingCfg)

	session, err := billingService.CreateCheckoutSession(ctx, user.ID, "vpn-monthly", "stripe")
	require.NoError(t, err)
	require.Equal(t, "sess_1", session.SessionID)

	subID := "sub_456"
	billingRepo.subscriptions[subID] = entities.Subscription{ID: uuid.New(), ProviderSubscriptionID: subID, UserID: user.ID, Status: "trialing", PlanID: plan.ID, Provider: "stripe"}
	provider.events = []*billing.WebhookEvent{{
		Type:             billing.WebhookTypeCheckoutCompleted,
		UserID:           user.ID,
		PlanCode:         "vpn-monthly",
		SubscriptionID:   subID,
		Status:           "active",
		CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
	}}

	require.NoError(t, billingService.HandleWebhook(ctx, "stripe", []byte("payload"), "sig"))
	sub, err := billingRepo.GetSubscriptionByProviderID(ctx, "stripe", subID)
	require.NoError(t, err)
	require.Equal(t, "active", sub.Status)

	node := &e2eNodeStore{node: entities.Node{PublicKey: "serverpk", Endpoint: "vpn.example.com:51820"}}
	peerRepo := newE2EPeerRepo()
	tokens := newE2ETokenStore()
	peerService := peers.NewService(peerRepo, node, tokens)

	peerOut, err := peerService.CreatePeer(ctx, peers.CreatePeerInput{
		UserID:     user.ID,
		NodeID:     uuid.New(),
		RegionID:   uuid.New(),
		DeviceName: "Laptop",
	})
	require.NoError(t, err)
	require.NotEmpty(t, peerOut.ConfigToken)

	cfgText, err := peerService.GetConfigByToken(ctx, user.ID, peerOut.ConfigToken)
	require.NoError(t, err)
	require.Contains(t, cfgText, "[Interface]")
}
