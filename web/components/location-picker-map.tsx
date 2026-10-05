"use client";

import "leaflet/dist/leaflet.css";
import { useEffect, useRef, useState } from "react";
import { MapContainer, Marker, Popup, TileLayer, useMap, useMapEvents } from "react-leaflet";
import L from "leaflet";

// Fix Leaflet's default marker icons when used with Next.js.
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png",
  iconUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png",
  shadowUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png",
});

interface LocationMapPickerProps {
  initialLat: number;
  initialLng: number;
  onLocationPicked: (lat: number, lng: number) => void;
}

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

function MapPositionController({ lat, lng }: { lat: number; lng: number }) {
  const map = useMap();

  useEffect(() => {
    map.setView([lat, lng]);
  }, [lat, lng, map]);

  return null;
}

export function LocationMapPicker({
  initialLat,
  initialLng,
  onLocationPicked,
}: LocationMapPickerProps) {
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

        <MapPositionController lat={initialLat} lng={initialLng} />

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

        <MapClickHandler onLocationPicked={updateLocation} />
      </MapContainer>
    </div>
  );
}
