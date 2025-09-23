package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for access tokens.
type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWT access tokens.
type Manager struct {
	secret         []byte
	issuer         string
	accessTokenTTL time.Duration
}

// NewManager constructs a Manager.
func NewManager(secret, issuer string, accessTokenTTL time.Duration) (*Manager, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	if accessTokenTTL <= 0 {
		return nil, errors.New("access token ttl must be positive")
	}

	return &Manager{
		secret:         []byte(secret),
		issuer:         issuer,
		accessTokenTTL: accessTokenTTL,
	}, nil
}

// GenerateAccessToken builds a signed JWT with the provided subject and issued time.
func (m *Manager) GenerateAccessToken(userID string, issuedAt time.Time) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(m.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return signed, nil
}

// Verify parses and validates the JWT, returning claims when valid.
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
