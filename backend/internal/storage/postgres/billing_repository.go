package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// BillingRepository handles plan, subscription and payment persistence.
type BillingRepository struct {
	pool *pgxpool.Pool
}

func NewBillingRepository(pool *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{pool: pool}
}

func (r *BillingRepository) UpsertPlan(ctx context.Context, plan entities.Plan) (entities.Plan, error) {
	const query = `
	INSERT INTO plans (code, name, description, price_cents, currency, billing_period, interval_count, device_limit, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (code)
	DO UPDATE SET
		name = EXCLUDED.name,
		description = EXCLUDED.description,
		price_cents = EXCLUDED.price_cents,
		currency = EXCLUDED.currency,
		billing_period = EXCLUDED.billing_period,
		interval_count = EXCLUDED.interval_count,
		device_limit = EXCLUDED.device_limit,
		is_active = EXCLUDED.is_active,
		updated_at = NOW()
	RETURNING id, code, name, description, price_cents, currency, billing_period, interval_count, device_limit, is_active, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		plan.Code,
		plan.Name,
		plan.Description,
		plan.PriceCents,
		plan.Currency,
		plan.BillingPeriod,
		plan.IntervalCount,
		plan.DeviceLimit,
		plan.IsActive,
	)

	return scanPlan(row)
}

func (r *BillingRepository) ListActivePlans(ctx context.Context) ([]entities.Plan, error) {
	const query = `
	SELECT id, code, name, description, price_cents, currency, billing_period, interval_count, device_limit, is_active, created_at, updated_at
	FROM plans
	WHERE is_active = true
	ORDER BY price_cents ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	var plans []entities.Plan
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *BillingRepository) GetPlanByCode(ctx context.Context, code string) (entities.Plan, error) {
	const query = `
	SELECT id, code, name, description, price_cents, currency, billing_period, interval_count, device_limit, is_active, created_at, updated_at
	FROM plans
	WHERE code = $1`

	row := r.pool.QueryRow(ctx, query, code)
	return scanPlan(row)
}

func (r *BillingRepository) UpsertSubscription(ctx context.Context, sub entities.Subscription) (entities.Subscription, error) {
	const query = `
	INSERT INTO subscriptions (
		user_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at,
		provider, provider_customer_id, provider_subscription_id, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())
	ON CONFLICT (provider, provider_subscription_id)
	DO UPDATE SET
		status = EXCLUDED.status,
		current_period_start = EXCLUDED.current_period_start,
		current_period_end = EXCLUDED.current_period_end,
		cancel_at_period_end = EXCLUDED.cancel_at_period_end,
		canceled_at = EXCLUDED.canceled_at,
		provider_customer_id = EXCLUDED.provider_customer_id,
		updated_at = NOW()
	RETURNING id, user_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, provider, provider_customer_id, provider_subscription_id, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		sub.UserID,
		sub.PlanID,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.CanceledAt,
		sub.Provider,
		sub.ProviderCustomerID,
		sub.ProviderSubscriptionID,
	)

	return scanSubscription(row)
}

func (r *BillingRepository) GetSubscriptionByProviderID(ctx context.Context, provider, providerSubscriptionID string) (entities.Subscription, error) {
	const query = `
	SELECT id, user_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, provider, provider_customer_id, provider_subscription_id, created_at, updated_at
	FROM subscriptions
	WHERE provider = $1 AND provider_subscription_id = $2`

	row := r.pool.QueryRow(ctx, query, provider, providerSubscriptionID)
	return scanSubscription(row)
}

func (r *BillingRepository) RecordPayment(ctx context.Context, payment entities.Payment) (entities.Payment, error) {
	const query = `
	INSERT INTO payments (subscription_id, provider, provider_payment_id, status, amount_cents, currency, paid_at, refunded_at, metadata, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
	ON CONFLICT (provider, provider_payment_id)
	DO UPDATE SET
		status = EXCLUDED.status,
		amount_cents = EXCLUDED.amount_cents,
		currency = EXCLUDED.currency,
		paid_at = EXCLUDED.paid_at,
		refunded_at = EXCLUDED.refunded_at,
		metadata = EXCLUDED.metadata,
		updated_at = NOW()
	RETURNING id, subscription_id, provider, provider_payment_id, status, amount_cents, currency, paid_at, refunded_at, metadata, created_at, updated_at`

	var metadataBytes any
	if payment.Metadata != nil {
		buf, err := json.Marshal(payment.Metadata)
		if err != nil {
			return entities.Payment{}, fmt.Errorf("marshal payment metadata: %w", err)
		}
		metadataBytes = buf
	}

	row := r.pool.QueryRow(ctx, query,
		payment.SubscriptionID,
		payment.Provider,
		payment.ProviderPaymentID,
		payment.Status,
		payment.AmountCents,
		payment.Currency,
		payment.PaidAt,
		payment.RefundedAt,
		metadataBytes,
	)

	return scanPayment(row)
}

func (r *BillingRepository) ListPaymentsByUser(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	const query = `
	SELECT p.id, p.subscription_id, p.provider, p.provider_payment_id, p.status, p.amount_cents, p.currency, p.paid_at, p.refunded_at, p.metadata, p.created_at, p.updated_at
	FROM payments p
	JOIN subscriptions s ON p.subscription_id = s.id
	WHERE s.user_id = $1
	ORDER BY COALESCE(p.paid_at, p.created_at) DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []entities.Payment
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, rows.Err()
}

func scanPlan(row pgx.Row) (entities.Plan, error) {
	var p entities.Plan
	if err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.PriceCents, &p.Currency, &p.BillingPeriod, &p.IntervalCount, &p.DeviceLimit, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return entities.Plan{}, translateError(err)
	}
	return p, nil
}

func scanSubscription(row pgx.Row) (entities.Subscription, error) {
	var s entities.Subscription
	if err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.PlanID,
		&s.Status,
		&s.CurrentPeriodStart,
		&s.CurrentPeriodEnd,
		&s.CancelAtPeriodEnd,
		&s.CanceledAt,
		&s.Provider,
		&s.ProviderCustomerID,
		&s.ProviderSubscriptionID,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		return entities.Subscription{}, translateError(err)
	}
	return s, nil
}

func scanPayment(row pgx.Row) (entities.Payment, error) {
	var (
		p        entities.Payment
		paidAt   sql.NullTime
		refunded sql.NullTime
		metadata []byte
	)

	if err := row.Scan(
		&p.ID,
		&p.SubscriptionID,
		&p.Provider,
		&p.ProviderPaymentID,
		&p.Status,
		&p.AmountCents,
		&p.Currency,
		&paidAt,
		&refunded,
		&metadata,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return entities.Payment{}, translateError(err)
	}

	if paidAt.Valid {
		p.PaidAt = &paidAt.Time
	}
	if refunded.Valid {
		p.RefundedAt = &refunded.Time
	}
	if len(metadata) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(metadata, &meta); err == nil {
			p.Metadata = meta
		}
	}

	return p, nil
}
