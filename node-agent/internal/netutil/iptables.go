package netutil

import (
	"fmt"
	"os/exec"
)

var runCommand = func(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// WithCommandRunner overrides the command runner for testing.
func WithCommandRunner(fn func(string, ...string) ([]byte, error)) (restore func()) {
	prev := runCommand
	runCommand = fn
	return func() { runCommand = prev }

}

// ApplyNATRules configures basic NAT masquerade and forwarding rules for an interface.
func ApplyNATRules(iface string) error {
	if iface == "" {
		return fmt.Errorf("iface required")
	}
	commands := [][]string{
		{"iptables", "-t", "nat", "-A", "POSTROUTING", "-o", iface, "-j", "MASQUERADE"},
		{"iptables", "-A", "FORWARD", "-i", iface, "-j", "ACCEPT"},
		{"iptables", "-A", "FORWARD", "-o", iface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"},
	}
	return runCommands(commands)
}

// EnableKillSwitch drops new outbound connections not using the WireGuard interface.
func EnableKillSwitch(iface string) error {
	if iface == "" {
		return fmt.Errorf("iface required")
	}
	commands := [][]string{
		{"iptables", "-A", "OUTPUT", "!", "-o", iface, "-m", "conntrack", "--ctstate", "NEW", "-j", "DROP"},
		{"iptables", "-A", "INPUT", "!", "-i", iface, "-m", "conntrack", "--ctstate", "NEW", "-j", "DROP"},
	}
	return runCommands(commands)
}

func runCommands(cmds [][]string) error {
	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		if _, err := runCommand(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("command %v failed: %w", cmd, err)
		}
	}
	return nil
}
