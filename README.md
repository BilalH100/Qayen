# Qayen

Healthcare web platform for finding medications and nearby pharmacies in Morocco, with availability alerts and pickup reservations.

## Current development

- Medication catalog and platform search; the **current** local database contained **9,567 medication records**.
- Pharmacy maps, pharmacist stock/location management, and role-based patient, pharmacist, and admin access.
- Email/password and Firebase Google sign-in, remembered account emails, and editable personal profiles.
- Sequential nearby-pharmacy alerts with requested quantities and live platform notifications.
- Pharmacist alert emails containing medication, quantity, and relevant patient contact details.
- Pickup reservations, pharmacist collection confirmation, automatic expiry and stock restoration, and patient email reminders.
- English, French, and Arabic, including right-to-left layouts; responsive navigation and localized Contact/footer details.

Parapharmacies, supplies/accessories, medical materials, and About currently use placeholder pages. Directory coverage and location accuracy vary; estimated coordinates and simulated inventory are local testing fixtures.

## Stack

Next.js 15 / React 19 / TypeScript / Tailwind CSS, Go / PostgreSQL, Firebase Authentication, SMTP email, and LibreTranslate through Docker.

## Local setup

Install Node.js/npm, Go 1.24.4 or newer, and PostgreSQL. Docker is needed when using the local automatic translation service.

1. Use the existing project PostgreSQL database or restore its backup. Base schema files are in `server/database/migrations/`; the backend initializes the pickup/email tables on startup, but requires the base schema first.
2. Create `server/.env` with your local values:

   ```dotenv
   DB_URL=postgres://USER:PASSWORD@localhost:5432/qayen?sslmode=disable
   PORT=:4000
   JWT_SECRECT=REPLACE_WITH_A_LONG_RANDOM_SIGNING_KEY
   FIREBASE_PROJECT_ID=YOUR_FIREBASE_PROJECT_ID
   GOOGLE_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
   ```

   `JWT_SECRECT` is the spelling currently expected by the code. Keep environment files and sender credentials private.

3. Copy `web/.env.local.example` to `web/.env.local` and fill in the Firebase web app settings. Configure SMTP using `server/email-reminders.env.example` and the email guide below.
4. From the repository root, start translation:

   ```powershell
   docker compose -f docker-compose.translate.yml up -d
   ```

5. In a backend terminal:

   ```powershell
   cd server
   go run main.go
   ```

6. In a separate frontend terminal:

   ```powershell
   cd web
   npm install --legacy-peer-deps
   npm run dev
   ```

Open `http://localhost:3000`. The API runs at `http://localhost:4000/api/v1`. Start only one backend on port 4000. Restart the affected server after changing its environment settings.

For LAN testing, add the frontend's full origin to `GOOGLE_ALLOWED_ORIGINS` and its hostname/IP without a scheme or port to Firebase's authorized domains. Changing Wi-Fi addresses requires updating both. Google token verification also needs internet access and an accurate system clock.

## Alert and email testing

- A patient creates an alert for a medication and quantity; nearby pharmacies are contacted sequentially. Linked pharmacist accounts respond through their dashboard; unmanaged entries follow the existing stock-based test workflow.
- Pharmacist emails go to the account linked to the targeted pharmacy. A real recipient address and configured SMTP sender are required; demo addresses are not real inboxes.
- Patient confirmation holds stock for **5 minutes** locally. Reminders occur at **3 and 4 minutes just during the testing phase**. The pharmacist marks collection; otherwise expiry cancels the hold and restores its quantity.
- `QAYEN_PICKUP_HOLD_SECONDS=86400` changes the hold to 24 hours; reminder thresholds remain 60% and 80% of the configured duration.
- `EMAIL_REMINDERS_ENABLED=true` enables configured SMTP sending. Pharmacist emails share the sender unless `PHARMACIST_ALERT_EMAILS_ENABLED=false`.
- `QAYEN_ENABLE_APPROXIMATE_ALERTS=true` enables isolated estimated-location/inventory fixtures when prepared in the database. These fixtures are not verified stock or precise locations.

The backend must stay running for email delivery and expiry processing. Check the patient's or targeted pharmacist's own inbox, including Spam.

## Checks

Backend, from the repository root:

```powershell
go test ./server/...
```

Frontend, from `web`:

```powershell
npm run test:translations
npm run test:maps
npm run test:accounts
```

Database integration tests are opt-in; see the guides for environment flags. Automated email tests use fake sending. Existing TypeScript issues in `menu-bar.tsx` and `theme-provider.tsx` remain; Next.js configuration currently allows build-time type/lint errors, so a successful build alone is not a clean type check.

## Development guides

- [Google sign-in setup](GOOGLE_SIGN_IN.md)
- [Patient and pharmacist email setup/testing](EMAIL_REMINDERS.md)
- [Pickup reservation testing](PICKUP_TESTING.md)
- [Catalog import and data limitations](server/cmd/catalog-import/README.md)
- [Pharmacy address enrichment](server/cmd/pharmacy-addresses/README.md)

Catalog/location imports should be previewed and reviewed before applying changes. Local directory snapshots do not establish current nationwide completeness, verified GPS positions, or real pharmacy stock.
