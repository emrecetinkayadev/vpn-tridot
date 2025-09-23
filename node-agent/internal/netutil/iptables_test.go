package netutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyNATRules(t *testing.T) {
	var commands [][]string
	orig := runCommand
	runCommand = func(name string, args ...string) ([]byte, error) {
		cmd := append([]string{name}, args...)
		commands = append(commands, cmd)
		return nil, nil
	}
	t.Cleanup(func() { runCommand = orig })

	require.NoError(t, ApplyNATRules("wg0"))
	require.Len(t, commands, 3)
}

func TestEnableKillSwitch(t *testing.T) {
	var commands [][]string
	orig := runCommand
	runCommand = func(name string, args ...string) ([]byte, error) {
		commands = append(commands, append([]string{name}, args...))
		return nil, nil
	}
	t.Cleanup(func() { runCommand = orig })

	require.NoError(t, EnableKillSwitch("wg0"))
	require.Len(t, commands, 2)
}

func TestRunCommandsError(t *testing.T) {
	orig := runCommand
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errors.New("fail") }
	t.Cleanup(func() { runCommand = orig })

	require.Error(t, ApplyNATRules("wg0"))
}
