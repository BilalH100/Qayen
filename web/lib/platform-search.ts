export interface SearchResult {
  id: string;
  title: string;
  description: string;
  href: string;
  category: "Medications" | "Pharmacies" | "Sections";
  keywords?: string;
}

export function normalizeSearch(value: string) {
  return value.normalize("NFKD").replace(/\p{M}/gu, "").toLowerCase()
    .replace(/[أإآٱ]/g, "ا").replace(/ى/g, "ي").replace(/[^\p{L}\p{N}]+/gu, " ").trim();
}

export const searchableSections: SearchResult[] = [
  { id: "medications", title: "Medications", description: "Find medications and medication information.", href: "/medications", category: "Sections", keywords: "medicine medicament medicaments médicaments médicament drugs ingredients dosage price médicaments الأدوية دواء ادوية مكونات جرعة سعر" },
  { id: "pharmacies", title: "Pharmacies", description: "Find pharmacies and addresses across Morocco.", href: "/pharmacies", category: "Sections", keywords: "pharmacy pharmacie map carte location address adresse Maroc Morocco صيدلية صيدليات الصيدليات خريطة عنوان المغرب" },
  { id: "parapharmacies", title: "Parapharmacies", description: "Coming soon", href: "/parapharmacies", category: "Sections", keywords: "para parapharmacie health beauty personal care santé beauté soins شبه صيدلية شبه الصيدليات تجميل عناية" },
  { id: "supplies", title: "Supplies & Accessories", description: "Coming soon", href: "/supplies-accessories", category: "Sections", keywords: "supplies accessories fournitures accessoires المستلزمات والملحقات مستلزمات ملحقات" },
  { id: "materials", title: "Medical Materials", description: "Coming soon", href: "/medical-materials", category: "Sections", keywords: "medical equipment materials matériel médical equipement المعدات الطبية معدات أجهزة" },
  { id: "about", title: "About", description: "Coming soon", href: "/about", category: "Sections", keywords: "Qayen about propos nous من نحن قايِن" },
  { id: "contact", title: "Contact", description: "Coming soon", href: "/contact", category: "Sections", keywords: "contact contacter email telephone اتصل بنا تواصل اتصال" },
];

// All query words must match the same item; accents and punctuation are ignored.
export function searchPlatform(items: SearchResult[], query: string): SearchResult[] {
  const normalized = normalizeSearch(query);
  const words = normalized.split(" ").filter(Boolean);
  if (!words.length) return [];
  return items.flatMap(item => {
    const title = normalizeSearch(item.title);
    const text = normalizeSearch(`${item.title} ${item.description} ${item.keywords ?? ""}`);
    if (!words.every(word => text.includes(word))) return [];
    const score = title === normalized ? 0 : title.startsWith(normalized) ? 1 : title.includes(normalized) ? 2 : 3;
    return [{ item, score }];
  }).sort((a, b) => a.score - b.score || a.item.title.localeCompare(b.item.title) || a.item.id.localeCompare(b.item.id))
    .map(({ item }) => item);
}
