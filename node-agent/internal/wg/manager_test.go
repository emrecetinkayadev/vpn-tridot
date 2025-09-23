package wg

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
)

func TestWritePeersCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(config.WireGuardConfig{
		InterfaceName:       "wg0",
		ListenPort:          51820,
		AddressCIDR:         "10.0.0.2/32",
		DNS:                 []string{"1.1.1.1", "8.8.8.8"},
		MTU:                 1420,
		PersistentKeepalive: 25,
		ConfigDirectory:     dir,
	})

	path, err := mgr.WritePeers([]Peer{{
		PublicKey:      "server-public",
		AllowedIPs:     []string{"0.0.0.0/0"},
		Endpoint:       "vpn.example.com:51820",
		PersistentKeep: 30,
	}})
	require.NoError(t, err)
	require.FileExists(t, path)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "[Interface]")
	require.Contains(t, string(content), "DNS = 1.1.1.1,8.8.8.8")
	require.Contains(t, string(content), "PersistentKeepalive = 30")
}

func TestEnsureBaseConfig(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(config.WireGuardConfig{
		InterfaceName:   "wg0",
		ListenPort:      51820,
		ConfigDirectory: dir,
	})

	path, err := mgr.EnsureBaseConfig()
	require.NoError(t, err)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "[Interface]")
}

func TestManagerValidation(t *testing.T) {
	mgr := NewManager(config.WireGuardConfig{})
	_, err := mgr.WritePeers(nil)
	require.Error(t, err)
}

func TestManagerStats(t *testing.T) {
	mgr := NewManager(config.WireGuardConfig{InterfaceName: "wg0"})
	mgr.WithClient(mockWGClient{stats: DeviceStats{PeerCount: 2, ActivePeers: 1, ReceiveBytes: 1234, TransmitBytes: 5678, LastHandshake: time.Unix(10, 0)}})

	stats, err := mgr.Stats()
	require.NoError(t, err)
	require.Equal(t, 2, stats.PeerCount)
	require.Equal(t, 1, stats.ActivePeers)
	require.Equal(t, uint64(1234), stats.ReceiveBytes)
	require.Equal(t, uint64(5678), stats.TransmitBytes)
	require.Equal(t, time.Unix(10, 0), stats.LastHandshake)
}

type mockWGClient struct {
	stats DeviceStats
	err   error
}

func (m mockWGClient) DeviceStats(string) (DeviceStats, error) {
	return m.stats, m.err
}
