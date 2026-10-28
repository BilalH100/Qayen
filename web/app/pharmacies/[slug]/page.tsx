"use client";

import { use, useEffect, useState } from "react";
import Link from "next/link";
import { BASE_URL } from "@/utils/api";
import dynamic from "next/dynamic";
import { pharmacyCoordinates } from "@/lib/pharmacy-location";

const PharmacyMap = dynamic(() => import("@/components/pharmacy-map").then(module => module.PharmacyMap), { ssr: false });

interface Pharmacy {
  id: number;
  name: string;
  address?: string;
  city?: string;
  phone?: string;
  latitude?: string;
  longitude?: string;
  address_source_url?: string;
  location_source_url?: string;
}

export default function PharmacyPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = use(params);
  const [pharmacy, setPharmacy] = useState<Pharmacy | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setFailed(false);
    setPharmacy(null);
    async function load() {
      try {
        const numeric = /^\d+$/.test(slug);
        const response = await fetch(`${BASE_URL}/pharmacies${numeric ? `/id/${slug}` : ""}`, { signal: controller.signal });
        if (!response.ok) throw new Error("Pharmacy request failed");
        const data = await response.json();
        if (numeric) setPharmacy(data.pharmacy ?? null);
        else {
          const rows: Pharmacy[] = Array.isArray(data) ? data : data.pharmacies ?? [];
          const matches = rows.filter(row => row.name.toLowerCase().replace(/\s+/g, "-") === slug);
          setPharmacy(matches.length === 1 ? matches[0] : null);
        }
      } catch {
        if (!controller.signal.aborted) setFailed(true);
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [slug]);

  const coordinates = pharmacyCoordinates(pharmacy ?? {});

  return (
    <main className="container mx-auto max-w-4xl space-y-6 px-4 py-10">
      <Link href="/pharmacies" className="text-primary underline">Back to pharmacies</Link>
      {loading ? <p>Loading pharmacy...</p> : !pharmacy ? (
        <section className="rounded-xl border p-6">
          <h1 className="text-2xl font-bold">{failed ? "Failed to load pharmacies" : "Pharmacy not found"}</h1>
          <p>Choose a pharmacy from the directory to view its available details.</p>
        </section>
      ) : (
        <>
          <header><h1 className="text-3xl font-bold">{pharmacy.name}</h1><p>{pharmacy.city}</p></header>
          <div className="grid gap-6 md:grid-cols-2">
            <section className="space-y-3 rounded-xl border p-6">
              <h2 className="text-xl font-semibold">Address</h2>
              <p>{pharmacy.address || "Exact address awaiting confirmation"}</p>
              {pharmacy.address_source_url?.startsWith("https://pharmacy.ma/pharmacie/") && (
                <p className="text-sm text-muted-foreground">
                  <span>Public directory address; confirm with the pharmacy.</span>{" "}
                  <a className="text-primary underline" href={pharmacy.address_source_url} target="_blank" rel="noopener noreferrer">Address source</a>
                </p>
              )}
              {coordinates ? <>
                <PharmacyMap lat={coordinates[0]} lng={coordinates[1]} name={pharmacy.name} address={pharmacy.address || ""} />
                <a className="text-primary underline" href={`https://www.google.com/maps?q=${coordinates[0]},${coordinates[1]}`} target="_blank" rel="noopener noreferrer">View on Map</a>
                {pharmacy.location_source_url?.startsWith("https://pharmap.ma/pharmacie/") && <p className="text-sm text-muted-foreground"><span>Published pharmacy location; confirm with the pharmacy.</span>{" "}<a className="text-primary underline" href={pharmacy.location_source_url} target="_blank" rel="noopener noreferrer">Map location source</a></p>}
              </> : <p>Location awaiting confirmation</p>}
            </section>
            <section className="space-y-3 rounded-xl border p-6">
              <h2 className="text-xl font-semibold">Contact</h2>
              {pharmacy.phone ? <a href={`tel:${pharmacy.phone.replace(/[^+\d]/g, "")}`}>{pharmacy.phone}</a> : <p>Phone number awaiting confirmation</p>}
              <p className="text-sm text-muted-foreground">Opening hours and medicine availability must be confirmed with the pharmacy.</p>
            </section>
          </div>
        </>
      )}
    </main>
  );
}
