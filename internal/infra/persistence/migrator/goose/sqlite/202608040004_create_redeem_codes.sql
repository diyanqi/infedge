-- +goose Up
CREATE TABLE IF NOT EXISTS of_redeem_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    plan_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'unused',
    used_by INTEGER,
    used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_redeem_codes_plan_id ON of_redeem_codes (plan_id);
CREATE INDEX IF NOT EXISTS idx_of_redeem_codes_status ON of_redeem_codes (status);

-- +goose Down
DROP TABLE IF EXISTS of_redeem_codes;
