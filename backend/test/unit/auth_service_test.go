package unit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/auth"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/jwt"
)

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakeHasher) Compare(encodedHash, password string) error {
	if encodedHash != "hashed:"+password {
		return errors.New("password mismatch")
	}
	return nil
}

type authRepoStub struct {
	users        map[string]entities.User
	tokens       []entities.UserToken
	sessions     []entities.Session
	nextUserID   uuid.UUID
	failCreate   error
	lookupErr    error
	tokenTypeErr error
}

var _ auth.Repository = (*authRepoStub)(nil)

func newAuthRepoStub() *authRepoStub {
	return &authRepoStub{
		users:      make(map[string]entities.User),
		nextUserID: uuid.New(),
	}
}

func (a *authRepoStub) CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error) {
	if a.failCreate != nil {
		return entities.User{}, a.failCreate
	}
	user := entities.User{
		ID:           a.nextUserID,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       "pending",
	}
	a.users[email] = user
	return user, nil
}

func (a *authRepoStub) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	if a.lookupErr != nil {
		return entities.User{}, a.lookupErr
	}
	user, ok := a.users[email]
	if !ok {
		return entities.User{}, errors.New("not found")
	}
	return user, nil
}

func (a *authRepoStub) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	for _, user := range a.users {
		if user.ID == id {
			return user, nil
		}
	}
	return entities.User{}, errors.New("not found")
}

func (a *authRepoStub) MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	for email, user := range a.users {
		if user.ID == userID {
			user.Status = "active"
			a.users[email] = user
			return nil
		}
	}
	return errors.New("not found")
}

func (a *authRepoStub) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return nil
}

func (a *authRepoStub) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

func (a *authRepoStub) UpsertTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return nil
}

func (a *authRepoStub) EnableTOTP(ctx context.Context, userID uuid.UUID, enabledAt time.Time) error {
	return nil
}

func (a *authRepoStub) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (a *authRepoStub) CreateSession(ctx context.Context, session entities.Session, userAgent, ipAddress string) (entities.Session, error) {
	a.sessions = append(a.sessions, session)
	return session, nil
}

func (a *authRepoStub) GetSessionByTokenHash(ctx context.Context, hash string) (entities.Session, error) {
	return entities.Session{}, errors.New("not implemented")
}

func (a *authRepoStub) UpdateSession(ctx context.Context, sessionID uuid.UUID, newHash string, expiresAt time.Time) (entities.Session, error) {
	return entities.Session{}, errors.New("not implemented")
}

func (a *authRepoStub) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (a *authRepoStub) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (a *authRepoStub) CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error) {
	if a.tokenTypeErr != nil {
		return entities.UserToken{}, a.tokenTypeErr
	}
	a.tokens = append(a.tokens, token)
	return token, nil
}

func (a *authRepoStub) GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error) {
	return entities.UserToken{}, errors.New("not implemented")
}

func (a *authRepoStub) ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error {
	return nil
}

func TestAuthServiceSignUpSuccess(t *testing.T) {
	repo := newAuthRepoStub()
	hasher := fakeHasher{}
	jwtManager, err := jwt.NewManager("test-secret", "vpn", time.Minute)
	require.NoError(t, err)

	cfg := config.AuthConfig{VerificationTokenTTL: time.Hour}
	service := auth.NewService(repo, hasher, jwtManager, cfg)

	now := time.Now().UTC()
	user, token, err := service.SignUp(context.Background(), "user@example.com", "supersecurepwd", now)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID)
	require.NotEmpty(t, token)
	require.Len(t, repo.tokens, 1)
	require.Equal(t, "email_verification", repo.tokens[0].TokenType)
}

func TestAuthServiceSignUpWeakPassword(t *testing.T) {
	repo := newAuthRepoStub()
	hasher := fakeHasher{}
	jwtManager, err := jwt.NewManager("test-secret", "vpn", time.Minute)
	require.NoError(t, err)

	service := auth.NewService(repo, hasher, jwtManager, config.AuthConfig{})

	_, _, err = service.SignUp(context.Background(), "weak@example.com", "short", time.Now())
	require.Error(t, err)
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	repo := newAuthRepoStub()
	hasher := fakeHasher{}
	jwtManager, err := jwt.NewManager("test-secret", "vpn", time.Minute)
	require.NoError(t, err)

	cfg := config.AuthConfig{RefreshTokenTTL: time.Hour}
	service := auth.NewService(repo, hasher, jwtManager, cfg)

	email := "login@example.com"
	password := "supersecretpass"
	hashVal, err := hasher.Hash(password)
	require.NoError(t, err)
	repo.users[strings.ToLower(email)] = entities.User{ID: uuid.New(), Email: email, PasswordHash: hashVal, Status: "active"}

	result, err := service.Login(context.Background(), email, password, "", auth.Metadata{UserAgent: "test", IP: "127.0.0.1"}, time.Now())
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.Equal(t, repo.users[strings.ToLower(email)].ID, result.UserID)
}

func TestAuthServiceLoginInvalidPassword(t *testing.T) {
	repo := newAuthRepoStub()
	hasher := fakeHasher{}
	jwtManager, err := jwt.NewManager("test-secret", "vpn", time.Minute)
	require.NoError(t, err)

	cfg := config.AuthConfig{RefreshTokenTTL: time.Hour}
	service := auth.NewService(repo, hasher, jwtManager, cfg)

	email := "login2@example.com"
	hashVal, err := hasher.Hash("supersecretpass")
	require.NoError(t, err)
	repo.users[strings.ToLower(email)] = entities.User{ID: uuid.New(), Email: email, PasswordHash: hashVal, Status: "active"}

	_, err = service.Login(context.Background(), email, "wrongpass", "", auth.Metadata{}, time.Now())
	require.Error(t, err)
}
