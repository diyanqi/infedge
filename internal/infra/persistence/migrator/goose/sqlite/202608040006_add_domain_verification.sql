-- +goose Up
ALTER TABLE of_zones ADD COLUMN claims_ownership BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE of_zones ADD COLUMN verification_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE of_zones ADD COLUMN verification_token TEXT NOT NULL DEFAULT '';
ALTER TABLE of_zones ADD COLUMN verified_at DATETIME NULL;
ALTER TABLE of_zone_domains ADD COLUMN verification_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE of_zone_domains ADD COLUMN verification_token TEXT NOT NULL DEFAULT '';
ALTER TABLE of_zone_domains ADD COLUMN verified_at DATETIME NULL;

-- +goose Down
ALTER TABLE of_zone_domains DROP COLUMN verified_at;
ALTER TABLE of_zone_domains DROP COLUMN verification_token;
ALTER TABLE of_zone_domains DROP COLUMN verification_status;
ALTER TABLE of_zones DROP COLUMN verified_at;
ALTER TABLE of_zones DROP COLUMN verification_token;
ALTER TABLE of_zones DROP COLUMN verification_status;
ALTER TABLE of_zones DROP COLUMN claims_ownership;
