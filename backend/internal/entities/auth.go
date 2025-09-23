package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	Status          string
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	TOTPSecret      *string
	TOTPEnabledAt   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	Status           string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UserToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	TokenType  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
	Metadata   *string
}
