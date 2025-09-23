package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// StripeProvider integrates Stripe Checkout and webhooks.
type StripeProvider struct {
	apiKey        string
	webhookSecret string
}

func NewStripeProvider(cfg config.StripeConfig) *StripeProvider {
	if cfg.APIKey == "" {
		return nil
	}
	return &StripeProvider{apiKey: cfg.APIKey, webhookSecret: cfg.WebhookSecret}
}

func (s *StripeProvider) Name() string {
	return "stripe"
}

func (s *StripeProvider) CreateCheckoutSession(ctx context.Context, plan entities.Plan, user entities.User, cfg ProviderCheckoutConfig) (CheckoutSession, error) {
	if s.apiKey == "" {
		return CheckoutSession{}, ErrProviderDisabled
	}

	stripe.Key = s.apiKey

	interval := stripe.PriceRecurringIntervalMonth
	intervalCount := int64(plan.IntervalCount)
	switch strings.ToLower(plan.BillingPeriod) {
	case "month":
		interval = stripe.PriceRecurringIntervalMonth
	case "quarter":
		interval = stripe.PriceRecurringIntervalMonth
		if intervalCount == 0 {
			intervalCount = 3
		}
	case "year":
		interval = stripe.PriceRecurringIntervalYear
	default:
		interval = stripe.PriceRecurringIntervalMonth
	}
	if intervalCount == 0 {
		intervalCount = 1
	}

	currency := strings.ToLower(plan.Currency)
	if currency == "" {
		currency = "try"
	}

	priceData := &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String(currency),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name: stripe.String(plan.Name),
			Metadata: map[string]string{
				"plan_code": plan.Code,
			},
		},
		UnitAmount: stripe.Int64(plan.PriceCents),
		Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval:      stripe.String(string(interval)),
			IntervalCount: stripe.Int64(intervalCount),
		},
	}

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(user.ID.String()),
		SuccessURL:        stripe.String(cfg.SuccessURL),
		CancelURL:         stripe.String(cfg.CancelURL),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{PriceData: priceData, Quantity: stripe.Int64(1)},
		},
		Metadata: map[string]string{
			"plan_code": plan.Code,
			"user_id":   user.ID.String(),
		},
		CustomerCreation: stripe.String(string(stripe.CheckoutSessionCustomerCreationIfRequired)),
	}

	params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
		Metadata: map[string]string{
			"plan_code": plan.Code,
			"user_id":   user.ID.String(),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("create stripe checkout session: %w", err)
	}

	return CheckoutSession{
		Provider:  s.Name(),
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

func (s *StripeProvider) ParseWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	if s.webhookSecret == "" {
		return nil, ErrProviderDisabled
	}

	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("construct stripe event: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return nil, fmt.Errorf("unmarshal checkout session: %w", err)
		}

		planCode := sess.Metadata["plan_code"]
		userIDStr := sess.Metadata["user_id"]
		if planCode == "" || userIDStr == "" {
			return nil, ErrInvalidWebhook
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, ErrInvalidWebhook
		}

		subscriptionID := ""
		if sess.Subscription != nil {
			subscriptionID = sess.Subscription.ID
		}

		status := string(sess.PaymentStatus)
		return &WebhookEvent{
			Provider:           s.Name(),
			Type:               WebhookTypeCheckoutCompleted,
			UserID:             userID,
			PlanCode:           planCode,
			SubscriptionID:     subscriptionID,
			CustomerID:         sess.Customer.ID,
			Status:             status,
			CurrentPeriodStart: time.Unix(event.Created, 0),
			OccurredAt:         time.Unix(event.Created, 0),
		}, nil

	case "customer.subscription.updated", "customer.subscription.created":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return nil, fmt.Errorf("unmarshal subscription: %w", err)
		}

		planCode := sub.Metadata["plan_code"]
		userIDStr := sub.Metadata["user_id"]
		var userID uuid.UUID
		if userIDStr != "" {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				userID = uid
			}
		}

		status := mapStripeSubscriptionStatus(sub.Status)
		cancelAt := sub.CanceledAt
		var canceledAt *time.Time
		if cancelAt > 0 {
			t := time.Unix(cancelAt, 0)
			canceledAt = &t
		}

		return &WebhookEvent{
			Provider:           s.Name(),
			Type:               WebhookTypeSubscriptionUpdated,
			UserID:             userID,
			PlanCode:           planCode,
			SubscriptionID:     sub.ID,
			CustomerID:         sub.Customer.ID,
			Status:             status,
			CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
			CurrentPeriodEnd:   time.Unix(sub.CurrentPeriodEnd, 0),
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
			CanceledAt:         canceledAt,
			OccurredAt:         time.Unix(event.Created, 0),
		}, nil

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return nil, fmt.Errorf("unmarshal subscription: %w", err)
		}

		return &WebhookEvent{
			Provider:          s.Name(),
			Type:              WebhookTypeSubscriptionCanceled,
			SubscriptionID:    sub.ID,
			CustomerID:        sub.Customer.ID,
			Status:            mapStripeSubscriptionStatus(sub.Status),
			OccurredAt:        time.Unix(event.Created, 0),
			CancelAtPeriodEnd: true,
			CanceledAt:        timePtr(time.Unix(event.Created, 0)),
		}, nil

	case "invoice.payment_succeeded":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			return nil, fmt.Errorf("unmarshal invoice: %w", err)
		}

		paidAt := time.Unix(inv.Created, 0)
		if inv.StatusTransitions != nil && inv.StatusTransitions.PaidAt > 0 {
			paidAt = time.Unix(inv.StatusTransitions.PaidAt, 0)
		}

		planCode := inv.Metadata["plan_code"]
		if planCode == "" && len(inv.Lines.Data) > 0 && inv.Lines.Data[0].Price != nil {
			planCode = inv.Lines.Data[0].Price.Metadata["plan_code"]
		}

		return &WebhookEvent{
			Provider:         s.Name(),
			Type:             WebhookTypePaymentSucceeded,
			SubscriptionID:   inv.Subscription.ID,
			CustomerID:       inv.Customer.ID,
			PaymentID:        inv.ID,
			AmountCents:      inv.AmountPaid,
			Currency:         strings.ToUpper(string(inv.Currency)),
			OccurredAt:       time.Unix(event.Created, 0),
			Status:           "succeeded",
			CurrentPeriodEnd: time.Unix(inv.PeriodEnd, 0),
			PlanCode:         planCode,
			PaidAt:           &paidAt,
		}, nil
	}

	return nil, nil
}

func mapStripeSubscriptionStatus(status stripe.SubscriptionStatus) string {
	switch status {
	case stripe.SubscriptionStatusActive:
		return "active"
	case stripe.SubscriptionStatusTrialing:
		return "trialing"
	case stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid, stripe.SubscriptionStatusIncomplete:
		return "past_due"
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusIncompleteExpired:
		return "canceled"
	default:
		return "past_due"
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
