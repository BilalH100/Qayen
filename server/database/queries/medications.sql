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

-- name: GetCategories :many
SELECT DISTINCT * FROM category;

-- name: SearchMedications :many
SELECT * FROM medications
WHERE
  presentation ILIKE $1 OR
  speciality ILIKE $1 OR
  active_substance ILIKE $1 OR
  therapeutic_class ILIKE $1 OR
  code ILIKE $1
ORDER BY id
LIMIT 20;

-- name: SearchMedicationsBySpeciality :many
SELECT * FROM medications
WHERE speciality ILIKE $1
ORDER BY 
  CASE 
    WHEN speciality ILIKE $1 THEN 1
    ELSE 2
  END,
  speciality
LIMIT 20;

-- name: SearchMedicationsWithPriority :many
SELECT * FROM medications
WHERE
  speciality ILIKE $1 OR
  active_substance ILIKE $1 OR
  presentation ILIKE $1 OR
  therapeutic_class ILIKE $1 OR
  code ILIKE $1
ORDER BY 
  CASE 
    WHEN speciality ILIKE $1 THEN 1
    WHEN active_substance ILIKE $1 THEN 2
    WHEN presentation ILIKE $1 THEN 3
    WHEN therapeutic_class ILIKE $1 THEN 4
    WHEN code ILIKE $1 THEN 5
    ELSE 6
  END,
  speciality
LIMIT 20;
