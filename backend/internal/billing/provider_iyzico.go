package billing

import (
	"context"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// IyzicoProvider is a placeholder for future Iyzico integration.
type IyzicoProvider struct{}

func NewIyzicoProvider(cfg config.IyzicoConfig) PaymentProvider {
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return nil
	}
	return &IyzicoProvider{}
}

func (p *IyzicoProvider) Name() string {
	return "iyzico"
}

func (p *IyzicoProvider) CreateCheckoutSession(ctx context.Context, plan entities.Plan, user entities.User, cfg ProviderCheckoutConfig) (CheckoutSession, error) {
	return CheckoutSession{}, ErrProviderDisabled
}

func (p *IyzicoProvider) ParseWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	return nil, ErrProviderDisabled
}
