"use client";

import {
  ArrowLeft,
  Clock,
  Package,
  Check,
  AlertCircle,
  MapPin,
  Navigation,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { MedicationCard } from "@/components/medication-card";
import { useEffect, useState, use } from "react";
import { BASE_URL } from "@/utils/api";
import { isMedicationAvailable } from "@/utils/types";

interface Medication {
  id: number;
  status: string;
  commercial_status: string;
  speciality: string;
  dosage: string;
  form: string;
  presentation: string;
  pp: string;
  active_substance: string;
  therapeutic_class: string;
  epi: string;
  ppv: string;
  ph: string;
  code: string;
  tva: string;
  created_at: string;
  pfht: string;
  description: string;
  common_sd: string[];
  serious_sd: string[];
  general_info: string[];
}

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

interface PharmacyWithDistance {
  distance: number;
  pharmacy: Pharmacy;
}

interface MedicationPageProps {
  params: {
    slug: string;
  };
}

export default function MedicationPage({ params }: MedicationPageProps) {
  const { slug } = use(params);
  const [medication, setMedication] = useState<Medication | null>(null);
  const [relatedMedications, setRelatedMedications] = useState<Medication[]>(
    [],
  );
  const [nearbyPharmacies, setNearbyPharmacies] = useState<
    PharmacyWithDistance[]
  >([]);
  // Hardcoded coordinates for Rabat, Morocco
  const [userCoordinates] = useState({
    latitude: "33.98978814461598",
    longitude: "-6.857842157069873",
  });
  const [loadingPharmacies, setLoadingPharmacies] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMedication = async () => {
      try {
        const response = await fetch(`${BASE_URL}/meds/code/${slug}`);

        if (!response.ok) {
          throw new Error("Medication not found");
        }

        const medicationData = await response.json();
        setMedication(medicationData);

        const relatedResponse = await fetch(`${BASE_URL}/meds/all`);
        if (relatedResponse.ok) {
          const allMedications = await relatedResponse.json();
          const related = allMedications
            .filter(
              (med: Medication) =>
                med.therapeutic_class === medicationData.therapeutic_class &&
                med.id !== medicationData.id,
            )
            .slice(0, 4);
          setRelatedMedications(related);
        }
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load medication",
        );
      } finally {
        setLoading(false);
      }
    };

    fetchMedication();

    // Using hardcoded coordinates for Rabat, Morocco
    console.log("Using hardcoded location: Rabat, Morocco");
  }, [slug]);

  // Fetch nearby pharmacies when we have both medication and user location
  useEffect(() => {
    const fetchNearbyPharmacies = async () => {
      if (!medication) return;

      setLoadingPharmacies(true);
      try {
        const url = `${BASE_URL}/pharmacies/closest?id=${medication.id}&latitude=${userCoordinates.latitude}&longitude=${userCoordinates.longitude}`;
        const response = await fetch(url);

        if (response.ok) {
          const data = await response.json();
          setNearbyPharmacies([data]); // Start with closest pharmacy
        }
      } catch (err) {
        console.error("Error fetching nearby pharmacies:", err);
      } finally {
        setLoadingPharmacies(false);
      }
    };

    fetchNearbyPharmacies();
  }, [medication, userCoordinates]);

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-teal-600 mx-auto mb-4"></div>
            <p className="text-slate-600 dark:text-slate-300">
              Loading medication...
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !medication) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Link
          href="/medications"
          className="flex items-center gap-2 text-teal-600 dark:text-teal-400 mb-6 hover:underline"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Medications
        </Link>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold mb-2">Medication Not Found</h1>
            <p className="text-slate-600 dark:text-slate-300 mb-4">
              {error || "The medication you're looking for doesn't exist."}
            </p>
            <Link href="/medications">
              <Button>Browse All Medications</Button>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Link
        href="/medications"
        className="flex items-center gap-2 text-teal-600 dark:text-teal-400 mb-6 hover:underline"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Medications
      </Link>

      <div className="grid md:grid-cols-2 gap-8 mb-12">
        <div className="bg-slate-50 dark:bg-slate-900 rounded-xl p-8 flex items-center justify-center">
          <div className="text-center">
            <Package className="h-24 w-24 text-slate-400 mx-auto mb-4" />
            <p className="text-slate-500">Medication Image</p>
          </div>
        </div>

        <div>
          <div className="flex flex-wrap gap-2 mb-4">
            <Badge className="bg-teal-600">
              {medication.therapeutic_class}
            </Badge>
            <Badge variant="outline">{medication.form}</Badge>
            {medication.status && (
              <Badge
                variant={
                  medication.status === "ACTIVE" ? "default" : "secondary"
                }
              >
                {medication.status}
              </Badge>
            )}
          </div>

          <h1 className="text-3xl font-bold mb-2">{medication.speciality}</h1>

          <div className="space-y-2 mb-6">
            <p className="text-slate-600 dark:text-slate-300">
              <span className="font-medium">Active Substance:</span>{" "}
              {medication.active_substance}
            </p>
            <p className="text-slate-600 dark:text-slate-300">
              <span className="font-medium">Dosage:</span> {medication.dosage}
            </p>
          </div>

          {medication.description && (
            <p className="text-slate-600 dark:text-slate-300 mb-6">
              {medication.description}
            </p>
          )}

          {medication.commercial_status && (
            <div className="flex items-center gap-1.5 text-sm mb-6">
              {isMedicationAvailable(medication.commercial_status) ? (
                <>
                  <div className="bg-green-100 dark:bg-green-900/30 p-0.5 rounded-full">
                    <Check className="h-3.5 w-3.5 text-green-600 dark:text-green-400" />
                  </div>
                  <span className="text-green-600 dark:text-green-400">
                    Available
                  </span>
                </>
              ) : (
                <>
                  <div className="bg-gray-100 dark:bg-gray-900/30 p-0.5 rounded-full">
                    <Clock className="h-3.5 w-3.5 text-gray-600 dark:text-gray-400" />
                  </div>
                  <span className="text-gray-600 dark:text-gray-400">
                    {medication.commercial_status}
                  </span>
                </>
              )}
            </div>
          )}

          <div className="space-y-3">
            {medication.pp && (
              <div className="text-sm">
                <span className="font-medium text-slate-700 dark:text-slate-300">
                  Public Price:
                </span>
                <span className="ml-2 text-slate-600 dark:text-slate-400">
                  {medication.pp} DZD
                </span>
              </div>
            )}
            {medication.ppv && (
              <div className="text-sm">
                <span className="font-medium text-slate-700 dark:text-slate-300">
                  Sale Price:
                </span>
                <span className="ml-2 text-slate-600 dark:text-slate-400">
                  {medication.ppv} DZD
                </span>
              </div>
            )}
            {medication.ph && (
              <div className="text-sm">
                <span className="font-medium text-slate-700 dark:text-slate-300">
                  Hospital Price:
                </span>
                <span className="ml-2 text-slate-600 dark:text-slate-400">
                  {medication.ph} DZD
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      <Tabs defaultValue="details" className="mb-12">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="dosage">Dosage</TabsTrigger>
          <TabsTrigger value="side-effects">Side Effects</TabsTrigger>
          <TabsTrigger value="general-info">General Info</TabsTrigger>
        </TabsList>
        <TabsContent value="details" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Product Details</h3>
          {medication.description ? (
            <p className="text-slate-600 dark:text-slate-300 mb-4">
              {medication.description}
            </p>
          ) : (
            <p className="text-slate-500 dark:text-slate-400 mb-4">
              No detailed description available.
            </p>
          )}
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h4 className="font-semibold mb-3">Basic Information</h4>
              <ul className="space-y-2 text-sm">
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Form:
                  </span>
                  <span className="font-medium">{medication.form}</span>
                </li>
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Status:
                  </span>
                  <span className="font-medium">{medication.status}</span>
                </li>
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Commercial Status:
                  </span>
                  <span className="font-medium">
                    {medication.commercial_status}
                  </span>
                </li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-3">Clinical Information</h4>
              <ul className="space-y-2 text-sm">
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Active Substance:
                  </span>
                  <span className="font-medium">
                    {medication.active_substance}
                  </span>
                </li>
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Therapeutic Class:
                  </span>
                  <span className="font-medium">
                    {medication.therapeutic_class}
                  </span>
                </li>
                <li className="flex justify-between">
                  <span className="text-slate-600 dark:text-slate-400">
                    Speciality:
                  </span>
                  <span className="font-medium">{medication.speciality}</span>
                </li>
              </ul>
            </div>
          </div>
        </TabsContent>
        <TabsContent value="dosage" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Dosage Information</h3>
          {medication.dosage ? (
            <div>
              <p className="text-slate-600 dark:text-slate-300 mb-4">
                <span className="font-medium">Dosage:</span> {medication.dosage}
              </p>
              <p className="text-slate-600 dark:text-slate-300 mb-4">
                <span className="font-medium">Presentation:</span>{" "}
                {medication.presentation}
              </p>
            </div>
          ) : (
            <p className="text-slate-500 dark:text-slate-400">
              No specific dosage information available.
            </p>
          )}
          <div className="mt-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-900 rounded-lg">
            <h4 className="font-semibold text-yellow-800 dark:text-yellow-400 mb-2">
              Important Note
            </h4>
            <p className="text-yellow-700 dark:text-yellow-300 text-sm">
              Always follow your doctor's instructions or consult with a
              healthcare professional for proper dosage. Do not self-medicate.
            </p>
          </div>
        </TabsContent>
        <TabsContent value="side-effects" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Side Effects</h3>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-4 border rounded-lg">
              <h4 className="font-semibold mb-2">Common Side Effects</h4>
              {medication.common_sd && medication.common_sd.length > 0 ? (
                <ul className="list-disc list-inside space-y-1 text-slate-600 dark:text-slate-300 text-sm">
                  {medication.common_sd.map((effect, index) => (
                    <li key={index}>{effect}</li>
                  ))}
                </ul>
              ) : (
                <p className="text-slate-500 dark:text-slate-400 text-sm">
                  No common side effects listed.
                </p>
              )}
            </div>
            <div className="p-4 border rounded-lg">
              <h4 className="font-semibold mb-2">Serious Side Effects</h4>
              {medication.serious_sd && medication.serious_sd.length > 0 ? (
                <ul className="list-disc list-inside space-y-1 text-slate-600 dark:text-slate-300 text-sm">
                  {medication.serious_sd.map((effect, index) => (
                    <li key={index}>{effect}</li>
                  ))}
                </ul>
              ) : (
                <p className="text-slate-500 dark:text-slate-400 text-sm">
                  No serious side effects listed.
                </p>
              )}
            </div>
          </div>
          <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900 rounded-lg">
            <h4 className="font-semibold text-red-800 dark:text-red-400 mb-2">
              Important Warning
            </h4>
            <p className="text-red-700 dark:text-red-300 text-sm">
              If you experience any severe or persistent side effects, stop
              taking this medication and consult your healthcare provider
              immediately.
            </p>
          </div>
        </TabsContent>
        <TabsContent value="general-info" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">General Information</h3>
          {medication.general_info && medication.general_info.length > 0 ? (
            <ul className="list-disc list-inside space-y-2 text-slate-600 dark:text-slate-300">
              {medication.general_info.map((info, index) => (
                <li key={index}>{info}</li>
              ))}
            </ul>
          ) : (
            <p className="text-slate-500 dark:text-slate-400">
              No additional general information available.
            </p>
          )}
          <div className="mt-6 grid md:grid-cols-2 gap-4">
            {medication.epi && (
              <div className="p-4 bg-slate-50 dark:bg-slate-800 rounded-lg">
                <h4 className="font-semibold mb-2">EPI Code</h4>
                <p className="text-slate-600 dark:text-slate-300 text-sm">
                  {medication.epi}
                </p>
              </div>
            )}
            {medication.tva && (
              <div className="p-4 bg-slate-50 dark:bg-slate-800 rounded-lg">
                <h4 className="font-semibold mb-2">TVA</h4>
                <p className="text-slate-600 dark:text-slate-300 text-sm">
                  {medication.tva}
                </p>
              </div>
            )}
          </div>
        </TabsContent>
      </Tabs>

      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-6">Nearby Pharmacies</h2>
        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg border border-blue-200 dark:border-blue-800 mb-4">
          <p className="text-blue-700 dark:text-blue-400 flex items-center">
            <MapPin className="h-4 w-4 mr-2" />
            Using location: Rabat, Morocco (33.9898°N, -6.8578°W)
          </p>
        </div>
        {loadingPharmacies ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-teal-600 mx-auto mb-4"></div>
            <p className="text-slate-600 dark:text-slate-300">
              Finding nearby pharmacies...
            </p>
          </div>
        ) : nearbyPharmacies.length > 0 ? (
          <div className="space-y-4">
            {nearbyPharmacies.map((item) => (
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
        ) : (
          <div className="text-center py-8 bg-slate-50 dark:bg-slate-800 rounded-lg">
            <p className="text-slate-600 dark:text-slate-300">
              No pharmacies with this medication found near you.
            </p>
          </div>
        )}
      </section>

      <section>
        <h2 className="text-2xl font-bold mb-6">Related Medications</h2>
        {relatedMedications.length > 0 ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
            {relatedMedications.map((med) => (
              <MedicationCard
                key={med.id}
                id={med.id}
                name={med.presentation}
                category={med.therapeutic_class}
                form={med.form}
                presentation={med.presentation}
                status={med.commercial_status}
                code={med.code}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-slate-500 dark:text-slate-400">
              No related medications found.
            </p>
          </div>
        )}
      </section>
    </div>
  );
}
