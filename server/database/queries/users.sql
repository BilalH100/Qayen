-- name: CreateUser :one
INSERT INTO users (email, name, phone)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET name = $2, phone = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
