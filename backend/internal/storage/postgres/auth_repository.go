package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/entities"
)

// AuthRepository provides persistence helpers for authentication flows.
type AuthRepository struct {
	pool *pgxpool.Pool
}

var ErrDuplicate = errors.New("duplicate record")

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	tx = nil
	return nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error) {
	const query = `
	INSERT INTO users (email, password_hash, status)
	VALUES ($1, $2, 'pending')
	RETURNING id, email, password_hash, status, email_verified_at, last_login_at, totp_secret, totp_enabled_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, email, passwordHash)
	return scanUser(row)
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	const query = `
    SELECT id, email, password_hash, status, email_verified_at, last_login_at, totp_secret, totp_enabled_at, created_at, updated_at
    FROM users
    WHERE email = $1`

	row := r.pool.QueryRow(ctx, query, email)
	return scanUser(row)
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	const query = `
    SELECT id, email, password_hash, status, email_verified_at, last_login_at, totp_secret, totp_enabled_at, created_at, updated_at
    FROM users
    WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (r *AuthRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	const query = `
	UPDATE users
	SET status = 'active', email_verified_at = $2, updated_at = NOW()
	WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID, verifiedAt)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	const query = `
	UPDATE users SET last_login_at = $2, updated_at = NOW() WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID, at)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	const query = `
	UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) UpsertTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	const query = `
	UPDATE users SET totp_secret = $2, totp_enabled_at = NULL, updated_at = NOW() WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID, secret)
	if err != nil {
		return fmt.Errorf("upsert totp secret: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) EnableTOTP(ctx context.Context, userID uuid.UUID, enabledAt time.Time) error {
	const query = `
	UPDATE users SET totp_enabled_at = $2, updated_at = NOW() WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID, enabledAt)
	if err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	const query = `
	UPDATE users SET totp_secret = NULL, totp_enabled_at = NULL, updated_at = NOW() WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("disable totp: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session entities.Session, userAgent, ipAddress string) (entities.Session, error) {
	const query = `
	INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, status, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, user_id, refresh_token_hash, status, expires_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, session.UserID, session.RefreshTokenHash, userAgent, ipAddress, session.Status, session.ExpiresAt)
	return scanSession(row)
}

func (r *AuthRepository) GetSessionByTokenHash(ctx context.Context, hash string) (entities.Session, error) {
	const query = `
	SELECT id, user_id, refresh_token_hash, status, expires_at, created_at, updated_at
	FROM sessions
	WHERE refresh_token_hash = $1`

	row := r.pool.QueryRow(ctx, query, hash)
	return scanSession(row)
}

func (r *AuthRepository) UpdateSession(ctx context.Context, sessionID uuid.UUID, newHash string, expiresAt time.Time) (entities.Session, error) {
	const query = `
	UPDATE sessions
	SET refresh_token_hash = $2, expires_at = $3, updated_at = NOW()
	WHERE id = $1 AND status = 'active'
	RETURNING id, user_id, refresh_token_hash, status, expires_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, sessionID, newHash, expiresAt)
	return scanSession(row)
}

func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	const query = `
	UPDATE sessions
	SET status = 'revoked', updated_at = NOW()
	WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	const query = `
	UPDATE sessions
	SET status = 'revoked', updated_at = NOW()
	WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

func (r *AuthRepository) CreateUserToken(ctx context.Context, token entities.UserToken) (entities.UserToken, error) {
	const query = `
INSERT INTO user_tokens (user_id, token_hash, token_type, expires_at, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token_hash, token_type, expires_at, consumed_at, created_at, metadata`

	row := r.pool.QueryRow(ctx, query, token.UserID, token.TokenHash, token.TokenType, token.ExpiresAt, token.Metadata)
	return scanUserToken(row)
}

func (r *AuthRepository) ConsumeUserToken(ctx context.Context, tokenID uuid.UUID) error {
	const query = `
	UPDATE user_tokens
	SET consumed_at = NOW()
	WHERE id = $1 AND consumed_at IS NULL`

	cmd, err := r.pool.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("consume user token: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AuthRepository) GetUserToken(ctx context.Context, tokenType, tokenHash string) (entities.UserToken, error) {
	const query = `
SELECT id, user_id, token_hash, token_type, expires_at, consumed_at, created_at, metadata
FROM user_tokens
WHERE token_type = $1 AND token_hash = $2`

	row := r.pool.QueryRow(ctx, query, tokenType, tokenHash)
	return scanUserToken(row)
}

func scanUser(row pgx.Row) (entities.User, error) {
	var (
		u      entities.User
		secret sql.NullString
	)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Status, &u.EmailVerifiedAt, &u.LastLoginAt, &secret, &u.TOTPEnabledAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return entities.User{}, translateError(err)
	}
	if secret.Valid {
		value := secret.String
		u.TOTPSecret = &value
	}
	return u, nil
}

func scanSession(row pgx.Row) (entities.Session, error) {
	var s entities.Session
	if err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.Status, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return entities.Session{}, translateError(err)
	}
	return s, nil
}

func scanUserToken(row pgx.Row) (entities.UserToken, error) {
	var (
		t        entities.UserToken
		metadata sql.NullString
	)
	if err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.TokenType, &t.ExpiresAt, &t.ConsumedAt, &t.CreatedAt, &metadata); err != nil {
		return entities.UserToken{}, translateError(err)
	}
	if metadata.Valid {
		value := metadata.String
		t.Metadata = &value
	}
	return t, nil
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicate
		}
	}
	return err
}
