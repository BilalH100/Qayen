"use client";
import { Search, Filter, MapPin } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { PharmacyCard } from "@/components/pharmacy-card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useState, useEffect } from "react"
import axios from "axios"
import { BASE_URL } from "@/utils/api"

interface Pharmacy {
  name: string;
  city: string;
  latitude: string;
  longitude: string;
  address: string;
  phone: string;
  activity?: string;
}

export default function PharmaciesPage() {
  const [pharmacies, setPharmacies] = useState<Pharmacy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [filteredPharmacies, setFilteredPharmacies] = useState<Pharmacy[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(9); // 3x3 grid

  // Calculate pagination
  const totalPages = Math.ceil(filteredPharmacies.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentPharmacies = filteredPharmacies.slice(startIndex, endIndex);

  useEffect(() => {
    const fetchPharmacies = async () => {
      try {
        setLoading(true);
        const response = await axios.get(`${BASE_URL}/pharmacies`);
        setPharmacies(response.data);
        setFilteredPharmacies(response.data);
      } catch (err) {
        console.error("Error fetching pharmacies:", err);
        setError("Failed to load pharmacies");
      } finally {
        setLoading(false);
      }
    };

    fetchPharmacies();
  }, []);

  useEffect(() => {
    const filtered = pharmacies.filter(pharmacy =>
      pharmacy.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pharmacy.address.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pharmacy.city.toLowerCase().includes(searchTerm.toLowerCase())
    );
    setFilteredPharmacies(filtered);
    setCurrentPage(1); // Reset to first page when search changes
  }, [searchTerm, pharmacies]);

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-6">Pharmacies</h1>
        <div className="flex justify-center items-center h-64">
          <div className="text-lg">Loading pharmacies...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-6">Pharmacies</h1>
        <div className="flex justify-center items-center h-64">
          <div className="text-lg text-red-600">{error}</div>
        </div>
      </div>
    );
  }
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Pharmacies</h1>
      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input 
            placeholder="Search pharmacies..." 
            className="pl-10" 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div className="relative md:w-1/3">
          <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input placeholder="Your location" className="pl-10" />
        </div>
        <div className="flex gap-2">
          <Select defaultValue="distance">
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="distance">Distance</SelectItem>
              <SelectItem value="rating">Rating</SelectItem>
              <SelectItem value="name">Name</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon">
            <Filter className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <Tabs defaultValue="grid" className="mb-8">
        <div className="flex justify-between items-center mb-4">
          <TabsList>
            <TabsTrigger value="grid">Grid View</TabsTrigger>
            <TabsTrigger value="map">Map View</TabsTrigger>
          </TabsList>
          {filteredPharmacies.length > 0 && (
            <div className="text-sm text-slate-600 dark:text-slate-400">
              Showing {startIndex + 1}-{Math.min(endIndex, filteredPharmacies.length)} of {filteredPharmacies.length} pharmacies
            </div>
          )}
        </div>
        <TabsContent value="grid">
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {currentPharmacies.length > 0 ? (
              currentPharmacies.map((pharmacy, index) => (
                <PharmacyCard
                  key={startIndex + index}
                  name={pharmacy.name}
                  image="/placeholder.svg?height=200&width=300"
                  address={pharmacy.address}
                  rating={4.5} // Default rating since not in backend data
                  openUntil="9:00 PM" // Default hours since not in backend data
                  distance="N/A" // Would need user location to calculate
                />
              ))
            ) : (
              <div className="col-span-full text-center py-8">
                <p className="text-lg text-slate-600 dark:text-slate-400">
                  {searchTerm ? "No pharmacies found matching your search." : "No pharmacies available."}
                </p>
              </div>
            )}
          </div>
        </TabsContent>
        <TabsContent value="map">
          <div className="bg-slate-100 dark:bg-slate-800 rounded-lg h-[600px] flex items-center justify-center">
            <p className="text-slate-500 dark:text-slate-400">Map view would be displayed here</p>
          </div>
        </TabsContent>
      </Tabs>
      {totalPages > 1 && (
        <div className="flex justify-center mt-12">
          <nav className="flex items-center gap-1">
            <Button 
              variant="outline" 
              size="icon" 
              disabled={currentPage === 1}
              onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
            >
              &lt;
            </Button>
            {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
              let pageNum;
              if (totalPages <= 5) {
                pageNum = i + 1;
              } else if (currentPage <= 3) {
                pageNum = i + 1;
              } else if (currentPage >= totalPages - 2) {
                pageNum = totalPages - 4 + i;
              } else {
                pageNum = currentPage - 2 + i;
              }
              
              return (
                <Button
                  key={pageNum}
                  variant="outline"
                  size="icon"
                  className={currentPage === pageNum ? "bg-teal-600 text-white" : ""}
                  onClick={() => setCurrentPage(pageNum)}
                >
                  {pageNum}
                </Button>
              );
            })}
            <Button 
              variant="outline" 
              size="icon" 
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
            >
              &gt;
            </Button>
          </nav>
        </div>
      )}
    </div>
  )
}
