package entities

import (
	"time"

	"github.com/google/uuid"
)

type Region struct {
	ID          uuid.UUID
	Code        string
	Name        string
	CountryCode string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RegionCapacity struct {
	Region        Region
	CapacityScore float64
	ActiveNodes   int
}

type Node struct {
	ID            uuid.UUID
	RegionID      uuid.UUID
	Hostname      string
	PublicIPv4    *string
	PublicIPv6    *string
	PublicKey     string
	Endpoint      string
	Status        string
	CapacityScore int
	TunnelPort    int
	LastSeenAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NodeHealth struct {
	ActivePeers     int
	CPUPercent      float64
	ThroughputMbps  float64
	PacketLossRatio float64
}
