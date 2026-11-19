-- +goose Up
CREATE TABLE IF NOT EXISTS pickup_email_reminders (
 id BIGSERIAL PRIMARY KEY,
 reservation_id INTEGER NOT NULL REFERENCES pickup_reservations(id),
 stage INTEGER NOT NULL CHECK(stage IN (1,2)),
 scheduled_at TIMESTAMPTZ NOT NULL,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','sent','skipped')),
 attempts INTEGER NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 sent_at TIMESTAMPTZ,
 UNIQUE(reservation_id,stage)
);
CREATE INDEX IF NOT EXISTS pickup_email_reminders_due ON pickup_email_reminders(next_attempt_at,scheduled_at) WHERE status='pending';
-- +goose Down
DROP TABLE pickup_email_reminders;
