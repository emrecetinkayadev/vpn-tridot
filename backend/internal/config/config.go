package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application-level configuration loaded from environment variables.
type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Log       LogConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
	Billing   BillingConfig
	Node      NodeConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Version string
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type LogConfig struct {
	Level string
}

type AuthConfig struct {
	JWTSigningKey         string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	VerificationTokenTTL  time.Duration
	PasswordResetTokenTTL time.Duration
	ArgonMemory           uint32
	ArgonIterations       uint32
	ArgonParallelism      uint8
	ArgonSaltLength       uint32
	ArgonKeyLength        uint32
	TOTP                  struct {
		Issuer        string
		SecretLength  int
		Digits        int
		PeriodSeconds uint
	}
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

type BillingConfig struct {
	DefaultCurrency string
	Stripe          StripeConfig
	Iyzico          IyzicoConfig
}

type NodeConfig struct {
	ProvisionToken string
}

type StripeConfig struct {
	APIKey        string
	WebhookSecret string
	SuccessURL    string
	CancelURL     string
}

type IyzicoConfig struct {
	APIKey    string
	SecretKey string
	BaseURL   string
}

// Load reads environment variables and applies sane defaults.
func Load() (Config, error) {
	cfg := Config{}

	cfg.App.Name = getEnv("APP_NAME", "vpn-backend")
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Version = getEnv("APP_VERSION", "dev")

	httpAddr := getEnv("HTTP_ADDR", ":8080")
	readTimeout, err := durationFromEnv("HTTP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_READ_TIMEOUT: %w", err)
	}
	writeTimeout, err := durationFromEnv("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_WRITE_TIMEOUT: %w", err)
	}
	shutdownTimeout, err := durationFromEnv("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_SHUTDOWN_TIMEOUT: %w", err)
	}

	reqPerSec, err := floatFromEnv("RATE_LIMIT_RPS", 20.0)
	if err != nil {
		return Config{}, fmt.Errorf("parse RATE_LIMIT_RPS: %w", err)
	}
	burst, err := intFromEnv("RATE_LIMIT_BURST", 40)
	if err != nil {
		return Config{}, fmt.Errorf("parse RATE_LIMIT_BURST: %w", err)
	}

	cfg.HTTP.Addr = httpAddr
	cfg.HTTP.ReadTimeout = readTimeout
	cfg.HTTP.WriteTimeout = writeTimeout
	cfg.HTTP.ShutdownTimeout = shutdownTimeout

	cfg.Log.Level = getEnv("LOG_LEVEL", "info")

	cfg.Database.DSN = getEnv("POSTGRES_DSN", "")
	cfg.Database.MaxOpenConns, err = intFromEnv("DB_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, fmt.Errorf("parse DB_MAX_OPEN_CONNS: %w", err)
	}
	cfg.Database.MaxIdleConns, err = intFromEnv("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, fmt.Errorf("parse DB_MAX_IDLE_CONNS: %w", err)
	}
	cfg.Database.ConnMaxLifetime, err = durationFromEnv("DB_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
	}

	cfg.Auth.JWTSigningKey = getEnv("JWT_SECRET", "")
	cfg.Auth.AccessTokenTTL, err = durationFromEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_ACCESS_TTL: %w", err)
	}
	cfg.Auth.RefreshTokenTTL, err = durationFromEnv("JWT_REFRESH_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_REFRESH_TTL: %w", err)
	}
	cfg.Auth.VerificationTokenTTL, err = durationFromEnv("AUTH_VERIFICATION_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_VERIFICATION_TTL: %w", err)
	}
	cfg.Auth.PasswordResetTokenTTL, err = durationFromEnv("AUTH_PASSWORD_RESET_TTL", 1*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_PASSWORD_RESET_TTL: %w", err)
	}

	cfg.Auth.ArgonMemory, err = uint32FromEnv("ARGON_MEMORY", 64*1024)
	if err != nil {
		return Config{}, fmt.Errorf("parse ARGON_MEMORY: %w", err)
	}
	cfg.Auth.ArgonIterations, err = uint32FromEnv("ARGON_ITERATIONS", 3)
	if err != nil {
		return Config{}, fmt.Errorf("parse ARGON_ITERATIONS: %w", err)
	}
	argonParallelism, err := uint32FromEnv("ARGON_PARALLELISM", 2)
	if err != nil {
		return Config{}, fmt.Errorf("parse ARGON_PARALLELISM: %w", err)
	}
	cfg.Auth.ArgonParallelism = uint8(argonParallelism)
	cfg.Auth.ArgonSaltLength, err = uint32FromEnv("ARGON_SALT_LENGTH", 16)
	if err != nil {
		return Config{}, fmt.Errorf("parse ARGON_SALT_LENGTH: %w", err)
	}
	cfg.Auth.ArgonKeyLength, err = uint32FromEnv("ARGON_KEY_LENGTH", 32)
	if err != nil {
		return Config{}, fmt.Errorf("parse ARGON_KEY_LENGTH: %w", err)
	}

	cfg.Auth.TOTP.Issuer = getEnv("TOTP_ISSUER", cfg.App.Name)
	cfg.Auth.TOTP.SecretLength, err = intFromEnv("TOTP_SECRET_LENGTH", 32)
	if err != nil {
		return Config{}, fmt.Errorf("parse TOTP_SECRET_LENGTH: %w", err)
	}
	cfg.Auth.TOTP.Digits, err = intFromEnv("TOTP_DIGITS", 6)
	if err != nil {
		return Config{}, fmt.Errorf("parse TOTP_DIGITS: %w", err)
	}
	totpPeriod, err := intFromEnv("TOTP_PERIOD_SECONDS", 30)
	if err != nil {
		return Config{}, fmt.Errorf("parse TOTP_PERIOD_SECONDS: %w", err)
	}
	cfg.Auth.TOTP.PeriodSeconds = uint(totpPeriod)

	cfg.RateLimit.RequestsPerSecond = reqPerSec
	cfg.RateLimit.Burst = burst

	cfg.Billing.DefaultCurrency = getEnv("BILLING_DEFAULT_CURRENCY", "TRY")
	cfg.Billing.Stripe.APIKey = getEnv("STRIPE_SECRET", "")
	cfg.Billing.Stripe.WebhookSecret = getEnv("STRIPE_WEBHOOK_SECRET", "")
	cfg.Billing.Stripe.SuccessURL = getEnv("STRIPE_SUCCESS_URL", "")
	cfg.Billing.Stripe.CancelURL = getEnv("STRIPE_CANCEL_URL", "")
	cfg.Billing.Iyzico.APIKey = getEnv("IYZICO_API_KEY", "")
	cfg.Billing.Iyzico.SecretKey = getEnv("IYZICO_SECRET_KEY", "")
	cfg.Billing.Iyzico.BaseURL = getEnv("IYZICO_BASE_URL", "")

	cfg.Node.ProvisionToken = getEnv("NODE_PROVISION_TOKEN", "")

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.RateLimit.RequestsPerSecond <= 0 {
		return errors.New("rate limit RPS must be greater than zero")
	}
	if cfg.RateLimit.Burst <= 0 {
		return errors.New("rate limit burst must be greater than zero")
	}
	if cfg.Auth.JWTSigningKey == "" {
		return errors.New("jwt secret is required")
	}
	if cfg.Database.DSN == "" {
		return errors.New("postgres dsn is required")
	}
	if cfg.Auth.AccessTokenTTL <= 0 {
		return errors.New("access token ttl must be greater than zero")
	}
	if cfg.Auth.RefreshTokenTTL <= 0 {
		return errors.New("refresh token ttl must be greater than zero")
	}
	if cfg.Auth.VerificationTokenTTL <= 0 {
		return errors.New("verification token ttl must be greater than zero")
	}
	if cfg.Auth.PasswordResetTokenTTL <= 0 {
		return errors.New("password reset token ttl must be greater than zero")
	}
	if cfg.Auth.ArgonMemory == 0 || cfg.Auth.ArgonIterations == 0 || cfg.Auth.ArgonSaltLength == 0 || cfg.Auth.ArgonKeyLength == 0 {
		return errors.New("argon parameters must be greater than zero")
	}
	if cfg.Auth.TOTP.SecretLength < 16 {
		return errors.New("totp secret length must be at least 16 bytes")
	}
	if cfg.Auth.TOTP.Digits != 6 && cfg.Auth.TOTP.Digits != 8 {
		return errors.New("totp digits must be 6 or 8")
	}
	if cfg.Auth.TOTP.PeriodSeconds == 0 {
		return errors.New("totp period must be greater than zero")
	}
	if cfg.Billing.DefaultCurrency == "" {
		return errors.New("billing default currency is required")
	}
	if cfg.Node.ProvisionToken == "" {
		return errors.New("node provision token is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	if value, ok := os.LookupEnv(key); ok {
		dur, err := time.ParseDuration(value)
		if err != nil {
			return 0, err
		}
		return dur, nil
	}
	return fallback, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	}
	return fallback, nil
}

func floatFromEnv(key string, fallback float64) (float64, error) {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	}
	return fallback, nil
}

func uint32FromEnv(key string, fallback uint32) (uint32, error) {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(parsed), nil
	}
	return fallback, nil
}
