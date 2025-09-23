package regions

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// Repository describes persistence operations used by the regions service.
type Repository interface {
	UpsertRegion(ctx context.Context, region entities.Region) (entities.Region, error)
	ListRegionsWithCapacity(ctx context.Context) ([]entities.RegionCapacity, error)
	RegisterOrUpdateNode(ctx context.Context, node entities.Node) (entities.Node, error)
	UpdateNodeHealth(ctx context.Context, nodeID uuid.UUID, capacityScore int) (entities.Node, error)
	GetRegionByCode(ctx context.Context, code string) (entities.Region, error)
	GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error)
}

// Service provides region listing and node orchestration.
type Service struct {
	repo Repository
	cfg  config.Config
}

func NewService(repo Repository, cfg config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// SeedDefaultRegions ensures predefined regions exist.
func (s *Service) SeedDefaultRegions(ctx context.Context) error {
	defaults := []entities.Region{
		{Code: "TR-IST", Name: "İstanbul", CountryCode: "TR", IsActive: true},
		{Code: "TR-IZM", Name: "İzmir", CountryCode: "TR", IsActive: true},
		{Code: "EU-FRA", Name: "Frankfurt", CountryCode: "DE", IsActive: true},
		{Code: "EU-NL", Name: "Amsterdam", CountryCode: "NL", IsActive: true},
	}

	for _, region := range defaults {
		if _, err := s.repo.UpsertRegion(ctx, region); err != nil {
			return fmt.Errorf("seed region %s: %w", region.Code, err)
		}
	}

	return nil
}

// List returns all regions with aggregated capacity information.
func (s *Service) List(ctx context.Context) ([]entities.RegionCapacity, error) {
	return s.repo.ListRegionsWithCapacity(ctx)
}

// RegisterNode registers or updates a node associated with a region code.
type RegisterNodeInput struct {
	RegionCode string
	Hostname   string
	PublicIPv4 *string
	PublicIPv6 *string
	PublicKey  string
	Endpoint   string
	TunnelPort int
}

func (s *Service) RegisterNode(ctx context.Context, input RegisterNodeInput) (entities.Node, error) {
	if input.RegionCode == "" || input.Hostname == "" {
		return entities.Node{}, errors.New("region code and hostname are required")
	}

	region, err := s.repo.GetRegionByCode(ctx, strings.ToUpper(input.RegionCode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Node{}, fmt.Errorf("region %s not found", input.RegionCode)
		}
		return entities.Node{}, err
	}

	node := entities.Node{
		RegionID:      region.ID,
		Hostname:      input.Hostname,
		PublicIPv4:    input.PublicIPv4,
		PublicIPv6:    input.PublicIPv6,
		PublicKey:     input.PublicKey,
		Endpoint:      input.Endpoint,
		Status:        "active",
		TunnelPort:    input.TunnelPort,
		CapacityScore: 100,
	}

	return s.repo.RegisterOrUpdateNode(ctx, node)
}

// ReportHealth updates node health metrics and recalculates capacity score.
type HealthReportInput struct {
	NodeID         uuid.UUID
	ActivePeers    int
	CPUPercent     float64
	ThroughputMbps float64
	PacketLoss     float64
}

func (s *Service) ReportHealth(ctx context.Context, input HealthReportInput) (entities.Node, error) {
	if input.NodeID == uuid.Nil {
		return entities.Node{}, errors.New("node id is required")
	}

	score := computeCapacityScore(input)
	return s.repo.UpdateNodeHealth(ctx, input.NodeID, score)
}

// GetNodeByID exposes node metadata for other services.
func (s *Service) GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error) {
	return s.repo.GetNodeByID(ctx, id)
}

func computeCapacityScore(health HealthReportInput) int {
	score := 100
	score -= minInt(60, health.ActivePeers*4)
	score -= int(math.Round(health.CPUPercent / 2))
	score -= int(math.Round(health.ThroughputMbps / 100))
	score -= int(math.Round(health.PacketLoss * 50))

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
