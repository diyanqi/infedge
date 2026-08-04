-- +goose Up
ALTER TABLE of_waf_rule_groups ADD COLUMN IF NOT EXISTS owner_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE of_waf_rule_groups ADD COLUMN IF NOT EXISTS host VARCHAR(255) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_of_waf_rule_groups_owner_id ON of_waf_rule_groups (owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_of_waf_rule_groups_owner_id;
ALTER TABLE of_waf_rule_groups DROP COLUMN IF EXISTS host;
ALTER TABLE of_waf_rule_groups DROP COLUMN IF EXISTS owner_id;
