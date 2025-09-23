package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

const (
	peersFile = "peers.json"
	drainFile = "drain"
)

// Store persists agent runtime state to disk for crash-safe recovery.
type Store struct {
	dir string
	mu  sync.Mutex
}

// New creates a state store rooted at dir. Directory is created if missing.
func New(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("state dir required")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) peersPath() string { return filepath.Join(s.dir, peersFile) }
func (s *Store) drainPath() string { return filepath.Join(s.dir, drainFile) }

// SavePeers persists the latest WireGuard peer definition for recovery.
func (s *Store) SavePeers(peers []wg.Peer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.peersPath(), data, 0o600)
}

// LoadPeers loads the previously saved peer definitions.
func (s *Store) LoadPeers() ([]wg.Peer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := os.ReadFile(s.peersPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var peers []wg.Peer
	if err := json.Unmarshal(content, &peers); err != nil {
		return nil, err
	}
	return peers, nil
}

// DrainEnabled reports whether drain mode is active.
func (s *Store) DrainEnabled() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := os.Stat(s.drainPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SetDrain toggles drain mode by creating/removing the drain marker file.
func (s *Store) SetDrain(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.drainPath()
	if enabled {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		return f.Close()
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Dir exposes the store directory (useful for logging/tests).
func (s *Store) Dir() string { return s.dir }
