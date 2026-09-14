from openpyxl import Workbook

wb = Workbook()
ws = wb.active
ws.title = "Bioequivalence Centers"

headers = ["NOM", "NOM DIRECTEUR", "ADRESSE", "TEL", "FAX", "TYPE", "ACTIVITE"]
ws.append(headers)

data_rows = [
    [
        "Centre des Etudes de Bioéquivalence Cheikh Zaid",
        "P. VAHIA CHEERAH",
        "BP 4533, Avenue Alidi El Farsi/Nadimir Ali Irène, Hay Road, RABAT",
        "003 37 68 68 66",
        "003 37 77 01 06",
        "",
        ""
    ],
    [
        "Centre Monormont VI des Etudes de Bioéquivalence",
        "P. LANCEN",
        "Campus arria city de la fondation, Batiment B, Serne étage, CASABLANCA",
        "003 29 08 91 15",
        "003 29 08 92 55",
        "",
        ""
    ]
]

for row in data_rows:
    ws.append(row)

wb.save("Bioequivalence_Centers.xlsx")
print("Excel file 'Bioequivalence_Centers.xlsx' has been created.")