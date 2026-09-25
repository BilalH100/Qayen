"use client";

import { useEffect, useRef } from "react";

declare global {
  interface Window {
    L?: any;
  }
}

interface LocationValue {
  lat: number;
  lng: number;
}

interface LocationPickerMapProps {
  value: LocationValue | null;
  onChange: (location: LocationValue) => void;
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

export default function LocationPickerMap({
  value,
  onChange,
}: LocationPickerMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<any>(null);
  const markerRef = useRef<any>(null);
  const onChangeRef = useRef(onChange);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

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
          center: value ? [value.lat, value.lng] : defaultCenter,
          zoom: value ? 15 : 12,
          zoomControl: true,
        });

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
          attribution:
            '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a> contributors',
          maxZoom: 19,
        }).addTo(map);

        const marker = L.marker(
          value ? [value.lat, value.lng] : defaultCenter,
          { draggable: true },
        ).addTo(map);

        marker.bindPopup(
          "Drag this pin or click anywhere on the map to choose the patient's location.",
        );

        const updateLocation = (lat: number, lng: number) => {
          marker.setLatLng([lat, lng]);
          onChangeRef.current({ lat, lng });
        };

        map.on("click", (event: any) => {
          updateLocation(event.latlng.lat, event.latlng.lng);
        });

        marker.on("dragend", () => {
          const position = marker.getLatLng();
          onChangeRef.current({
            lat: position.lat,
            lng: position.lng,
          });
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
      }
    };
  }, []);

  useEffect(() => {
    if (!value || !mapRef.current || !markerRef.current) {
      return;
    }

    const latLng = [value.lat, value.lng];

    markerRef.current.setLatLng(latLng);
    mapRef.current.setView(latLng, Math.max(mapRef.current.getZoom(), 14));
  }, [value]);

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
