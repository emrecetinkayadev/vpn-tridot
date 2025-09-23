package hcaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

var (
	// ErrTokenMissing indicates that the captcha token is missing or empty.
	ErrTokenMissing = errors.New("captcha token required")
	// ErrVerificationFailed indicates that hCaptcha verification failed.
	ErrVerificationFailed = errors.New("captcha verification failed")
)

// Verifier validates hCaptcha tokens against the remote API.
type Verifier struct {
	cfg        config.HCaptchaConfig
	httpClient *http.Client
}

// New creates a verifier using the provided configuration.
func New(cfg config.HCaptchaConfig) *Verifier {
	client := &http.Client{Timeout: 5 * time.Second}
	return &Verifier{cfg: cfg, httpClient: client}
}

// Enabled reports whether the verifier should run.
func (v *Verifier) Enabled() bool {
	return v != nil && v.cfg.Enabled
}

// Verify validates a token with the hCaptcha API.
func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) error {
	if !v.Enabled() {
		return nil
	}
	if token = stringTrim(token); token == "" {
		return ErrTokenMissing
	}

	form := url.Values{}
	form.Set("secret", v.cfg.Secret)
	form.Set("response", token)
	if remote := stringTrim(remoteIP); remote != "" {
		form.Set("remoteip", remote)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.cfg.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrVerificationFailed
	}

	var payload response
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}

	if !payload.Success {
		return ErrVerificationFailed
	}

	if payload.Score != nil && v.cfg.ScoreThreshold > 0 {
		if *payload.Score < v.cfg.ScoreThreshold {
			return ErrVerificationFailed
		}
	}

	return nil
}

// WithHTTPClient replaces the default HTTP client. Useful for tests.
func (v *Verifier) WithHTTPClient(client *http.Client) *Verifier {
	if client != nil {
		v.httpClient = client
	}
	return v
}

type response struct {
	Success    bool     `json:"success"`
	Score      *float64 `json:"score"`
	Challenge  string   `json:"challenge_ts"`
	Hostname   string   `json:"hostname"`
	Credit     bool     `json:"credit"`
	ErrorCodes []string `json:"error-codes"`
}

func stringTrim(s string) string {
	return strings.TrimSpace(s)
}
