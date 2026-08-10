-- +goose Up
DROP INDEX IF EXISTS idx_of_zones_domain;
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zones_owner_domain ON of_zones (domain, owner_id);
DROP INDEX IF EXISTS idx_of_zone_domains_domain;
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zone_domains_zone_domain ON of_zone_domains (zone_id, domain);

-- +goose Down
DROP INDEX IF EXISTS idx_of_zone_domains_zone_domain;
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zone_domains_domain ON of_zone_domains (domain);
DROP INDEX IF EXISTS idx_of_zones_owner_domain;
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zones_domain ON of_zones (domain);
