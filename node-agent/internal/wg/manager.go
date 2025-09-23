package wg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
)

// Peer describes a WireGuard peer entry.
type Peer struct {
	PublicKey      string   `json:"public_key"`
	PresharedKey   string   `json:"preshared_key"`
	AllowedIPs     []string `json:"allowed_ips"`
	Endpoint       string   `json:"endpoint"`
	PersistentKeep int      `json:"persistent_keepalive"`
}

// Manager writes WireGuard configuration files to disk.
type Manager struct {
	cfg    config.WireGuardConfig
	client wireGuardClient
}

// NewManager creates a config manager for WireGuard interface.
func NewManager(cfg config.WireGuardConfig) *Manager {
	return &Manager{cfg: cfg, client: defaultClient{}}
}

// WithClient allows injecting a custom WireGuard client (useful for tests).
func (m *Manager) WithClient(client wireGuardClient) {
	m.client = client
}

type wireGuardClient interface {
	DeviceStats(interfaceName string) (DeviceStats, error)
}

type DeviceStats struct {
	LastHandshake time.Time
	ReceiveBytes  uint64
	TransmitBytes uint64
	PeerCount     int
	ActivePeers   int
}

// Stats returns aggregated WireGuard device metrics.
func (m *Manager) Stats() (DeviceStats, error) {
	if m.client == nil {
		return DeviceStats{}, fmt.Errorf("wireguard client not configured")
	}
	return m.client.DeviceStats(m.cfg.InterfaceName)
}

// EnsureBaseConfig writes the base interface configuration (without peers).
func (m *Manager) EnsureBaseConfig() (string, error) {
	return m.WritePeers(nil)
}

// WritePeers renders WireGuard config with provided peers and writes to config directory.
func (m *Manager) WritePeers(peers []Peer) (string, error) {
	if m.cfg.InterfaceName == "" {
		return "", fmt.Errorf("interface name required")
	}
	if m.cfg.ConfigDirectory == "" {
		return "", fmt.Errorf("config directory required")
	}

	if err := os.MkdirAll(m.cfg.ConfigDirectory, 0o750); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	filePath := filepath.Join(m.cfg.ConfigDirectory, fmt.Sprintf("%s.conf", m.cfg.InterfaceName))
	content := m.render(peers)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return filePath, nil
}

func (m *Manager) render(peers []Peer) string {
	var b strings.Builder
	b.WriteString("[Interface]\n")
	if m.cfg.AddressCIDR != "" {
		b.WriteString("Address = ")
		b.WriteString(m.cfg.AddressCIDR)
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("ListenPort = %d\n", m.cfg.ListenPort))
	if m.cfg.MTU > 0 {
		b.WriteString(fmt.Sprintf("MTU = %d\n", m.cfg.MTU))
	}
	if len(m.cfg.DNS) > 0 {
		b.WriteString("DNS = ")
		b.WriteString(strings.Join(m.cfg.DNS, ","))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	for _, peer := range peers {
		b.WriteString("[Peer]\n")
		b.WriteString("PublicKey = ")
		b.WriteString(peer.PublicKey)
		b.WriteString("\n")
		if peer.PresharedKey != "" {
			b.WriteString("PresharedKey = ")
			b.WriteString(peer.PresharedKey)
			b.WriteString("\n")
		}
		if len(peer.AllowedIPs) > 0 {
			b.WriteString("AllowedIPs = ")
			b.WriteString(strings.Join(peer.AllowedIPs, ","))
			b.WriteString("\n")
		}
		if peer.Endpoint != "" {
			b.WriteString("Endpoint = ")
			b.WriteString(peer.Endpoint)
			b.WriteString("\n")
		}
		keepalive := peer.PersistentKeep
		if keepalive == 0 {
			keepalive = m.cfg.PersistentKeepalive
		}
		if keepalive > 0 {
			b.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", keepalive))
		}
		b.WriteString("\n")
	}

	return b.String()
}
