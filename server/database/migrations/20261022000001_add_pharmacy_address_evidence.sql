-- +goose Up
CREATE TABLE IF NOT EXISTS pharmacy_address_evidence (
  id BIGSERIAL PRIMARY KEY,
  pharmacy_id INTEGER NOT NULL REFERENCES pharmacies(id),
  provider TEXT NOT NULL,
  source_url TEXT NOT NULL,
  page_url TEXT NOT NULL,
  page_sha256 TEXT NOT NULL,
  record_sha256 TEXT NOT NULL,
  retrieved_at TIMESTAMPTZ NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  previous_address TEXT NOT NULL,
  address TEXT NOT NULL,
  match_method TEXT NOT NULL,
  business_verified BOOLEAN NOT NULL DEFAULT false,
  UNIQUE(pharmacy_id,provider,record_sha256)
);

-- +goose Down
DROP TABLE IF EXISTS pharmacy_address_evidence;
