-- name: CreateUser :one
INSERT INTO users (email, name, phone, password, user_role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET name = $2, phone = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * 
FROM users
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUserRole :one 
UPDATE users 
SET user_role = $2
WHERE id = $1
RETURNING *;

