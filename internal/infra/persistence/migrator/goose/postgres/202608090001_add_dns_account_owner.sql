-- +goose Up
ALTER TABLE of_dns_accounts ADD COLUMN IF NOT EXISTS owner_id BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_of_dns_accounts_owner ON of_dns_accounts (owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_of_dns_accounts_owner;
ALTER TABLE of_dns_accounts DROP COLUMN IF EXISTS owner_id;
