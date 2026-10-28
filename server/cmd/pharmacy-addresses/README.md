# Fill blank pharmacy addresses from public evidence

This command fills only blank addresses. It does not create pharmacies, guess GPS coordinates, populate stock, or claim that a pharmacy is independently verified or currently open.

## Sources and collection

`server/tools/collect_pharmacy_addresses.py` reads the public Pharmacy.ma directory sequentially, respecting its robots policy and waiting one second between requests. It caches each page with a SHA256 and retrieval timestamp. Re-running uses cached pages; use a new output directory for a fresh collection. The script uses only Python's standard library.

```powershell
python server/tools/collect_pharmacy_addresses.py --output-dir tmp/pharmacy-addresses-20261022
```

Source: https://pharmacy.ma/pharmacies

Methodology: https://pharmacy.ma/home/methodology

Pharmacy.ma describes its records as information assembled from public sources, directories, professionals and users. Collection time does not establish when the address was last verified. Its listings can be outdated. A directory match is not proof that a business currently exists, is licensed, is open, or has medicines in stock. Each saved address retains its source URL and is marked `business_verified=false`.

The government address workbook at data.gov.ma was also inspected, but its stated update date is 12 May 2016; it was not used alone to fill addresses. OpenStreetMap endpoints were unavailable and no OSM data was imported.

## Matching and preservation

Match only a unique pharmacy name and city, ignoring accents, punctuation, the leading word `Pharmacie`, and the trailing country label. The explicit city spelling aliases Fez/Fes and Tangier/Tanger are supported. There is no fuzzy name matching. Multiple database identities, multiple source profiles, conflicting addresses and unnamed source rows are excluded.

An address must include a numbered street/building/lot or a street with an explicit landmark description. A neighborhood, city, postcode or street name alone is held for review. This check measures address specificity; it does not independently verify the physical location. All existing nonblank addresses are preserved.

The tool requires a newly created, verified pg_dump archive before applying. Schema and address changes are committed together in a serializable transaction. Before committing, it compares all other pharmacy fields and hashes the medication and stock tables to confirm preservation. It records evidence in `pharmacy_address_evidence`. Re-running skips already populated addresses.

## Preview and apply

Run from the repository root; database configuration is read from `server/.env` without printing credentials.

```powershell
go run ./server/cmd/pharmacy-addresses -source tmp/pharmacy-addresses-20261022/directory-addresses.json -report tmp/pharmacy-addresses-20261022/preview.json
```

To apply, supply `-apply`, `-expected-sha256` from the reviewed report, `-backup` with a new destination, and optionally `-pg-dump` with the executable path. Reports, downloaded evidence and backups stay in ignored `tmp`.

The 22 October 2026 run collected 3,257 records from 131 pages and filled 1,347 blank addresses across 32 cities. The original 100 addresses stayed unchanged. There are now 1,447 recorded addresses and 7,341 missing addresses among 8,788 pharmacy records. The medication count remains 9,567 and stock count 291. Full decisions and city totals are in `tmp/pharmacy-addresses-20261022/applied.json`.

Restart the backend to expose `address_source_url` on individual pharmacy details. The detail page shows the source and asks users to confirm directory addresses with the pharmacy. If a pharmacist subsequently changes the address, the old source URL is not shown for the new address.

Tests: `go test ./server/cmd/pharmacy-addresses`. Set `CATALOG_TEST_DB_ENV=../../.env` to enable the PostgreSQL integration test, which writes only TEMP tables and rolls back.
