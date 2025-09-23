package main

import (
	"fmt"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/agent"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/metrics"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/netutil"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/state"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/transport"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

func newAgent(cfg config.Config) (*agent.Agent, *metrics.Exporter, error) {
	client, err := transport.NewMTLSClient(cfg.MTLS, cfg.ControlPlane.Timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("init mtls client: %w", err)
	}

	if cfg.Agent.PollInterval <= 0 {
		cfg.Agent.PollInterval = 30 * time.Second
	}

	wgManager := wg.NewManager(cfg.WireGuard)
	configPath, err := wgManager.EnsureBaseConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("write wireguard config: %w", err)
	}
	if cfg.WireGuard.EnableNAT {
		if err := netutil.ApplyNATRules(cfg.WireGuard.InterfaceName); err != nil {
			return nil, nil, fmt.Errorf("apply nat rules: %w", err)
		}
	}
	if cfg.WireGuard.EnableKillSwitch {
		if err := netutil.EnableKillSwitch(cfg.WireGuard.InterfaceName); err != nil {
			return nil, nil, fmt.Errorf("enable kill switch: %w", err)
		}
	}

	ag, err := agent.New(cfg, client)
	if err != nil {
		return nil, nil, err
	}
	stateStore, err := state.New(cfg.Agent.StateDirectory)
	if err != nil {
		return nil, nil, fmt.Errorf("init state store: %w", err)
	}
	ag.WithState(stateStore)
	ag.WithWireGuard(wgManager, configPath, wg.SetupInterface, func(path string) error {
		return wg.SyncPeers(cfg.WireGuard.InterfaceName, path)
	})
	exporter := metrics.New()
	ag.WithMetrics(exporter)
	if peers, err := stateStore.LoadPeers(); err != nil {
		return nil, nil, fmt.Errorf("load persisted peers: %w", err)
	} else if len(peers) > 0 {
		if err := ag.ApplyPeers(peers); err != nil {
			return nil, nil, fmt.Errorf("restore peers: %w", err)
		}
	}

	return ag, exporter, nil
}
