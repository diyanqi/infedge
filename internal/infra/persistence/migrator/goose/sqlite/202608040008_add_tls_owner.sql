-- +goose Up
ALTER TABLE of_tls_certificates ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_of_tls_certificates_owner_id ON of_tls_certificates (owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_of_tls_certificates_owner_id;
ALTER TABLE of_tls_certificates DROP COLUMN owner_id;
