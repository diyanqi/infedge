-- +goose Up
ALTER TABLE of_zones ADD COLUMN IF NOT EXISTS claims_ownership BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE of_zones ADD COLUMN IF NOT EXISTS verification_status VARCHAR(32) NOT NULL DEFAULT 'pending';
ALTER TABLE of_zones ADD COLUMN IF NOT EXISTS verification_token VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE of_zones ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP NULL;
ALTER TABLE of_zone_domains ADD COLUMN IF NOT EXISTS verification_status VARCHAR(32) NOT NULL DEFAULT 'pending';
ALTER TABLE of_zone_domains ADD COLUMN IF NOT EXISTS verification_token VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE of_zone_domains ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP NULL;

-- +goose Down
ALTER TABLE of_zone_domains DROP COLUMN IF EXISTS verified_at;
ALTER TABLE of_zone_domains DROP COLUMN IF EXISTS verification_token;
ALTER TABLE of_zone_domains DROP COLUMN IF EXISTS verification_status;
ALTER TABLE of_zones DROP COLUMN IF EXISTS verified_at;
ALTER TABLE of_zones DROP COLUMN IF EXISTS verification_token;
ALTER TABLE of_zones DROP COLUMN IF EXISTS verification_status;
ALTER TABLE of_zones DROP COLUMN IF EXISTS claims_ownership;
