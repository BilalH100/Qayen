-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_subject TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS users_google_subject_key ON users(google_subject);

-- +goose Down
DROP INDEX IF EXISTS users_google_subject_key;
ALTER TABLE users DROP COLUMN IF EXISTS google_subject;
