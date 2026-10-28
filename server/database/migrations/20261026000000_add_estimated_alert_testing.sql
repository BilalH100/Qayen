-- +goose Up
CREATE TABLE IF NOT EXISTS catalog_test_settings (
  id BOOLEAN PRIMARY KEY DEFAULT true CHECK(id),
  enabled BOOLEAN NOT NULL DEFAULT false
);
INSERT INTO catalog_test_settings(id,enabled) VALUES(true,false) ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS catalog_test_locations (
  pharmacy_id INTEGER PRIMARY KEY REFERENCES pharmacies(id),
  matched_address TEXT NOT NULL,
  latitude DOUBLE PRECISION NOT NULL CHECK(latitude BETWEEN -90 AND 90),
  longitude DOUBLE PRECISION NOT NULL CHECK(longitude BETWEEN -180 AND 180),
  basis TEXT NOT NULL DEFAULT 'city_area_estimate'
);
CREATE TABLE IF NOT EXISTS catalog_test_stock (
  id SERIAL PRIMARY KEY,
  pharmacy_id INTEGER NOT NULL REFERENCES pharmacies(id),
  medication_id INTEGER NOT NULL REFERENCES medications(id),
  quantity INTEGER NOT NULL CHECK(quantity>=0),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE(pharmacy_id,medication_id)
);
CREATE TABLE IF NOT EXISTS catalog_test_alerts (
  alert_id INTEGER PRIMARY KEY REFERENCES medication_alerts(id) ON DELETE CASCADE
);
CREATE OR REPLACE VIEW alert_pharmacy_locations AS
 SELECT p.id,p.name,p.address,
  COALESCE(p.latitude,CASE WHEN s.enabled AND current_setting('qayen.test_alerts',true)='on' AND p.address=t.matched_address THEN t.latitude END) AS latitude,
  COALESCE(p.longitude,CASE WHEN s.enabled AND current_setting('qayen.test_alerts',true)='on' AND p.address=t.matched_address THEN t.longitude END) AS longitude,
  p.city,p.phone,p.created_at
 FROM pharmacies p
 LEFT JOIN catalog_test_locations t ON t.pharmacy_id=p.id AND p.latitude IS NULL AND p.longitude IS NULL
 LEFT JOIN catalog_test_settings s ON s.id=true;
CREATE OR REPLACE VIEW alert_stock AS
 SELECT s.id,s.pharmacy_id,s.medication_id,s.quantity,s.updated_at FROM stock s
 UNION ALL
 SELECT -t.id,t.pharmacy_id,t.medication_id,t.quantity,t.updated_at FROM catalog_test_stock t
 JOIN catalog_test_settings c ON c.id=true AND c.enabled
 WHERE current_setting('qayen.test_alerts',true)='on'
 AND NOT EXISTS(SELECT 1 FROM stock s WHERE s.pharmacy_id=t.pharmacy_id AND s.medication_id=t.medication_id);
-- +goose Down
DROP VIEW IF EXISTS alert_stock;
DROP VIEW IF EXISTS alert_pharmacy_locations;
DROP TABLE IF EXISTS catalog_test_alerts;
DROP TABLE IF EXISTS catalog_test_stock;
DROP TABLE IF EXISTS catalog_test_locations;
DROP TABLE IF EXISTS catalog_test_settings;
