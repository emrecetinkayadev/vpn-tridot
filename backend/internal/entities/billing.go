package entities

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID            uuid.UUID
	Code          string
	Name          string
	Description   string
	PriceCents    int64
	Currency      string
	BillingPeriod string
	IntervalCount int
	DeviceLimit   int
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Subscription struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	PlanID                 uuid.UUID
	Status                 string
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
	Provider               string
	ProviderCustomerID     string
	ProviderSubscriptionID string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Payment struct {
	ID                uuid.UUID
	SubscriptionID    uuid.UUID
	Provider          string
	ProviderPaymentID string
	Status            string
	AmountCents       int64
	Currency          string
	PaidAt            *time.Time
	RefundedAt        *time.Time
	Metadata          map[string]any
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
