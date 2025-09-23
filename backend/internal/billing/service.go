package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// Repository defines persistence methods required by the billing service.
type Repository interface {
	UpsertPlan(ctx context.Context, plan entities.Plan) (entities.Plan, error)
	ListActivePlans(ctx context.Context) ([]entities.Plan, error)
	GetPlanByCode(ctx context.Context, code string) (entities.Plan, error)
	UpsertSubscription(ctx context.Context, sub entities.Subscription) (entities.Subscription, error)
	GetSubscriptionByProviderID(ctx context.Context, provider, providerSubscriptionID string) (entities.Subscription, error)
	RecordPayment(ctx context.Context, payment entities.Payment) (entities.Payment, error)
	ListPaymentsByUser(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error)
}

// UserStore exposes read access to user entities.
type UserStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error)
}

// Service orchestrates plan management, checkout, and webhook handling.
type Service struct {
	repo      Repository
	users     UserStore
	providers map[string]PaymentProvider
	cfg       config.BillingConfig
}

func NewService(repo Repository, users UserStore, providers map[string]PaymentProvider, cfg config.BillingConfig) *Service {
	prov := make(map[string]PaymentProvider)
	for key, provider := range providers {
		if provider == nil {
			continue
		}
		prov[strings.ToLower(key)] = provider
	}
	return &Service{repo: repo, users: users, providers: prov, cfg: cfg}
}

// SeedDefaultPlans ensures core plans exist in the database.
func (s *Service) SeedDefaultPlans(ctx context.Context) error {
	defaults := []struct {
		Code        string
		Name        string
		Description string
		PriceCents  int64
		Period      string
		Interval    int
		Devices     int
	}{
		{
			Code:        "vpn-monthly",
			Name:        "Aylık",
			Description: "Aylık abonelik",
			PriceCents:  14900,
			Period:      "month",
			Interval:    1,
			Devices:     5,
		},
		{
			Code:        "vpn-quarterly",
			Name:        "3 Aylık",
			Description: "3 aylık abonelik",
			PriceCents:  39900,
			Period:      "quarter",
			Interval:    1,
			Devices:     5,
		},
		{
			Code:        "vpn-annual",
			Name:        "Yıllık",
			Description: "12 aylık abonelik",
			PriceCents:  129900,
			Period:      "year",
			Interval:    1,
			Devices:     5,
		},
	}

	for _, item := range defaults {
		plan := entities.Plan{
			Code:          item.Code,
			Name:          item.Name,
			Description:   item.Description,
			PriceCents:    item.PriceCents,
			Currency:      strings.ToUpper(s.cfg.DefaultCurrency),
			BillingPeriod: item.Period,
			IntervalCount: item.Interval,
			DeviceLimit:   item.Devices,
			IsActive:      true,
		}
		if _, err := s.repo.UpsertPlan(ctx, plan); err != nil {
			return fmt.Errorf("seed plan %s: %w", plan.Code, err)
		}
	}

	return nil
}

// ListPlans returns all active plans ordered by price.
func (s *Service) ListPlans(ctx context.Context) ([]entities.Plan, error) {
	return s.repo.ListActivePlans(ctx)
}

// CreateCheckoutSession creates a checkout session for the given user and plan.
func (s *Service) CreateCheckoutSession(ctx context.Context, userID uuid.UUID, planCode, provider string) (CheckoutSession, error) {
	if provider == "" {
		provider = "stripe"
	}

	providerKey := strings.ToLower(provider)
	paymentProvider, ok := s.providers[providerKey]
	if !ok {
		return CheckoutSession{}, ErrUnsupportedProvider
	}

	plan, err := s.repo.GetPlanByCode(ctx, planCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CheckoutSession{}, ErrPlanNotFound
		}
		return CheckoutSession{}, err
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return CheckoutSession{}, err
	}

	checkoutCfg := ProviderCheckoutConfig{
		SuccessURL: s.cfg.Stripe.SuccessURL,
		CancelURL:  s.cfg.Stripe.CancelURL,
	}

	session, err := paymentProvider.CreateCheckoutSession(ctx, plan, user, checkoutCfg)
	if err != nil {
		return CheckoutSession{}, err
	}

	session.Provider = providerKey
	return session, nil
}

// HandleWebhook processes provider webhook payloads.
func (s *Service) HandleWebhook(ctx context.Context, provider string, payload []byte, signature string) error {
	providerKey := strings.ToLower(provider)
	paymentProvider, ok := s.providers[providerKey]
	if !ok {
		return ErrUnsupportedProvider
	}

	event, err := paymentProvider.ParseWebhook(payload, signature)
	if err != nil {
		return err
	}
	if event == nil {
		return nil
	}

	switch event.Type {
	case WebhookTypeCheckoutCompleted:
		return s.handleCheckoutCompleted(ctx, providerKey, event)
	case WebhookTypeSubscriptionUpdated:
		return s.handleSubscriptionUpdated(ctx, providerKey, event)
	case WebhookTypeSubscriptionCanceled:
		return s.handleSubscriptionCanceled(ctx, providerKey, event)
	case WebhookTypePaymentSucceeded:
		return s.handlePaymentSucceeded(ctx, providerKey, event)
	default:
		return nil
	}
}

// ListPayments returns historic payments for a user.
func (s *Service) ListPayments(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	return s.repo.ListPaymentsByUser(ctx, userID)
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, provider string, event *WebhookEvent) error {
	if event.UserID == uuid.Nil {
		return ErrInvalidWebhook
	}
	if event.SubscriptionID == "" {
		return nil
	}

	plan, err := s.repo.GetPlanByCode(ctx, event.PlanCode)
	if err != nil {
		return err
	}

	periodStart := event.CurrentPeriodStart
	if periodStart.IsZero() {
		periodStart = event.OccurredAt
	}
	periodEnd := computePeriodEnd(plan, periodStart)

	status := canonicalStatus(event.Status)

	sub := entities.Subscription{
		UserID:                 event.UserID,
		PlanID:                 plan.ID,
		Status:                 status,
		CurrentPeriodStart:     periodStart,
		CurrentPeriodEnd:       periodEnd,
		CancelAtPeriodEnd:      false,
		Provider:               provider,
		ProviderCustomerID:     event.CustomerID,
		ProviderSubscriptionID: event.SubscriptionID,
	}

	_, err = s.repo.UpsertSubscription(ctx, sub)
	return err
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, provider string, event *WebhookEvent) error {
	existing, err := s.repo.GetSubscriptionByProviderID(ctx, provider, event.SubscriptionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	planID := uuid.Nil
	var plan entities.Plan
	var planAvailable bool
	if event.PlanCode != "" {
		p, err := s.repo.GetPlanByCode(ctx, event.PlanCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanNotFound
			}
			return err
		}
		plan = p
		planAvailable = true
		planID = p.ID
	}
	if planID == uuid.Nil && existing.ID != uuid.Nil {
		planID = existing.PlanID
	}
	if planID == uuid.Nil && existing.ID == uuid.Nil {
		return ErrPlanNotFound
	}

	userID := event.UserID
	if userID == uuid.Nil && existing.ID != uuid.Nil {
		userID = existing.UserID
	}
	if userID == uuid.Nil {
		return ErrInvalidWebhook
	}

	periodStart := event.CurrentPeriodStart
	if periodStart.IsZero() {
		if existing.ID != uuid.Nil {
			periodStart = existing.CurrentPeriodStart
		} else {
			periodStart = event.OccurredAt
		}
	}

	periodEnd := event.CurrentPeriodEnd
	if periodEnd.IsZero() {
		if existing.ID != uuid.Nil {
			periodEnd = existing.CurrentPeriodEnd
		} else if planAvailable {
			periodEnd = computePeriodEnd(plan, periodStart)
		}
	}

	status := canonicalStatus(event.Status)
	if event.Status == "" && existing.ID != uuid.Nil {
		status = existing.Status
	}

	sub := entities.Subscription{
		UserID:                 userID,
		PlanID:                 planID,
		Status:                 status,
		CurrentPeriodStart:     periodStart,
		CurrentPeriodEnd:       periodEnd,
		CancelAtPeriodEnd:      event.CancelAtPeriodEnd,
		CanceledAt:             event.CanceledAt,
		Provider:               provider,
		ProviderCustomerID:     event.CustomerID,
		ProviderSubscriptionID: event.SubscriptionID,
	}

	_, err = s.repo.UpsertSubscription(ctx, sub)
	return err
}

func (s *Service) handleSubscriptionCanceled(ctx context.Context, provider string, event *WebhookEvent) error {
	existing, err := s.repo.GetSubscriptionByProviderID(ctx, provider, event.SubscriptionID)
	if err != nil {
		return err
	}

	canceledAt := event.CanceledAt
	if canceledAt == nil {
		t := event.OccurredAt
		canceledAt = &t
	}

	sub := entities.Subscription{
		ID:                     existing.ID,
		UserID:                 existing.UserID,
		PlanID:                 existing.PlanID,
		Status:                 "canceled",
		CurrentPeriodStart:     existing.CurrentPeriodStart,
		CurrentPeriodEnd:       event.OccurredAt,
		CancelAtPeriodEnd:      true,
		CanceledAt:             canceledAt,
		Provider:               provider,
		ProviderCustomerID:     existing.ProviderCustomerID,
		ProviderSubscriptionID: existing.ProviderSubscriptionID,
	}

	_, err = s.repo.UpsertSubscription(ctx, sub)
	return err
}

func (s *Service) handlePaymentSucceeded(ctx context.Context, provider string, event *WebhookEvent) error {
	sub, err := s.repo.GetSubscriptionByProviderID(ctx, provider, event.SubscriptionID)
	if err != nil {
		return err
	}

	metadata := map[string]any{}
	if event.PlanCode != "" {
		metadata["plan_code"] = event.PlanCode
	}
	if event.CustomerID != "" {
		metadata["customer_id"] = event.CustomerID
	}

	payment := entities.Payment{
		SubscriptionID:    sub.ID,
		Provider:          provider,
		ProviderPaymentID: event.PaymentID,
		Status:            event.Status,
		AmountCents:       event.AmountCents,
		Currency:          event.Currency,
		Metadata:          metadata,
	}
	if event.PaidAt != nil {
		payment.PaidAt = event.PaidAt
	}

	if _, err := s.repo.RecordPayment(ctx, payment); err != nil {
		return err
	}

	if !event.CurrentPeriodEnd.IsZero() {
		periodStart := sub.CurrentPeriodStart
		if !event.CurrentPeriodStart.IsZero() {
			periodStart = event.CurrentPeriodStart
		}
		updates := entities.Subscription{
			UserID:                 sub.UserID,
			PlanID:                 sub.PlanID,
			Status:                 sub.Status,
			CurrentPeriodStart:     periodStart,
			CurrentPeriodEnd:       event.CurrentPeriodEnd,
			CancelAtPeriodEnd:      sub.CancelAtPeriodEnd,
			CanceledAt:             sub.CanceledAt,
			Provider:               provider,
			ProviderCustomerID:     sub.ProviderCustomerID,
			ProviderSubscriptionID: sub.ProviderSubscriptionID,
		}
		_, err = s.repo.UpsertSubscription(ctx, updates)
		return err
	}

	return nil
}

func computePeriodEnd(plan entities.Plan, start time.Time) time.Time {
	interval := plan.IntervalCount
	if interval <= 0 {
		interval = 1
	}

	switch strings.ToLower(plan.BillingPeriod) {
	case "month":
		return start.AddDate(0, interval, 0)
	case "quarter":
		return start.AddDate(0, interval*3, 0)
	case "year":
		return start.AddDate(interval, 0, 0)
	default:
		return start.AddDate(0, interval, 0)
	}
}

func canonicalStatus(status string) string {
	s := strings.ToLower(status)
	switch s {
	case "active":
		return "active"
	case "trialing":
		return "trialing"
	case "past_due", "unpaid", "incomplete":
		return "past_due"
	case "canceled", "incomplete_expired":
		return "canceled"
	default:
		return "past_due"
	}
}
