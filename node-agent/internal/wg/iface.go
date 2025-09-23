package wg

import (
	"fmt"
	"os/exec"
)

// SetupInterface ensures the WireGuard interface is created using wg-quick.
func SetupInterface(configPath string) error {
	cmd := exec.Command("wg-quick", "up", configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wg-quick up failed: %w: %s", err, string(output))
	}
	return nil
}

// TeardownInterface brings down a WireGuard interface via wg-quick.
func TeardownInterface(configPath string) error {
	cmd := exec.Command("wg-quick", "down", configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wg-quick down failed: %w: %s", err, string(output))
	}
	return nil
}

// SyncPeers reloads the configuration to apply peer changes.
func SyncPeers(interfaceName, configPath string) error {
	cmd := exec.Command("wg", "syncconf", interfaceName, configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wg syncconf failed: %w: %s", err, string(output))
	}
	return nil
}
