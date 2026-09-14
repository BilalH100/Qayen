from openpyxl import Workbook
from openpyxl.styles import Font

wb = Workbook()
ws = wb.active
ws.title = "Autorités de Régulation"

headers = ["Pays", "Autorité de régulation", "Statut OMS"]
ws.append(headers)
for cell in ws[1]:
    cell.font = Font(bold=True)

data = [
    ["Argentine", "ANMAT", "D"],
    ["Australie", "TGA", "B CE"],
    ["Belgique", "FAMPH", "B CE"],
    ["Brésil", "ANYSA", "D E"],
    ["Bulgarie", "BOA", "B E"],
    ["Canada", "Health Canada", "B CD E"],
    ["Chine", "NMPA", "E"],
    ["Cuba", "CECMEO", "D E"],
    ["Danemark", "DKMA", "B E"],
    ["Egypte", "EDA", "A (ML3)"],
    ["France", "ANSM", "B CE"],
    ["Allemagne", "PEI", "C E"],
    ["Inde", "COSCO", "A (ML3)"],
    ["Indonésie", "BAJAM POM", "A (ML3)"],
    ["Iran (République Islamique de l')", "IFDA", "E"],
    ["Italie", "AIFA", "B CE"],
    ["Japon", "PMDA", "B CE"],
    ["Mexique", "COFEPRIS", "B E"],
    ["Pays-Bas", "MEB", "D E"],
    ["Fédération de Russie", "Ministry of Health", "B CE"],
    ["Serbie", "ALIMS", "E"],
    ["Suède", "MPA", "A (ML3)"],
    ["Thaïlande", "Thai-FDA", "B E"],
    ["Royaume-Uni", "MHRA", "A (ML3)"],
    ["États-Unis", "FDA", "B CE"],
    ["Viet Nam", "DAV", "B CD E"],
    ["Réseau européen", "EMA", "A (ML3)"]
]

for row in data:
    ws.append(row)

ws.column_dimensions['A'].width = 30  # Pays
ws.column_dimensions['B'].width = 30  # Autorité
ws.column_dimensions['C'].width = 15  # Statut

ws.append([])
ws.append(["Légende:"])
ws.append(["A : ML3/ML4 ANR (médicaments et/ou vaccins)"])
ws.append(["B : Autorités réglementaires strictes (SSA) (médicaments)"])
ws.append(["C : Autorité ANR fonctionnelle (vaccins)"])

wb.save("Autorites_Regulation_Pharmaceutique.xlsx")
print("Fichier Excel créé avec succès : 'Autorites_Regulation_Pharmaceutique.xlsx'")
