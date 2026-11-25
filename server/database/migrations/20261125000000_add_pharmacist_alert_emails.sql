-- +goose Up
CREATE TABLE IF NOT EXISTS pharmacist_alert_emails (
 id BIGSERIAL PRIMARY KEY,
 notification_id INTEGER NOT NULL REFERENCES pharmacy_alert_notifications(id) ON DELETE CASCADE,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','sent','skipped')),
 attempts INTEGER NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 sent_at TIMESTAMPTZ,
 UNIQUE(notification_id,user_id)
);
CREATE INDEX IF NOT EXISTS pharmacist_alert_emails_due ON pharmacist_alert_emails(next_attempt_at) WHERE status='pending';
-- +goose Down
DROP TABLE pharmacist_alert_emails;
