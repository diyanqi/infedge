-- +goose Up
ALTER TABLE of_origins ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_proxy_routes ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_zones ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_pages_projects ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_config_versions ADD COLUMN created_by_user_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE of_nodes ADD COLUMN node_group_id INTEGER;
ALTER TABLE of_nodes ADD COLUMN monthly_bytes_limit INTEGER NOT NULL DEFAULT 0;
ALTER TABLE w_system_configs ADD COLUMN smtp_from_email TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_of_origins_owner_id ON of_origins (owner_id);
CREATE INDEX IF NOT EXISTS idx_of_proxy_routes_owner_id ON of_proxy_routes (owner_id);
CREATE INDEX IF NOT EXISTS idx_of_zones_owner_id ON of_zones (owner_id);
CREATE INDEX IF NOT EXISTS idx_of_pages_projects_owner_id ON of_pages_projects (owner_id);
CREATE INDEX IF NOT EXISTS idx_of_config_versions_created_by_user_id ON of_config_versions (created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_of_nodes_node_group_id ON of_nodes (node_group_id);

CREATE TABLE IF NOT EXISTS of_subscription_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_fen INTEGER NOT NULL DEFAULT 0,
    billing_months INTEGER NOT NULL DEFAULT 1,
    high_speed_bytes INTEGER NOT NULL DEFAULT 0,
    throttle_bytes_per_sec INTEGER NOT NULL DEFAULT 0,
    daily_publish_limit INTEGER NOT NULL DEFAULT 1,
    max_zones INTEGER NOT NULL DEFAULT 1,
    max_origins INTEGER NOT NULL DEFAULT 1,
    max_routes INTEGER NOT NULL DEFAULT 1,
    max_pages INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_subscription_plans_enabled ON of_subscription_plans (enabled);

CREATE TABLE IF NOT EXISTS of_user_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    plan_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    starts_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_user_subscriptions_user_status ON of_user_subscriptions (user_id, status);

CREATE TABLE IF NOT EXISTS of_payment_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    gateway TEXT NOT NULL,
    pid TEXT NOT NULL,
    secret_key TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS of_payment_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    plan_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    amount_fen INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    trade_no TEXT NOT NULL DEFAULT '',
    paid_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_payment_orders_user_status ON of_payment_orders (user_id, status);

CREATE TABLE IF NOT EXISTS of_node_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    monthly_bytes_limit INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS of_node_group_nodes (
    node_group_id INTEGER NOT NULL,
    node_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (node_group_id, node_id)
);
CREATE INDEX IF NOT EXISTS idx_of_node_group_nodes_node_id ON of_node_group_nodes (node_id);

INSERT OR IGNORE INTO of_subscription_plans
    (name, description, price_fen, billing_months, high_speed_bytes, throttle_bytes_per_sec,
     daily_publish_limit, max_zones, max_origins, max_routes, max_pages, enabled)
VALUES ('基础版', '适合体验的基础套餐', 0, 1, 10737418240, 1048576, 1, 1, 1, 1, 0, TRUE);

INSERT OR IGNORE INTO w_system_configs (key, value, type, visibility, description, created_at, updated_at)
VALUES ('smtp_from_email', '', 'system', 0, 'SMTP 发件邮箱', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- +goose Down
DROP TABLE IF EXISTS of_node_group_nodes;
DROP TABLE IF EXISTS of_node_groups;
DROP TABLE IF EXISTS of_payment_orders;
DROP TABLE IF EXISTS of_payment_channels;
DROP TABLE IF EXISTS of_user_subscriptions;
DROP TABLE IF EXISTS of_subscription_plans;
ALTER TABLE of_nodes DROP COLUMN node_group_id;
ALTER TABLE of_nodes DROP COLUMN monthly_bytes_limit;
ALTER TABLE of_config_versions DROP COLUMN created_by_user_id;
ALTER TABLE of_pages_projects DROP COLUMN owner_id;
ALTER TABLE of_zones DROP COLUMN owner_id;
ALTER TABLE of_proxy_routes DROP COLUMN owner_id;
ALTER TABLE of_origins DROP COLUMN owner_id;
ALTER TABLE w_system_configs DROP COLUMN smtp_from_email;
