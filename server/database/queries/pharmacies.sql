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

-- name: GetClosestPharmacyWithMedication :one
SELECT
  p.*,
  (
    6371 * acos(
      cos(radians($1)) *
      cos(radians(p.latitude)) *
      cos(radians(p.longitude) - radians($2)) +
      sin(radians($1)) *
      sin(radians(p.latitude))
    )
  ) AS distance_km
FROM pharmacies p
JOIN stock pm ON pm.pharmacy_id = p.id
WHERE pm.medication_id = $3
  AND pm.quantity > 0
ORDER BY distance_km ASC
LIMIT 1;
