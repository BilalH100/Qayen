"use client";

import "leaflet/dist/leaflet.css";
import { useEffect, useMemo, useRef, useState } from "react";
import { MapContainer, Marker, Popup, TileLayer, useMap, useMapEvents } from "react-leaflet";
import L from "leaflet";
import { approximatePharmacyAreas, pharmacyCoordinates } from "@/lib/pharmacy-location";
import { ApproximatePharmacyAreas } from "@/components/approximate-pharmacy-areas";

// Fix Leaflet's default marker icons when used with Next.js.
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png",
  iconUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png",
  shadowUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png",
});

export interface MapPharmacy {
  id: number;
  name: string;
  address: string;
  city?: string;
  latitude: string;
  longitude: string;
}

interface LocationMapPickerProps {
  initialLat: number;
  initialLng: number;
  onLocationPicked: (lat: number, lng: number) => void;
  /** All pharmacies, shown as small teal markers. */
  pharmacies?: MapPharmacy[];
  /** The searched pharmacy: the map zooms to it and its marker turns red. */
  selectedPharmacyId?: number | null;
}

// Pharmacy marker: round badge with a white "+" cross.
const pin = (color: string, size: number) =>
  L.divIcon({
    className: "",
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
    popupAnchor: [0, -size / 2],
    html: `<svg width="${size}" height="${size}" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg"><circle cx="16" cy="16" r="14" fill="${color}" stroke="white" stroke-width="2.5"/><path d="M13 8h6v5h5v6h-5v5h-6v-5H8v-6h5z" fill="white"/></svg>`,
  });
const pharmacyIcon = pin("#0d9488", 30); // teal
const selectedPharmacyIcon = pin("#dc2626", 44); // red + bigger

function MapClickHandler({
  onLocationPicked,
}: {
  onLocationPicked: (lat: number, lng: number) => void;
}) {
  useMapEvents({
    click(event) {
      onLocationPicked(event.latlng.lat, event.latlng.lng);
    },
  });

  return null;
}

function MapPositionController({
  lat,
  lng,
  target,
  targetZoom,
}: {
  lat: number;
  lng: number;
  target: [number, number] | null;
  targetZoom: number;
}) {
  const map = useMap();

  useEffect(() => {
    if (target) {
      map.flyTo(target, targetZoom, { duration: 1 });
    } else {
      map.setView([lat, lng]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [lat, lng, target?.[0], target?.[1], targetZoom, map]);

  return null;
}

export function LocationMapPicker({
  initialLat,
  initialLng,
  onLocationPicked,
  pharmacies = [],
  selectedPharmacyId = null,
}: LocationMapPickerProps) {
  const validPharmacies = pharmacies.filter(
    (p) => pharmacyCoordinates(p) !== null,
  );
  const selected = validPharmacies.find((p) => p.id === selectedPharmacyId) || null;
  const approximate = useMemo(() => approximatePharmacyAreas(pharmacies), [pharmacies]);
  const selectedArea = approximate.find(area => area.pharmacies.some(p => p.id === selectedPharmacyId));
  const target: [number, number] | null = selected
    ? [parseFloat(selected.latitude), parseFloat(selected.longitude)]
    : selectedArea?.point ?? null;

  const [markerPosition, setMarkerPosition] = useState<[number, number]>([
    initialLat,
    initialLng,
  ]);
  const markerRef = useRef<L.Marker | null>(null);

  useEffect(() => {
    setMarkerPosition([initialLat, initialLng]);
  }, [initialLat, initialLng]);

  const updateLocation = (lat: number, lng: number) => {
    setMarkerPosition([lat, lng]);
    onLocationPicked(lat, lng);
  };

  const handleMarkerDragEnd = () => {
    const marker = markerRef.current;
    if (!marker) return;

    const position = marker.getLatLng();
    updateLocation(position.lat, position.lng);
  };

  return (
    <div className="space-y-2">
      {approximate.length > 0 && <p className="text-xs text-amber-800 dark:text-amber-300">Amber groups show approximate city areas, not exact pharmacy locations.</p>}
    <div className="h-72 w-full overflow-hidden rounded-md border border-slate-200 dark:border-slate-700">
      <MapContainer
        center={[initialLat, initialLng]}
        zoom={13}
        scrollWheelZoom={true}
        style={{ height: "100%", width: "100%" }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        <MapPositionController lat={initialLat} lng={initialLng} target={target} targetZoom={selected ? 17 : 10} />
        <ApproximatePharmacyAreas areas={approximate} selectedId={selectedPharmacyId} onUseArea={point => updateLocation(point[0], point[1])} />

        <Marker
          ref={markerRef}
          position={markerPosition}
          draggable={true}
          eventHandlers={{
            dragend: handleMarkerDragEnd,
          }}
        >
          <Popup>
            Your selected location. Drag this pin or click somewhere else on the map.
          </Popup>
        </Marker>

        {validPharmacies.map((p) => (
          <Marker
            key={p.id}
            position={[parseFloat(p.latitude), parseFloat(p.longitude)]}
            icon={p.id === selectedPharmacyId ? selectedPharmacyIcon : pharmacyIcon}
            zIndexOffset={p.id === selectedPharmacyId ? 1000 : 0}
          >
            <Popup>
              <strong>{p.name}</strong>
              <br />
              {p.address}
            </Popup>
          </Marker>
        ))}

        <MapClickHandler onLocationPicked={updateLocation} />
      </MapContainer>
    </div>
    </div>
  );
}
