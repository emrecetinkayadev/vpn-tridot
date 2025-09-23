package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

// Agent represents the node agent runtime.
type Agent struct {
	cfg          config.Config
	client       *http.Client
	wgManager    wireGuardManager
	wgConfigPath string
	wgUp         func(string) error
	wgSync       func(string) error
	metrics      metricsExporter
	prevStats    wg.DeviceStats
	prevStatsAt  time.Time
	state        stateStore
	maxRetry     time.Duration
	retryBase    time.Duration
}

type wireGuardManager interface {
	WritePeers([]wg.Peer) (string, error)
	Stats() (wg.DeviceStats, error)
}

type metricsExporter interface {
	Update(wg.DeviceStats)
	Handler() http.Handler
}

type stateStore interface {
	SavePeers([]wg.Peer) error
	LoadPeers() ([]wg.Peer, error)
	DrainEnabled() (bool, error)
}

// New creates a node agent with the provided configuration and HTTP client.
func New(cfg config.Config, client *http.Client) (*Agent, error) {
	if client == nil {
		return nil, fmt.Errorf("http client required")
	}
	maxRetry := cfg.Agent.MaxRetryInterval
	if maxRetry <= 0 {
		maxRetry = 30 * time.Second
	}
	return &Agent{cfg: cfg, client: client, maxRetry: maxRetry, retryBase: time.Second}, nil
}

// WithWireGuard configures WireGuard helpers for the agent.
func (a *Agent) WithWireGuard(manager wireGuardManager, configPath string, upFn, syncFn func(string) error) {
	a.wgManager = manager
	a.wgConfigPath = configPath
	a.wgUp = upFn
	a.wgSync = syncFn
}

// WithMetrics sets the metrics exporter for Prometheus reporting.
func (a *Agent) WithMetrics(exporter metricsExporter) {
	a.metrics = exporter
}

// WithState configures persistent state handling for crash-safe recovery.
func (a *Agent) WithState(store stateStore) {
	a.state = store
}

// ApplyPeers writes new peer configuration and triggers sync command if configured.
func (a *Agent) ApplyPeers(peers []wg.Peer) error {
	if a.wgManager == nil {
		return fmt.Errorf("wireguard manager not configured")
	}
	path, err := a.wgManager.WritePeers(peers)
	if err != nil {
		return err
	}
	if a.wgSync != nil {
		if err := a.wgSync(path); err != nil {
			return err
		}
	}
	if a.state != nil {
		if err := a.state.SavePeers(peers); err != nil {
			log.Printf("agent: persist peers failed: %v", err)
		}
	}
	return nil
}

// Run starts the agent loop until the context is cancelled.
func (a *Agent) Run(ctx context.Context) error {
	if a.wgUp != nil && a.wgConfigPath != "" {
		if err := a.wgUp(a.wgConfigPath); err != nil {
			log.Printf("agent: wireguard setup failed: %v", err)
		}
	}
	if err := a.registerWithRetry(ctx); err != nil {
		return err
	}
	if err := a.reportHealthWithRetry(ctx); err != nil {
		log.Printf("agent: initial health report failed: %v", err)
	}

	ticker := time.NewTicker(a.cfg.Agent.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.reportHealthWithRetry(ctx); err != nil {
				log.Printf("agent: health report failed: %v", err)
			}
		}
	}
}

func (a *Agent) registerWithRetry(ctx context.Context) error {
	return a.withRetry(ctx, "register", a.doRegister)
}

func (a *Agent) reportHealthWithRetry(ctx context.Context) error {
	return a.withRetry(ctx, "health", func(ctx context.Context) error { return a.reportHealth(ctx) })
}

func (a *Agent) doRegister(ctx context.Context) error {
	registerURL, err := JoinURL(a.cfg.ControlPlane.URL, a.cfg.ControlPlane.RegisterPath)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, nil)
	if err != nil {
		return err
	}
	addAuthHeaders(req, a.cfg.Provision.Token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("register failed with status %d", resp.StatusCode)
	}

	if a.wgSync != nil && a.wgConfigPath != "" {
		if err := a.wgSync(a.wgConfigPath); err != nil {
			log.Printf("agent: wireguard sync failed: %v", err)
		}
	}

	return nil
}

func (a *Agent) reportHealth(ctx context.Context) error {
	healthURL, err := JoinURL(a.cfg.ControlPlane.URL, a.cfg.ControlPlane.HealthPath)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, healthURL, strings.NewReader("{}"))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	addAuthHeaders(req, a.cfg.Provision.Token)

	body := map[string]any{
		"timestamp": time.Now().UTC(),
	}

	if a.wgManager != nil {
		if stats, err := a.wgManager.Stats(); err == nil {
			now := time.Now()
			rxBps, txBps := a.computeThroughput(now, stats)
			if a.metrics != nil {
				a.metrics.Update(stats)
			}
			ratio := 0.0
			if stats.PeerCount > 0 {
				ratio = float64(stats.ActivePeers) / float64(stats.PeerCount)
			}
			wgState := map[string]any{
				"peer_count":        stats.PeerCount,
				"active_peer_count": stats.ActivePeers,
				"handshake_ratio":   ratio,
				"last_handshake":    stats.LastHandshake,
				"rx_bytes":          stats.ReceiveBytes,
				"tx_bytes":          stats.TransmitBytes,
				"rx_bps":            rxBps,
				"tx_bps":            txBps,
			}
			if a.state != nil {
				if drain, err := a.state.DrainEnabled(); err == nil {
					wgState["drain"] = drain
				} else {
					log.Printf("agent: drain state read failed: %v", err)
				}
			}
			body["wireguard"] = wgState
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(payload))
	req.ContentLength = int64(len(payload))

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (a *Agent) computeThroughput(now time.Time, stats wg.DeviceStats) (float64, float64) {
	if a.prevStatsAt.IsZero() {
		a.prevStats = stats
		a.prevStatsAt = now
		return 0, 0
	}
	elapsed := now.Sub(a.prevStatsAt).Seconds()
	if elapsed <= 0 {
		a.prevStats = stats
		a.prevStatsAt = now
		return 0, 0
	}
	rxDelta := counterDelta(stats.ReceiveBytes, a.prevStats.ReceiveBytes)
	txDelta := counterDelta(stats.TransmitBytes, a.prevStats.TransmitBytes)
	a.prevStats = stats
	a.prevStatsAt = now
	return float64(rxDelta) * 8 / elapsed, float64(txDelta) * 8 / elapsed
}

func addAuthHeaders(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "vpn-node-agent/1.0")
}

// JoinURL joins control plane base URL with a relative path.
func JoinURL(base, p string) (string, error) {
	if p == "" {
		return base, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	joined, err := url.JoinPath(u.String(), p)
	if err != nil {
		return "", err
	}
	return joined, nil
}

func counterDelta(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	return current
}

var sleepDelay = func(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (a *Agent) withRetry(ctx context.Context, label string, fn func(context.Context) error) error {
	backoff := a.retryBase
	if backoff <= 0 {
		backoff = time.Second
	}
	for {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		log.Printf("agent: %s attempt failed: %v", label, err)
		wait := backoff
		if backoff < a.maxRetry {
			wait = nextBackoff(backoff)
		}
		if err := sleepDelay(ctx, wait); err != nil {
			return err
		}
		backoff *= 2
		if backoff > a.maxRetry {
			backoff = a.maxRetry
		}
	}
}

func nextBackoff(base time.Duration) time.Duration {
	if base <= 0 {
		return time.Millisecond
	}
	min := base / 2
	if min <= 0 {
		min = time.Millisecond
	}
	randRange := base - min
	if randRange <= 0 {
		randRange = min
	}
	return min + time.Duration(rand.Int63n(int64(randRange)+1))
}
