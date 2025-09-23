package entities

import "time"

// UsageSummary represents aggregated traffic stats for a user's peers.
type UsageSummary struct {
	TotalBytesTX    int64
	TotalBytesRX    int64
	PeerCount       int
	ActivePeerCount int
	LastHandshakeAt *time.Time
}

func (u UsageSummary) TotalBytes() int64 {
	return u.TotalBytesTX + u.TotalBytesRX
}
