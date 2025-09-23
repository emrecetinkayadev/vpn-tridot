package billing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// CheckoutSession holds provider-specific session identifiers and URLs.
type CheckoutSession struct {
	Provider  string
	SessionID string
	URL       string
}

// WebhookEvent represents normalized payment events consumed by the billing service.
type WebhookEvent struct {
	Provider           string
	Type               string
	UserID             uuid.UUID
	PlanCode           string
	SubscriptionID     string
	CustomerID         string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	PaymentID          string
	AmountCents        int64
	Currency           string
	OccurredAt         time.Time
	PaidAt             *time.Time
}

const (
	WebhookTypeCheckoutCompleted    = "checkout_completed"
	WebhookTypeSubscriptionUpdated  = "subscription_updated"
	WebhookTypeSubscriptionCanceled = "subscription_canceled"
	WebhookTypePaymentSucceeded     = "payment_succeeded"
)

// PaymentProvider describes third-party gateway integrations.
type PaymentProvider interface {
	CreateCheckoutSession(ctx context.Context, plan entities.Plan, user entities.User, cfg ProviderCheckoutConfig) (CheckoutSession, error)
	ParseWebhook(payload []byte, signature string) (*WebhookEvent, error)
	Name() string
}

// ProviderCheckoutConfig carries provider-agnostic options.
type ProviderCheckoutConfig struct {
	SuccessURL string
	CancelURL  string
}
