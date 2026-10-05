-- +goose Up
-- +goose StatementBegin
ALTER TABLE medication_alerts ADD COLUMN IF NOT EXISTS requested_quantity INT NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE medication_alerts DROP COLUMN IF EXISTS requested_quantity;
-- +goose StatementEnd