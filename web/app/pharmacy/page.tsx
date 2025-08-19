'use client';

import { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Building, MapPin, Phone, Loader2 } from 'lucide-react';
import Link from 'next/link';
import { BASE_URL } from '@/utils/api';

interface Pharmacy {
  id: number;
  name: string;
  address: string;
  phone: string;
  latitude: number;
  longitude: number;
  city: string;
}

export default function PharmacySelection() {
  const [pharmacies, setPharmacies] = useState<Pharmacy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPharmacies = async () => {
      try {
        const response = await fetch(`${BASE_URL}/pharmacies`);
        if (response.ok) {
          const data = await response.json();
          setPharmacies(data.slice(0, 12)); // Limit to first 12 pharmacies for better UX
        } else {
          throw new Error('Failed to fetch pharmacies');
        }
      } catch (error) {
        console.error('Error fetching pharmacies:', error);
        setError(error instanceof Error ? error.message : 'Unknown error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchPharmacies();
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white dark:from-slate-900 dark:to-slate-800 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600 dark:text-gray-300">Loading pharmacies...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white dark:from-slate-900 dark:to-slate-800 flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-100 dark:bg-red-900/20 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-red-800 dark:text-red-400 mb-2">
              Error Loading Pharmacies
            </h2>
            <p className="text-red-600 dark:text-red-300 mb-4">{error}</p>
            <Button 
              onClick={() => window.location.reload()} 
              variant="outline"
              className="border-red-300 text-red-700 hover:bg-red-50 dark:border-red-600 dark:text-red-400"
            >
              Try Again
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Select Your Pharmacy
          </h1>
          <p className="text-gray-600 dark:text-gray-300">
            Choose your pharmacy to access the dashboard and manage medication alerts
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {pharmacies.map((pharmacy) => (
            <Card key={pharmacy.id} className="hover:shadow-lg transition-shadow border-slate-200 dark:border-slate-700">
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Building className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                  <span className="truncate">{pharmacy.name}</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-start gap-2 text-sm text-gray-600 dark:text-gray-300">
                    <MapPin className="h-4 w-4 mt-0.5 flex-shrink-0 text-gray-400" />
                    <span className="line-clamp-2">{pharmacy.address}</span>
                  </div>
                  
                  {pharmacy.city && (
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      <span className="font-medium">City:</span> {pharmacy.city}
                    </div>
                  )}

                  {pharmacy.phone && (
                    <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                      <Phone className="h-4 w-4 text-gray-400" />
                      <span>{pharmacy.phone}</span>
                    </div>
                  )}

                  <div className="pt-3">
                    <Link href={`/pharmacy/${pharmacy.id}/dashboard`}>
                      <Button className="w-full bg-blue-600 hover:bg-blue-700 dark:bg-blue-600 dark:hover:bg-blue-700">
                        Access Dashboard
                      </Button>
                    </Link>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {pharmacies.length === 0 && !loading && !error && (
          <div className="text-center py-12">
            <Building className="h-16 w-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              No pharmacies found
            </h3>
            <p className="text-gray-600 dark:text-gray-300">
              Please make sure the backend server is running and pharmacies are properly configured.
            </p>
          </div>
        )}
        
        {pharmacies.length > 0 && (
          <div className="text-center mt-8">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Showing {pharmacies.length} pharmacies
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
