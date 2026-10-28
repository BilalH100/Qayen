export function pharmacyCoordinates(pharmacy: {
  latitude?: string | number | null;
  longitude?: string | number | null;
}): [number, number] | null {
  const { latitude, longitude } = pharmacy;
  if (latitude === null || latitude === undefined || longitude === null || longitude === undefined) return null;
  if (String(latitude).trim() === "" || String(longitude).trim() === "") return null;
  const lat = Number(latitude);
  const lng = Number(longitude);
  if (!Number.isFinite(lat) || !Number.isFinite(lng) || Math.abs(lat) > 90 || Math.abs(lng) > 180 || (lat === 0 && lng === 0)) return null;
  return [lat, lng];
}

export interface AreaPharmacy {
  id: number;
  name: string;
  address: string;
  city?: string;
  latitude?: string | number | null;
  longitude?: string | number | null;
}

export interface ApproximatePharmacyArea {
  city: string;
  point: [number, number];
  pharmacies: AreaPharmacy[];
  sourceUrl?: string;
}

function cityKey(city: string): string {
  return city.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toUpperCase()
    .replace(/\s*[-,]?\s*MAROC$/, "").replace(/[^A-Z0-9]/g, "");
}

// Gazetteer city points, used only for labelled city-area groups. Never
// return these through pharmacyCoordinates or save them as pharmacy GPS.
const cityAreas: Record<string, { point: [number, number]; sourceUrl: string }> = {
  MARRAKECH: { point: [31.63416, -7.99994], sourceUrl: "https://www.geonames.org/2542997/marrakesh.html" },
  DAKHLA: { point: [23.68477, -15.95798], sourceUrl: "https://www.geonames.org/2463447/dakhla.html" },
  AZEMMOUR: { point: [33.289518, -8.342498], sourceUrl: "https://www.geonames.org/2556657/azemmour.html" },
  BOUZNIKA: { point: [33.78942, -7.15968], sourceUrl: "https://www.geonames.org/2553751/bouznika.html" },
  YOUSSOUFIA: { point: [32.246342, -8.529407], sourceUrl: "https://www.geonames.org/search.html?country=MA&q=&startRow=250" },
};

export function approximatePharmacyAreas(visible: AreaPharmacy[], reference: AreaPharmacy[] = visible): ApproximatePharmacyArea[] {
  const known = new Map<string, [number, number][]>();
  for (const pharmacy of reference) {
    const point = pharmacyCoordinates(pharmacy);
    const key = cityKey(pharmacy.city || "");
    if (point && key) known.set(key, [...(known.get(key) || []), point]);
  }
  const median = (values: number[]) => {
    const sorted = [...values].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
  };
  const areas = new Map<string, ApproximatePharmacyArea>();
  for (const pharmacy of visible) {
    if (pharmacyCoordinates(pharmacy) || !pharmacy.address.trim()) continue;
    const key = cityKey(pharmacy.city || "");
    if (!key) continue;
    const fallback = cityAreas[key];
    const points = known.get(key);
    if (!fallback && !points?.length) continue;
    if (!areas.has(key)) {
      areas.set(key, {
        city: pharmacy.city || "",
        point: fallback ? [...fallback.point] : [median(points!.map(p => p[0])), median(points!.map(p => p[1]))],
        sourceUrl: fallback?.sourceUrl,
        pharmacies: [],
      });
    }
    areas.get(key)!.pharmacies.push(pharmacy);
  }
  return [...areas.values()];
}
