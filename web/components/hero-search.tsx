import type React from "react";

import { useState, useEffect, useRef } from "react";
import { Search, MapPin, X, Loader2, Bell, CheckCircle, AlertTriangle } from "lucide-react";
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
import dynamic from "next/dynamic";

const LocationMapPicker = dynamic(
  () => import("./location-picker-map").then((mod) => mod.LocationPickerMap),
  { ssr: false }
);

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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";

interface Medication {
  id: number;
  speciality: string;
  presentation: string;
  dosage: string;
  form: string;
  therapeutic_class: string;
}

interface PharmacyResponse {
  pharmacy_id: number;
  pharmacy_name: string;
  pharmacy_address: string;
  pharmacy_phone: string;
  latitude: number;
  longitude: number;
  distance_km: number;
  response_type: string;
  substitute_brand?: string;
  substitute_notes?: string;
  available_quantity: number;
}

interface AlertResult {
  alert_id: number;
  medication_id: number;
  status: string;
  pharmacies: PharmacyResponse[];
  created_at: string;
  expires_at: string;
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
  const [showAlertDialog, setShowAlertDialog] = useState(false);
  const [alertResult, setAlertResult] = useState<AlertResult | null>(null);
  const [creatingAlert, setCreatingAlert] = useState(false);
  const [waitingForResponses, setWaitingForResponses] = useState(false);
  const ATTEMPT_SECONDS = 30; // time each pharmacy gets before we try the next one
  const [timeRemaining, setTimeRemaining] = useState(ATTEMPT_SECONDS);
  // How many units the patient needs. A pharmacy is tried if it has the medication
  // in positive stock; the requested quantity is checked when it responds.
  const [requestedQuantity, setRequestedQuantity] = useState("1");
  const [statusNote, setStatusNote] = useState<string | null>(null);
  const waitCapTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // Hardcoded coordinates for Rabat, Morocco
  const [userCoordinates, setUserCoordinates] = useState<{
    latitude: string;
    longitude: string;
  }>({
    latitude: "33.98978814461598",
    longitude: "-6.857842157069873",
  });
  const [locationError, setLocationError] = useState<string | null>(null);
  // Quantity the patient has typed for each pharmacy's "Confirm & take"
  // input, keyed by pharmacy_id, plus which pharmacy (if any) is currently
  // being confirmed and which one has already been confirmed.
  const [confirmQuantities, setConfirmQuantities] = useState<Record<number, string>>({});
  const [confirmingPharmacyId, setConfirmingPharmacyId] = useState<number | null>(null);
  const [confirmedPharmacyId, setConfirmedPharmacyId] = useState<number | null>(null);
  const { user, isAuthenticated } = useAuth();
  const { toast } = useToast();
  const eventSourceRef = useRef<EventSource | null>(null);
  const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const resultsPollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const seenPharmacyIdsRef = useRef<Set<number>>(new Set());

  // Always close any open SSE connection / countdown when the component
  // unmounts.
  useEffect(() => {
    return () => {
      eventSourceRef.current?.close();
      if (countdownIntervalRef.current) clearInterval(countdownIntervalRef.current);
      if (resultsPollingRef.current) clearInterval(resultsPollingRef.current);
    };
  }, []);

  const getUserLocation = () => {
    setLocationError(null);

    if (!navigator.geolocation) {
      setLocationError("Geolocation is not supported by this browser.");
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (position) => {
        const latitude = position.coords.latitude.toFixed(6);
        const longitude = position.coords.longitude.toFixed(6);
        setUserCoordinates({ latitude, longitude });
        setLocation(`Current location (${latitude}, ${longitude})`);
      },
      (error) => {
        console.error("Geolocation error:", error);
        setLocationError(
          "Could not get your current location. You can choose a location directly on the map."
        );
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 60000 }
    );
  };

  const handleMapLocationPicked = (lat: number, lng: number) => {
    const latitude = lat.toFixed(6);
    const longitude = lng.toFixed(6);
    setUserCoordinates({ latitude, longitude });
    setLocation(`Selected location (${latitude}, ${longitude})`);
    setLocationError(null);
  };

  // Search medications as user types
  useEffect(() => {
    const searchMedications = async () => {
      if (query.length < 2 || searchType !== "medication") {
        setSearchResults([]);
        return;
      }

      setSearching(true);
      try {
        const response = await fetch(
          `${BASE_URL}/meds/search?q=${encodeURIComponent(query)}`,
        );
        if (response.ok) {
          const data = await response.json();
          setSearchResults(Array.isArray(data) ? data : []);
        } else {
          setSearchResults([]);
        }
      } catch (error) {
        console.error("Error searching medications:", error);
        setSearchResults([]);
      } finally {
        setSearching(false);
      }
    };

    const debounce = setTimeout(() => {
      searchMedications();
    }, 300);

    return () => clearTimeout(debounce);
  }, [query, searchType]);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();

    if (searchType === "medication" && selectedMedication) {
      if (!isAuthenticated) {
        toast({
          title: "Authentication Required",
          description: "Please log in to create medication alerts",
          variant: "destructive",
        });
        return;
      }
      await createMedicationAlert();
    } else {
      console.log(`Searching for ${searchType}: ${query} in ${location}`);
      // Handle pharmacy search here
    }
  };

  const createMedicationAlert = async () => {
    if (!selectedMedication || !isAuthenticated || !user) return;

    setCreatingAlert(true);
    // Fully reset everything from the previous request so the new one starts
    // clean (this was why "Pickup confirmed" showed instead of the quantity
    // selector on 2nd+ requests until the page was refreshed).
    eventSourceRef.current?.close();
    eventSourceRef.current = null;
    if (countdownIntervalRef.current) {
      clearInterval(countdownIntervalRef.current);
      countdownIntervalRef.current = null;
    }
    seenPharmacyIdsRef.current = new Set();
    setConfirmedPharmacyId(null);
    setConfirmingPharmacyId(null);
    setConfirmQuantities({});
    setAlertResult(null);
    setStatusNote(null);
    if (waitCapTimeoutRef.current) clearTimeout(waitCapTimeoutRef.current);
    try {
      const response = await fetch(`${BASE_URL}/alerts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          customer_id: user.id,
          medication_id: selectedMedication.id,
          latitude: parseFloat(userCoordinates.latitude),
          longitude: parseFloat(userCoordinates.longitude),
          search_radius_km: 50,
          max_response_time_minutes: 2,
          requested_quantity: Math.max(parseInt(requestedQuantity, 10) || 1, 1),
        }),
      });

      if (response.ok) {
        const responseData = await response.json();
        const alertData = responseData.data;
        setAlertResult({
          alert_id: alertData.alert_id,
          medication_id: alertData.medication_id,
          status: "waiting",
          pharmacies: [],
          created_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + 2 * 60 * 1000).toISOString(), // 2 minutes from now
        });
        setConfirmQuantities({});
        setShowAlertDialog(true);
        setWaitingForResponses(true);
        setTimeRemaining(ATTEMPT_SECONDS);
        // Safety net in case the live connection drops: stop waiting after 5 min.
        waitCapTimeoutRef.current = setTimeout(() => {
        setWaitingForResponses(false);
        if (resultsPollingRef.current) {
          clearInterval(resultsPollingRef.current);
          resultsPollingRef.current = null;
        }
      }, 5 * 60 * 1000);
        
        toast({
          title: "Alert Created",
          description: "Your medication alert has been sent to nearby pharmacies",
        });

        // Start the visual countdown (a progress indicator only — it no
        // longer decides when the "waiting" state ends).
        startCountdownTimer();
        // Listen for pharmacist responses in real time instead of polling,
        // so the waiting box disappears the instant a pharmacy responds.
        listenForResponses(alertData.alert_id);
      } else {
        const errorData = await response.json();
        toast({
          title: "Error Creating Alert",
          description:
            (errorData.error && typeof errorData.error === "object"
              ? [errorData.error.message, errorData.error.details].filter(Boolean).join(" — ")
              : errorData.error) ||
            errorData.message ||
            "Failed to create alert. Please try again.",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error("Error creating alert:", error);
      toast({
        title: "Network Error",
        description: "Unable to create alert. Please check your connection.",
        variant: "destructive",
      });
    } finally {
      setCreatingAlert(false);
    }
  };

  const startCountdownTimer = () => {
    if (countdownIntervalRef.current) clearInterval(countdownIntervalRef.current);
    const interval = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 1) {
          // Just the per-pharmacy timer: the server decides when waiting ends
          // (it moves to the next pharmacy and tells us over SSE).
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    countdownIntervalRef.current = interval;
  };

  const fetchAlertResults = async (alertId: number) => {
    try {
      const response = await fetch(
        `${BASE_URL}/alerts/${alertId}/results?lat=${userCoordinates.latitude}&lng=${userCoordinates.longitude}`,
      );
      if (response.ok) {
        const data = await response.json();
        return data.data;
      }
    } catch (error) {
      console.error("Error fetching alert results:", error);
    }
    return undefined;
  };

  // Real-time updates: as soon as a pharmacist picks "available", "not
  // available", or "substitute" on their dashboard, the backend pushes a
  // "new_response" event over SSE. We use that (instead of polling every
  // 5s) to close the waiting box immediately and tell the patient exactly
  // which pharmacy (name + id) responded and with what.
  const listenForResponses = (alertId: number) => {
    console.log("[alert] connecting to SSE stream for alert", alertId, `${BASE_URL}/alerts/${alertId}/stream`);
    const eventSource = new EventSource(`${BASE_URL}/alerts/${alertId}/stream`);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      console.log("[alert] SSE connection OPEN for alert", alertId);
    };

    eventSource.onmessage = async (event) => {
      console.log("[alert] SSE message received:", event.data);
      if (eventSourceRef.current !== eventSource) return; // stale stream
      const updateData = JSON.parse(event.data);
      if (updateData.type === "trying_next") {
        setTimeRemaining(ATTEMPT_SECONDS);
        setStatusNote("Previous pharmacy couldn't help — trying the next nearest pharmacy...");
        return;
      }
      if (updateData.type === "exhausted") {
        if (countdownIntervalRef.current) {
          clearInterval(countdownIntervalRef.current);
          countdownIntervalRef.current = null;
        }
        if (resultsPollingRef.current) {
          clearInterval(resultsPollingRef.current);
          resultsPollingRef.current = null;
        }
        setWaitingForResponses(false);
        setStatusNote("No nearby pharmacy could fulfill this request. Try a smaller quantity or try again later.");
        toast({
          title: "No pharmacy available",
          description: "All nearby pharmacies carrying the medication were tried.",
          variant: "destructive",
        });
        return;
      }
      if (updateData.type !== "new_response") return;

      const result = await fetchAlertResults(alertId);
      if (!result) return;

      const pharmacies: PharmacyResponse[] = result.pharmacies || [];
      setAlertResult((prev) =>
        prev
          ? { ...prev, status: result.status || prev.status, pharmacies }
          : null,
      );

      // Only an "available" / "substitute" answer ends the wait. An
      // "unavailable" one just means the server moves to the next pharmacy.
      const hasOffer = pharmacies.some((p) => p.response_type !== "unavailable");
      if (hasOffer) {
        if (countdownIntervalRef.current) {
          clearInterval(countdownIntervalRef.current);
          countdownIntervalRef.current = null;
        }
        if (resultsPollingRef.current) {
          clearInterval(resultsPollingRef.current);
          resultsPollingRef.current = null;
        }
        setWaitingForResponses(false);
        setStatusNote(null);
      }

      const justArrived = pharmacies.filter(
        (p) => !seenPharmacyIdsRef.current.has(p.pharmacy_id),
      );
      pharmacies.forEach((p) => seenPharmacyIdsRef.current.add(p.pharmacy_id));

      justArrived.forEach((p) => {
        if (p.response_type === "available") {
          toast({
            title: "Medication available!",
            description: `${p.pharmacy_name} (ID: ${p.pharmacy_id}) has it in stock.`,
          });
        } else if (p.response_type === "substitute") {
          toast({
            title: "Alternative available",
            description: `${p.pharmacy_name} (ID: ${p.pharmacy_id}) suggested a substitute.`,
          });
        } else {
          toast({
            title: "Not available",
            description: `${p.pharmacy_name} (ID: ${p.pharmacy_id}) cannot fulfill the requested quantity.`,
            variant: "destructive",
          });
        }
      });
    };

    // Backup polling protects against a response being created before the SSE
    // connection is fully established (important for automatic responses from
    // pharmacies without accounts). SSE remains the fast path.
    if (resultsPollingRef.current) clearInterval(resultsPollingRef.current);
    resultsPollingRef.current = setInterval(async () => {
      const result = await fetchAlertResults(alertId);
      if (!result || eventSourceRef.current !== eventSource) return;

      const pharmacies: PharmacyResponse[] = result.pharmacies || [];
      setAlertResult((prev) =>
        prev
          ? { ...prev, status: result.status || prev.status, pharmacies }
          : null,
      );

      const hasOffer = pharmacies.some((p) => p.response_type !== "unavailable");
      const searchEnded = ["expired", "completed", "cancelled"].includes(result.status);
      if (hasOffer || searchEnded) {
        if (countdownIntervalRef.current) {
          clearInterval(countdownIntervalRef.current);
          countdownIntervalRef.current = null;
        }
        if (resultsPollingRef.current) {
          clearInterval(resultsPollingRef.current);
          resultsPollingRef.current = null;
        }
        setWaitingForResponses(false);
        setStatusNote(
          hasOffer
            ? null
            : "No nearby pharmacy could fulfill the requested quantity. Try a smaller quantity or a larger search radius.",
        );
      }
    }, 2000);

    eventSource.onerror = (err) => {
      console.error("[alert] SSE connection error for alert", alertId, err, "readyState:", eventSource.readyState);
    };

    // Safety net: close the stream after 2 minutes regardless.
    setTimeout(() => {
      eventSource.close();
    }, 2 * 60 * 1000);
  };

  const handleConfirmPickup = async (pharmacy: PharmacyResponse) => {
    if (!alertResult) return;

    const rawQuantity = confirmQuantities[pharmacy.pharmacy_id] ?? String(Math.max(parseInt(requestedQuantity, 10) || 1, 1));
    const quantity = parseInt(rawQuantity, 10);

    if (!rawQuantity || isNaN(quantity) || quantity <= 0) {
      toast({
        title: "Enter a quantity",
        description: "Please enter how many you'd like to take.",
        variant: "destructive",
      });
      return;
    }

    if (pharmacy.available_quantity > 0 && quantity > pharmacy.available_quantity) {
      toast({
        title: "Not enough in stock",
        description: `${pharmacy.pharmacy_name} only has ${pharmacy.available_quantity} available.`,
        variant: "destructive",
      });
      return;
    }

    setConfirmingPharmacyId(pharmacy.pharmacy_id);
    try {
      const response = await fetch(
        `${BASE_URL}/alerts/${alertResult.alert_id}/confirm`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            pharmacy_id: pharmacy.pharmacy_id,
            quantity,
          }),
        },
      );

      if (response.ok) {
        setConfirmedPharmacyId(pharmacy.pharmacy_id);
        toast({
          title: "Pickup confirmed",
          description: `You're all set to pick up ${quantity} from ${pharmacy.pharmacy_name}.`,
        });
      } else {
        const errorData = await response.json();
        toast({
          title: "Couldn't confirm pickup",
          description:
            errorData.error?.message || errorData.error || "Please try again.",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error("Error confirming pickup:", error);
      toast({
        title: "Network Error",
        description: "Unable to confirm pickup. Please check your connection.",
        variant: "destructive",
      });
    } finally {
      setConfirmingPharmacyId(null);
    }
  };

  const handleMedicationSelect = (medication: Medication) => {
    setSelectedMedication(medication);
    setQuery(medication.speciality);
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
                searchResults &&
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
                                      {medication.speciality}
                                    </p>
                                    <p className="text-xs text-slate-500 dark:text-slate-400">
                                      <span className="font-medium text-teal-600 dark:text-teal-400">
                                        {medication.presentation}
                                      </span>{" "}
                                      • {medication.dosage} • {medication.form} •{" "}
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

        <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm dark:border-slate-700 dark:bg-slate-900">
          <div className="mb-2 flex items-center justify-between gap-3">
            <div>
              <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                Choose your location on the map
              </p>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Click anywhere or drag the pin to set your location.
              </p>
            </div>
            <MapPin className="h-5 w-5 shrink-0 text-teal-600" />
          </div>

          <LocationMapPicker
            initialLat={parseFloat(userCoordinates.latitude)}
            initialLng={parseFloat(userCoordinates.longitude)}
            onLocationPicked={handleMapLocationPicked}
          />

          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
            Selected: {userCoordinates.latitude}, {userCoordinates.longitude}
          </p>
        </div>

        <div className="flex justify-end">
          <Button
            type="button"
            variant="link"
            size="sm"
            className="h-auto p-0 text-teal-700"
            onClick={getUserLocation}
          >
            Use my current location instead
          </Button>
        </div>

        {locationError && (
          <p className="text-sm text-red-500">{locationError}</p>
        )}

        {searchType === "medication" && (
          <div className="flex items-center gap-3">
            <label htmlFor="requested-quantity" className="text-sm text-slate-600 dark:text-slate-300">
              Quantity needed
            </label>
            <Input
              id="requested-quantity"
              type="number"
              min={1}
              className="h-10 w-24"
              value={requestedQuantity}
              onChange={(e) => setRequestedQuantity(e.target.value)}
            />
          </div>
        )}

        <Button
          type="submit"
          className="h-12 bg-teal-600 hover:bg-teal-700"
          disabled={
            (searchType === "medication" && !selectedMedication) || 
            creatingAlert
          }
        >
          {creatingAlert ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Creating Alert...
            </>
          ) : (
            <>
              {searchType === "medication" ? (
                <Bell className="mr-2 h-4 w-4" />
              ) : (
                <Search className="mr-2 h-4 w-4" />
              )}
              {searchType === "medication" ? "Create Alert" : "Search"}
            </>
          )}
        </Button>
      </form>

      {/* Alert Dialog */}
      <Dialog open={showAlertDialog} onOpenChange={setShowAlertDialog}>
        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5 text-teal-600" />
              Medication Alert: {selectedMedication?.speciality}
            </DialogTitle>
          </DialogHeader>
          
          {alertResult && (
            <div className="space-y-4">
              {/* Alert Status */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {waitingForResponses ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin text-teal-600" />
                      <span className="text-sm text-slate-600 dark:text-slate-300">
                        {statusNote || "Waiting for the nearest pharmacy to respond..."}
                      </span>
                    </>
                  ) : alertResult.pharmacies.length > 0 ? (
                    <>
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      <span className="text-sm text-green-600">
                        {alertResult.pharmacies.length} pharmacy(ies) responded
                      </span>
                    </>
                  ) : (
                    <>
                      <AlertTriangle className="h-4 w-4 text-amber-600" />
                      <span className="text-sm text-amber-600">
                        {statusNote || "No responses yet"}
                      </span>
                    </>
                  )}
                </div>
                
                {waitingForResponses && (
                  <div className="text-sm text-slate-500">
                    {Math.floor(timeRemaining / 60)}:{(timeRemaining % 60).toString().padStart(2, '0')}
                  </div>
                )}
              </div>

              {/* Progress indicator */}
              {waitingForResponses && (
                <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                  <div 
                    className="bg-teal-600 h-2 rounded-full transition-all duration-1000"
                    style={{ width: `${((ATTEMPT_SECONDS - timeRemaining) / ATTEMPT_SECONDS) * 100}%` }}
                  />
                </div>
              )}

              {/* Pharmacy Responses */}
              {alertResult.pharmacies.length > 0 && (
                <div className="space-y-3">
                  <h3 className="font-medium text-slate-900 dark:text-slate-100">
                    Pharmacy Responses
                  </h3>
                  {alertResult.pharmacies.map((pharmacy, index) => (
                    <Card key={index} className="border-l-4 border-l-teal-600">
                      <CardHeader className="pb-2">
                        <div className="flex items-start justify-between">
                          <div>
                            <CardTitle className="text-lg">
                              {pharmacy.pharmacy_name}{" "}
                              <span className="text-sm font-normal text-slate-500">
                                (ID: {pharmacy.pharmacy_id})
                              </span>
                            </CardTitle>
                            <p className="text-sm text-slate-600 dark:text-slate-300">
                              {pharmacy.pharmacy_address}
                            </p>
                            <p className="text-sm text-slate-500">
                              {pharmacy.distance_km.toFixed(1)} km away
                            </p>
                          </div>
                          <Badge
                            variant={pharmacy.response_type === 'unavailable' ? 'secondary' : 'default'}
                            className={
                              pharmacy.response_type === 'available'
                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                                : pharmacy.response_type === 'substitute'
                                ? 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300'
                                : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
                            }
                          >
                            {pharmacy.response_type === 'available'
                              ? 'In Stock'
                              : pharmacy.response_type === 'substitute'
                              ? 'Substitute Available'
                              : 'Not Available'}
                          </Badge>
                        </div>
                      </CardHeader>
                      {pharmacy.response_type !== 'unavailable' && (
                        <CardContent className="pt-0">
                          {pharmacy.substitute_brand && (
                            <p className="text-sm text-slate-600 dark:text-slate-300">
                              <span className="font-medium">Substitute:</span> {pharmacy.substitute_brand}
                            </p>
                          )}
                          {pharmacy.substitute_notes && (
                            <p className="text-sm text-slate-600 dark:text-slate-300 mt-1">
                              <span className="font-medium">Notes:</span> {pharmacy.substitute_notes}
                            </p>
                          )}
                          <div className="mt-3 flex flex-wrap items-center gap-2">
                            <Button 
                              size="sm" 
                              className="bg-teal-600 hover:bg-teal-700"
                              onClick={() => window.open(`tel:${pharmacy.pharmacy_phone}`, '_self')}
                            >
                              Call Pharmacy
                            </Button>
                            <Button 
                              size="sm" 
                              variant="outline"
                              onClick={() => window.open(
                                `https://maps.google.com?q=${pharmacy.latitude},${pharmacy.longitude}`,
                                '_blank'
                              )}
                            >
                              View on Map
                            </Button>
                          </div>
                          {pharmacy.response_type === 'available' && (
                            confirmedPharmacyId === pharmacy.pharmacy_id ? (
                              <div className="mt-3 flex items-center gap-2 text-sm text-green-700 dark:text-green-400">
                                <CheckCircle className="h-4 w-4" />
                                Pickup confirmed
                              </div>
                            ) : (
                              <div className="mt-3 flex items-center gap-2">
                                <Input
                                  type="number"
                                  min={1}
                                  max={pharmacy.available_quantity > 0 ? pharmacy.available_quantity : undefined}
                                  placeholder="Qty"
                                  className="h-8 w-20"
                                  value={confirmQuantities[pharmacy.pharmacy_id] ?? String(pharmacy.available_quantity > 0 ? Math.min(Math.max(parseInt(requestedQuantity, 10) || 1, 1), pharmacy.available_quantity) : (parseInt(requestedQuantity, 10) || 1))}
                                  onChange={(e) =>
                                    setConfirmQuantities((prev) => ({
                                      ...prev,
                                      [pharmacy.pharmacy_id]: e.target.value,
                                    }))
                                  }
                                />
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="border-teal-600 text-teal-700 hover:bg-teal-50 dark:text-teal-400"
                                  disabled={confirmingPharmacyId === pharmacy.pharmacy_id}
                                  onClick={() => handleConfirmPickup(pharmacy)}
                                >
                                  {confirmingPharmacyId === pharmacy.pharmacy_id ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                  ) : (
                                    "Confirm & Take"
                                  )}
                                </Button>
                                {pharmacy.available_quantity > 0 && (
                                  <span className="text-xs text-slate-500">
                                    {pharmacy.available_quantity} available
                                  </span>
                                )}
                              </div>
                            )
                          )}
                        </CardContent>
                      )}
                    </Card>
                  ))}
                </div>
              )}

              {/* No responses and time is up */}
              {!waitingForResponses && alertResult.pharmacies.length === 0 && (
                <Alert>
                  <AlertTriangle className="h-4 w-4" />
                  <AlertDescription>
                    No pharmacies responded to your alert. You may want to try again or search for alternative medications.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
