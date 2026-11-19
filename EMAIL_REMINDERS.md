# Pickup reminder emails

Both reminder emails go to the email address of the patient's Qayen account, including Google accounts. In local 5-minute mode they are scheduled 3 and 4 minutes after reservation. The existing in-app reminders remain active. The backend must be running; the browser can be closed.

## Gmail setup

1. Use a Gmail account as Qayen's sender. Enable Google 2-Step Verification and create an App Password: https://myaccount.google.com/apppasswords (some managed accounts restrict this).
2. Copy the settings from `server/email-reminders.env.example` into `server/.env`. Set `SMTP_USER` and `SMTP_FROM` to the sender address, and `SMTP_PASSWORD` to the sender's App Password without spaces. Do not use the normal Gmail password. Keep all credentials local; never commit or paste them into chat.
3. Set `EMAIL_REMINDERS_ENABLED=true` and restart your normal Go backend. Startup should say `Pickup email reminders enabled`.
4. Sign in as a patient with an inbox you control, reserve medication, and check email at 3 and 4 minutes. Test collection/expiry separately. Gmail may place messages in Spam; SMTP acceptance cannot guarantee Inbox placement.

SMTP uses encrypted STARTTLS on port 587 or implicit TLS on port 465. Only the sending account needs SMTP credentials; patients never supply mailbox passwords.

Delivery jobs are persisted with one job per reservation/reminder, retry failures every 15 seconds while the hold remains active, and skip picked-up, canceled, expired or stale reminders. After downtime past the first reminder's window, only the latest relevant reminder is sent. A rare SMTP-accepted/database-commit failure can cause a retry; SMTP cannot guarantee exactly-once delivery. Stable Message-ID values identify repeated attempts.

The schema is initialized at backend startup, and the additive migration is also supplied for Goose. No sender is enabled by default and no live email is sent by automated tests.

Validation: `go test ./server/...`; isolated PostgreSQL delivery tests: set `EMAIL_TEST_DB_ENV=../.env` and run `go test ./server/services -run TestPickupEmailDelivery -count=1` from the repository root. Those tests use temporary tables and a fake sender.

Files added:
- `server/database/migrations/20261111000000_add_pickup_email_reminders.sql`
- `server/services/pickup_emails.go`
- `server/services/reminder_smtp.go`
- `server/services/pickup_emails_test.go`
- `server/email-reminders.env.example`x`
- `EMAIL_REMINDERS.md`

Files changed:
- `server/database/pickups.go`
- `server/main.go`
