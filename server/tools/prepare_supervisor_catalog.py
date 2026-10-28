"""Read the supervisor's workbooks into a validated, traceable import bundle.

Requires openpyxl and xlrd. Original workbooks are never edited.
"""
import argparse
import hashlib
import json
import sys
from datetime import datetime, timezone
from decimal import Decimal, InvalidOperation
from pathlib import Path


def text(value):
    return "" if value is None else str(value).strip()


def amount(value, row, field):
    if value is None or text(value) == "":
        return None
    try:
        number = Decimal(str(value))
    except InvalidOperation as exc:
        raise ValueError(f"Row {row}: invalid {field}") from exc
    if not number.is_finite() or number < 0:
        raise ValueError(f"Row {row}: invalid {field}")
    return format(number, "f")


def source(path, sheet):
    return {"filename": path.name, "sha256": hashlib.sha256(path.read_bytes()).hexdigest(), "sheet": sheet}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--medications", type=Path, required=True)
    parser.add_argument("--pharmacies", type=Path, required=True)
    parser.add_argument("--xls-reader-dir", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.xls_reader_dir:
        sys.path.insert(0, str(args.xls_reader_dir))
    import openpyxl
    import xlrd

    book = openpyxl.load_workbook(args.medications, read_only=True, data_only=True)
    sheet = book["Réf. des médicaments"]
    iterator = sheet.iter_rows(values_only=True)
    headers = next(iterator)
    required = {"CODE", "NOM", "DCI1", "DOSAGE1", "UNITE_DOSAGE1", "FORME", "PRESENTATION", "PPV", "PH", "PRIX_BR", "PRINCEPS_GENERIQUE", "TAUX_REMBOURSEMENT"}
    if set(headers) != required:
        raise ValueError("Unexpected CNOPS columns; review the source before importing")
    medications = []
    codes = set()
    for row_number, values in enumerate(iterator, 2):
        if all(value is None for value in values):
            continue
        row = dict(zip(headers, values))
        code = text(row["CODE"])
        if isinstance(row["CODE"], (int, float)):
            numeric_code = Decimal(str(row["CODE"]))
            if numeric_code != numeric_code.to_integral_value():
                raise ValueError(f"CNOPS row {row_number}: fractional product code")
            code = str(int(numeric_code))
        if not code.isdigit() or len(code) != 13 or code in codes:
            raise ValueError(f"CNOPS row {row_number}: invalid or duplicate product code")
        if any(not text(row[field]) for field in ("NOM", "DCI1", "FORME", "PRESENTATION")):
            raise ValueError(f"CNOPS row {row_number}: missing required product identity")
        codes.add(code)
        medications.append({
            "source_row": row_number, "code": code, "name": text(row["NOM"]),
            "ingredient": text(row["DCI1"]), "dosage": text(row["DOSAGE1"]),
            "dosage_unit": text(row["UNITE_DOSAGE1"]), "form": text(row["FORME"]),
            "presentation": text(row["PRESENTATION"]),
            "public_price": amount(row["PPV"], row_number, "PPV"),
            "hospital_price": amount(row["PH"], row_number, "PH"),
            "reimbursement_base": amount(row["PRIX_BR"], row_number, "PRIX_BR"),
            "product_type": text(row["PRINCEPS_GENERIQUE"]),
            "reimbursement_rate": text(row["TAUX_REMBOURSEMENT"]),
        })
    medication_source = source(args.medications, sheet.title)
    book.close()

    book = xlrd.open_workbook(str(args.pharmacies), on_demand=True)
    sheet = book.sheet_by_name("Pharmacies Maroc")
    headers = sheet.row_values(3)
    if headers[:4] != ["N°", "Nom de la pharmacie", "Ville", "Province / précision"]:
        raise ValueError("Unexpected pharmacy columns; review the source before importing")
    pharmacies = []
    for index in range(4, sheet.nrows):
        values = sheet.row_values(index)
        if not text(values[1]):
            continue
        pharmacies.append({"source_row": index + 1, "name": text(values[1]), "city": text(values[2]), "province": text(values[3])})
    pharmacy_source = source(args.pharmacies, sheet.name)
    book.release_resources()
    if not medications or not pharmacies:
        raise ValueError("An empty source cannot be imported")
    bundle = {"prepared_at": datetime.now(timezone.utc).isoformat(), "medication_source": medication_source, "pharmacy_source": pharmacy_source, "medications": medications, "pharmacies": pharmacies}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(bundle, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"medications": len(medications), "pharmacies": len(pharmacies), "pharmacies_missing_city": sum(not p["city"] for p in pharmacies), "bundle": str(args.output)}))


if __name__ == "__main__":
    main()
