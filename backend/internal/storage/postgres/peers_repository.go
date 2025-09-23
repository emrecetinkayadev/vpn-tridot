package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// PeersRepository handles CRUD for WireGuard peers.
type PeersRepository struct {
	pool *pgxpool.Pool
}

func NewPeersRepository(pool *pgxpool.Pool) *PeersRepository {
	return &PeersRepository{pool: pool}
}

func (r *PeersRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Peer, error) {
	const query = `
	SELECT id, user_id, node_id, region_id, device_name, public_key, preshared_key,
	       allowed_ips, dns_servers, keepalive, mtu, status, created_at, updated_at,
	       last_handshake_at, bytes_tx, bytes_rx
	FROM peers
	WHERE user_id = $1
	ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list peers: %w", err)
	}
	defer rows.Close()

	var peers []entities.Peer
	for rows.Next() {
		peer, err := scanPeer(rows)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	return peers, rows.Err()
}

func (r *PeersRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM peers WHERE user_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count peers: %w", err)
	}
	return count, nil
}

func (r *PeersRepository) Create(ctx context.Context, peer entities.Peer) (entities.Peer, error) {
	const query = `
	INSERT INTO peers (
		user_id, node_id, region_id, device_name, public_key, preshared_key,
		allowed_ips, dns_servers, keepalive, mtu, status, bytes_tx, bytes_rx
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	RETURNING id, user_id, node_id, region_id, device_name, public_key, preshared_key,
	          allowed_ips, dns_servers, keepalive, mtu, status, created_at, updated_at,
	          last_handshake_at, bytes_tx, bytes_rx`

	dns := pgStringArray(peer.DNSServers)
	row := r.pool.QueryRow(ctx, query,
		peer.UserID,
		peer.NodeID,
		peer.RegionID,
		peer.DeviceName,
		peer.PublicKey,
		peer.PresharedKey,
		peer.AllowedIPs,
		dns,
		peer.Keepalive,
		peer.MTU,
		peer.Status,
		peer.BytesTX,
		peer.BytesRX,
	)

	return scanPeer(row)
}

func (r *PeersRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entities.Peer, error) {
	const query = `
	SELECT id, user_id, node_id, region_id, device_name, public_key, preshared_key,
	       allowed_ips, dns_servers, keepalive, mtu, status, created_at, updated_at,
	       last_handshake_at, bytes_tx, bytes_rx
	FROM peers
	WHERE id = $1 AND user_id = $2`

	row := r.pool.QueryRow(ctx, query, id, userID)
	return scanPeer(row)
}

func (r *PeersRepository) Rename(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (entities.Peer, error) {
	const query = `
	UPDATE peers
	SET device_name = $3, updated_at = NOW()
	WHERE id = $1 AND user_id = $2
	RETURNING id, user_id, node_id, region_id, device_name, public_key, preshared_key,
	          allowed_ips, dns_servers, keepalive, mtu, status, created_at, updated_at,
	          last_handshake_at, bytes_tx, bytes_rx`

	row := r.pool.QueryRow(ctx, query, id, userID, name)
	return scanPeer(row)
}

func (r *PeersRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const query = `DELETE FROM peers WHERE id = $1 AND user_id = $2`

	cmd, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete peer: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PeersRepository) UsageSummaryByUser(ctx context.Context, userID uuid.UUID) (entities.UsageSummary, error) {
	const query = `
	SELECT
	       COALESCE(SUM(bytes_tx), 0) AS total_tx,
	       COALESCE(SUM(bytes_rx), 0) AS total_rx,
	       COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_count,
	       COUNT(*) AS peer_count,
	       MAX(last_handshake_at) AS last_handshake
	FROM peers
	WHERE user_id = $1`

	var (
		totalTX   int64
		totalRX   int64
		activeCnt int
		peerCount int
		lastSeen  sql.NullTime
	)

	if err := r.pool.QueryRow(ctx, query, userID).Scan(&totalTX, &totalRX, &activeCnt, &peerCount, &lastSeen); err != nil {
		return entities.UsageSummary{}, fmt.Errorf("usage summary: %w", err)
	}

	summary := entities.UsageSummary{
		TotalBytesTX:    totalTX,
		TotalBytesRX:    totalRX,
		PeerCount:       peerCount,
		ActivePeerCount: activeCnt,
	}

	if lastSeen.Valid {
		value := lastSeen.Time
		summary.LastHandshakeAt = &value
	}

	return summary, nil
}

func scanPeer(row pgx.Row) (entities.Peer, error) {
	var (
		peer       entities.Peer
		preshared  sql.NullString
		dnsServers []string
		keepalive  sql.NullInt32
		mtu        sql.NullInt32
		lastSeen   sql.NullTime
	)

	if err := row.Scan(
		&peer.ID,
		&peer.UserID,
		&peer.NodeID,
		&peer.RegionID,
		&peer.DeviceName,
		&peer.PublicKey,
		&preshared,
		&peer.AllowedIPs,
		&dnsServers,
		&keepalive,
		&mtu,
		&peer.Status,
		&peer.CreatedAt,
		&peer.UpdatedAt,
		&lastSeen,
		&peer.BytesTX,
		&peer.BytesRX,
	); err != nil {
		return entities.Peer{}, err
	}

	if preshared.Valid {
		val := preshared.String
		peer.PresharedKey = &val
	}
	peer.DNSServers = dnsServers
	if keepalive.Valid {
		val := int(keepalive.Int32)
		peer.Keepalive = &val
	}
	if mtu.Valid {
		val := int(mtu.Int32)
		peer.MTU = &val
	}
	if lastSeen.Valid {
		val := lastSeen.Time
		peer.LastHandshakeAt = &val
	}

	return peer, nil
}

func pgStringArray(values []string) interface{} {
	if len(values) == 0 {
		return []string{}
	}
	return values
}

var ErrDuplicatePeer = errors.New("peer already exists")
