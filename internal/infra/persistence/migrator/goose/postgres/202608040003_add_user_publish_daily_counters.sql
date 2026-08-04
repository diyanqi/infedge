-- +goose Up
CREATE TABLE IF NOT EXISTS of_user_publish_daily_counters (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    day_start TIMESTAMPTZ NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_of_user_publish_daily_user_day UNIQUE (user_id, day_start)
);

-- +goose Down
DROP TABLE IF EXISTS of_user_publish_daily_counters;
