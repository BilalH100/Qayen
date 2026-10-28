-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catalog_import_sources (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  filename TEXT NOT NULL,
  sheet_name TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, sha256)
);

CREATE TABLE IF NOT EXISTS medication_reference_entries (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL DEFAULT 'CNOPS',
  product_code TEXT NOT NULL,
  name TEXT NOT NULL,
  active_ingredient TEXT NOT NULL,
  dosage TEXT,
  dosage_unit TEXT,
  form TEXT NOT NULL,
  presentation TEXT NOT NULL,
  public_price NUMERIC,
  hospital_price NUMERIC,
  reimbursement_base NUMERIC,
  product_type TEXT,
  reimbursement_rate TEXT,
  currency TEXT NOT NULL DEFAULT 'MAD',
  content_sha256 TEXT NOT NULL,
  medication_id INTEGER REFERENCES medications(id) ON DELETE SET NULL,
  match_status TEXT NOT NULL CHECK (match_status IN ('matched', 'unmatched', 'ambiguous')),
  source_id BIGINT NOT NULL REFERENCES catalog_import_sources(id),
  source_row INTEGER NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, product_code)
);
CREATE INDEX IF NOT EXISTS idx_medication_references_medication ON medication_reference_entries(medication_id);

CREATE TABLE IF NOT EXISTS pharmacy_directory_entries (
  id BIGSERIAL PRIMARY KEY,
  source_id BIGINT NOT NULL REFERENCES catalog_import_sources(id),
  source_row INTEGER NOT NULL,
  name TEXT NOT NULL,
  city TEXT,
  province TEXT,
  pharmacy_id INTEGER REFERENCES pharmacies(id) ON DELETE SET NULL,
  match_status TEXT NOT NULL CHECK (match_status IN ('matched', 'inserted', 'missing_city', 'ambiguous')),
  UNIQUE(source_id, source_row)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pharmacy_directory_entries;
DROP TABLE IF EXISTS medication_reference_entries;
DROP TABLE IF EXISTS catalog_import_sources;
-- +goose StatementEnd
