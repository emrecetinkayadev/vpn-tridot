package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/billing"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

type billingRepoStub struct {
	plans         map[string]entities.Plan
	subscriptions map[string]entities.Subscription
	payments      []entities.Payment
}

func newBillingRepoStub() *billingRepoStub {
	return &billingRepoStub{
		plans:         make(map[string]entities.Plan),
		subscriptions: make(map[string]entities.Subscription),
	}
}

func (r *billingRepoStub) UpsertPlan(ctx context.Context, plan entities.Plan) (entities.Plan, error) {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	r.plans[plan.Code] = plan
	return plan, nil
}

func (r *billingRepoStub) ListActivePlans(ctx context.Context) ([]entities.Plan, error) {
	var out []entities.Plan
	for _, plan := range r.plans {
		if plan.IsActive {
			out = append(out, plan)
		}
	}
	return out, nil
}

func (r *billingRepoStub) GetPlanByCode(ctx context.Context, code string) (entities.Plan, error) {
	plan, ok := r.plans[code]
	if !ok {
		return entities.Plan{}, errors.New("plan not found")
	}
	return plan, nil
}

func (r *billingRepoStub) UpsertSubscription(ctx context.Context, sub entities.Subscription) (entities.Subscription, error) {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	r.subscriptions[sub.ProviderSubscriptionID] = sub
	return sub, nil
}

func (r *billingRepoStub) GetSubscriptionByProviderID(ctx context.Context, provider, providerSubscriptionID string) (entities.Subscription, error) {
	sub, ok := r.subscriptions[providerSubscriptionID]
	if !ok {
		return entities.Subscription{}, errors.New("sub not found")
	}
	return sub, nil
}

func (r *billingRepoStub) RecordPayment(ctx context.Context, payment entities.Payment) (entities.Payment, error) {
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	r.payments = append(r.payments, payment)
	return payment, nil
}

func (r *billingRepoStub) ListPaymentsByUser(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	return r.payments, nil
}

type billingUserStoreStub struct {
	users map[uuid.UUID]entities.User
}

func (u *billingUserStoreStub) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	user, ok := u.users[id]
	if !ok {
		return entities.User{}, errors.New("user not found")
	}
	return user, nil
}

type providerStub struct {
	checkoutCalls []struct {
		plan entities.Plan
		user entities.User
	}
	events []*billing.WebhookEvent
}

func (p *providerStub) Name() string { return "stripe" }

func (p *providerStub) CreateCheckoutSession(ctx context.Context, plan entities.Plan, user entities.User, cfg billing.ProviderCheckoutConfig) (billing.CheckoutSession, error) {
	p.checkoutCalls = append(p.checkoutCalls, struct {
		plan entities.Plan
		user entities.User
	}{plan: plan, user: user})
	return billing.CheckoutSession{Provider: p.Name(), SessionID: "sess_123", URL: "https://checkout.example"}, nil
}

func (p *providerStub) ParseWebhook(payload []byte, signature string) (*billing.WebhookEvent, error) {
	if len(p.events) == 0 {
		return nil, nil
	}
	event := p.events[0]
	p.events = p.events[1:]
	return event, nil
}

func setupBillingService() (*billing.Service, *billingRepoStub, *providerStub, uuid.UUID) {
	repo := newBillingRepoStub()
	userID := uuid.New()
	plan, _ := repo.UpsertPlan(context.Background(), entities.Plan{Code: "vpn-monthly", Name: "Monthly", PriceCents: 1000, Currency: "TRY", BillingPeriod: "month", IntervalCount: 1, DeviceLimit: 5, IsActive: true})
	users := &billingUserStoreStub{users: map[uuid.UUID]entities.User{userID: {ID: userID, Email: "user@example.com", Status: "active"}}}
	provider := &providerStub{}
	providers := map[string]billing.PaymentProvider{"stripe": provider}
	cfg := config.BillingConfig{DefaultCurrency: "TRY"}
	cfg.Stripe.SuccessURL = "https://app.example.com/success"
	cfg.Stripe.CancelURL = "https://app.example.com/cancel"

	service := billing.NewService(repo, users, providers, cfg)
	repo.plans[plan.Code] = plan
	return service, repo, provider, userID
}

func TestBillingCreateCheckoutSession(t *testing.T) {
	service, _, provider, userID := setupBillingService()

	session, err := service.CreateCheckoutSession(context.Background(), userID, "vpn-monthly", "stripe")
	require.NoError(t, err)
	require.Equal(t, "sess_123", session.SessionID)
	require.Len(t, provider.checkoutCalls, 1)
}

func TestBillingHandleWebhookCheckoutCompleted(t *testing.T) {
	service, repo, provider, userID := setupBillingService()

	subID := "sub_123"
	repo.subscriptions[subID] = entities.Subscription{
		ID:                     uuid.New(),
		UserID:                 userID,
		Provider:               "stripe",
		ProviderSubscriptionID: subID,
		Status:                 "trialing",
	}

	provider.events = []*billing.WebhookEvent{
		{
			Type:             billing.WebhookTypeCheckoutCompleted,
			UserID:           userID,
			PlanCode:         "vpn-monthly",
			SubscriptionID:   subID,
			Status:           "active",
			CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
		},
	}

	err := service.HandleWebhook(context.Background(), "stripe", []byte("payload"), "sig")
	require.NoError(t, err)
	updated, err := repo.GetSubscriptionByProviderID(context.Background(), "stripe", subID)
	require.NoError(t, err)
	require.Equal(t, "active", updated.Status)
}

func TestBillingHandleWebhookUnsupportedProvider(t *testing.T) {
	service, _, _, _ := setupBillingService()
	err := service.HandleWebhook(context.Background(), "unknown", nil, "")
	require.ErrorIs(t, err, billing.ErrUnsupportedProvider)
}
