-- +goose Up
CREATE TABLE IF NOT EXISTS of_user_publish_daily_counters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    day_start DATETIME NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, day_start)
);

-- +goose Down
DROP TABLE IF EXISTS of_user_publish_daily_counters;
