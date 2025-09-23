package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents runtime configuration for the node agent.
type Config struct {
	ControlPlane ControlPlaneConfig `yaml:"controlPlane"`
	Provision    ProvisionConfig    `yaml:"provision"`
	MTLS         MTLSConfig         `yaml:"mtls"`
	Agent        AgentConfig        `yaml:"agent"`
	WireGuard    WireGuardConfig    `yaml:"wireguard"`
}

type ControlPlaneConfig struct {
	URL string `yaml:"url"
 json:"url"`
	RegisterPath string        `yaml:"registerPath" json:"register_path"`
	HealthPath   string        `yaml:"healthPath" json:"health_path"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`
}

type ProvisionConfig struct {
	Token string `yaml:"token" json:"token"`
}

type MTLSConfig struct {
	CACert     string `yaml:"caPEM" json:"ca_pem"`
	CACertFile string `yaml:"caFile" json:"ca_file"`
	Cert       string `yaml:"certPEM" json:"cert_pem"`
	CertFile   string `yaml:"certFile" json:"cert_file"`
	Key        string `yaml:"keyPEM" json:"key_pem"`
	KeyFile    string `yaml:"keyFile" json:"key_file"`
}

type AgentConfig struct {
	PollInterval     time.Duration `yaml:"pollInterval" json:"poll_interval"`
	MetricsAddress   string        `yaml:"metricsAddr" json:"metrics_addr"`
	StateDirectory   string        `yaml:"stateDir" json:"state_dir"`
	MaxRetryInterval time.Duration `yaml:"maxRetryInterval" json:"max_retry_interval"`
}

type WireGuardConfig struct {
	InterfaceName       string   `yaml:"interfaceName" json:"interface_name"`
	ListenPort          int      `yaml:"listenPort" json:"listen_port"`
	AddressCIDR         string   `yaml:"address" json:"address"`
	DNS                 []string `yaml:"dns" json:"dns"`
	MTU                 int      `yaml:"mtu" json:"mtu"`
	PersistentKeepalive int      `yaml:"persistentKeepalive" json:"persistent_keepalive"`
	ConfigDirectory     string   `yaml:"configDir" json:"config_dir"`
	EnableNAT           bool     `yaml:"enableNAT" json:"enable_nat"`
	EnableKillSwitch    bool     `yaml:"enableKillSwitch" json:"enable_kill_switch"`
}

// Load reads configuration from YAML file (optional) and environment variables.
func Load() (Config, error) {
	var cfg Config
	cfg.Agent.PollInterval = 30 * time.Second
	cfg.Agent.MetricsAddress = ":9102"
	cfg.Agent.StateDirectory = "/var/lib/vpn-agent"
	cfg.Agent.MaxRetryInterval = 2 * time.Minute
	cfg.ControlPlane.Timeout = 10 * time.Second
	cfg.ControlPlane.RegisterPath = "/api/v1/nodes/register"
	cfg.ControlPlane.HealthPath = "/api/v1/nodes/health"
	cfg.WireGuard.InterfaceName = "wg0"
	cfg.WireGuard.ListenPort = 51820
	cfg.WireGuard.ConfigDirectory = "/etc/wireguard"
	cfg.WireGuard.PersistentKeepalive = 25

	if path := os.Getenv("NODE_AGENT_CONFIG_FILE"); path != "" {
		fileCfg, err := fromYAML(path)
		if err != nil {
			return Config{}, err
		}
		cfg = merge(cfg, fileCfg)
	}

	overlayEnv(&cfg)

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func fromYAML(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}
	dir := filepath.Dir(path)
	if cfg.MTLS.CACertFile != "" && !filepath.IsAbs(cfg.MTLS.CACertFile) {
		cfg.MTLS.CACertFile = filepath.Join(dir, cfg.MTLS.CACertFile)
	}
	if cfg.MTLS.CertFile != "" && !filepath.IsAbs(cfg.MTLS.CertFile) {
		cfg.MTLS.CertFile = filepath.Join(dir, cfg.MTLS.CertFile)
	}
	if cfg.MTLS.KeyFile != "" && !filepath.IsAbs(cfg.MTLS.KeyFile) {
		cfg.MTLS.KeyFile = filepath.Join(dir, cfg.MTLS.KeyFile)
	}
	if cfg.WireGuard.ConfigDirectory != "" && !filepath.IsAbs(cfg.WireGuard.ConfigDirectory) {
		cfg.WireGuard.ConfigDirectory = filepath.Join(dir, cfg.WireGuard.ConfigDirectory)
	}
	if cfg.Agent.StateDirectory != "" && !filepath.IsAbs(cfg.Agent.StateDirectory) {
		cfg.Agent.StateDirectory = filepath.Join(dir, cfg.Agent.StateDirectory)
	}
	return cfg, nil
}

func overlayEnv(cfg *Config) {
	if v := os.Getenv("CONTROL_PLANE_URL"); v != "" {
		cfg.ControlPlane.URL = v
	}
	if v := os.Getenv("CONTROL_PLANE_REGISTER_PATH"); v != "" {
		cfg.ControlPlane.RegisterPath = v
	}
	if v := os.Getenv("CONTROL_PLANE_HEALTH_PATH"); v != "" {
		cfg.ControlPlane.HealthPath = v
	}
	if v := os.Getenv("CONTROL_PLANE_TIMEOUT"); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			cfg.ControlPlane.Timeout = dur
		}
	}

	if v := os.Getenv("NODE_PROVISION_TOKEN"); v != "" {
		cfg.Provision.Token = v
	}

	if v := os.Getenv("MTLS_CA_PEM"); v != "" {
		cfg.MTLS.CACert = v
	}
	if v := os.Getenv("MTLS_CA_FILE"); v != "" {
		cfg.MTLS.CACertFile = v
	}
	if v := os.Getenv("MTLS_CLIENT_CERT"); v != "" {
		cfg.MTLS.Cert = v
	}
	if v := os.Getenv("MTLS_CLIENT_CERT_FILE"); v != "" {
		cfg.MTLS.CertFile = v
	}
	if v := os.Getenv("MTLS_CLIENT_KEY"); v != "" {
		cfg.MTLS.Key = v
	}
	if v := os.Getenv("MTLS_CLIENT_KEY_FILE"); v != "" {
		cfg.MTLS.KeyFile = v
	}

	if v := os.Getenv("AGENT_POLL_INTERVAL"); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			cfg.Agent.PollInterval = dur
		}
	}
	if v := os.Getenv("AGENT_METRICS_ADDR"); v != "" {
		cfg.Agent.MetricsAddress = v
	}
	if v := os.Getenv("AGENT_STATE_DIR"); v != "" {
		cfg.Agent.StateDirectory = v
	}
	if v := os.Getenv("AGENT_MAX_RETRY_INTERVAL"); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			cfg.Agent.MaxRetryInterval = dur
		}
	}

	if v := os.Getenv("WG_INTERFACE"); v != "" {
		cfg.WireGuard.InterfaceName = v
	}
	if v := os.Getenv("WG_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.ListenPort = port
		}
	}
	if v := os.Getenv("WG_ADDRESS"); v != "" {
		cfg.WireGuard.AddressCIDR = v
	}
	if v := os.Getenv("WG_DNS"); v != "" {
		cfg.WireGuard.DNS = strings.Split(v, ",")
	}
	if v := os.Getenv("WG_MTU"); v != "" {
		if mtu, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.MTU = mtu
		}
	}
	if v := os.Getenv("WG_KEEPALIVE"); v != "" {
		if ka, err := strconv.Atoi(v); err == nil {
			cfg.WireGuard.PersistentKeepalive = ka
		}
	}
	if v := os.Getenv("WG_CONFIG_DIR"); v != "" {
		cfg.WireGuard.ConfigDirectory = v
	}
	if v := os.Getenv("WG_ENABLE_NAT"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.WireGuard.EnableNAT = b
		}
	}
	if v := os.Getenv("WG_ENABLE_KILLSWITCH"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.WireGuard.EnableKillSwitch = b
		}
	}
}

func validate(cfg Config) error {
	if cfg.ControlPlane.URL == "" {
		return errors.New("control plane url is required")
	}
	if cfg.Provision.Token == "" {
		return errors.New("provision token is required")
	}
	if cfg.MTLS.CACert == "" && cfg.MTLS.CACertFile == "" {
		return errors.New("mtls ca cert or file required")
	}
	if cfg.MTLS.Cert == "" && cfg.MTLS.CertFile == "" {
		return errors.New("mtls client cert required")
	}
	if cfg.MTLS.Key == "" && cfg.MTLS.KeyFile == "" {
		return errors.New("mtls client key required")
	}
	if cfg.Agent.PollInterval <= 0 {
		return errors.New("poll interval must be greater than zero")
	}
	if cfg.Agent.MetricsAddress == "" {
		return errors.New("agent metrics address required")
	}
	if cfg.Agent.StateDirectory == "" {
		return errors.New("agent state directory required")
	}
	if cfg.Agent.MaxRetryInterval <= 0 {
		return errors.New("agent max retry interval must be greater than zero")
	}
	if cfg.WireGuard.InterfaceName == "" {
		return errors.New("wireguard interface name required")
	}
	if cfg.WireGuard.ListenPort <= 0 || cfg.WireGuard.ListenPort > 65535 {
		return errors.New("wireguard listen port invalid")
	}
	return nil
}

func merge(base, override Config) Config {
	cfg := base
	if override.ControlPlane.URL != "" {
		cfg.ControlPlane.URL = override.ControlPlane.URL
	}
	if override.ControlPlane.RegisterPath != "" {
		cfg.ControlPlane.RegisterPath = override.ControlPlane.RegisterPath
	}
	if override.ControlPlane.HealthPath != "" {
		cfg.ControlPlane.HealthPath = override.ControlPlane.HealthPath
	}
	if override.ControlPlane.Timeout != 0 {
		cfg.ControlPlane.Timeout = override.ControlPlane.Timeout
	}
	if override.Provision.Token != "" {
		cfg.Provision.Token = override.Provision.Token
	}
	if override.MTLS.CACert != "" {
		cfg.MTLS.CACert = override.MTLS.CACert
	}
	if override.MTLS.CACertFile != "" {
		cfg.MTLS.CACertFile = override.MTLS.CACertFile
	}
	if override.MTLS.Cert != "" {
		cfg.MTLS.Cert = override.MTLS.Cert
	}
	if override.MTLS.CertFile != "" {
		cfg.MTLS.CertFile = override.MTLS.CertFile
	}
	if override.MTLS.Key != "" {
		cfg.MTLS.Key = override.MTLS.Key
	}
	if override.MTLS.KeyFile != "" {
		cfg.MTLS.KeyFile = override.MTLS.KeyFile
	}
	if override.Agent.PollInterval != 0 {
		cfg.Agent.PollInterval = override.Agent.PollInterval
	}
	if override.Agent.MetricsAddress != "" {
		cfg.Agent.MetricsAddress = override.Agent.MetricsAddress
	}
	if override.Agent.StateDirectory != "" {
		cfg.Agent.StateDirectory = override.Agent.StateDirectory
	}
	if override.Agent.MaxRetryInterval != 0 {
		cfg.Agent.MaxRetryInterval = override.Agent.MaxRetryInterval
	}
	if override.WireGuard.InterfaceName != "" {
		cfg.WireGuard.InterfaceName = override.WireGuard.InterfaceName
	}
	if override.WireGuard.ListenPort != 0 {
		cfg.WireGuard.ListenPort = override.WireGuard.ListenPort
	}
	if override.WireGuard.AddressCIDR != "" {
		cfg.WireGuard.AddressCIDR = override.WireGuard.AddressCIDR
	}
	if len(override.WireGuard.DNS) > 0 {
		cfg.WireGuard.DNS = override.WireGuard.DNS
	}
	if override.WireGuard.MTU != 0 {
		cfg.WireGuard.MTU = override.WireGuard.MTU
	}
	if override.WireGuard.PersistentKeepalive != 0 {
		cfg.WireGuard.PersistentKeepalive = override.WireGuard.PersistentKeepalive
	}
	if override.WireGuard.ConfigDirectory != "" {
		cfg.WireGuard.ConfigDirectory = override.WireGuard.ConfigDirectory
	}
	if override.WireGuard.EnableNAT {
		cfg.WireGuard.EnableNAT = true
	}
	if override.WireGuard.EnableKillSwitch {
		cfg.WireGuard.EnableKillSwitch = true
	}
	return cfg
}
