-- name: CreateMedication :exec
INSERT INTO medications (
  status, commercial_status, speciality, dosage,
  form, presentation, pp, active_substance, therapeutic_class,
  epi, ppv, ph, pfht, code, tva, description, common_sd, serious_sd, general_info
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19); 

-- name: GetMedicationByID :one
SELECT * FROM medications WHERE id = $1;

-- name: GetMedicationByCode :one
SELECT * FROM medications WHERE code = $1;

-- name: ListMedications :many
SELECT * FROM medications ORDER BY speciality;

-- name: UpdateMedication :one
UPDATE medications
SET
  status = $2, commercial_status = $3, speciality = $4, dosage = $5,
  form = $6, presentation = $7, pp = $8, active_substance = $9,
  therapeutic_class = $10, epi = $11, ppv = $12, ph = $13, pfht = $14, code = $15, tva = $16, description = $17, 
  common_sd = $18, serious_sd = $19, general_info = $20
WHERE id = $1
RETURNING *;

-- name: DeleteMedication :exec
DELETE FROM medications WHERE id = $1;

-- name: GetMedications :many
SELECT * FROM medications
ORDER BY id
LIMIT $1 OFFSET $2;
