package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

func TestStorePeersPersistence(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	require.NoError(t, err)

	peers := []wg.Peer{{PublicKey: "pk", AllowedIPs: []string{"0.0.0.0/0"}}}
	require.NoError(t, store.SavePeers(peers))

	loaded, err := store.LoadPeers()
	require.NoError(t, err)
	require.Equal(t, peers, loaded)

	// Ensure file written with restrictive permissions.
	info, err := os.Stat(filepath.Join(dir, "peers.json"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestStoreDrainToggle(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	require.NoError(t, err)

	enabled, err := store.DrainEnabled()
	require.NoError(t, err)
	require.False(t, enabled)

	require.NoError(t, store.SetDrain(true))
	enabled, err = store.DrainEnabled()
	require.NoError(t, err)
	require.True(t, enabled)

	require.NoError(t, store.SetDrain(false))
	enabled, err = store.DrainEnabled()
	require.NoError(t, err)
	require.False(t, enabled)
}

func TestLoadPeersMissingFile(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	require.NoError(t, err)

	peers, err := store.LoadPeers()
	require.NoError(t, err)
	require.Nil(t, peers)
}
