-- +goose Up
-- +goose StatementBegin

-- Add pharmacy operating hours
ALTER TABLE pharmacies ADD COLUMN opening_time TIME DEFAULT '08:00:00';
ALTER TABLE pharmacies ADD COLUMN closing_time TIME DEFAULT '22:00:00';
ALTER TABLE pharmacies ADD COLUMN is_24h BOOLEAN DEFAULT FALSE;
ALTER TABLE pharmacies ADD COLUMN closed_days INTEGER[] DEFAULT '{}'; -- 0=Sunday, 1=Monday, etc.

-- Table for medication search alerts
CREATE TABLE medication_alerts (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES users(id) ON DELETE CASCADE,
  medication_id INT REFERENCES medications(id) ON DELETE CASCADE,
  customer_latitude DOUBLE PRECISION NOT NULL,
  customer_longitude DOUBLE PRECISION NOT NULL,
  search_radius_km DOUBLE PRECISION DEFAULT 10.0,
  max_response_time_minutes INT DEFAULT 2,
  status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'expired', 'cancelled')),
  created_at TIMESTAMP DEFAULT now(),
  expires_at TIMESTAMP DEFAULT (now() + INTERVAL '2 minutes'),
  
  -- Analytics fields
  total_pharmacies_notified INT DEFAULT 0,
  total_responses_received INT DEFAULT 0
);

-- Table for pharmacy alert notifications sent
CREATE TABLE pharmacy_alert_notifications (
  id SERIAL PRIMARY KEY,
  alert_id INT REFERENCES medication_alerts(id) ON DELETE CASCADE,
  pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
  sent_at TIMESTAMP DEFAULT now(),
  
  -- Rate limiting
  UNIQUE(alert_id, pharmacy_id)
);

-- Table for pharmacist responses to alerts
CREATE TABLE pharmacist_responses (
  id SERIAL PRIMARY KEY,
  alert_id INT REFERENCES medication_alerts(id) ON DELETE CASCADE,
  pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
  response_type TEXT NOT NULL CHECK (response_type IN ('available', 'unavailable', 'substitute')),
  substitute_medication_id INT REFERENCES medications(id),
  substitute_brand TEXT,
  substitute_notes TEXT,
  response_time_seconds INT, -- Time taken to respond
  responded_at TIMESTAMP DEFAULT now(),
  expires_at TIMESTAMP DEFAULT (now() + INTERVAL '1 hour'), -- Response validity
  
  -- Ensure one response per pharmacy per alert
  UNIQUE(alert_id, pharmacy_id)
);

-- Table for rate limiting repeated alerts
CREATE TABLE alert_rate_limits (
  id SERIAL PRIMARY KEY,
  pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
  medication_id INT REFERENCES medications(id) ON DELETE CASCADE,
  last_alert_sent TIMESTAMP DEFAULT now(),
  alert_count_today INT DEFAULT 1,
  
  UNIQUE(pharmacy_id, medication_id)
);

-- Table for pharmacy analytics
CREATE TABLE pharmacy_analytics (
  id SERIAL PRIMARY KEY,
  pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
  date DATE DEFAULT CURRENT_DATE,
  total_alerts_received INT DEFAULT 0,
  total_responses_sent INT DEFAULT 0,
  avg_response_time_seconds DOUBLE PRECISION DEFAULT 0,
  availability_rate DOUBLE PRECISION DEFAULT 0, -- percentage of "available" responses
  
  UNIQUE(pharmacy_id, date)
);

-- Indexes for performance
CREATE INDEX idx_medication_alerts_status_expires ON medication_alerts(status, expires_at);
CREATE INDEX idx_medication_alerts_customer_medication ON medication_alerts(customer_id, medication_id);
CREATE INDEX idx_pharmacy_responses_alert_pharmacy ON pharmacist_responses(alert_id, pharmacy_id);
CREATE INDEX idx_pharmacy_responses_expires ON pharmacist_responses(expires_at);
CREATE INDEX idx_pharmacy_location ON pharmacies(latitude, longitude);
CREATE INDEX idx_alert_rate_limits_pharmacy_med ON alert_rate_limits(pharmacy_id, medication_id);
CREATE INDEX idx_pharmacy_analytics_date ON pharmacy_analytics(date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pharmacy_analytics;
DROP TABLE IF EXISTS alert_rate_limits;
DROP TABLE IF EXISTS pharmacist_responses;
DROP TABLE IF EXISTS pharmacy_alert_notifications;
DROP TABLE IF EXISTS medication_alerts;

ALTER TABLE pharmacies DROP COLUMN IF EXISTS opening_time;
ALTER TABLE pharmacies DROP COLUMN IF EXISTS closing_time;
ALTER TABLE pharmacies DROP COLUMN IF EXISTS is_24h;
ALTER TABLE pharmacies DROP COLUMN IF EXISTS closed_days;

-- +goose StatementEnd
