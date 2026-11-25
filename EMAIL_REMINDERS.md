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


## Pharmacist alert emails

Pharmacist notification emails use the same configured Gmail/SMTP sender. They are enabled alongside `EMAIL_REMINDERS_ENABLED=true`; optionally set `PHARMACIST_ALERT_EMAILS_ENABLED=false` to disable only these emails. Restart your usual backend after this code update; the new table is created automatically on startup.

Recipients are pharmacist user accounts whose `managed_pharmacy_id` matches the pharmacy currently targeted by the sequential alert. The recipient must have a real working email address. Catalog pharmacies without a linked pharmacist account cannot receive email. Demo addresses such as `pharmacist@demo.com` are not a Gmail inbox.

The email includes alert reference, medication name, dosage, form, ingredient, quantity, patient name and phone (when provided), and Morocco timestamps. Set optional `QAYEN_PUBLIC_URL` to your reachable frontend address for a clickable dashboard link. Patient email and GPS coordinates are not included.

Test using separate patient and pharmacist sessions. Use a pharmacist account linked to a nearby pharmacy with an inbox you control, then create a patient alert near that pharmacy with a selected medication and quantity. The targeted pharmacist should receive an email promptly, alongside the existing platform alert. Do not answer immediately if testing email delivery. The existing 30-second response/escalation window remains in effect; email arrival speed depends on the provider. Patient pickup reminders remain at 3 and 4 minutes after reservation.

Each targeted notification/account pair has one durable delivery job. SMTP failures retry every 5 seconds while that pharmacy still has an active unanswered alert. Completed, expired, responded, reassigned and old notifications are skipped. Email sending is separate from stock changes and patient reminders. SMTP acceptance cannot guarantee Inbox placement or exactly-once delivery after a rare acceptance/database-commit failure.

Isolated PostgreSQL tests (fake mail only): set `EMAIL_TEST_DB_ENV=../.env`, then run `go test ./server/services -run TestPharmacistAlertEmail -count=1` from the repository root.

Files added for pharmacist emails:
- `server/database/migrations/20261125000000_add_pharmacist_alert_emails.sql`
- `server/services/pharmacist_emails.go`
- `server/services/pharmacist_emails_test.go`

Files changed for pharmacist emails:
- `server/database/pickups.go`
- `server/main.go`
- `server/email-reminders.env.example`
- `EMAIL_REMINDERS.md`
