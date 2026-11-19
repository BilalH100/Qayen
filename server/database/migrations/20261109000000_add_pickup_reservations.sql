-- +goose Up
CREATE TABLE IF NOT EXISTS pickup_reservations (
 id SERIAL PRIMARY KEY,
 alert_id INTEGER NOT NULL UNIQUE REFERENCES medication_alerts(id),
 customer_id INTEGER NOT NULL REFERENCES users(id),
 pharmacy_id INTEGER NOT NULL REFERENCES pharmacies(id),
 medication_id INTEGER NOT NULL REFERENCES medications(id),
 stock_id INTEGER NOT NULL,
 quantity INTEGER NOT NULL CHECK(quantity > 0),
 status TEXT NOT NULL DEFAULT 'awaiting' CHECK(status IN ('awaiting','picked_up','canceled')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 expires_at TIMESTAMPTZ NOT NULL,
 reminder_stage INTEGER NOT NULL DEFAULT 0 CHECK(reminder_stage BETWEEN 0 AND 2),
 resolved_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS pickup_reservations_due ON pickup_reservations(expires_at) WHERE status='awaiting';
CREATE INDEX IF NOT EXISTS pickup_reservations_customer ON pickup_reservations(customer_id,created_at DESC);
CREATE INDEX IF NOT EXISTS pickup_reservations_pharmacy ON pickup_reservations(pharmacy_id,created_at DESC);

-- +goose Down
DROP TABLE pickup_reservations;
