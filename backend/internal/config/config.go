package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application-level configuration loaded from environment variables.
type Config struct {
	App           AppConfig
	HTTP          HTTPConfig
	Database      DatabaseConfig
	Log           LogConfig
	Auth          AuthConfig
	RateLimit     RateLimitConfig
	Billing       BillingConfig
	Node          NodeConfig
	Security      SecurityConfig
	Observability ObservabilityConfig
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
	Level   string
	Request RequestLogConfig
}

type RequestLogConfig struct {
	Enabled         bool
	Headers         []string
	MaskHeaders     []string
	QueryParams     []string
	MaskQueryParams []string
}

type ObservabilityConfig struct {
	Metrics MetricsConfig
}

type MetricsConfig struct {
	Enabled   bool
	Path      string
	Namespace string
	Subsystem string
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
	Auth              RateLimitRuleConfig
	Checkout          RateLimitRuleConfig
	Peers             RateLimitRuleConfig
}

type RateLimitRuleConfig struct {
	Enabled           bool
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

type SecurityConfig struct {
	HCaptcha HCaptchaConfig
	CORS     CORSConfig
	CSRF     CSRFConfig
	Secrets  SecretsConfig
}

type HCaptchaConfig struct {
	Enabled        bool
	Secret         string
	SiteKey        string
	ScoreThreshold float64
	Endpoint       string
}

type CORSConfig struct {
	Enabled          bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type CSRFConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	ProtectedMethods []string
	AllowNoOrigin    bool
}

type SecretsConfig struct {
	SOPS  SOPSConfig
	Vault VaultConfig
}

type SOPSConfig struct {
	Enabled bool
	Path    string
	Format  string
}

type VaultConfig struct {
	Enabled       bool
	Address       string
	Token         string
	Path          string
	Namespace     string
	Timeout       time.Duration
	TLSSkipVerify bool
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
	logRequestsEnabled, err := boolFromEnv("LOG_REQUESTS_ENABLED", true)
	if err != nil {
		return Config{}, fmt.Errorf("parse LOG_REQUESTS_ENABLED: %w", err)
	}
	cfg.Log.Request.Enabled = logRequestsEnabled
	cfg.Log.Request.Headers = stringSliceFromEnv("LOG_REQUEST_HEADERS", "x-request-id,authorization")
	cfg.Log.Request.MaskHeaders = stringSliceFromEnv("LOG_MASK_HEADERS", "authorization,proxy-authorization,x-api-key,cookie,set-cookie")
	cfg.Log.Request.QueryParams = stringSliceFromEnv("LOG_REQUEST_QUERY_PARAMS", "token")
	cfg.Log.Request.MaskQueryParams = stringSliceFromEnv("LOG_MASK_QUERY_PARAMS", "token,auth,code,password")

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
	cfg.RateLimit.Auth = loadRateLimitRule("RATE_LIMIT_AUTH", cfg.RateLimit)
	cfg.RateLimit.Checkout = loadRateLimitRule("RATE_LIMIT_CHECKOUT", cfg.RateLimit)
	cfg.RateLimit.Peers = loadRateLimitRule("RATE_LIMIT_PEERS", cfg.RateLimit)

	cfg.Billing.DefaultCurrency = getEnv("BILLING_DEFAULT_CURRENCY", "TRY")
	cfg.Billing.Stripe.APIKey = getEnv("STRIPE_SECRET", "")
	cfg.Billing.Stripe.WebhookSecret = getEnv("STRIPE_WEBHOOK_SECRET", "")
	cfg.Billing.Stripe.SuccessURL = getEnv("STRIPE_SUCCESS_URL", "")
	cfg.Billing.Stripe.CancelURL = getEnv("STRIPE_CANCEL_URL", "")
	cfg.Billing.Iyzico.APIKey = getEnv("IYZICO_API_KEY", "")
	cfg.Billing.Iyzico.SecretKey = getEnv("IYZICO_SECRET_KEY", "")
	cfg.Billing.Iyzico.BaseURL = getEnv("IYZICO_BASE_URL", "")

	cfg.Node.ProvisionToken = getEnv("NODE_PROVISION_TOKEN", "")

	hCaptchaEnabled, err := boolFromEnv("HCAPTCHA_ENABLED", false)
	if err != nil {
		return Config{}, fmt.Errorf("parse HCAPTCHA_ENABLED: %w", err)
	}
	cfg.Security.HCaptcha.Enabled = hCaptchaEnabled
	cfg.Security.HCaptcha.Secret = getEnv("HCAPTCHA_SECRET", "")
	cfg.Security.HCaptcha.SiteKey = getEnv("HCAPTCHA_SITEKEY", "")
	threshold, err := floatFromEnv("HCAPTCHA_SCORE_THRESHOLD", 0.5)
	if err != nil {
		return Config{}, fmt.Errorf("parse HCAPTCHA_SCORE_THRESHOLD: %w", err)
	}
	cfg.Security.HCaptcha.ScoreThreshold = threshold
	cfg.Security.HCaptcha.Endpoint = getEnv("HCAPTCHA_ENDPOINT", "https://hcaptcha.com/siteverify")

	corsEnabled, err := boolFromEnv("CORS_ENABLED", true)
	if err != nil {
		return Config{}, fmt.Errorf("parse CORS_ENABLED: %w", err)
	}
	cfg.Security.CORS.Enabled = corsEnabled
	cfg.Security.CORS.AllowOrigins = stringSliceFromEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.Security.CORS.AllowMethods = upperSlice(stringSliceFromEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"))
	cfg.Security.CORS.AllowHeaders = stringSliceFromEnv("CORS_ALLOWED_HEADERS", "Authorization,Content-Type,X-CSRF-Token")
	cfg.Security.CORS.ExposeHeaders = stringSliceFromEnv("CORS_EXPOSE_HEADERS", "Link")
	allowCredentials, err := boolFromEnv("CORS_ALLOW_CREDENTIALS", false)
	if err != nil {
		return Config{}, fmt.Errorf("parse CORS_ALLOW_CREDENTIALS: %w", err)
	}
	cfg.Security.CORS.AllowCredentials = allowCredentials
	maxAge, err := durationFromEnv("CORS_MAX_AGE", 10*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("parse CORS_MAX_AGE: %w", err)
	}
	cfg.Security.CORS.MaxAge = maxAge

	defaultCSRFOrigins := strings.Join(cfg.Security.CORS.AllowOrigins, ",")
	csrfEnabled, err := boolFromEnv("CSRF_ENABLED", true)
	if err != nil {
		return Config{}, fmt.Errorf("parse CSRF_ENABLED: %w", err)
	}
	cfg.Security.CSRF.Enabled = csrfEnabled
	cfg.Security.CSRF.AllowedOrigins = stringSliceFromEnv("CSRF_ALLOWED_ORIGINS", defaultCSRFOrigins)
	cfg.Security.CSRF.ProtectedMethods = upperSlice(stringSliceFromEnv("CSRF_PROTECTED_METHODS", "POST,PUT,PATCH,DELETE"))
	allowNoOrigin, err := boolFromEnv("CSRF_ALLOW_NO_ORIGIN", true)
	if err != nil {
		return Config{}, fmt.Errorf("parse CSRF_ALLOW_NO_ORIGIN: %w", err)
	}
	cfg.Security.CSRF.AllowNoOrigin = allowNoOrigin

	sopsEnabled, err := boolFromEnv("SOPS_SECRETS_ENABLED", false)
	if err != nil {
		return Config{}, fmt.Errorf("parse SOPS_SECRETS_ENABLED: %w", err)
	}
	cfg.Security.Secrets.SOPS.Enabled = sopsEnabled
	cfg.Security.Secrets.SOPS.Path = getEnv("SOPS_SECRETS_PATH", "")
	cfg.Security.Secrets.SOPS.Format = strings.ToLower(getEnv("SOPS_SECRETS_FORMAT", "json"))

	vaultEnabled, err := boolFromEnv("VAULT_ENABLED", false)
	if err != nil {
		return Config{}, fmt.Errorf("parse VAULT_ENABLED: %w", err)
	}
	cfg.Security.Secrets.Vault.Enabled = vaultEnabled
	cfg.Security.Secrets.Vault.Address = getEnv("VAULT_ADDR", "")
	cfg.Security.Secrets.Vault.Token = getEnv("VAULT_TOKEN", "")
	cfg.Security.Secrets.Vault.Path = strings.TrimPrefix(getEnv("VAULT_PATH", ""), "/")
	cfg.Security.Secrets.Vault.Namespace = getEnv("VAULT_NAMESPACE", "")
	vaultTimeout, err := durationFromEnv("VAULT_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse VAULT_TIMEOUT: %w", err)
	}
	cfg.Security.Secrets.Vault.Timeout = vaultTimeout
	tlsSkip, err := boolFromEnv("VAULT_TLS_SKIP_VERIFY", false)
	if err != nil {
		return Config{}, fmt.Errorf("parse VAULT_TLS_SKIP_VERIFY: %w", err)
	}
	cfg.Security.Secrets.Vault.TLSSkipVerify = tlsSkip

	metricsEnabled, err := boolFromEnv("METRICS_ENABLED", true)
	if err != nil {
		return Config{}, fmt.Errorf("parse METRICS_ENABLED: %w", err)
	}
	cfg.Observability.Metrics.Enabled = metricsEnabled
	cfg.Observability.Metrics.Path = getEnv("METRICS_PATH", "/metrics")
	defaultNamespace := sanitizePrometheusName(getEnv("METRICS_NAMESPACE", cfg.App.Name))
	if defaultNamespace == "" {
		defaultNamespace = "vpn_backend"
	}
	cfg.Observability.Metrics.Namespace = defaultNamespace
	cfg.Observability.Metrics.Subsystem = sanitizePrometheusName(getEnv("METRICS_SUBSYSTEM", "http"))
	if cfg.Observability.Metrics.Subsystem == "" {
		cfg.Observability.Metrics.Subsystem = "http"
	}

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
	if cfg.Observability.Metrics.Enabled {
		if cfg.Observability.Metrics.Path == "" {
			return errors.New("metrics path is required when metrics are enabled")
		}
		if !strings.HasPrefix(cfg.Observability.Metrics.Path, "/") {
			return errors.New("metrics path must start with /")
		}
	}
	if cfg.Security.HCaptcha.Enabled {
		if cfg.Security.HCaptcha.Secret == "" {
			return errors.New("hcaptcha secret is required when hcaptcha is enabled")
		}
		if cfg.Security.HCaptcha.ScoreThreshold < 0 {
			return errors.New("hcaptcha score threshold must be >= 0")
		}
	}
	if cfg.Security.Secrets.SOPS.Enabled {
		switch cfg.Security.Secrets.SOPS.Format {
		case "json", "yaml", "yml", "env":
		default:
			return fmt.Errorf("unsupported sops format: %s", cfg.Security.Secrets.SOPS.Format)
		}
		if cfg.Security.Secrets.SOPS.Path == "" {
			return errors.New("sops secrets path is required when sops secrets are enabled")
		}
	}
	if cfg.Security.Secrets.Vault.Enabled {
		if cfg.Security.Secrets.Vault.Address == "" {
			return errors.New("vault address is required when vault is enabled")
		}
		if cfg.Security.Secrets.Vault.Token == "" {
			return errors.New("vault token is required when vault is enabled")
		}
		if cfg.Security.Secrets.Vault.Path == "" {
			return errors.New("vault path is required when vault is enabled")
		}
		if cfg.Security.Secrets.Vault.Timeout <= 0 {
			return errors.New("vault timeout must be greater than zero")
		}
	}
	if cfg.Security.CORS.Enabled {
		if len(cfg.Security.CORS.AllowOrigins) == 0 {
			return errors.New("cors allow origins must not be empty")
		}
		if len(cfg.Security.CORS.AllowMethods) == 0 {
			return errors.New("cors allow methods must not be empty")
		}
		if cfg.Security.CORS.MaxAge < 0 {
			return errors.New("cors max age must be >= 0")
		}
	}
	if cfg.Security.CSRF.Enabled {
		if len(cfg.Security.CSRF.AllowedOrigins) == 0 {
			return errors.New("csrf allowed origins must not be empty")
		}
		if len(cfg.Security.CSRF.ProtectedMethods) == 0 {
			return errors.New("csrf protected methods must not be empty")
		}
	}
	for name, rule := range map[string]RateLimitRuleConfig{
		"auth":     cfg.RateLimit.Auth,
		"checkout": cfg.RateLimit.Checkout,
		"peers":    cfg.RateLimit.Peers,
	} {
		if !rule.Enabled {
			continue
		}
		if rule.RequestsPerSecond <= 0 {
			return fmt.Errorf("rate limit %s rps must be greater than zero", name)
		}
		if rule.Burst <= 0 {
			return fmt.Errorf("rate limit %s burst must be greater than zero", name)
		}
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

func boolFromEnv(key string, fallback bool) (bool, error) {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return false, err
		}
		return parsed, nil
	}
	return fallback, nil
}

func stringSliceFromEnv(key, fallback string) []string {
	value := getEnv(key, fallback)
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func upperSlice(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strings.ToUpper(value))
	}
	return result
}

func sanitizePrometheusName(v string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ToLower(trimmed)
	return prometheusNameReplacer.Replace(trimmed)
}

var prometheusNameReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_")

func (c RateLimitConfig) DefaultRule() RateLimitRuleConfig {
	return RateLimitRuleConfig{
		Enabled:           true,
		RequestsPerSecond: c.RequestsPerSecond,
		Burst:             c.Burst,
	}
}

func loadRateLimitRule(prefix string, base RateLimitConfig) RateLimitRuleConfig {
	enabled, err := boolFromEnv(prefix+"_ENABLED", true)
	if err != nil {
		enabled = true
	}
	rps, err := floatFromEnv(prefix+"_RPS", base.RequestsPerSecond)
	if err != nil {
		rps = base.RequestsPerSecond
	}
	burst, err := intFromEnv(prefix+"_BURST", base.Burst)
	if err != nil {
		burst = base.Burst
	}
	return RateLimitRuleConfig{
		Enabled:           enabled,
		RequestsPerSecond: rps,
		Burst:             burst,
	}
}
