# Estimated locations in the local alert test workflow

The local test setup enables the 846 addressed pharmacies without recorded GPS to participate using their displayed city-area estimates. Together with the 601 recorded locations, 1,447 entries are eligible for nearby alert matching in test mode. These estimates describe a city area, not the physical pharmacy door.

## Activation and isolation

The activation command adds only `catalog_test_locations`, `catalog_test_stock`, `catalog_test_settings` and `catalog_test_alerts`, plus two effective-data views. Original pharmacy latitude/longitude, medication rows, user accounts and stock are preserved. Both `QAYEN_ENABLE_APPROXIMATE_ALERTS=true` in the backend environment and the database test setting must be enabled. The backend sets `qayen.test_alerts=on` on its database connections only with this environment flag. Other sessions do not see estimated locations or simulated stock through the views.

Restart the backend after activation. To disable this feature, set `QAYEN_ENABLE_APPROXIMATE_ALERTS=false` in `server/.env` and restart. The normal workflow then uses recorded GPS and original stock. Archived test alerts explicitly require test mode when retrieving results.

The prepared input is `tmp/estimated-alert-testing-20261026/locations.json`, generated from the same city-area grouping helper used by the maps. Preview from the repository root:

```powershell
go run ./server/cmd/estimated-alert-testing -locations tmp/estimated-alert-testing-20261026/locations.json -report tmp/estimated-alert-testing-20261026/preview.json
```

Applying requires `-apply`, the reviewed input's `-expected-sha256`, a new `-backup` archive and optionally `-pg-dump` with the executable path. A full database archive is verified before writes. The command checks every pharmacy's current address/city and missing GPS, then verifies hashes of the original catalog and inventory before committing. Existing test stock is not reset on repeat activation. Prepared environment changes retain existing credentials without printing them. Reports and backups stay in ignored `tmp`.

## Test stock and responses

Each of the 846 test pharmacies receives 20 simulated units of each existing test SKU, stored separately:

| Medication ID | Name | Presentation |
| --- | --- | --- |
| 540 | Codoliprane | Boite de 16 |
| 1526 | Rhumix | Boite de 10 |
| 4160 | Rinomicine | Boite de 10 |

This produces 2,538 isolated stock rows. Existing real stock takes precedence for the same pharmacy/medication pair, including zero quantities. Other medicines do not automatically gain inventory.

Pharmacies without linked accounts follow the existing automatic stock response path. A linked pharmacist retains the dashboard/SSE/manual response path; assign the existing pharmacist through the admin tools to test a specific entry. No extra user accounts or passwords were created. Manual and automatic responses enforce the requested quantity using effective stock. Pickup decrements the simulated table when the response used test stock; real stock is otherwise handled by the existing atomic decrement.

The home map's amber popup has **Use this city area as my search location**, which selects that estimate for the patient's search location. Results explicitly label estimated locations and simulated stock. Navigation to an estimated pharmacy pin is disabled. The map still distinguishes recorded pins from approximate city-area groups; no estimated values were written into recorded GPS columns.

## Verification on 26 October 2026

- Backend suite passed: `go test ./server/...`.
- PostgreSQL integration passed with TEMP tables only: `CATALOG_TEST_DB_ENV=../.env go test ./server/services -run TestEstimatedAlertWorkflowWithIsolatedInventory`. It covers automatic response, manual dashboard/response, pickup, preserving real stock/GPS, session gating and the database kill-switch.
- Map and translation tests passed. The pharmacist dashboard's Next.js params type was corrected for the manual workflow. Other existing frontend type errors in menu-bar and theme-provider remain outside this change.
- Applied 846 test locations and 2,538 test stock rows. Backup and report: `tmp/estimated-alert-testing-20261026/before-testing.dump` and `applied.json`.
- Live database smoke test created test alert 409 for the demo patient near Marrakech. Pharmacy 4678 answered automatically; pickup completed and changed only its simulated Codoliprane quantity from 20 to 19. The original catalog/stock hash stayed identical. Details: `tmp/estimated-alert-testing-20261026/smoke.json`.
- Effective counts: 1,447 locations with test mode enabled, 601 with it disabled; original stock remains 291 rows.
