# Import the AMMPS catalog

Run commands from the repository root. The tool reads `DB_URL` from `server/.env`
without printing credentials. It uses the existing AMMPS JSON export schema in
`server/database/data/medications.json` and `pharmacies.json`.

## Preview

```powershell
go run ./server/cmd/catalog-import -medications server/database/data/medications.json -pharmacies server/database/data/pharmacies.json -report tmp/catalog-import-report.json
```

The default is a **read-only transaction**. Review insertion/update counts,
missing fields, conflicting products, duplicate records, and missing locations
in the JSON report. Existing source files are historical snapshots, not a claim
of current nationwide completeness. Obtain refreshed exports from AMMPS when
available and use their paths with the same flags.

## Apply a reviewed export

```powershell
go run ./server/cmd/catalog-import -medications server/database/data/medications.json -pharmacies server/database/data/pharmacies.json -report tmp/catalog-import-applied.json -apply
```

The import commits in one transaction. Medication matches use name, strength,
form, packaging and manufacturer, preserving existing IDs and clinical fields.
Conflicting identities/prices and duplicate matches are reported for review.
The tool never deletes records or creates stock quantities.

Pharmacy directory entries are matched on name and city, with normalization for
existing `Rabat - Maroc` city labels. New entries have no address, coordinates or
inventory. Their location must be supplied and checked by the pharmacy; they
cannot participate in location/stock-based matching until those fields exist.
Repeated name/city rows cannot prove how many distinct physical pharmacies
exist and require address-level verification.

Assign a pharmacist account through the existing admin dashboard. That pharmacist
can edit the exact address, city, phone and map pin on `/pharmacy`, then maintain
real quantities in **My Stock**. Saving a map pin is a user correction, not
independent verification of the pharmacy's physical location or authority.
Existing demo stock must be replaced with actual inventory before public use.
