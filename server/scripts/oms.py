from openpyxl import Workbook

# Create a new workbook and select the active worksheet
wb = Workbook()
ws = wb.active
ws.title = "WHO Prequalified Labs"

# Define headers
headers = ["Pays", "Laboratoire de contrôle de médicaments", "Date de pré-qualification", "Région"]
ws.append(headers)

# Add all data rows (combined from both images)
data = [
    # Previous data
    ["Afrique du sud", "Mardi 1", "18.07.2017", "AFRO"],
    ["Afrique du sud", "Adcock R-D", "15.01.2008", ""],
    ["Afrique du sud", "BIP", "05.07.2005", ""],
    ["Allemagne", "Impha", "01.04.2010", "EURO"],
    ["Bangladesh", "NCL", "16.03.2020", "SEARO"],
    ["Belarus", "RCAI", "21.06.2012", "EURO"],
    ["Belgique", "SCALDGO", "30.10.2015", "EURO"],
    ["Belgique", "LUG", "13.12.2016", ""],
    ["Belgique", "SGS", "31.05.2011", ""],
    ["Brésil", "LACEN-GO", "26.07.2018", "AMRO"],
    ["Brésil", "FUNED", "20.10.2011", ""],
    ["Brésil", "INCOS", "11.03.2014", ""],
    ["Canada", "KARS", "10.02.2010", "AMRO"],
    ["Chine", "MFBC", "20.11.2011", "WPRO"],
    ["Chine", "SZDC", "01.05.2018", ""],
    ["Colombie", "HIVIMA", "07.09.2020", "AMRO"],
    ["Croatie", "HALMED ONCE", "16.06.2016", "EURO"],
    ["France", "Gimopham", "26.07.2018", "EURO"],
    ["France", "APTYS", "26.07.2018", ""],
    ["France", "CHMP", "08.10.2008", ""],
    ["Ghana", "CDA_OCL", "28.11.2021", "AFRO"],
    ["Inde", "SGS", "02.05.2023", "SEARO"],
    ["Inde", "VIMTA", "17.07.2008", ""],
    ["Inde", "ILL", "15.06.2015", ""],
    ["Indonésie", "SPIRA", "07.09.2020", ""],
    ["Iran", "NGCDF", "10.12.2019", "EMRO"],
    ["Jordanie", "FOL", "01.03.2016", "EMRO"],
    ["Hongrie", "RSE", "14.03.2020", "EURO"],
    ["Kazakhstan", "NOCL", "17.07.2008", "EURO"],
    ["Kenya", "MEDS", "23.03.2009", "AFRO"],
    ["Liban", "Awara OCI", "16.03.2020", "EMRO"],
    ["Maroc", "LINCM", "17.07.2008", "EMRO"],
    ["Mexique", "CCVAC", "13.11.2015", "AMRO"],
    ["Nigeria", "IMPDAC OCI", "30.10.2023", "AFRO"],
    ["Pakistan", "DTK Mullan", "29.03.2023", "EMRO"],
    ["Pakistan", "PDTRC", "03.07.2019", ""],
    ["Pakistan", "DTLF", "16.03.2020", ""],
    ["Pakistan", "DTL Rewalpindi", "08.11.2022", ""],
    
    # New data from PAGE 1
    ["OTL Punjab labore", "", "30.10.2023", ""],
    ["Pays-Bas", "SBU", "23.09.2014", "EURO"],
    ["Portugal", "IMEANKED", "13.08.2011", "EURO"],
    ["Portugal", "Laboratorio Basi", "12.06.2013", ""],
    ["Fédération de Russie", "SCEEMP Rotter", "11.03.2014", "EURO"],
    ["Singapour", "TUV SUD PSB", "21.05.2009", "WPRO"],
    ["Stamm", "Intertek", "27.10.2014", "EURO"],
    ["Tanzanie", "TMDA-GCL", "17.01.2011", "AFRO"],
    ["Thaïlande", "BDI", "02.11.2012", "SEARO"],  # Corrected SFARO to SEARO
    ["Urania", "MDA-GCL", "22.01.2015", "AFRO"],
    ["Ukraine", "CLCCM", "16.04.2010", "EURO"],
    ["Uruguay", "CCCM", "18.09.2010", "AMRO"],
    ["Viet Nam", "NICOLE", "28.11.2008", "WPRO"],
    ["Zimbabwe", "MCAC-GCL", "13.09.2014", "AFRO"]
]

# Add all data rows to the worksheet
for row in data:
    ws.append(row)

# Save the workbook
wb.save("WHO_Prequalified_Medicine_Labs_Combined.xlsx")
print("Excel file 'WHO_Prequalified_Medicine_Labs_Combined.xlsx' has been created successfully!")