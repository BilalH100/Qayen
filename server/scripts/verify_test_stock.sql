-- Qayen test data verification

-- 1) Number of pharmacies displayed on the map (coordinates available)
SELECT COUNT(*) AS map_ready_pharmacies
FROM pharmacies
WHERE latitude IS NOT NULL
  AND longitude IS NOT NULL;

-- 2) Confirm the 3 test medications exist
SELECT id, speciality
FROM medications
WHERE LOWER(TRIM(speciality)) IN (
    'codoliprane',
    'rhumix',
    'rinomicine'
)
ORDER BY speciality;

-- 3) Check stock coverage and quantity range for each test medication
SELECT
    m.speciality,
    COUNT(s.id) AS pharmacies_with_stock,
    MIN(s.quantity) AS minimum_quantity,
    MAX(s.quantity) AS maximum_quantity
FROM medications m
LEFT JOIN stock s ON s.medication_id = m.id
WHERE LOWER(TRIM(m.speciality)) IN (
    'codoliprane',
    'rhumix',
    'rinomicine'
)
GROUP BY m.id, m.speciality
ORDER BY m.speciality;

-- 4) Pharmacies missing one or more of the 3 test medications.
-- Expected result: 0 rows.
SELECT
    p.id,
    p.name,
    COUNT(s.id) AS test_medication_count
FROM pharmacies p
JOIN medications m
    ON LOWER(TRIM(m.speciality)) IN ('codoliprane', 'rhumix', 'rinomicine')
LEFT JOIN stock s
    ON s.pharmacy_id = p.id
   AND s.medication_id = m.id
WHERE p.latitude IS NOT NULL
  AND p.longitude IS NOT NULL
GROUP BY p.id, p.name
HAVING COUNT(s.id) < 3
ORDER BY p.id;
