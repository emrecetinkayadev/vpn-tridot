-- +goose Up
ALTER TYPE user_token_type ADD VALUE IF NOT EXISTS 'peer_config';

-- +goose Down
DELETE FROM user_tokens WHERE token_type = 'peer_config';

ALTER TYPE user_token_type RENAME TO user_token_type_old;
CREATE TYPE user_token_type AS ENUM ('email_verification', 'password_reset');
ALTER TABLE user_tokens
	ALTER COLUMN token_type TYPE user_token_type USING token_type::text::user_token_type;
DROP TYPE user_token_type_old;
