Pickup reservation testing
==========================
Patient: search a medication, receive an available pharmacy response, then confirm pickup. The requested quantity is held immediately for 5 minutes.

In-app reminders appear at 3 and 4 minutes in the patient reservation panel, available on all pages while signed in. Browser notifications, emails and SMS are not sent. Returning to the app shows the current reminder/status.

Pharmacist: open the pharmacy dashboard, find Awaiting pickup, and click Mark as picked up when the patient collects. Stock is already reserved; collection does not reduce it a second time.

If the deadline passes first, the backend cancels the reservation and restores only its quantity to its original real or simulated inventory. The worker runs every second while the backend is running, and catches up after a restart.

Both dashboards show Awaiting pickup, Picked up and Canceled requests. The lists refresh every 2 seconds. Use separate browsers or browser profiles for patient/pharmacist testing.

The default hold is 300 seconds. Set QAYEN_PICKUP_HOLD_SECONDS=86400 in server/.env and restart the backend for 24 hours; reminders occur at 60% and 80% of the configured hold. The pickup schema is initialized on backend startup and is also provided as a Goose migration. Existing completed alerts are not converted into reservations.
