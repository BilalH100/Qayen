"use client";

import { Suspense, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { PlatformSearchForm } from "@/components/platform-search-form";
import { Button } from "@/components/ui/button";
import { BASE_URL } from "@/utils/api";
import { searchableSections, searchPlatform, type SearchResult } from "@/lib/platform-search";

interface Medication {
  id: number; speciality: string; code?: string; dosage?: string; active_substance?: string;
  therapeutic_class?: string; form?: string; presentation?: string; description?: string;
}
interface Pharmacy { id: number; name: string; city?: string; address?: string; phone?: string }

async function loadMedications(signal: AbortSignal): Promise<SearchResult[]> {
  const items: SearchResult[] = [];
  const batchSize = 1000;
  for (let offset = 0; ; offset += batchSize) {
    const response = await fetch(`${BASE_URL}/meds/all?offset=${offset}&limit=${batchSize}`, { signal });
    if (!response.ok) throw new Error("Medication search unavailable");
    const rows: Medication[] = await response.json();
    if (!Array.isArray(rows)) throw new Error("Invalid medication results");
    items.push(...rows.map(med => ({
      id: `medication-${med.id}`, title: med.speciality,
      description: [med.dosage, med.active_substance, med.form, med.presentation].filter(Boolean).join(" · "),
      keywords: [med.code, med.therapeutic_class, med.description].filter(Boolean).join(" "),
      href: `/medications/${encodeURIComponent(med.code || String(med.id))}`, category: "Medications" as const,
    })));
    if (rows.length < batchSize) return items;
  }
}

async function loadPharmacies(signal: AbortSignal): Promise<SearchResult[]> {
  const response = await fetch(`${BASE_URL}/pharmacies`, { signal });
  if (!response.ok) throw new Error("Pharmacy search unavailable");
  const data = await response.json();
  const rows: Pharmacy[] = Array.isArray(data) ? data : data.pharmacies;
  if (!Array.isArray(rows)) throw new Error("Invalid pharmacy results");
  return rows.map(pharmacy => ({
    id: `pharmacy-${pharmacy.id}`, title: pharmacy.name,
    description: [pharmacy.address, pharmacy.city].filter(Boolean).join(" · "),
    keywords: pharmacy.phone, href: `/pharmacies/${pharmacy.id}`, category: "Pharmacies",
  }));
}

function SearchResults() {
  const params = useSearchParams();
  const query = (params.get("q") ?? "").trim().slice(0, 200);
  const [items, setItems] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);
  const [attempt, setAttempt] = useState(0);
  const [page, setPage] = useState(1);
  const [category, setCategory] = useState("All");
  const hasQuery = !!query;

  useEffect(() => {
    if (!hasQuery) return;
    const controller = new AbortController();
    setLoading(true);
    setErrors([]);
    void Promise.allSettled([loadMedications(controller.signal), loadPharmacies(controller.signal)]).then(results => {
      if (controller.signal.aborted) return;
      setItems(results.flatMap(result => result.status === "fulfilled" ? result.value : []));
      setErrors(results.flatMap((result, index) => result.status === "rejected" ? [index === 0 ? "Medication search is temporarily unavailable." : "Pharmacy search is temporarily unavailable."] : []));
      setLoading(false);
    });
    return () => controller.abort();
  }, [hasQuery, attempt]);

  useEffect(() => { setPage(1); }, [query, category]);
  const results = useMemo(() => searchPlatform([...searchableSections, ...items], query), [items, query]);
  const filtered = results.filter(result => category === "All" || result.category === category);
  const pageSize = 20;
  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const currentPage = Math.min(page, totalPages);

  return (
    <main className="container mx-auto max-w-4xl space-y-6 px-4 py-10">
      <h1 className="text-3xl font-bold">Search Qayen</h1>
      <PlatformSearchForm key={query} initialQuery={query} />
      <p className="text-muted-foreground">Search medications, ingredients, pharmacies, cities, and platform sections.</p>
      {!query ? <p>Enter a search term to get started.</p> : <>
        <div className="flex flex-wrap gap-2" aria-label="Search categories">
          {["All", "Medications", "Pharmacies", "Sections"].map(label => (
            <Button key={label} variant={category === label ? "default" : "outline"} aria-pressed={category === label} onClick={() => setCategory(label)}>{label}</Button>
          ))}
        </div>
        {errors.length > 0 && <div role="alert" className="rounded-lg border p-4">
          {errors.map(error => <p key={error}>{error}</p>)}
          <Button variant="outline" className="mt-2" onClick={() => setAttempt(value => value + 1)}>Try again</Button>
        </div>}
        <div role="status" aria-live="polite">
          {loading ? <p>Searching...</p> : <p><span>{filtered.length}</span> <span>Search results</span></p>}
        </div>
        {!loading && !filtered.length && <p>No results found. Try another name, ingredient, or city.</p>}
        <div className="space-y-3">
          {filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize).map(result => (
            <Link key={result.id} href={result.href} className="block rounded-xl border p-5 transition-colors hover:border-teal-500 hover:bg-teal-50/50 dark:hover:bg-teal-950/20">
              <span className="text-xs font-medium text-teal-600 dark:text-teal-400">{result.category}</span>
              <h2 className="mt-1 text-lg font-semibold">{result.title}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{result.description}</p>
            </Link>
          ))}
        </div>
        {totalPages > 1 && <div className="flex items-center justify-center gap-4">
          <Button variant="outline" disabled={currentPage === 1} onClick={() => setPage(currentPage - 1)}>Previous</Button>
          <span>{currentPage} / {totalPages}</span>
          <Button variant="outline" disabled={currentPage === totalPages} onClick={() => setPage(currentPage + 1)}>Next</Button>
        </div>}
      </>}
    </main>
  );
}

export default function SearchPage() {
  return <Suspense fallback={<p className="container py-10">Searching...</p>}><SearchResults /></Suspense>;
}
