-- +goose Up
-- +goose StatementBegin

-- Create enum type for user roles
CREATE TYPE user_role AS ENUM ('regular', 'pharmacist', 'admin');

-- Add user_role column to users table with default value 'regular'
ALTER TABLE users ADD COLUMN user_role user_role DEFAULT 'regular' NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the user_role column
ALTER TABLE users DROP COLUMN IF EXISTS user_role;

-- Drop the enum type
DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd