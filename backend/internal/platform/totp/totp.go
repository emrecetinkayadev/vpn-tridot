package totp

import (
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Config controls TOTP generation parameters.
type Config struct {
	Issuer       string
	SecretLength int
	Digits       otp.Digits
	Period       uint
}

// GenerateSecret creates a random TOTP secret and provisioning URI.
func GenerateSecret(cfg Config, accountName string) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      cfg.Issuer,
		AccountName: accountName,
		SecretSize:  uint(cfg.SecretLength),
	})
	if err != nil {
		return "", "", fmt.Errorf("generate totp secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// Validate verifies a TOTP code against the given secret.
func Validate(secret, code string, cfg Config, at time.Time) (bool, error) {
	code = strings.TrimSpace(code)
	valid, err := totp.ValidateCustom(code, secret, at, totp.ValidateOpts{
		Period:    cfg.Period,
		Skew:      1,
		Digits:    cfg.Digits,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false, fmt.Errorf("validate totp: %w", err)
	}

	return valid, nil
}
