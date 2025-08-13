"use client";

import type React from "react";

import { useState, useEffect } from "react";
import { Search, MapPin, X, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/contexts/auth-context";
import { BASE_URL } from "@/utils/api";
import { PharmacyList } from "./pharmacy-list";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

interface Medication {
  id: number;
  speciality: string;
  presentation: string;
  dosage: string;
  form: string;
  therapeutic_class: string;
}

export function HeroSearch() {
  const [searchType, setSearchType] = useState<"medication" | "pharmacy">(
    "medication",
  );
  const [location, setLocation] = useState("");
  const [query, setQuery] = useState("");
  const [searching, setSearching] = useState(false);
  const [searchResults, setSearchResults] = useState<Medication[]>([]);
  const [selectedMedication, setSelectedMedication] =
    useState<Medication | null>(null);
  const [showPharmacyDialog, setShowPharmacyDialog] = useState(false);
  // Hardcoded coordinates for Rabat, Morocco
  const [userCoordinates, setUserCoordinates] = useState<{
    latitude: string;
    longitude: string;
  }>({
    latitude: "33.98978814461598",
    longitude: "-6.857842157069873",
  });
  const [locationError, setLocationError] = useState<string | null>(null);
  const { user, isAuthenticated } = useAuth();

  // For testing - always use hardcoded coordinates for Rabat
  const getUserLocation = () => {
    setLocationError(null);
    // We keep the coordinates we already have in state
    // Just update the location display text
    setLocation("Rabat, Morocco");
  };

  // Search medications as user types
  useEffect(() => {
    const searchMedications = async () => {
      if (query.length < 2 || searchType !== "medication") return;

      setSearching(true);
      try {
        const response = await fetch(
          `${BASE_URL}/meds/search?q=${encodeURIComponent(query)}`,
        );
        if (response.ok) {
          const data = await response.json();
          setSearchResults(data);
        }
      } catch (error) {
        console.error("Error searching medications:", error);
      } finally {
        setSearching(false);
      }
    };

    const debounce = setTimeout(() => {
      searchMedications();
    }, 300);

    return () => clearTimeout(debounce);
  }, [query, searchType]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();

    if (searchType === "medication" && selectedMedication) {
      // Coordinates are always available as they're hardcoded
      setShowPharmacyDialog(true);
    } else {
      console.log(`Searching for ${searchType}: ${query} in ${location}`);
      // Handle pharmacy search here
    }
  };

  const handleMedicationSelect = (medication: Medication) => {
    setSelectedMedication(medication);
    setQuery(medication.presentation);
  };

  const clearSelection = () => {
    setSelectedMedication(null);
    setQuery("");
  };

  return (
    <div className="space-y-4">
      {isAuthenticated && user && (
        <div className="mb-4">
          <p className="text-lg text-slate-600 dark:text-slate-300">
            Welcome back,{" "}
            <span className="font-semibold text-teal-600">
              {user.name.split(" ")[0]}
            </span>
            !
          </p>
        </div>
      )}

      <form onSubmit={handleSearch} className="flex flex-col gap-4">
        <div className="relative flex-1">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                className="absolute left-0 top-0 h-full px-3 py-2 rounded-r-none border-r-0 z-10"
              >
                {searchType === "medication" ? "Medication" : "Pharmacy"}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem
                onClick={() => {
                  setSearchType("medication");
                  setSelectedMedication(null);
                  setQuery("");
                }}
              >
                Medication
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => {
                  setSearchType("pharmacy");
                  setSelectedMedication(null);
                  setQuery("");
                }}
              >
                Pharmacy
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {searchType === "medication" && (
            <div className="relative">
              <Input
                placeholder="Search for medications..."
                className="pl-32 h-12 pr-10"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
              {selectedMedication && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-2 top-1/2 transform -translate-y-1/2"
                  onClick={clearSelection}
                  type="button"
                >
                  <X className="h-4 w-4" />
                </Button>
              )}

              {query.length >= 2 &&
                !selectedMedication &&
                searchResults.length > 0 && (
                  <div className="absolute w-full bg-white dark:bg-slate-800 shadow-lg rounded-md mt-1 border border-slate-200 dark:border-slate-700 z-50 max-h-64 overflow-y-auto">
                    <Command>
                      <CommandList>
                        {searching ? (
                          <div className="flex items-center justify-center p-4">
                            <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
                          </div>
                        ) : (
                          <>
                            <CommandEmpty>No medications found</CommandEmpty>
                            <CommandGroup heading="Medications">
                              {searchResults.map((medication) => (
                                <CommandItem
                                  key={medication.id}
                                  onSelect={() =>
                                    handleMedicationSelect(medication)
                                  }
                                  className="cursor-pointer"
                                >
                                  <div>
                                    <p className="font-medium">
                                      {medication.presentation}
                                    </p>
                                    <p className="text-xs text-slate-500 dark:text-slate-400">
                                      {medication.dosage} • {medication.form} •{" "}
                                      {medication.therapeutic_class}
                                    </p>
                                  </div>
                                </CommandItem>
                              ))}
                            </CommandGroup>
                          </>
                        )}
                      </CommandList>
                    </Command>
                  </div>
                )}
            </div>
          )}

          {searchType === "pharmacy" && (
            <Input
              placeholder="Search for pharmacies..."
              className="pl-32 h-12"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          )}
        </div>

        <div className="relative">
          <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input
            placeholder="Your location"
            className="pl-10 h-12 pr-24"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
          />
          <Button
            type="button"
            variant="ghost"
            className="absolute right-0 top-0 h-full px-3"
            onClick={getUserLocation}
          >
            Use current
          </Button>
        </div>

        {locationError && (
          <p className="text-sm text-red-500">{locationError}</p>
        )}

        <Button
          type="submit"
          className="h-12 bg-teal-600 hover:bg-teal-700"
          disabled={searchType === "medication" && !selectedMedication}
        >
          <Search className="mr-2 h-4 w-4" />
          Search
        </Button>
      </form>

      <Dialog open={showPharmacyDialog} onOpenChange={setShowPharmacyDialog}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              Pharmacies with {selectedMedication?.presentation}
            </DialogTitle>
          </DialogHeader>
          {selectedMedication && (
            <PharmacyList
              medicationId={selectedMedication.id}
              userCoordinates={userCoordinates}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
