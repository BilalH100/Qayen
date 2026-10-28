# Supervisor reference import

The importer reads the validated JSON produced by `server/tools/prepare_supervisor_catalog.py` from the two supplied workbooks. The originals are never modified. The converter requires openpyxl and xlrd (the latter can be supplied through `--xls-reader-dir`). Other supplier/manufacturer/veterinary files are outside this import.

Run from the repository root:

```powershell
go run ./server/cmd/supervisor-import -bundle tmp/supervisor-import-20261022/bundle.json -report tmp/supervisor-import-20261022/preview.json
```

Preview is read-only. Applying requires `-apply`, `-expected-sha256` matching the reviewed bundle, `-backup` pointing to a new archive, and optionally `-pg-dump` pointing to the PostgreSQL executable. A full custom-format backup is created and checked with pg_restore before any writes. The import and its schema changes use one serializable transaction. Reports and backups contain database/source information and belong in ignored `tmp`, not source control.

CNOPS records are stored in `medication_reference_entries` with source hashes, row numbers, prices and reimbursement data. Only exact name, ingredient, strength, form and packaging matches link to existing medications. Unknown or multicomponent strengths remain unmatched. Prices never overwrite the existing medication catalog; the source does not establish a current effective date. Changed existing source codes are held for review.

Pharmacy names and cities enter the directory once per normalized identity. Missing-city and ambiguous rows remain in `pharmacy_directory_entries` for review. Unknown address/phone strings stay blank for compatibility with existing SQLC models; coordinates stay NULL. No stock, user accounts, alerts, contact details or existing pharmacy records are created or altered. Directory presence does not confirm opening hours, operating status or stock.

The 2026-10-22 import added 5,917 CNOPS references (6 exact links; 5,911 requiring review) and 8,688 pharmacies. All 8,781 pharmacy source rows are retained, including 34 without a city. Existing totals of 9,567 medications and 291 stock rows were preserved. There are now 8,788 pharmacy directory records. Detailed provenance, price differences and backup location are in `tmp/supervisor-import-20261022/applied.json`. A second preview confirmed zero new insertions.

After restarting the backend, linked medication details expose `references` for the reference panel. Pharmacy detail pages fetch real records and show missing information explicitly. Reconciliation queries:

```sql
SELECT * FROM medication_reference_entries WHERE match_status <> 'matched';
SELECT * FROM pharmacy_directory_entries WHERE pharmacy_id IS NULL;
```

Run the import tests with `CATALOG_TEST_DB_ENV=../../.env` set for `go test ./server/cmd/supervisor-import`. Integration tests create TEMP tables and roll back; they do not modify permanent data. Keep the backup for recovery; a full restore replaces database state and should be performed only against an explicitly chosen recovery database.
