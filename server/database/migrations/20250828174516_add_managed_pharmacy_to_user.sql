-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN managed_pharmacy_id INTEGER REFERENCES pharmacies(id) ON DELETE SET NULL;
-- +goose StatementEnd




-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN managed_pharmacy_id;
-- +goose StatementEnd
1