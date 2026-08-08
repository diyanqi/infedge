-- +goose Up
ALTER TABLE of_proxy_routes ADD COLUMN IF NOT EXISTS upstream_weights TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE of_proxy_routes DROP COLUMN IF EXISTS upstream_weights;
