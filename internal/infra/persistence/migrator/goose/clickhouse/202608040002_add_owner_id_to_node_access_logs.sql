-- +goose Up
ALTER TABLE of_node_access_logs ADD COLUMN IF NOT EXISTS owner_id UInt64 DEFAULT 0 AFTER node_id;

-- +goose Down
ALTER TABLE of_node_access_logs DROP COLUMN IF EXISTS owner_id;
