package wg

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var activePeerWindow = 3 * time.Minute

type defaultClient struct{}

func (defaultClient) DeviceStats(interfaceName string) (DeviceStats, error) {
	if interfaceName == "" {
		return DeviceStats{}, fmt.Errorf("interface name required")
	}
	out, err := exec.Command("wg", "show", interfaceName, "dump").Output()
	if err != nil {
		return DeviceStats{}, fmt.Errorf("wg show dump: %w", err)
	}
	stats := DeviceStats{}
	now := time.Now()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue // interface line
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}
		handshake, _ := strconv.ParseInt(fields[5], 10, 64)
		rx, _ := strconv.ParseUint(fields[6], 10, 64)
		tx, _ := strconv.ParseUint(fields[7], 10, 64)
		stats.PeerCount++
		stats.ReceiveBytes += rx
		stats.TransmitBytes += tx
		if handshake > 0 {
			t := time.Unix(handshake, 0)
			if t.After(stats.LastHandshake) {
				stats.LastHandshake = t
			}
			if now.Sub(t) <= activePeerWindow {
				stats.ActivePeers++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return DeviceStats{}, err
	}
	return stats, nil
}
