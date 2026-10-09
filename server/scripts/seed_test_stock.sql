-- Qayen test data: give every map-ready pharmacy the 3 test medications.
-- Quantity is randomized from 10 to 100 inclusive.

BEGIN;

WITH test_meds AS (
    SELECT id, speciality
    FROM medications
    WHERE LOWER(TRIM(speciality)) IN (
        'codoliprane',
        'rhumix',
        'rinomicine'
    )
)
INSERT INTO stock (pharmacy_id, medication_id, quantity)
SELECT
    p.id,
    m.id,
    10 + FLOOR(RANDOM() * 91)::INT
FROM pharmacies p
CROSS JOIN test_meds m
WHERE p.latitude IS NOT NULL
  AND p.longitude IS NOT NULL
ON CONFLICT (pharmacy_id, medication_id)
DO UPDATE SET
    quantity = EXCLUDED.quantity,
    updated_at = NOW();

COMMIT;
