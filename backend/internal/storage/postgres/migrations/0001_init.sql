-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE account_status AS ENUM ('pending', 'active', 'suspended', 'disabled');
CREATE TYPE subscription_status AS ENUM ('trialing', 'active', 'past_due', 'canceled');
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed', 'refunded');
CREATE TYPE payment_provider AS ENUM ('stripe', 'iyzico');
CREATE TYPE node_status AS ENUM ('active', 'draining', 'disabled');
CREATE TYPE peer_status AS ENUM ('active', 'revoked');
CREATE TYPE session_status AS ENUM ('active', 'revoked');
CREATE TYPE user_token_type AS ENUM ('email_verification', 'password_reset');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    status          account_status NOT NULL DEFAULT 'pending',
    email_verified_at TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    totp_secret     TEXT,
    totp_enabled_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    description     TEXT,
    price_cents     INTEGER NOT NULL CHECK (price_cents >= 0),
    currency        CHAR(3) NOT NULL,
    billing_period  TEXT NOT NULL CHECK (billing_period IN ('month', 'quarter', 'year')),
    interval_count  SMALLINT NOT NULL DEFAULT 1 CHECK (interval_count > 0),
    device_limit    SMALLINT NOT NULL DEFAULT 5 CHECK (device_limit > 0),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id             UUID NOT NULL REFERENCES plans(id),
    status              subscription_status NOT NULL DEFAULT 'trialing',
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end  TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    canceled_at         TIMESTAMPTZ,
    provider            payment_provider NOT NULL DEFAULT 'stripe',
    provider_customer_id TEXT,
    provider_subscription_id TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    provider        payment_provider NOT NULL,
    provider_payment_id TEXT NOT NULL,
    status          payment_status NOT NULL,
    amount_cents    INTEGER NOT NULL CHECK (amount_cents >= 0),
    currency        CHAR(3) NOT NULL,
    paid_at         TIMESTAMPTZ,
    refunded_at     TIMESTAMPTZ,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_payment_id)
);

CREATE TABLE regions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    country_code    CHAR(2) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nodes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id       UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    hostname        TEXT NOT NULL UNIQUE,
    public_ipv4     INET,
    public_ipv6     INET,
    public_key      TEXT NOT NULL,
    endpoint        TEXT NOT NULL,
    status          node_status NOT NULL DEFAULT 'active',
    capacity_score  INTEGER NOT NULL DEFAULT 0,
    tunnel_port     INTEGER NOT NULL CHECK (tunnel_port BETWEEN 1 AND 65535),
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE peers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    node_id         UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    region_id       UUID NOT NULL REFERENCES regions(id),
    device_name     TEXT NOT NULL,
    public_key      TEXT NOT NULL UNIQUE,
    preshared_key   TEXT,
    allowed_ips     TEXT NOT NULL,
    dns_servers     TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    keepalive       INTEGER CHECK (keepalive IS NULL OR keepalive > 0),
    mtu             INTEGER CHECK (mtu IS NULL OR mtu > 0),
    status          peer_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_handshake_at TIMESTAMPTZ,
    bytes_tx        BIGINT NOT NULL DEFAULT 0 CHECK (bytes_tx >= 0),
    bytes_rx        BIGINT NOT NULL DEFAULT 0 CHECK (bytes_rx >= 0),
    UNIQUE (user_id, device_name)
);

CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent      TEXT,
    ip_address      INET,
    status          session_status NOT NULL DEFAULT 'active',
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (refresh_token_hash)
);

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_type      TEXT NOT NULL,
    action          TEXT NOT NULL,
    target_type     TEXT,
    target_id       TEXT,
    metadata        JSONB,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    token_type      user_token_type NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (token_type, token_hash)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_users_email ON users USING BTREE (email);
CREATE INDEX idx_subscriptions_user ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_plan ON subscriptions (plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);
CREATE UNIQUE INDEX uq_subscriptions_user_active ON subscriptions (user_id) WHERE status IN ('trialing', 'active');
CREATE UNIQUE INDEX uq_subscriptions_provider_id ON subscriptions (provider, provider_subscription_id) WHERE provider_subscription_id IS NOT NULL;
CREATE INDEX idx_payments_subscription ON payments (subscription_id);
CREATE INDEX idx_regions_active ON regions (is_active);
CREATE INDEX idx_nodes_region ON nodes (region_id);
CREATE INDEX idx_nodes_status ON nodes (status);
CREATE INDEX idx_peers_user ON peers (user_id);
CREATE INDEX idx_peers_node ON peers (node_id);
CREATE INDEX idx_peers_status ON peers (status);
CREATE INDEX idx_sessions_user ON sessions (user_id);
CREATE INDEX idx_sessions_status ON sessions (status);
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id);
CREATE INDEX idx_user_tokens_user ON user_tokens (user_id);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_tokens_expires_at;
DROP INDEX IF EXISTS idx_user_tokens_user;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_sessions_status;
DROP INDEX IF EXISTS idx_sessions_user;
DROP INDEX IF EXISTS idx_peers_status;
DROP INDEX IF EXISTS idx_peers_node;
DROP INDEX IF EXISTS idx_peers_user;
DROP INDEX IF EXISTS idx_nodes_status;
DROP INDEX IF EXISTS idx_nodes_region;
DROP INDEX IF EXISTS idx_regions_active;
DROP INDEX IF EXISTS idx_payments_subscription;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS uq_subscriptions_user_active;
DROP INDEX IF EXISTS uq_subscriptions_provider_id;
DROP INDEX IF EXISTS idx_subscriptions_plan;
DROP INDEX IF EXISTS idx_subscriptions_user;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS peers;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS session_status;
DROP TYPE IF EXISTS peer_status;
DROP TYPE IF EXISTS node_status;
DROP TYPE IF EXISTS payment_provider;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS account_status;
DROP TYPE IF EXISTS user_token_type;
-- +goose StatementEnd
