package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("CONTROL_PLANE_URL", "https://api.example.com")
	t.Setenv("NODE_PROVISION_TOKEN", "token-123")
	t.Setenv("MTLS_CA_PEM", "ca-pem")
	t.Setenv("MTLS_CLIENT_CERT", "cert-pem")
	t.Setenv("MTLS_CLIENT_KEY", "key-pem")
	t.Setenv("AGENT_POLL_INTERVAL", "45s")
	t.Setenv("WG_INTERFACE", "wg-test")
	t.Setenv("WG_PORT", "51821")
	t.Setenv("WG_ADDRESS", "10.10.0.2/32")
	t.Setenv("WG_DNS", "1.1.1.1,8.8.8.8")
	t.Setenv("WG_MTU", "1300")
	t.Setenv("WG_KEEPALIVE", "20")
	t.Setenv("WG_CONFIG_DIR", "/tmp/wg")
	t.Setenv("WG_ENABLE_NAT", "true")
	t.Setenv("WG_ENABLE_KILLSWITCH", "true")
	t.Setenv("AGENT_METRICS_ADDR", "127.0.0.1:9200")
	t.Setenv("AGENT_STATE_DIR", "/tmp/vpn-agent-state")
	t.Setenv("AGENT_MAX_RETRY_INTERVAL", "45s")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "https://api.example.com", cfg.ControlPlane.URL)
	require.Equal(t, "token-123", cfg.Provision.Token)
	require.Equal(t, "ca-pem", cfg.MTLS.CACert)
	require.Equal(t, "cert-pem", cfg.MTLS.Cert)
	require.Equal(t, "key-pem", cfg.MTLS.Key)
	require.Equal(t, "45s", cfg.Agent.PollInterval.String())
	require.Equal(t, "wg-test", cfg.WireGuard.InterfaceName)
	require.Equal(t, 51821, cfg.WireGuard.ListenPort)
	require.Equal(t, "10.10.0.2/32", cfg.WireGuard.AddressCIDR)
	require.ElementsMatch(t, []string{"1.1.1.1", "8.8.8.8"}, cfg.WireGuard.DNS)
	require.Equal(t, 1300, cfg.WireGuard.MTU)
	require.Equal(t, 20, cfg.WireGuard.PersistentKeepalive)
	require.Equal(t, "/tmp/wg", cfg.WireGuard.ConfigDirectory)
	require.True(t, cfg.WireGuard.EnableNAT)
	require.True(t, cfg.WireGuard.EnableKillSwitch)
	require.Equal(t, "127.0.0.1:9200", cfg.Agent.MetricsAddress)
	require.Equal(t, "/tmp/vpn-agent-state", cfg.Agent.StateDirectory)
	require.Equal(t, "45s", cfg.Agent.MaxRetryInterval.String())
}

func TestLoadFromFile(t *testing.T) {
	yamlContent := []byte(`controlPlane:
  url: https://cp.local
  registerPath: /register
  healthPath: /health
  timeout: 5s
provision:
  token: file-token
mtls:
  caFile: ca.pem
  certFile: cert.pem
  keyFile: key.pem
agent:
  pollInterval: 20s
  metricsAddr: 127.0.0.1:9300
  stateDir: ./state
  maxRetryInterval: 1m
wireguard:
  interfaceName: wg1
  listenPort: 51830
  address: 10.20.0.2/32
  dns:
    - 9.9.9.9
  mtu: 1280
  persistentKeepalive: 33
  configDir: wgconf
`)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "agent.yaml")
	require.NoError(t, os.WriteFile(configPath, yamlContent, 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.pem"), []byte("CA"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cert.pem"), []byte("CERT"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "key.pem"), []byte("KEY"), 0o600))

	t.Setenv("NODE_AGENT_CONFIG_FILE", configPath)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "https://cp.local", cfg.ControlPlane.URL)
	require.Equal(t, "/register", cfg.ControlPlane.RegisterPath)
	require.Equal(t, "file-token", cfg.Provision.Token)
	require.Equal(t, filepath.Join(dir, "ca.pem"), cfg.MTLS.CACertFile)
	require.Equal(t, filepath.Join(dir, "cert.pem"), cfg.MTLS.CertFile)
	require.Equal(t, filepath.Join(dir, "key.pem"), cfg.MTLS.KeyFile)
	require.Equal(t, "20s", cfg.Agent.PollInterval.String())
	require.Equal(t, "127.0.0.1:9300", cfg.Agent.MetricsAddress)
	require.Equal(t, filepath.Join(dir, "state"), cfg.Agent.StateDirectory)
	require.Equal(t, "1m0s", cfg.Agent.MaxRetryInterval.String())
	require.Equal(t, "wg1", cfg.WireGuard.InterfaceName)
	require.Equal(t, 51830, cfg.WireGuard.ListenPort)
	require.Equal(t, "10.20.0.2/32", cfg.WireGuard.AddressCIDR)
	require.Equal(t, []string{"9.9.9.9"}, cfg.WireGuard.DNS)
	require.Equal(t, 1280, cfg.WireGuard.MTU)
	require.Equal(t, 33, cfg.WireGuard.PersistentKeepalive)
	require.Equal(t, filepath.Join(dir, "wgconf"), cfg.WireGuard.ConfigDirectory)
	require.False(t, cfg.WireGuard.EnableNAT)
	require.False(t, cfg.WireGuard.EnableKillSwitch)
}

func TestValidateMissingValues(t *testing.T) {
	_, err := config.Load()
	require.Error(t, err)
}
