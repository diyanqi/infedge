-- +goose Up
CREATE TABLE IF NOT EXISTS of_user_traffic_monthly (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    month_start DATETIME NOT NULL,
    bytes_sent INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, month_start)
);
CREATE INDEX IF NOT EXISTS idx_of_user_traffic_monthly_month ON of_user_traffic_monthly (month_start);

-- +goose Down
DROP TABLE IF EXISTS of_user_traffic_monthly;
