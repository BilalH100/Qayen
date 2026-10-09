"use client";

import { useEffect, useRef } from "react";
import { BASE_URL } from "@/utils/api";

declare global {
  interface Window {
    L?: any;
  }
}

interface LocationPickerMapProps {
  initialLat: number;
  initialLng: number;
  onLocationPicked: (lat: number, lng: number) => void;
}

interface PharmacyLocation {
  id: number;
  name: string;
  city: string;
  address: string;
  latitude: string | number;
  longitude: string | number;
}

const LEAFLET_VERSION = "1.9.4";
const LEAFLET_CSS = `https://unpkg.com/leaflet@${LEAFLET_VERSION}/dist/leaflet.css`;
const LEAFLET_JS = `https://unpkg.com/leaflet@${LEAFLET_VERSION}/dist/leaflet.js`;

let leafletPromise: Promise<any> | null = null;

function loadLeaflet(): Promise<any> {
  if (typeof window === "undefined") {
    return Promise.reject(new Error("Leaflet can only load in the browser."));
  }

  if (window.L) {
    return Promise.resolve(window.L);
  }

  if (leafletPromise) {
    return leafletPromise;
  }

  leafletPromise = new Promise((resolve, reject) => {
    const existingCss = document.querySelector(
      `link[href="${LEAFLET_CSS}"]`,
    );

    if (!existingCss) {
      const link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = LEAFLET_CSS;
      document.head.appendChild(link);
    }

    const existingScript = document.querySelector(
      `script[src="${LEAFLET_JS}"]`,
    ) as HTMLScriptElement | null;

    if (existingScript) {
      existingScript.addEventListener("load", () => resolve(window.L));
      existingScript.addEventListener("error", () =>
        reject(new Error("Failed to load Leaflet.")),
      );
      return;
    }

    const script = document.createElement("script");
    script.src = LEAFLET_JS;
    script.async = true;
    script.onload = () => {
      if (window.L) {
        resolve(window.L);
      } else {
        reject(new Error("Leaflet loaded but was not available."));
      }
    };
    script.onerror = () => reject(new Error("Failed to load Leaflet."));
    document.body.appendChild(script);
  });

  return leafletPromise;
}

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function LocationPickerMap({
  initialLat,
  initialLng,
  onLocationPicked,
}: LocationPickerMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<any>(null);
  const markerRef = useRef<any>(null);
  const pharmacyLayerRef = useRef<any>(null);
  const onLocationPickedRef = useRef(onLocationPicked);

  useEffect(() => {
    onLocationPickedRef.current = onLocationPicked;
  }, [onLocationPicked]);

  useEffect(() => {
    let cancelled = false;

    const initializeMap = async () => {
      try {
        const L = await loadLeaflet();

        if (cancelled || !mapContainerRef.current || mapRef.current) {
          return;
        }

        const defaultCenter: [number, number] = [34.0209, -6.8416];

        const map = L.map(mapContainerRef.current, {
          center: [initialLat, initialLng],
          zoom: 13,
          zoomControl: true,
        });

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
          attribution:
            '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a> contributors',
          maxZoom: 19,
        }).addTo(map);

        // Pharmacy markers are visual reference points only. They never replace
        // the patient's draggable/clickable location marker.
        const pharmacyLayer = L.layerGroup().addTo(map);
        pharmacyLayerRef.current = pharmacyLayer;

        const pharmacyIcon = L.divIcon({
          className: "qayen-pharmacy-marker",
          html: `
            <div style="
              width: 28px;
              height: 28px;
              border-radius: 50%;
              background: #0f766e;
              border: 3px solid white;
              box-shadow: 0 2px 6px rgba(0,0,0,0.35);
              display: flex;
              align-items: center;
              justify-content: center;
              color: white;
              font-size: 15px;
              font-weight: 700;
            ">+</div>
          `,
          iconSize: [28, 28],
          iconAnchor: [14, 14],
          popupAnchor: [0, -14],
        });

        try {
          const response = await fetch(`${BASE_URL}/pharmacies`, {
            cache: "no-store",
          });

          if (response.ok && !cancelled) {
            const payload = await response.json();
            const pharmacies: PharmacyLocation[] = Array.isArray(payload?.pharmacies)
              ? payload.pharmacies
              : [];

            pharmacies.forEach((pharmacy) => {
              const lat = Number(pharmacy.latitude);
              const lng = Number(pharmacy.longitude);

              if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;

              const pharmacyMarker = L.marker([lat, lng], {
                icon: pharmacyIcon,
                interactive: true,
              }).addTo(pharmacyLayer);

              pharmacyMarker.bindPopup(`
                <div style="min-width: 180px">
                  <strong>${escapeHtml(pharmacy.name)}</strong><br />
                  <span>${escapeHtml(pharmacy.address || "Rabat")}</span>
                </div>
              `);
            });
          }
        } catch (error) {
          console.error("Failed to load pharmacy locations:", error);
        }

        const marker = L.marker(
          [initialLat, initialLng],
          { draggable: true },
        ).addTo(map);

        marker.bindPopup(
          "Drag this pin or click anywhere on the map to choose the patient's location.",
        );

        const updateLocation = (lat: number, lng: number) => {
          marker.setLatLng([lat, lng]);
          onLocationPickedRef.current(lat, lng);
        };

        map.on("click", (event: any) => {
          updateLocation(event.latlng.lat, event.latlng.lng);
        });

        marker.on("dragend", () => {
          const position = marker.getLatLng();
          onLocationPickedRef.current(position.lat, position.lng);
        });

        mapRef.current = map;
        markerRef.current = marker;

        setTimeout(() => map.invalidateSize(), 100);
      } catch (error) {
        console.error("Failed to initialize location map:", error);
      }
    };

    initializeMap();

    return () => {
      cancelled = true;

      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
        markerRef.current = null;
        pharmacyLayerRef.current = null;
      }
    };
  }, []);

  return (
    <div className="space-y-2">
      <div className="overflow-hidden rounded-lg border">
        <div
          ref={mapContainerRef}
          className="h-[360px] w-full"
          aria-label="Interactive map for selecting the patient's location"
        />
      </div>

      <p className="text-xs text-muted-foreground">
        Click anywhere on the map or drag the pin to set the patient's exact
        location in Rabat.
      </p>
    </div>
  );
}

export { LocationPickerMap };
export default LocationPickerMap;
