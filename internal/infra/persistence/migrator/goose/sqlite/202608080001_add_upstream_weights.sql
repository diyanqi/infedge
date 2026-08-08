-- +goose Up
ALTER TABLE of_proxy_routes ADD COLUMN upstream_weights TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE of_proxy_routes DROP COLUMN upstream_weights;
