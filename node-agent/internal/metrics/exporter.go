package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/wg"
)

// nowFunc allows tests to control the perceived current time.
var nowFunc = time.Now

// Exporter exposes WireGuard metrics for Prometheus scraping.
type Exporter struct {
	mu          sync.Mutex
	registry    *prometheus.Registry
	handler     http.Handler
	peerGauge   prometheus.Gauge
	activeGauge prometheus.Gauge
	ratioGauge  prometheus.Gauge
	rxCounter   prometheus.Counter
	txCounter   prometheus.Counter
	rxBpsGauge  prometheus.Gauge
	txBpsGauge  prometheus.Gauge
	handshake   prometheus.Gauge

	lastRx     uint64
	lastTx     uint64
	lastSample time.Time
}

// New creates a metrics exporter with in-memory registry.
func New() *Exporter {
	r := prometheus.NewRegistry()
	exp := &Exporter{
		registry:    r,
		peerGauge:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_peers", Help: "Current WireGuard peer count"}),
		activeGauge: prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_active_peers", Help: "Peers with a recent handshake"}),
		ratioGauge:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_handshake_ratio", Help: "Active peers / total peers"}),
		rxCounter:   prometheus.NewCounter(prometheus.CounterOpts{Name: "node_agent_wireguard_rx_bytes_total", Help: "Cumulative received bytes"}),
		txCounter:   prometheus.NewCounter(prometheus.CounterOpts{Name: "node_agent_wireguard_tx_bytes_total", Help: "Cumulative transmitted bytes"}),
		rxBpsGauge:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_rx_throughput_bps", Help: "Receive throughput in bits per second"}),
		txBpsGauge:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_tx_throughput_bps", Help: "Transmit throughput in bits per second"}),
		handshake:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "node_agent_wireguard_last_handshake", Help: "Timestamp of the latest peer handshake"}),
	}

	r.MustRegister(exp.peerGauge, exp.activeGauge, exp.ratioGauge, exp.rxCounter, exp.txCounter, exp.rxBpsGauge, exp.txBpsGauge, exp.handshake)
	exp.handler = promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	return exp
}

// Update records the latest WireGuard statistics.
func (e *Exporter) Update(stats wg.DeviceStats) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := nowFunc()

	e.peerGauge.Set(float64(stats.PeerCount))
	e.activeGauge.Set(float64(stats.ActivePeers))
	if stats.PeerCount > 0 {
		e.ratioGauge.Set(float64(stats.ActivePeers) / float64(stats.PeerCount))
	} else {
		e.ratioGauge.Set(0)
	}
	if stats.LastHandshake.IsZero() {
		e.handshake.Set(0)
	} else {
		e.handshake.Set(float64(stats.LastHandshake.Unix()))
	}

	e.updateByteSeries(now, stats.ReceiveBytes, &e.lastRx, e.rxCounter, e.rxBpsGauge)
	e.updateByteSeries(now, stats.TransmitBytes, &e.lastTx, e.txCounter, e.txBpsGauge)
	e.lastSample = now
}

func (e *Exporter) updateByteSeries(now time.Time, current uint64, last *uint64, counter prometheus.Counter, throughput prometheus.Gauge) {
	if *last == 0 {
		counter.Add(float64(current))
		throughput.Set(0)
		*last = current
		return
	}

	if current < *last {
		counter.Add(float64(current))
		throughput.Set(0)
		*last = current
		return
	}

	delta := current - *last
	if delta > 0 {
		counter.Add(float64(delta))
	}
	if !e.lastSample.IsZero() {
		elapsed := now.Sub(e.lastSample).Seconds()
		if elapsed > 0 {
			throughput.Set(float64(delta) * 8 / elapsed)
		} else {
			throughput.Set(0)
		}
	} else {
		throughput.Set(0)
	}
	*last = current
}

// Handler returns an HTTP handler for Prometheus scraping.
func (e *Exporter) Handler() http.Handler {
	return e.handler
}
