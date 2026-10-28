"use client";

import "leaflet/dist/leaflet.css";
import { useEffect, useMemo } from "react";
import Link from "next/link";
import { CircleMarker, MapContainer, Popup, TileLayer, useMap } from "react-leaflet";
import { approximatePharmacyAreas, pharmacyCoordinates } from "@/lib/pharmacy-location";
import { ApproximatePharmacyAreas } from "@/components/approximate-pharmacy-areas";

interface DirectoryPharmacy {
  id: number;
  name: string;
  address: string;
  city: string;
  latitude: string;
  longitude: string;
}

function FitPharmacies({ points, approximate }: { points: [number, number][]; approximate: boolean }) {
  const map = useMap();
  useEffect(() => {
    map.invalidateSize();
    if (points.length === 1) map.setView(points[0], approximate ? 10 : 16);
    else if (points.length > 1) map.fitBounds(points, { padding: [35, 35], maxZoom: 14 });
    else map.setView([30.7, -7.6], 5);
  }, [map, points, approximate]);
  return null;
}

export function PharmacyDirectoryMap({ pharmacies, referencePharmacies = pharmacies }: { pharmacies: DirectoryPharmacy[]; referencePharmacies?: DirectoryPharmacy[] }) {
  const located = useMemo(() => pharmacies.flatMap(pharmacy => {
    const point = pharmacyCoordinates(pharmacy);
    return point ? [{ pharmacy, point }] : [];
  }), [pharmacies]);
  const approximate = useMemo(() => approximatePharmacyAreas(pharmacies, referencePharmacies), [pharmacies, referencePharmacies]);
  const points = useMemo(() => [...located.map(row => row.point), ...approximate.map(area => area.point)], [located, approximate]);

  return (
    <section className="space-y-3">
      <p className="text-sm text-slate-600 dark:text-slate-300">
        <span>Pharmacies with map locations:</span>{" "}{located.length}{" / "}{pharmacies.length}
      </p>
      <p className="text-sm text-amber-800 dark:text-amber-300"><span>City-area estimates:</span>{" "}{approximate.reduce((count, area) => count + area.pharmacies.length, 0)}</p>
      <p className="text-sm text-slate-500 dark:text-slate-400">Teal pins show recorded GPS. Amber groups show approximate city areas, not pharmacy locations or directions.</p>
      <div className="relative z-0 h-[600px] overflow-hidden rounded-lg border">
        <MapContainer center={[30.7, -7.6]} zoom={5} preferCanvas scrollWheelZoom style={{ height: "100%", width: "100%" }}>
          <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors' />
          <FitPharmacies points={points} approximate={approximate.length > 0} />
          <ApproximatePharmacyAreas areas={approximate} />
          {located.map(({ pharmacy, point }) => (
            <CircleMarker key={pharmacy.id} center={point} radius={7} pathOptions={{ color: "#ffffff", weight: 2, fillColor: "#0d9488", fillOpacity: 0.95 }}>
              <Popup>
                <strong>{pharmacy.name}</strong><br />
                {pharmacy.address}<br />{pharmacy.city}<br />
                <Link href={`/pharmacies/${pharmacy.id}`}>View pharmacy details</Link>
              </Popup>
            </CircleMarker>
          ))}
        </MapContainer>
      </div>
      {points.length === 0 && <p>No pharmacies with coordinates match this search.</p>}
    </section>
  );
}
