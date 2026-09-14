"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { SearchIcon, MapPinIcon, ClockIcon, PhoneIcon, CheckIcon, AlertTriangleIcon } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/contexts/auth-context";
import { BASE_URL } from "@/utils/api";

interface Medication {
  id: number;
  speciality: string;
  active_substance: string;
  dosage: string;
  form: string;
}

interface PharmacyAvailability {
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
  response_time_seconds: number;
}

interface AlertResult {
  alert_id: number;
  status: string;
  pharmacies: PharmacyAvailability[];
  created_at: string;
  expires_at: string;
}

export default function MedicationAlertSearch() {
  const [searchTerm, setSearchTerm] = useState("");
  const [medications, setMedications] = useState<Medication[]>([]);
  const [selectedMedication, setSelectedMedication] = useState<Medication | null>(null);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [locationLabel, setLocationLabel] = useState<string>("");
  const [addressInput, setAddressInput] = useState("");
  const [geocoding, setGeocoding] = useState(false);
  const [addressResults, setAddressResults] = useState<
    { label: string; lat: number; lng: number }[]
  >([]);
  const [searchRadius, setSearchRadius] = useState(10);
  const [alertResult, setAlertResult] = useState<AlertResult | null>(null);
  const [isSearching, setIsSearching] = useState(false);
  const [isWaiting, setIsWaiting] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState(0);
  const { toast } = useToast();
  const { user } = useAuth();

  // Try to get the user's real browser location as a starting point.
  // This is only a convenience default — since Kayena's seeded pharmacies
  // are currently all in Rabat, use the address search below to test
  // from anywhere else.
  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setLocation({
            lat: position.coords.latitude,
            lng: position.coords.longitude,
          });
          setLocationLabel("Your current location");
        },
        () => {
          // Silently ignore — the address search below covers this case.
        }
      );
    }
  }, []);

  // Debounced address -> coordinates lookup (OpenStreetMap Nominatim, free, no API key)
  useEffect(() => {
    if (addressInput.trim().length < 3) {
      setAddressResults([]);
      return;
    }
    const timeoutId = setTimeout(async () => {
      try {
        setGeocoding(true);
        const response = await fetch(
          `https://nominatim.openstreetmap.org/search?format=json&limit=5&q=${encodeURIComponent(
            addressInput
          )}`
        );
        const data = await response.json();
        setAddressResults(
          data.map((item: any) => ({
            label: item.display_name,
            lat: parseFloat(item.lat),
            lng: parseFloat(item.lon),
          }))
        );
      } catch (error) {
        console.error("Geocoding failed:", error);
      } finally {
        setGeocoding(false);
      }
    }, 500);
    return () => clearTimeout(timeoutId);
  }, [addressInput]);

  const chooseAddress = (result: { label: string; lat: number; lng: number }) => {
    setLocation({ lat: result.lat, lng: result.lng });
    setLocationLabel(result.label);
    setAddressInput("");
    setAddressResults([]);
  };

  // Search for medications
  const searchMedications = async (term: string) => {
    if (term.length < 2) {
      setMedications([]);
      return;
    }

    try {
      const response = await fetch(`${BASE_URL}/meds/search?q=${encodeURIComponent(term)}`);
      const data = await response.json();
      
      if (Array.isArray(data)) {
        setMedications(data.slice(0, 10)); // Limit to 10 results
      }
    } catch (error) {
      console.error("Failed to search medications:", error);
    }
  };

  // Debounced search
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      searchMedications(searchTerm);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchTerm]);

  // Create medication alert
  const createAlert = async () => {
    if (!selectedMedication || !location) {
      toast({
        title: "Missing Information",
        description: "Please select a medication and enable location services",
        variant: "destructive",
      });
      return;
    }

    setIsSearching(true);
    
    try {
      const response = await fetch(`${BASE_URL}/alerts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          customer_id: user?.id,
          medication_id: selectedMedication.id,
          latitude: location.lat,
          longitude: location.lng,
          search_radius_km: searchRadius,
          max_response_time_minutes: 2,
        }),
      });

      const data = await response.json();
      
      if (data.success) {
        setAlertResult(data.data);
        setIsWaiting(true);
        setTimeRemaining(2 * 60); // 2 minutes in seconds
        
        // Start countdown
        const countdownInterval = setInterval(() => {
          setTimeRemaining(prev => {
            if (prev <= 1) {
              clearInterval(countdownInterval);
              setIsWaiting(false);
              return 0;
            }
            return prev - 1;
          });
        }, 1000);

        // Set up real-time updates
        const eventSource = new EventSource(`/api/v1/alerts/${data.data.alert_id}/stream`);
        
        eventSource.onmessage = (event) => {
          const updateData = JSON.parse(event.data);
          if (updateData.type === "new_response") {
            // Refresh alert results
            fetchAlertResults(data.data.alert_id);
          }
        };

        // Clean up event source when component unmounts or search completes
        setTimeout(() => {
          eventSource.close();
        }, 2 * 60 * 1000); // 2 minutes

        toast({
          title: "Alert Sent",
          description: `Notifying nearby pharmacies about ${selectedMedication.speciality}`,
        });
      } else {
        throw new Error(data.error || "Failed to create alert");
      }
    } catch (error) {
      console.error("Failed to create alert:", error);
      toast({
        title: "Error",
        description: "Failed to send alert to pharmacies",
        variant: "destructive",
      });
    } finally {
      setIsSearching(false);
    }
  };

  // Fetch alert results
  const fetchAlertResults = async (alertId: number) => {
    if (!location) return;

    try {
      const response = await fetch(
        `${BASE_URL}/alerts/${alertId}/results?lat=${location.lat}&lng=${location.lng}`
      );
      const data = await response.json();
      
      if (data.success) {
        setAlertResult(data.data);
      }
    } catch (error) {
      console.error("Failed to fetch alert results:", error);
    }
  };

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  const getResponseIcon = (responseType: string) => {
    switch (responseType) {
      case "available":
        return <CheckIcon className="h-4 w-4 text-green-600" />;
      case "substitute":
        return <AlertTriangleIcon className="h-4 w-4 text-yellow-600" />;
      default:
        return null;
    }
  };

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="space-y-6">
        <div className="text-center space-y-2">
          <h1 className="text-3xl font-bold">Find Medication Nearby</h1>
          <p className="text-muted-foreground">
            Get real-time availability from pharmacies in your area
          </p>
        </div>

        {/* Search Interface */}
        {!alertResult && (
          <Card>
            <CardHeader>
              <CardTitle>Search for Medication</CardTitle>
              <CardDescription>
                Enter the medication name to find nearby pharmacies with availability
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="relative">
                <SearchIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Enter medication name..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>

              {/* Medication Results */}
              {medications.length > 0 && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Select Medication:</label>
                  <div className="grid gap-2">
                    {medications.map((med) => (
                      <div
                        key={med.id}
                        onClick={() => setSelectedMedication(med)}
                        className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                          selectedMedication?.id === med.id
                            ? "border-primary bg-primary/5"
                            : "hover:bg-muted"
                        }`}
                      >
                        <div className="font-medium">{med.speciality}</div>
                        <div className="text-sm text-muted-foreground">
                          {med.active_substance} • {med.dosage} • {med.form}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Location */}
              <div className="space-y-2">
                <label className="text-sm font-medium">Search location</label>
                <div className="relative">
                  <MapPinIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Type a city or address (e.g. Rabat, Morocco)..."
                    value={addressInput}
                    onChange={(e) => setAddressInput(e.target.value)}
                    className="pl-10"
                  />
                </div>
                {geocoding && (
                  <p className="text-xs text-muted-foreground">Searching...</p>
                )}
                {addressResults.length > 0 && (
                  <div className="border rounded-md divide-y max-h-48 overflow-y-auto">
                    {addressResults.map((result, idx) => (
                      <div
                        key={idx}
                        onClick={() => chooseAddress(result)}
                        className="p-2 text-sm cursor-pointer hover:bg-muted"
                      >
                        {result.label}
                      </div>
                    ))}
                  </div>
                )}
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span className="flex items-center">
                    <MapPinIcon className="h-4 w-4 mr-1" />
                    {location
                      ? locationLabel || "Custom location set"
                      : "No location set yet"}
                  </span>
                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    className="h-auto p-0"
                    onClick={() => {
                      if (navigator.geolocation) {
                        navigator.geolocation.getCurrentPosition(
                          (position) => {
                            setLocation({
                              lat: position.coords.latitude,
                              lng: position.coords.longitude,
                            });
                            setLocationLabel("Your current location");
                          },
                          () => {
                            toast({
                              title: "Location Error",
                              description:
                                "Couldn't access your current location. Try typing an address above instead.",
                              variant: "destructive",
                            });
                          }
                        );
                      }
                    }}
                  >
                    Use my current location
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Tip: Kayena's pharmacies are currently all in Rabat — search
                  "Rabat, Morocco" above to test regardless of where you
                  actually are.
                </p>
              </div>

              {/* Search Settings */}
              <div className="flex items-center gap-4">
                <div className="flex-1">
                  <label className="text-sm font-medium">Search Radius (km)</label>
                  <Input
                    type="number"
                    value={searchRadius}
                    onChange={(e) => setSearchRadius(parseInt(e.target.value) || 10)}
                    min="1"
                    max="50"
                    className="mt-1"
                  />
                </div>
              </div>

              <Button
                onClick={createAlert}
                disabled={!selectedMedication || !location || isSearching}
                className="w-full"
                size="lg"
              >
                {isSearching ? "Sending Alert..." : "Find Available Pharmacies"}
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Waiting for Responses */}
        {isWaiting && alertResult && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ClockIcon className="h-5 w-5" />
                Waiting for Pharmacy Responses
              </CardTitle>
              <CardDescription>
                Time remaining: {formatTime(timeRemaining)}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Alert>
                <AlertDescription>
                  We've notified nearby pharmacies about your request for{" "}
                  <strong>{selectedMedication?.speciality}</strong>. 
                  Responses will appear below as they come in.
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {alertResult && alertResult.pharmacies.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Available Pharmacies</CardTitle>
              <CardDescription>
                Pharmacies that have confirmed availability
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {alertResult.pharmacies.map((pharmacy) => (
                  <div
                    key={pharmacy.pharmacy_id}
                    className="border rounded-lg p-4 space-y-3"
                  >
                    <div className="flex items-start justify-between">
                      <div className="space-y-1">
                        <h3 className="font-semibold flex items-center gap-2">
                          {pharmacy.pharmacy_name}
                          {getResponseIcon(pharmacy.response_type)}
                          <Badge variant={pharmacy.response_type === "available" ? "default" : "secondary"}>
                            {pharmacy.response_type === "available" ? "Available" : "Substitute"}
                          </Badge>
                        </h3>
                        <p className="text-sm text-muted-foreground flex items-center gap-1">
                          <MapPinIcon className="h-3 w-3" />
                          {pharmacy.pharmacy_address} • {pharmacy.distance_km.toFixed(1)}km away
                        </p>
                        {pharmacy.substitute_brand && (
                          <p className="text-sm">
                            <strong>Alternative:</strong> {pharmacy.substitute_brand}
                            {pharmacy.substitute_notes && ` - ${pharmacy.substitute_notes}`}
                          </p>
                        )}
                      </div>
                      <div className="text-right space-y-1">
                        <div className="text-sm text-muted-foreground">
                          Responded in {pharmacy.response_time_seconds}s
                        </div>
                        <Button size="sm" className="flex items-center gap-1">
                          <PhoneIcon className="h-3 w-3" />
                          {pharmacy.pharmacy_phone}
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* No Results */}
        {alertResult && !isWaiting && alertResult.pharmacies.length === 0 && (
          <Card>
            <CardContent className="text-center py-8">
              <div className="space-y-2">
                <h3 className="font-semibold">No Responses Received</h3>
                <p className="text-muted-foreground">
                  Unfortunately, no nearby pharmacies confirmed availability for{" "}
                  <strong>{selectedMedication?.speciality}</strong> within the time limit.
                </p>
                <div className="pt-4">
                  <Button
                    onClick={() => {
                      setAlertResult(null);
                      setSelectedMedication(null);
                      setSearchTerm("");
                    }}
                    variant="outline"
                  >
                    Search Again
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
