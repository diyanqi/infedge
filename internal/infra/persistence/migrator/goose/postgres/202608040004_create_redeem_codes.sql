-- +goose Up
CREATE TABLE IF NOT EXISTS of_redeem_codes (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    plan_id BIGINT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'unused',
    used_by BIGINT,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_redeem_codes_plan_id ON of_redeem_codes (plan_id);
CREATE INDEX IF NOT EXISTS idx_of_redeem_codes_status ON of_redeem_codes (status);

-- +goose Down
DROP TABLE IF EXISTS of_redeem_codes;
