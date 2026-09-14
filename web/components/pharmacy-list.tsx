"use client";

import { useState, useEffect } from "react";
import { MapPin, Clock, Navigation } from "lucide-react";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { BASE_URL } from "@/utils/api";

interface Pharmacy {
  id: number;
  name: string;
  address: string;
  latitude: string;
  longitude: string;
  city: string;
  phone: string;
  created_at: string;
  activity?: string;
}

interface PharmacyListProps {
  medicationId: number;
  userCoordinates: {
    latitude: string;
    longitude: string;
  };
}

interface PharmacyWithDistance {
  distance: number;
  pharmacy: Pharmacy;
}

export function PharmacyList({
  medicationId,
  userCoordinates,
}: PharmacyListProps) {
  const [pharmacies, setPharmacies] = useState<PharmacyWithDistance[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPharmacies = async () => {
      if (!medicationId) return;

      setLoading(true);
      setError(null);

      try {
        const url = `${BASE_URL}/pharmacies/closest?id=${medicationId}&latitude=${33.98943231678375}&longitude=${-6.857927987754053}`;
        const response = await fetch(url);

        if (!response.ok) {
          throw new Error("Failed to fetch pharmacy data");
        }
        const data = await response.json();
        setPharmacies([data]);
      } catch (err) {
        console.error("Error fetching pharmacies:", err);
        setError(
          err instanceof Error ? err.message : "Failed to load pharmacies",
        );
      } finally {
        setLoading(false);
      }
    };

    fetchPharmacies();
  }, [medicationId, userCoordinates]);

  if (loading) {
    return (
      <div className="w-full py-8">
        <div className="flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-teal-600 mx-auto mb-4"></div>
            <p className="text-slate-600 dark:text-slate-300">
              Finding pharmacies near you...
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full py-4">
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-red-700 dark:text-red-400">{error}</p>
          <p className="text-sm text-red-600 dark:text-red-500 mt-2">
            Please check your connection or try again later.
          </p>
        </div>
      </div>
    );
  }

  // We always have coordinates as they're hardcoded

  if (pharmacies.length === 0) {
    return (
      <div className="w-full py-4">
        <div className="bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-4">
          <p className="text-slate-700 dark:text-slate-300">
            No pharmacies with this medication found near your location.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4 w-full">
      <h3 className="text-xl font-semibold text-slate-900 dark:text-white mb-4">
        Pharmacies with this medication
      </h3>

      {pharmacies.map((item) => (
        <Card
          key={item.pharmacy.id}
          className="overflow-hidden transition-all duration-300 hover:shadow-lg"
        >
          <CardContent className="p-4">
            <div className="flex justify-between items-start">
              <h3 className="text-lg font-semibold text-slate-900 dark:text-white">
                {item.pharmacy.name}
              </h3>
              <Badge
                variant="outline"
                className="bg-teal-50 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400 border-teal-200 dark:border-teal-800"
              >
                {item.distance} km away
              </Badge>
            </div>

            <div className="flex items-center gap-1 text-slate-500 dark:text-slate-400 mt-2 text-sm">
              <MapPin className="h-4 w-4" />
              <span>
                {item.pharmacy.address}, {item.pharmacy.city}
              </span>
            </div>

            <div className="flex items-center gap-1 text-slate-500 dark:text-slate-400 mt-1 text-sm">
              <Clock className="h-4 w-4" />
              <span>Phone: {item.pharmacy.phone}</span>
            </div>
          </CardContent>

          <CardFooter className="p-4 pt-0 flex justify-between">
            <Link
              href={`https://www.google.com/maps/dir/?api=1&destination=${item.pharmacy.latitude},${item.pharmacy.longitude}`}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Button
                variant="outline"
                size="sm"
                className="flex items-center gap-2"
              >
                <Navigation className="h-4 w-4" />
                Directions
              </Button>
            </Link>

            <Link href={`/pharmacies/${item.pharmacy.id}`}>
              <Button size="sm" className="bg-teal-600 hover:bg-teal-700">
                View Details
              </Button>
            </Link>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
