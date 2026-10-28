"use client";

import Link from "next/link";
import { CircleMarker, Popup } from "react-leaflet";
import type { ApproximatePharmacyArea } from "@/lib/pharmacy-location";

export function ApproximatePharmacyAreas({ areas, selectedId, onUseArea }: { areas: ApproximatePharmacyArea[]; selectedId?: number | null; onUseArea?: (point: [number, number]) => void }) {
  return <>{areas.map(area => (
    <CircleMarker key={area.city} center={area.point} radius={area.pharmacies.some(p => p.id === selectedId) ? 24 : 17} bubblingMouseEvents={false}
      pathOptions={{ color: "#92400e", weight: 2, dashArray: "4 4", fillColor: "#f59e0b", fillOpacity: 0.45 }}>
      <Popup>
        <strong>{area.city}</strong><br />
        <strong>Approximate city area only</strong>
        <p>This area represents a group of pharmacies. It is not a pharmacy address or a navigation destination.</p>
        <p><span>Pharmacies awaiting exact coordinates:</span>{" "}{area.pharmacies.length}</p>
        {onUseArea && <button type="button" className="mb-2 rounded bg-amber-100 px-2 py-1 font-medium text-amber-900" onClick={() => onUseArea(area.point)}>Use this city area as my search location</button>}
        <div className="max-h-48 space-y-2 overflow-y-auto">
          {area.pharmacies.map(pharmacy => <div key={pharmacy.id} className={pharmacy.id === selectedId ? "rounded bg-amber-100 p-1" : ""}>
            <Link href={`/pharmacies/${pharmacy.id}`}>{pharmacy.name}</Link><br />{pharmacy.address}
          </div>)}
        </div>
        {area.sourceUrl && <p><a href={area.sourceUrl} target="_blank" rel="noopener noreferrer">City location source</a></p>}
      </Popup>
    </CircleMarker>
  ))}</>;
}
