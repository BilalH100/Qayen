# Google sign-in for Qayen

## Current integration: Firebase popup (localhost and LAN)

The Google button uses Firebase Authentication's `signInWithPopup`, then sends the Firebase ID token to Qayen. The backend verifies the signature, project, issuer, expiry, verified email and Google provider, then issues the normal Qayen session. Firebase uses temporary in-memory authentication; account data, roles and alerts remain in PostgreSQL. Google provider IDs preserve existing Qayen Google accounts. Existing password accounts are never linked based only on a matching email.

Firebase project: `qayen-2c858`. Enable Google under **Authentication → Sign-in method**. Under **Authentication → Settings → Authorized domains**, authorize `localhost`, `127.0.0.1`, and `192.168.3.114` as needed (no scheme or port). The supplied project already lists `localhost` and `192.168.3.114`.

Frontend settings in `web/.env.local`:

```dotenv
NEXT_PUBLIC_FIREBASE_API_KEY=YOUR_FIREBASE_WEB_API_KEY
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=qayen-2c858.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=qayen-2c858
NEXT_PUBLIC_FIREBASE_APP_ID=YOUR_FIREBASE_WEB_APP_ID
```

Backend settings in `server/.env`:

```dotenv
FIREBASE_PROJECT_ID=qayen-2c858
GOOGLE_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,http://192.168.3.114:3000
```

Run `npm install --legacy-peer-deps` in `web` if dependencies are missing. Run `go run ./server/cmd/google-auth-setup` from the project root if the Google identity column has not been added. Restart both servers, open `http://192.168.3.114:3000`, and click **Continue with Google**. Allow the popup and choose an account. A real account login needs to be tested interactively; no real Google account is simulated by the automated tests.

If the laptop's network IP changes, update both Firebase's authorized domains and `GOOGLE_ALLOWED_ORIGINS`. Use HTTPS for deployment. No service-account key, analytics, storage, or Firebase database is needed for this integration.

References: [Firebase Google popup](https://firebase.google.com/docs/auth/web/google-signin), [Firebase token verification](https://firebase.google.com/docs/auth/admin/verify-id-tokens).

## Legacy direct Google Identity Services integration

The old `/users/google` endpoint remains compatible with previously issued Google ID tokens, but the UI now uses Firebase. The following setup applies only to that older direct flow:

Google Identity Services handles the account picker; Qayen verifies the signed ID token on the backend and starts its usual session. No client secret or additional npm/Go dependency is needed.

1. In [Google Cloud Console](https://console.cloud.google.com/), select/create your project. Under **Google Auth Platform**, configure Branding, Audience (add test users while in Testing), and create a **Web application** client under Clients.
2. Add authorized JavaScript origins `http://localhost:3000` and `http://127.0.0.1:3000`. Add your real HTTPS origin before deployment. This popup integration does not need a redirect URI.
3. Add to `web/.env.local`:

   ```dotenv
   NEXT_PUBLIC_GOOGLE_CLIENT_ID=YOUR_CLIENT_ID.apps.googleusercontent.com
   ```

4. Add to `server/.env` (same client ID):

   ```dotenv
   GOOGLE_CLIENT_ID=YOUR_CLIENT_ID.apps.googleusercontent.com
   GOOGLE_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
   ```

5. From the project root, run `go run ./server/cmd/google-auth-setup` once. This adds only the Google identity column and unique index. Alternatively, apply migration `20261101000000_add_google_sign_in.sql` with your migration runner; the setup command is idempotent and does not update the Goose version table.
6. Restart backend and frontend. Open `http://localhost:3000` and select **Continue with Google** in either Sign In or Create Account. First use creates a regular account; subsequent uses sign in to that account. Google does not supply a phone number; it remains blank until provided separately.

Without a client ID, the Google option is disabled with an explanatory message. Existing password login remains available. An existing account with the same email requires its password: automatic email-based account linking is deliberately disabled. Linking accounts can be added later through an authenticated profile flow.

For deployment use HTTPS, list the exact frontend origin in `GOOGLE_ALLOWED_ORIGINS` and the Google console, and set `GOOGLE_COOKIE_SECURE=true` behind an HTTPS reverse proxy. Use the same site for frontend/backend so the short-lived nonce cookie is sent. Plain HTTP LAN addresses such as `http://192.168.x.x:3000` are not a supported Google sign-in development origin; test on localhost or an HTTPS development domain. Never put a client secret in frontend environment variables.

Verification: `go test ./server/utils ./server/http/handlers ./server/repository` and the existing frontend translation/account tests. A real Google login requires the client ID and an authorized test account.

References: [Google setup](https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid), [token verification](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token), [JavaScript API](https://developers.google.com/identity/gsi/web/reference/js-reference).
