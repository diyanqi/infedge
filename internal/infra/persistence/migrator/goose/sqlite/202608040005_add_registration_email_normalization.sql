-- +goose Up
ALTER TABLE w_users ADD COLUMN email_normalized TEXT NOT NULL DEFAULT '';
UPDATE w_users SET email_normalized = lower(trim(email)) WHERE email_normalized = '' AND email <> '';
CREATE INDEX IF NOT EXISTS idx_w_users_email_normalized ON w_users (email_normalized);
INSERT OR IGNORE INTO w_system_configs (key, value, type, visibility, description, created_at, updated_at)
VALUES ('registration_email_domain_allowlist', '', 'system', 0, '注册邮箱域名白名单，逗号分隔；留空表示不限制', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- +goose Down
DROP INDEX IF EXISTS idx_w_users_email_normalized;
ALTER TABLE w_users DROP COLUMN email_normalized;
DELETE FROM w_system_configs WHERE key = 'registration_email_domain_allowlist';
