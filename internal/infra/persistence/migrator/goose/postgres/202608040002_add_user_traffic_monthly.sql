-- +goose Up
CREATE TABLE IF NOT EXISTS of_user_traffic_monthly (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    month_start TIMESTAMPTZ NOT NULL,
    bytes_sent BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_of_user_traffic_monthly_user_month UNIQUE (user_id, month_start)
);
CREATE INDEX IF NOT EXISTS idx_of_user_traffic_monthly_month ON of_user_traffic_monthly (month_start);

-- +goose Down
DROP TABLE IF EXISTS of_user_traffic_monthly;
