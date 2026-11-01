-- +goose Up
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at
    ON idempotency_keys (expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_idempotency_keys_expires_at;
