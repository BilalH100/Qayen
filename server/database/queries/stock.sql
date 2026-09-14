-- name: CreateStock :one
INSERT INTO stock (pharmacy_id, medication_id, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetStock :one
SELECT * FROM stock WHERE pharmacy_id = $1 AND medication_id = $2;

-- name: ListStockByPharmacy :many
SELECT s.*, m.speciality
FROM stock s
JOIN medications m ON m.id = s.medication_id
WHERE s.pharmacy_id = $1;

-- name: UpdateStockQuantity :one
UPDATE stock
SET quantity = $3, updated_at = now()
WHERE pharmacy_id = $1 AND medication_id = $2
RETURNING *;

-- name: DeleteStockEntry :exec
DELETE FROM stock WHERE pharmacy_id = $1 AND medication_id = $2;
