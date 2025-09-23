package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// RegionsRepository provides access to regions and nodes tables.
type RegionsRepository struct {
	pool *pgxpool.Pool
}

func NewRegionsRepository(pool *pgxpool.Pool) *RegionsRepository {
	return &RegionsRepository{pool: pool}
}

func (r *RegionsRepository) UpsertRegion(ctx context.Context, region entities.Region) (entities.Region, error) {
	const query = `
	INSERT INTO regions (code, name, country_code, is_active)
	VALUES ($1,$2,$3,$4)
	ON CONFLICT (code)
	DO UPDATE SET
		name = EXCLUDED.name,
		country_code = EXCLUDED.country_code,
		is_active = EXCLUDED.is_active,
		updated_at = NOW()
	RETURNING id, code, name, country_code, is_active, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		region.Code,
		region.Name,
		region.CountryCode,
		region.IsActive,
	)

	return scanRegion(row)
}

func (r *RegionsRepository) ListRegionsWithCapacity(ctx context.Context) ([]entities.RegionCapacity, error) {
	const query = `
	SELECT r.id,
	       r.code,
	       r.name,
	       r.country_code,
	       r.is_active,
	       r.created_at,
	       r.updated_at,
	       COALESCE(AVG(CASE WHEN n.status = 'active' THEN n.capacity_score END), 0) AS capacity_score,
	       COALESCE(SUM(CASE WHEN n.status = 'active' THEN 1 ELSE 0 END), 0)      AS active_nodes
	FROM regions r
	LEFT JOIN nodes n ON n.region_id = r.id
	GROUP BY r.id
	ORDER BY r.code`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}
	defer rows.Close()

	var result []entities.RegionCapacity
	for rows.Next() {
		var (
			region      entities.Region
			capacity    sql.NullFloat64
			activeNodes int
		)

		if err := rows.Scan(
			&region.ID,
			&region.Code,
			&region.Name,
			&region.CountryCode,
			&region.IsActive,
			&region.CreatedAt,
			&region.UpdatedAt,
			&capacity,
			&activeNodes,
		); err != nil {
			return nil, err
		}

		result = append(result, entities.RegionCapacity{
			Region:        region,
			CapacityScore: capacity.Float64,
			ActiveNodes:   activeNodes,
		})
	}

	return result, rows.Err()
}

func (r *RegionsRepository) RegisterOrUpdateNode(ctx context.Context, node entities.Node) (entities.Node, error) {
	const query = `
	INSERT INTO nodes (region_id, hostname, public_ipv4, public_ipv6, public_key, endpoint, status, tunnel_port, capacity_score, last_seen_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	ON CONFLICT (hostname)
	DO UPDATE SET
		region_id = EXCLUDED.region_id,
		public_ipv4 = EXCLUDED.public_ipv4,
		public_ipv6 = EXCLUDED.public_ipv6,
		public_key = EXCLUDED.public_key,
		endpoint = EXCLUDED.endpoint,
		status = EXCLUDED.status,
		tunnel_port = EXCLUDED.tunnel_port,
		capacity_score = EXCLUDED.capacity_score,
		last_seen_at = EXCLUDED.last_seen_at,
		updated_at = NOW()
	RETURNING id, region_id, hostname, public_ipv4, public_ipv6, public_key, endpoint, status, capacity_score, tunnel_port, last_seen_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		node.RegionID,
		node.Hostname,
		node.PublicIPv4,
		node.PublicIPv6,
		node.PublicKey,
		node.Endpoint,
		node.Status,
		node.TunnelPort,
		node.CapacityScore,
		time.Now().UTC(),
	)

	return scanNode(row)
}

func (r *RegionsRepository) UpdateNodeHealth(ctx context.Context, nodeID uuid.UUID, capacityScore int) (entities.Node, error) {
	const query = `
	UPDATE nodes
	SET capacity_score = $2,
	    last_seen_at = NOW(),
	    updated_at = NOW()
	WHERE id = $1
	RETURNING id, region_id, hostname, public_ipv4, public_ipv6, public_key, endpoint, status, capacity_score, tunnel_port, last_seen_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, nodeID, capacityScore)
	return scanNode(row)
}

func (r *RegionsRepository) GetRegionByCode(ctx context.Context, code string) (entities.Region, error) {
	const query = `
	SELECT id, code, name, country_code, is_active, created_at, updated_at
	FROM regions
	WHERE code = $1`

	row := r.pool.QueryRow(ctx, query, code)
	return scanRegion(row)
}

func (r *RegionsRepository) GetNodeByID(ctx context.Context, id uuid.UUID) (entities.Node, error) {
	const query = `
	SELECT id, region_id, hostname, public_ipv4, public_ipv6, public_key, endpoint, status, capacity_score, tunnel_port, last_seen_at, created_at, updated_at
	FROM nodes
	WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	return scanNode(row)
}

func scanRegion(row pgx.Row) (entities.Region, error) {
	var region entities.Region
	if err := row.Scan(&region.ID, &region.Code, &region.Name, &region.CountryCode, &region.IsActive, &region.CreatedAt, &region.UpdatedAt); err != nil {
		return entities.Region{}, err
	}
	return region, nil
}

func scanNode(row pgx.Row) (entities.Node, error) {
	var (
		node       entities.Node
		ipv4, ipv6 sql.NullString
		lastSeen   sql.NullTime
	)

	if err := row.Scan(
		&node.ID,
		&node.RegionID,
		&node.Hostname,
		&ipv4,
		&ipv6,
		&node.PublicKey,
		&node.Endpoint,
		&node.Status,
		&node.CapacityScore,
		&node.TunnelPort,
		&lastSeen,
		&node.CreatedAt,
		&node.UpdatedAt,
	); err != nil {
		return entities.Node{}, err
	}

	if ipv4.Valid {
		value := ipv4.String
		node.PublicIPv4 = &value
	}
	if ipv6.Valid {
		value := ipv6.String
		node.PublicIPv6 = &value
	}
	if lastSeen.Valid {
		value := lastSeen.Time
		node.LastSeenAt = &value
	}

	return node, nil
}
