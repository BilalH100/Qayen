interface Medication {
  id: number;
  status: string;
  commercial_status: string;
  speciality: string;
  dosage: string;
  form: string;
  presentation: string;
  pp: string;
  active_Substance: string;
  therapeutic_class: string;
  epi: string;
  ppv: string;
  ph: string;
  code: string;
  tva: string;
  createdAt: string;
  pfht: string;
  description: string;
  common_sd: string[];
  serious_sd: string[];
  general_info: string[];
}

export const ROLES = {
  REGULAR: "regular",
  PHARMACIST: "pharmacist",
  ADMIN: "admin",
} as const;

// The real seeded data stores commercial status as French free-text values
// like "Commercialisé", "Commercialisé / AO", "Commercialisé EXPORT",
// "Non Commercialisé", "Retiré du Marché", "Suspendu du Marché", etc.
// A medication is available whenever the value *starts with* "Commercialisé"
// (any suffix), and NOT when it starts with "Non Commercialisé".
export function isMedicationAvailable(commercialStatus?: string | null) {
  const normalized = (commercialStatus ?? "").trim().toLowerCase();
  return (
    normalized.startsWith("commercialisé") ||
    normalized.startsWith("commercialise") ||
    normalized === "active"
  );
}
