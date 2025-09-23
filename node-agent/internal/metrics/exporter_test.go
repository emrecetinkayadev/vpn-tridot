package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

func TestExporterUpdate(t *testing.T) {
	exp := New()

	originalNow := nowFunc
	nowFunc = func() time.Time { return time.Unix(100, 0) }
	t.Cleanup(func() { nowFunc = originalNow })

	exp.Update(wg.DeviceStats{PeerCount: 4, ActivePeers: 1, ReceiveBytes: 100, TransmitBytes: 200, LastHandshake: time.Unix(95, 0)})

	nowFunc = func() time.Time { return time.Unix(110, 0) }
	exp.Update(wg.DeviceStats{PeerCount: 4, ActivePeers: 2, ReceiveBytes: 150, TransmitBytes: 260, LastHandshake: time.Unix(120, 0)})

	rec := httptest.NewRecorder()
	exp.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "node_agent_wireguard_peers 4")
	require.Contains(t, body, "node_agent_wireguard_active_peers 2")
	require.Contains(t, body, "node_agent_wireguard_handshake_ratio 0.5")
	require.Contains(t, body, "node_agent_wireguard_last_handshake 120")
	require.Contains(t, body, "node_agent_wireguard_rx_bytes_total 150")
	require.Contains(t, body, "node_agent_wireguard_tx_bytes_total 260")
	require.Contains(t, body, "node_agent_wireguard_rx_throughput_bps 40")
	require.Contains(t, body, "node_agent_wireguard_tx_throughput_bps 48")
}
