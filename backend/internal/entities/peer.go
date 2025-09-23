package entities

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Peer struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	NodeID          uuid.UUID
	RegionID        uuid.UUID
	DeviceName      string
	PublicKey       string
	PresharedKey    *string
	AllowedIPs      string
	DNSServers      []string
	Keepalive       *int
	MTU             *int
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastHandshakeAt *time.Time
	BytesTX         int64
	BytesRX         int64
}

func (p Peer) IsActive() bool {
	return stringsEqualFold(p.Status, "active")
}

func stringsEqualFold(a, b string) bool {
	return strings.EqualFold(a, b)
}
