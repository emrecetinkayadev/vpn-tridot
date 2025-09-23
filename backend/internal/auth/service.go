package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/hash"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/jwt"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/random"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/totp"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/storage/postgres"
)

const (
	statusActive  = "active"
	statusPending = "pending"

	sessionActive = "active"

	tokenTypeEmailVerification = "email_verification"
	tokenTypePasswordReset     = "password_reset"
)

// Repository abstracts persistence for the auth service.
type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpsertTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error
	EnableTOTP(ctx context.Context, userID uuid.UUID, enabledAt time.Time) error
	DisableTOTP(ctx context.Context, userID uuid.UUID) error

	CreateSession(ctx context.Context, session entities.Session, userAgent, ipAddress string) (entities.Session, error)
	GetSessionByTokenHash(ctx context.Context, hash string) (entities.Session, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID, newHash string, expiresAt time.Time) (entities.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error

	CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error)
	GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error)
	ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error
}

// Service encapsulates authentication flows.
type Service struct {
	repo       Repository
	hasher     hash.Hasher
	jwtManager *jwt.Manager
	cfg        config.AuthConfig
	totpCfg    totp.Config
}

// Metadata holds request context info for session management.
type Metadata struct {
	UserAgent string
	IP        string
}

// LoginResult wraps tokens returned to the client.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	UserID       uuid.UUID
}

func NewService(repo Repository, hasher hash.Hasher, jwtManager *jwt.Manager, cfg config.AuthConfig) *Service {
	digits := otp.DigitsSix
	if cfg.TOTP.Digits == 8 {
		digits = otp.DigitsEight
	}

	totpCfg := totp.Config{
		Issuer:       cfg.TOTP.Issuer,
		SecretLength: cfg.TOTP.SecretLength,
		Digits:       digits,
		Period:       cfg.TOTP.PeriodSeconds,
	}

	return &Service{repo: repo, hasher: hasher, jwtManager: jwtManager, cfg: cfg, totpCfg: totpCfg}
}

// SignUp creates a pending user and issues an email verification token.
func (s *Service) SignUp(ctx context.Context, email, password string, now time.Time) (entities.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if len(password) < 10 {
		return entities.User{}, "", fmt.Errorf("password must be at least 10 characters")
	}

	hashValue, err := s.hasher.Hash(password)
	if err != nil {
		return entities.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, hashValue)
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicate) {
			return entities.User{}, "", ErrEmailAlreadyUsed
		}
		return entities.User{}, "", err
	}

	token, _, err := s.generateUserToken(ctx, user.ID, tokenTypeEmailVerification, s.cfg.VerificationTokenTTL, now)
	if err != nil {
		return entities.User{}, "", err
	}

	return user, token, nil
}

// VerifyEmail marks the user as active using the provided verification token.
func (s *Service) VerifyEmail(ctx context.Context, token string, now time.Time) error {
	record, err := s.lookupToken(ctx, tokenTypeEmailVerification, token, now)
	if err != nil {
		return err
	}

	if err := s.repo.MarkEmailVerified(ctx, record.UserID, now); err != nil {
		return err
	}

	return s.repo.ConsumeUserToken(ctx, record.ID)
}

// Login authenticates the user and returns signed tokens.
func (s *Service) Login(ctx context.Context, email, password, otpCode string, meta Metadata, now time.Time) (LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if user.Status != statusActive {
		return LoginResult{}, ErrEmailNotVerified
	}

	if user.TOTPEnabledAt != nil {
		if otpCode == "" {
			return LoginResult{}, ErrTOTPRequired
		}
		if user.TOTPSecret == nil {
			return LoginResult{}, ErrInvalidTOTP
		}
		valid, err := totp.Validate(*user.TOTPSecret, otpCode, s.totpCfg, now)
		if err != nil {
			return LoginResult{}, err
		}
		if !valid {
			return LoginResult{}, ErrInvalidTOTP
		}
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), now)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	session := entities.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		Status:           sessionActive,
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
	}

	if _, err := s.repo.CreateSession(ctx, session, meta.UserAgent, meta.IP); err != nil {
		return LoginResult{}, fmt.Errorf("create session: %w", err)
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID, now); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.AccessTokenTTL,
		UserID:       user.ID,
	}, nil
}

// Refresh exchanges a valid refresh token for a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string, meta Metadata, now time.Time) (LoginResult, error) {
	tokenHash := hashToken(refreshToken)

	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, ErrTokenNotFound
		}
		return LoginResult{}, err
	}

	if session.Status != sessionActive || session.ExpiresAt.Before(now) {
		return LoginResult{}, ErrTokenNotFound
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, ErrTokenNotFound
		}
		return LoginResult{}, err
	}

	if user.Status != statusActive {
		return LoginResult{}, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), now)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, newRefreshHash, err := generateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	if _, err := s.repo.UpdateSession(ctx, session.ID, newRefreshHash, now.Add(s.cfg.RefreshTokenTTL)); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.cfg.AccessTokenTTL,
		UserID:       user.ID,
	}, nil
}

// RequestPasswordReset creates a password reset token for the user.
func (s *Service) RequestPasswordReset(ctx context.Context, email string, now time.Time) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	token, _, err := s.generateUserToken(ctx, user.ID, tokenTypePasswordReset, s.cfg.PasswordResetTokenTTL, now)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword updates the user's password using a valid reset token and revokes sessions.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string, now time.Time) error {
	if len(newPassword) < 10 {
		return fmt.Errorf("password must be at least 10 characters")
	}

	record, err := s.lookupToken(ctx, tokenTypePasswordReset, token, now)
	if err != nil {
		return err
	}

	hashValue, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, record.UserID, hashValue); err != nil {
		return err
	}

	if err := s.repo.ConsumeUserToken(ctx, record.ID); err != nil {
		return err
	}

	return s.repo.RevokeUserSessions(ctx, record.UserID)
}

// SetupTOTP generates a new TOTP secret for the user and stores it disabled.
func (s *Service) SetupTOTP(ctx context.Context, userID uuid.UUID, now time.Time) (string, string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if user.TOTPEnabledAt != nil {
		return "", "", ErrTOTPAlreadyEnabled
	}

	accountName := user.Email
	secret, uri, err := totp.GenerateSecret(s.totpCfg, accountName)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.UpsertTOTPSecret(ctx, userID, secret); err != nil {
		return "", "", err
	}

	return secret, uri, nil
}

// ConfirmTOTP validates the provided code and enables TOTP for the user.
func (s *Service) ConfirmTOTP(ctx context.Context, userID uuid.UUID, code string, now time.Time) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	if user.TOTPSecret == nil {
		return ErrTOTPNotSetup
	}

	valid, err := totp.Validate(*user.TOTPSecret, code, s.totpCfg, now)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidTOTP
	}

	return s.repo.EnableTOTP(ctx, userID, now)
}

// DisableTOTP removes the stored TOTP secret and disables enforcement.
func (s *Service) DisableTOTP(ctx context.Context, userID uuid.UUID, code string, now time.Time) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	if user.TOTPSecret == nil && user.TOTPEnabledAt == nil {
		return ErrTOTPNotSetup
	}

	if user.TOTPSecret != nil && user.TOTPEnabledAt != nil {
		if code == "" {
			return ErrTOTPRequired
		}
		valid, err := totp.Validate(*user.TOTPSecret, code, s.totpCfg, now)
		if err != nil {
			return err
		}
		if !valid {
			return ErrInvalidTOTP
		}
	}

	return s.repo.DisableTOTP(ctx, userID)
}

func (s *Service) lookupToken(ctx context.Context, tokenType, token string, now time.Time) (entities.UserToken, error) {
	tokenHash := hashToken(token)

	record, err := s.repo.GetUserToken(ctx, tokenType, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.UserToken{}, ErrTokenNotFound
		}
		return entities.UserToken{}, err
	}

	if record.ConsumedAt != nil || record.ExpiresAt.Before(now) {
		return entities.UserToken{}, ErrTokenNotFound
	}

	return record, nil
}

func (s *Service) generateUserToken(ctx context.Context, userID uuid.UUID, tokenType string, ttl time.Duration, now time.Time) (string, string, error) {
	rawToken, err := random.String(32)
	if err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}

	tokenHash := hashToken(rawToken)

	entity := entities.UserToken{
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: tokenType,
		ExpiresAt: now.Add(ttl),
	}

	if _, err := s.repo.CreateUserToken(ctx, entity); err != nil {
		return "", "", err
	}

	return rawToken, tokenHash, nil
}

func generateRefreshToken() (string, string, error) {
	token, err := random.String(48)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
