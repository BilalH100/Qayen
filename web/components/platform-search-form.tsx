"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export function PlatformSearchForm({ initialQuery = "", autoFocus = false, onClose }: {
  initialQuery?: string;
  autoFocus?: boolean;
  onClose?: () => void;
}) {
  const [query, setQuery] = useState(initialQuery);
  const router = useRouter();

  return (
    <form role="search" className="flex w-full items-center gap-1" onSubmit={event => {
      event.preventDefault();
      if (!query.trim()) return;
      router.push(`/search?q=${encodeURIComponent(query.trim())}`);
      onClose?.();
    }} onKeyDown={event => {
      if (event.key === "Escape") onClose?.();
    }}>
      <Input name="q" type="search" aria-label="Search Qayen" placeholder="Search Qayen..."
        value={query} onChange={event => setQuery(event.target.value)} autoFocus={autoFocus}
        className="min-w-0 flex-1" required maxLength={200} />
      <Button type="submit" variant="ghost" size="icon" disabled={!query.trim()} aria-label="Search">
        <Search className="h-4 w-4" />
      </Button>
      {onClose && <Button type="button" variant="ghost" size="icon" onClick={onClose} aria-label="Close search">
        <X className="h-4 w-4" />
      </Button>}
    </form>
  );
}
