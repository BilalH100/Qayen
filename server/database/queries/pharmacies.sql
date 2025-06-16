-- name: CreatePharmacy :one
INSERT INTO pharmacies (name, address, latitude, longitude, city, phone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPharmacyByID :one
SELECT * FROM pharmacies WHERE id = $1;

-- name: ListPharmacies :many
SELECT * FROM pharmacies ORDER BY id;

-- name: UpdatePharmacy :one
UPDATE pharmacies
SET name = $2, address = $3, latitude = $4, longitude = $5, city = $6, phone = $7
WHERE id = $1
RETURNING *;

-- name: DeletePharmacy :exec
DELETE FROM pharmacies WHERE id = $1;
