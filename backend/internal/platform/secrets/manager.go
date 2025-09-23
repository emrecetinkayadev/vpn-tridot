package secrets

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
)

// Manager loads secrets from configured providers (SOPS, Vault) and exposes them.
type Manager struct {
	cfg        config.SecretsConfig
	httpClient *http.Client
}

// NewManager constructs a secrets manager for the given configuration.
func NewManager(cfg config.SecretsConfig) *Manager {
	client := &http.Client{Timeout: cfg.Vault.Timeout}
	if cfg.Vault.TLSSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}
	if client.Timeout == 0 {
		client.Timeout = 5 * time.Second
	}
	return &Manager{cfg: cfg, httpClient: client}
}

// Enabled reports whether any secret provider is active.
func (m *Manager) Enabled() bool {
	return m.cfg.SOPS.Enabled || m.cfg.Vault.Enabled
}

// Load fetches secrets from enabled providers and merges them into a flat map.
func (m *Manager) Load(ctx context.Context) (map[string]string, error) {
	secrets := make(map[string]string)

	if m.cfg.SOPS.Enabled {
		sopsSecrets, err := m.loadSOPS()
		if err != nil {
			return nil, fmt.Errorf("load sops secrets: %w", err)
		}
		mergeMaps(secrets, sopsSecrets)
	}

	if m.cfg.Vault.Enabled {
		vaultSecrets, err := m.loadVault(ctx)
		if err != nil {
			return nil, fmt.Errorf("load vault secrets: %w", err)
		}
		mergeMaps(secrets, vaultSecrets)
	}

	return secrets, nil
}

// Apply writes secret key/value pairs into the process environment.
func (m *Manager) Apply(values map[string]string) {
	for key, value := range values {
		_ = os.Setenv(key, value)
	}
}

func (m *Manager) loadSOPS() (map[string]string, error) {
	path := m.cfg.SOPS.Path
	if path == "" {
		return nil, errors.New("sops path is empty")
	}

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(m.cfg.SOPS.Format) {
	case "json":
		return decodeMapJSON(content)
	case "yaml", "yml":
		return decodeMapYAML(content)
	case "env":
		return decodeDotEnv(content)
	default:
		return nil, fmt.Errorf("unsupported sops format: %s", m.cfg.SOPS.Format)
	}
}

func (m *Manager) loadVault(ctx context.Context) (map[string]string, error) {
	addr := strings.TrimSuffix(m.cfg.Vault.Address, "/")
	path := strings.TrimPrefix(m.cfg.Vault.Path, "/")
	if addr == "" || path == "" {
		return nil, errors.New("vault address or path is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr+"/v1/"+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", m.cfg.Vault.Token)
	if ns := strings.TrimSpace(m.cfg.Vault.Namespace); ns != "" {
		req.Header.Set("X-Vault-Namespace", ns)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("vault request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	data, err := extractVaultData(payload["data"])
	if err != nil {
		return nil, err
	}

	return data, nil
}

func decodeMapJSON(content []byte) (map[string]string, error) {
	var raw map[string]any
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, err
	}
	return flattenStringMap(raw)
}

func decodeMapYAML(content []byte) (map[string]string, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, err
	}
	return flattenStringMap(raw)
}

func decodeDotEnv(content []byte) (map[string]string, error) {
	result := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env line: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty env key in line: %s", line)
		}
		result[key] = strings.Trim(value, "\"'")
	}
	return result, nil
}

func flattenStringMap(raw map[string]any) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range raw {
		switch v := value.(type) {
		case string:
			result[key] = v
		case fmt.Stringer:
			result[key] = v.String()
		case float64, float32, int, int64, int32, uint64, uint32, bool:
			result[key] = fmt.Sprint(v)
		case map[string]any:
			nested, err := flattenStringMap(v)
			if err != nil {
				return nil, err
			}
			for nk, nv := range nested {
				result[key+"_"+nk] = nv
			}
		default:
			return nil, fmt.Errorf("unsupported value type for key %s", key)
		}
	}
	return result, nil
}

func extractVaultData(raw any) (map[string]string, error) {
	if raw == nil {
		return nil, errors.New("vault data is empty")
	}

	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("vault data is not an object")
	}

	// KV v2 nests data twice (data.data).
	if nested, exists := obj["data"]; exists {
		return extractVaultData(nested)
	}

	flat := make(map[string]string)
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			flat[key] = v
		case fmt.Stringer:
			flat[key] = v.String()
		case float64, float32, int, int64, int32, uint64, uint32, bool:
			flat[key] = fmt.Sprint(v)
		default:
			return nil, fmt.Errorf("unsupported vault value type for key %s", key)
		}
	}

	if len(flat) == 0 {
		return nil, errors.New("vault payload did not contain key/value pairs")
	}

	return flat, nil
}

func mergeMaps(dst, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}
