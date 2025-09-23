package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAgentRun(t *testing.T) {
	tr := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})

	client := &http.Client{Transport: tr}
	cfg := config.Config{
		ControlPlane: config.ControlPlaneConfig{
			URL:          "https://control",
			RegisterPath: "/register",
			HealthPath:   "/health",
			Timeout:      5 * time.Second,
		},
		Provision: config.ProvisionConfig{Token: "test"},
		MTLS:      config.MTLSConfig{CACert: "ca", Cert: "cert", Key: "key"},
		Agent:     config.AgentConfig{PollInterval: 10 * time.Millisecond},
	}

	a, err := New(cfg, client)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	require.NoError(t, a.Run(ctx))
}

func TestJoinURL(t *testing.T) {
	joined, err := JoinURL("https://api.example.com", "/register")
	require.NoError(t, err)
	require.Equal(t, "https://api.example.com/register", joined)
}

func TestApplyPeers(t *testing.T) {
	mgr := &wgManagerStub{}
	a := &Agent{}
	state := &stateStub{}
	a.WithState(state)
	a.WithWireGuard(mgr, "/etc/wireguard/wg0.conf", nil, func(path string) error {
		mgr.synced = append(mgr.synced, path)
		return nil
	})

	err := a.ApplyPeers([]wg.Peer{{PublicKey: "pk", AllowedIPs: []string{"0.0.0.0/0"}}})
	require.NoError(t, err)
	require.Len(t, mgr.configs, 1)
	require.Equal(t, "/etc/wireguard/wg0.conf", mgr.synced[0])
	require.Len(t, state.savedPeers, 1)
}

func TestReportHealthIncludesDrain(t *testing.T) {
	stats := wg.DeviceStats{PeerCount: 2, ActivePeers: 1, ReceiveBytes: 10, TransmitBytes: 20, LastHandshake: time.Unix(100, 0)}
	mgr := &wgManagerStub{stats: stats}
	state := &stateStub{drain: true}
	reqBody := make(chan []byte, 1)
	tr := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		reqBody <- body
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})
	client := &http.Client{Transport: tr}
	cfg := config.Config{
		ControlPlane: config.ControlPlaneConfig{URL: "https://cp", HealthPath: "/health"},
		Provision:    config.ProvisionConfig{Token: "tok"},
		Agent:        config.AgentConfig{PollInterval: time.Second},
	}
	a, err := New(cfg, client)
	require.NoError(t, err)
	a.WithWireGuard(mgr, "", nil, nil)
	a.WithState(state)
	require.NoError(t, a.reportHealth(context.Background()))
	select {
	case body := <-reqBody:
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		wgPayload, ok := payload["wireguard"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, true, wgPayload["drain"])
	case <-time.After(time.Second):
		t.Fatal("expected health request")
	}
}

func TestWithRetryRetriesUntilSuccess(t *testing.T) {
	originalSleep := sleepDelay
	sleepDelay = func(context.Context, time.Duration) error { return nil }
	defer func() { sleepDelay = originalSleep }()

	attempts := 0
	a := &Agent{maxRetry: 50 * time.Millisecond, retryBase: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := a.withRetry(ctx, "test", func(context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("fail")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

type wgManagerStub struct {
	configs []wg.Peer
	synced  []string
	stats   wg.DeviceStats
}

func (m *wgManagerStub) WritePeers(peers []wg.Peer) (string, error) {
	path := "/etc/wireguard/wg0.conf"
	m.configs = append(m.configs, peers...)
	return path, nil
}

func (m *wgManagerStub) Stats() (wg.DeviceStats, error) {
	return m.stats, nil
}

type stateStub struct {
	savedPeers [][]wg.Peer
	drain      bool
}

func (s *stateStub) SavePeers(peers []wg.Peer) error {
	s.savedPeers = append(s.savedPeers, peers)
	return nil
}

func (s *stateStub) LoadPeers() ([]wg.Peer, error) { return nil, nil }

func (s *stateStub) DrainEnabled() (bool, error) { return s.drain, nil }
