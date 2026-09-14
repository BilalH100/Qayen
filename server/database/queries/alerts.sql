-- name: CreateMedicationAlert :one
INSERT INTO medication_alerts (
  customer_id, medication_id, customer_latitude, customer_longitude, 
  search_radius_km, max_response_time_minutes
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMedicationAlert :one
SELECT * FROM medication_alerts WHERE id = $1;

-- name: UpdateMedicationAlertStatus :one
UPDATE medication_alerts 
SET status = $2, total_responses_received = $3
WHERE id = $1
RETURNING *;

-- name: ExpireMedicationAlert :exec
UPDATE medication_alerts 
SET status = 'expired'
WHERE id = $1 AND status = 'pending';

-- name: GetActiveMedicationAlerts :many
SELECT * FROM medication_alerts 
WHERE status = 'pending' AND expires_at > now()
ORDER BY created_at DESC;

-- name: GetNearbyPharmaciesForAlert :many
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
WHERE 
  p.latitude IS NOT NULL 
  AND p.longitude IS NOT NULL
  AND (
    6371 * acos(
      cos(radians($1)) *
      cos(radians(p.latitude)) *
      cos(radians(p.longitude) - radians($2)) +
      sin(radians($1)) *
      sin(radians(p.latitude))
    )
  ) <= $3
ORDER BY distance_km ASC;

-- name: CheckPharmacyIsOpen :one
SELECT 
  CASE 
    WHEN p.is_24h = true THEN true
    WHEN p.is_24h IS NULL AND p.opening_time IS NULL AND p.closing_time IS NULL AND (p.closed_days IS NULL OR array_length(p.closed_days, 1) IS NULL) THEN true
    WHEN EXTRACT(DOW FROM now()) = ANY(p.closed_days) THEN false
    WHEN p.opening_time IS NOT NULL AND p.closing_time IS NOT NULL AND now()::time BETWEEN p.opening_time AND p.closing_time THEN true
    ELSE false
  END as is_open
FROM pharmacies p 
WHERE p.id = $1;

-- name: CreatePharmacyAlertNotification :one
INSERT INTO pharmacy_alert_notifications (alert_id, pharmacy_id)
VALUES ($1, $2)
ON CONFLICT (alert_id, pharmacy_id) DO NOTHING
RETURNING *;

-- name: CheckAlertRateLimit :one
SELECT 
  id,
  last_alert_sent,
  alert_count_today,
  CASE 
    WHEN DATE(last_alert_sent) = CURRENT_DATE AND alert_count_today >= $3 THEN false
    WHEN last_alert_sent > (now() - INTERVAL '15 minutes') THEN false
    ELSE true
  END as can_send_alert
FROM alert_rate_limits 
WHERE pharmacy_id = $1 AND medication_id = $2;

-- name: UpdateAlertRateLimit :one
INSERT INTO alert_rate_limits (pharmacy_id, medication_id, last_alert_sent, alert_count_today)
VALUES ($1, $2, now(), 1)
ON CONFLICT (pharmacy_id, medication_id) 
DO UPDATE SET 
  last_alert_sent = now(),
  alert_count_today = CASE 
    WHEN DATE(alert_rate_limits.last_alert_sent) = CURRENT_DATE 
    THEN alert_rate_limits.alert_count_today + 1
    ELSE 1
  END
RETURNING *;

-- name: CreatePharmacistResponse :one
INSERT INTO pharmacist_responses (
  alert_id, pharmacy_id, response_type, substitute_medication_id, 
  substitute_brand, substitute_notes, response_time_seconds
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (alert_id, pharmacy_id) 
DO UPDATE SET 
  response_type = $3,
  substitute_medication_id = $4,
  substitute_brand = $5,
  substitute_notes = $6,
  response_time_seconds = $7,
  responded_at = now(),
  expires_at = now() + INTERVAL '1 hour'
RETURNING *;

-- name: GetPharmacistResponsesForAlert :many
SELECT 
  pr.*,
  p.name as pharmacy_name,
  p.address as pharmacy_address,
  p.phone as pharmacy_phone,
  p.latitude as pharmacy_latitude,
  p.longitude as pharmacy_longitude,
  m.speciality as substitute_medication_name,
  (
    6371 * acos(
      cos(radians($2)) *
      cos(radians(p.latitude)) *
      cos(radians(p.longitude) - radians($3)) +
      sin(radians($2)) *
      sin(radians(p.latitude))
    )
  ) AS distance_km
FROM pharmacist_responses pr
JOIN pharmacies p ON p.id = pr.pharmacy_id
LEFT JOIN medications m ON m.id = pr.substitute_medication_id
WHERE pr.alert_id = $1 
  AND pr.expires_at > now()
  AND pr.response_type IN ('available', 'substitute')
ORDER BY distance_km ASC;

-- name: GetExpiredResponses :many
SELECT * FROM pharmacist_responses 
WHERE expires_at <= now();

-- name: CleanupExpiredResponses :exec
DELETE FROM pharmacist_responses 
WHERE expires_at <= now();

-- name: UpdatePharmacyAnalytics :one
INSERT INTO pharmacy_analytics (
  pharmacy_id, date, total_alerts_received, total_responses_sent, 
  avg_response_time_seconds, availability_rate
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (pharmacy_id, date) 
DO UPDATE SET 
  total_alerts_received = pharmacy_analytics.total_alerts_received + $3,
  total_responses_sent = pharmacy_analytics.total_responses_sent + $4,
  avg_response_time_seconds = (
    (pharmacy_analytics.avg_response_time_seconds * pharmacy_analytics.total_responses_sent + $5 * $4) / 
    (pharmacy_analytics.total_responses_sent + $4)
  ),
  availability_rate = (
    (pharmacy_analytics.availability_rate * pharmacy_analytics.total_responses_sent + $6 * $4) / 
    (pharmacy_analytics.total_responses_sent + $4)
  )
RETURNING *;

-- name: GetPharmacyAnalytics :many
SELECT * FROM pharmacy_analytics 
WHERE pharmacy_id = $1 
ORDER BY date DESC
LIMIT $2;

-- name: GetPharmacyDashboardAlerts :many
SELECT 
  ma.id,
  ma.created_at,
  ma.expires_at,
  m.speciality as medication_name,
  m.active_substance,
  u.name as customer_name,
  (
    6371 * acos(
      cos(radians(ma.customer_latitude)) *
      cos(radians(p.latitude)) *
      cos(radians(p.longitude) - radians(ma.customer_longitude)) +
      sin(radians(ma.customer_latitude)) *
      sin(radians(p.latitude))
    )
  ) AS customer_distance_km,
  CASE 
    WHEN pr.id IS NOT NULL THEN true 
    ELSE false 
  END as already_responded
FROM medication_alerts ma
JOIN medications m ON m.id = ma.medication_id
JOIN users u ON u.id = ma.customer_id
JOIN pharmacy_alert_notifications pan ON pan.alert_id = ma.id
JOIN pharmacies p ON p.id = pan.pharmacy_id
LEFT JOIN pharmacist_responses pr ON pr.alert_id = ma.id AND pr.pharmacy_id = $1
WHERE pan.pharmacy_id = $1 
  AND ma.status = 'pending'
  AND ma.expires_at > now()
ORDER BY ma.created_at DESC;