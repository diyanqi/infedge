-- +goose Up
ALTER TABLE of_waf_rule_groups ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_waf_rule_groups ADD COLUMN host TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_of_waf_rule_groups_owner_id ON of_waf_rule_groups (owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_of_waf_rule_groups_owner_id;
ALTER TABLE of_waf_rule_groups DROP COLUMN host;
ALTER TABLE of_waf_rule_groups DROP COLUMN owner_id;
