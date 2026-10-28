# Published pharmacy map locations

This command adds published pharmacy POI coordinates only when a unique name and city match the existing pharmacy and its saved address is corroborated. It never geocodes a street or city centre, invents a pin, changes an existing GPS coordinate, or creates inventory. Published points are not independently verified on site.

## Collection

Source: https://pharmap.ma/ and https://pharmap.ma/sources

`server/tools/collect_pharmacy_locations.py` reads public profile URLs advertised in the site's sitemap whose names match addressed pharmacies lacking GPS. It respects robots rules, uses a single connection, limits requests to 2.5 per second, caches responses, and stops on HTTP 403/429. No provider API key is used. Structured pharmacy name, city, address, telephone and geo fields are extracted; a map link containing an address search is not accepted as coordinates.

The collector requires the prior pharmacy API snapshot (`{"pharmacies":[...]}`), plus downloaded `pharmap-robots.txt` and `pharmap-sitemap.xml` in its output directory:

```powershell
python server/tools/collect_pharmacy_locations.py --directory tmp/pharmacy-gps-20261022 --pharmacies tmp/pharmacy-gps-20261022/pharmacies-before.json
```

The 22 October 2026 run collected all 1,879 candidate profiles. Eighteen transient failures were retried successfully from cache. Each record retains its public source URL, page SHA256 and capture time. Capture time is not proof of the source's last on-site verification. The source did not cover every pharmacy or city in the supplied database.

## Matching

Names and cities use explicit normalization of accents, punctuation and common spelling aliases. Multiple branches with the same name/city are held. An accepted point requires one of:

- An equivalent full address after normalizing abbreviations and articles.
- The same numbered street, with additional district detail at a complete token boundary.
- The same business telephone, explicit street number and at least two meaningful address terms. A telephone from the prior address reference is usable only when that reference matches the current saved address.

Changed building numbers, conflicting addresses, missing addresses, shared coordinates for different identities, incomplete points, out-of-region coordinates and coarse rounded positions are held. Existing or partially populated coordinates remain unchanged. A statistical review found no accepted point over 40 km from the median of other accepted points in the same city (for cities with at least three matches); medians are never saved as locations.

## Preview and apply

From the repository root:

```powershell
go run ./server/cmd/pharmacy-locations -source tmp/pharmacy-gps-20261022/published-locations.json -address-reference tmp/pharmacy-addresses-20261022/directory-addresses.json -report tmp/pharmacy-gps-20261022/preview.json
```

Preview is read-only. Applying requires `-apply`, the completed collection, a new `-backup` destination, and `-expected-sha256` matching the combined source/reference hash in the reviewed report. Set `-pg-dump` to the executable path when needed. A full database backup is created and checked using pg_restore before writes. One serializable transaction stores coordinates and `pharmacy_location_evidence`, marking `business_verified=false`. All other pharmacy columns, medication rows and inventory are compared before commit. Reports, cached evidence and backups stay in ignored `tmp`.

The applied run added **506 locations across 27 cities**, preserving the original **95 locations**. The live API now exposes **601 pharmacies with coordinates**. Of 1,447 addressed pharmacies, **846 still lack coordinates**: 400 had no matching published point, 369 had uncorroborated addresses, 74 shared points with other identities, and 3 had ambiguous source identities. There remain 8,788 pharmacies, 9,567 medications and 291 stock rows. A second preview added zero locations. Reports: `tmp/pharmacy-gps-20261022/applied.json` and `after-verification.json`.

## Map display and checks

### Optional city-area approximation (26 October 2026)

At the user's request, the home map and directory map now also show amber city-area groups for addressed pharmacies lacking coordinates. A group lists each pharmacy and its recorded address; it does not represent an individual pharmacy pin. Estimates use the median of recorded locations in the same city, or a linked GeoNames city point for Marrakech, Dakhla, Azemmour, Bouznika and Youssoufia. Searches retain the complete directory as the reference so filtering does not move an area estimate.

The database check found 601 recorded GPS locations and all 846 remaining addressed pharmacies represented in 30 approximate city groups. The approximation code does not update pharmacy latitude/longitude, return estimates through the GPS validator, generate navigation links, or provide coordinates to stock/distance/alert matching. City groups are explicitly labelled and open at city zoom when selected from the home search. Map and translation tests passed; full TypeScript checking still reports only the previously listed unrelated errors. No additional GPS records were imported for this display feature.

The directory's Map View now displays valid coordinates from the API, fits the filtered set, and opens pharmacy details from each popup. The home map uses the same strict coordinate validation. Individual pharmacy details include a map. Restart the backend to expose `location_source_url`, which links to matching evidence only while the saved coordinates and address remain unchanged. Refresh the page to fetch the new pins.

Checks: `go test ./server/cmd/pharmacy-locations` (set `CATALOG_TEST_DB_ENV=../../.env` for rollback-only TEMP-table integration), `npm run test:maps`, and `npm run test:translations` in `web`. Directory and detail HTTP routes returned 200. Full TypeScript checking still reports the pre-existing dashboard, menu-bar and theme-provider errors; no errors were reported in the changed map files. Visual browser checking was unavailable because no browser was connected.
