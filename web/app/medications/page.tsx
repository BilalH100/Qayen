"use client";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { MedicationCard } from "@/components/medication-card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import axios from "axios";
import { BASE_URL } from "@/utils/api";
import React, { useEffect, useState, useCallback } from "react";

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

export default function MedicationsPage() {
  const [meds, setMeds] = React.useState<Medication[] | null>([]);
  const [allMeds, setAllMeds] = React.useState<Medication[] | null>([]);
  const [currentPage, setCurrentPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const [loading, setLoading] = React.useState(false);
  const [searchTerm, setSearchTerm] = React.useState("");
  const [filteredMeds, setFilteredMeds] = React.useState<Medication[] | null>([]);
  const [sortBy, setSortBy] = React.useState("relevance");
  const [selectedCategory, setSelectedCategory] = React.useState("all");
  const itemsPerPage = 20;

  const fetchMedications = async (page: number = 1) => {
    try {
      setLoading(true);
      const offset = (page - 1) * itemsPerPage;
      const response = await axios.get(
        `${BASE_URL}/meds/all?offset=${offset}&limit=${itemsPerPage}`,
      );

      if (!response.data || response.data.length === 0) {
        setMeds([]);
        if (page === 1) {
          setTotalPages(1);
        }
      } else {
        setMeds(response.data);
        if (page === 1 && !searchTerm) {
          setAllMeds(response.data);
        }
        if (response.data.length < itemsPerPage) {
          setTotalPages(page);
        } else {
          setTotalPages(Math.max(page + 1, totalPages));
        }
      }
      setCurrentPage(page);
    } catch (error) {
      console.error(error);
      setMeds([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchAllMedications = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`${BASE_URL}/meds/all?offset=0&limit=10000`);
      if (response.data && response.data.length > 0) {
        setAllMeds(response.data);
        return response.data;
      }
      return [];
    } catch (error) {
      console.error(error);
      return [];
    } finally {
      setLoading(false);
    }
  };

  // Handle sorting
  const handleSort = (medications: Medication[]) => {
    if (!medications) return medications;
    
    const sorted = [...medications];
    switch (sortBy) {
      case "name-asc":
        return sorted.sort((a, b) => a.speciality.localeCompare(b.speciality));
      case "name-desc":
        return sorted.sort((a, b) => b.speciality.localeCompare(a.speciality));
      default:
        return sorted;
    }
  };

  // Handle category filtering
  const handleCategoryFilter = (medications: Medication[]) => {
    if (!medications || selectedCategory === "all") return medications;
    
    return medications.filter(med => 
      med.therapeutic_class?.toLowerCase().includes(selectedCategory.toLowerCase())
    );
  };

  const debouncedSearch = useCallback(
    debounce((term: string) => {
      handleSearch(term);
    }, 300),
    [allMeds],
  );

  const handleSearchInput = (value: string) => {
    setSearchTerm(value);
    debouncedSearch(value);
  };

  const handleSearch = (term: string) => {
    if (!term.trim()) {
      setFilteredMeds([]);
      fetchMedications(1);
      return;
    }

    if (allMeds && allMeds.length > 0) {
      let filtered = allMeds.filter(
        (med) =>
          med.speciality?.toLowerCase().includes(term.toLowerCase()) ||
          med.active_substance?.toLowerCase().includes(term.toLowerCase()) ||
          med.therapeutic_class?.toLowerCase().includes(term.toLowerCase()) ||
          med.form?.toLowerCase().includes(term.toLowerCase()) ||
          med.presentation?.toLowerCase().includes(term.toLowerCase()) ||
          med.code?.toLowerCase().includes(term.toLowerCase()),
      );
      
      // Apply category filter
      filtered = handleCategoryFilter(filtered);
      
      // Apply sorting
      filtered = handleSort(filtered);
      
      setFilteredMeds(filtered);
      const totalFilteredPages = Math.ceil(filtered.length / itemsPerPage);
      setTotalPages(totalFilteredPages || 1);
      setCurrentPage(1);
      const firstPageFiltered = filtered.slice(0, itemsPerPage);
      setMeds(firstPageFiltered);
    }
  };

  // Handle category change
  const handleCategoryChange = (category: string) => {
    setSelectedCategory(category);
    setCurrentPage(1);
    
    if (searchTerm) {
      handleSearch(searchTerm);
    } else {
      // Apply filter to current medications
      if (allMeds && allMeds.length > 0) {
        let filtered = handleCategoryFilter(allMeds);
        filtered = handleSort(filtered);
        
        const totalFilteredPages = Math.ceil(filtered.length / itemsPerPage);
        setTotalPages(totalFilteredPages || 1);
        
        const firstPageFiltered = filtered.slice(0, itemsPerPage);
        setMeds(firstPageFiltered);
        setFilteredMeds(filtered);
      }
    }
  };

  // Handle sort change
  const handleSortChange = (sort: string) => {
    setSortBy(sort);
    
    if (searchTerm) {
      handleSearch(searchTerm);
    } else if (filteredMeds && filteredMeds.length > 0) {
      const sorted = handleSort(filteredMeds);
      setFilteredMeds(sorted);
      const firstPageSorted = sorted.slice(0, itemsPerPage);
      setMeds(firstPageSorted);
    } else if (allMeds && allMeds.length > 0) {
      let filtered = handleCategoryFilter(allMeds);
      filtered = handleSort(filtered);
      
      const totalFilteredPages = Math.ceil(filtered.length / itemsPerPage);
      setTotalPages(totalFilteredPages || 1);
      
      const firstPageFiltered = filtered.slice(0, itemsPerPage);
      setMeds(firstPageFiltered);
      setFilteredMeds(filtered);
    }
  };

  function debounce(func: Function, wait: number) {
    let timeout: NodeJS.Timeout;
    return function executedFunction(...args: any[]) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  const handleSearchPagination = (page: number) => {
    if (searchTerm && filteredMeds) {
      const startIndex = (page - 1) * itemsPerPage;
      const endIndex = startIndex + itemsPerPage;
      const pageData = filteredMeds.slice(startIndex, endIndex);
      setMeds(pageData);
      setCurrentPage(page);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
    fetchMedications(1);
    fetchAllMedications();
  }, []);

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages && page !== currentPage) {
      window.scrollTo({ top: 0, behavior: "smooth" });
      
      if (searchTerm) {
        handleSearchPagination(page);
      } else {
        fetchMedications(page);
      }
    }
  };

  const generatePageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;

    if (totalPages <= maxVisiblePages) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      let startPage = Math.max(1, currentPage - 2);
      let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

      if (endPage - startPage < maxVisiblePages - 1) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }

      for (let i = startPage; i <= endPage; i++) {
        pages.push(i);
      }
    }

    return pages;
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Medications</h1>
      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input
            placeholder="Search medications..."
            className="pl-10"
            value={searchTerm}
            onChange={(e) => handleSearchInput(e.target.value)}
          />
        </div>
        <div className="flex gap-2">
          <Select value={selectedCategory} onValueChange={handleCategoryChange}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Category" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Categories</SelectItem>
              <SelectItem value="pain relief">Pain Relief</SelectItem>
              <SelectItem value="antibiotics">Antibiotics</SelectItem>
              <SelectItem value="vitamins">Vitamins</SelectItem>
              <SelectItem value="diabetes">Diabetes</SelectItem>
              <SelectItem value="cardiovascular">Cardiovascular</SelectItem>
              <SelectItem value="dermatology">Dermatology</SelectItem>
            </SelectContent>
          </Select>
          <Select value={sortBy} onValueChange={handleSortChange}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="relevance">Relevance</SelectItem>
              <SelectItem value="name-asc">Name: A to Z</SelectItem>
              <SelectItem value="name-desc">Name: Z to A</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      
      <div className="flex flex-wrap gap-3 mb-8">
        {searchTerm ? (
          <div className="flex items-center gap-2 px-3 py-1 bg-teal-100 text-teal-800 rounded-full text-sm">
            Searching for: "{searchTerm}"
            <button
              onClick={() => {
                setSearchTerm("");
                handleSearch("");
              }}
              className="ml-1 text-teal-600 hover:text-teal-800"
            >
              ✕
            </button>
          </div>
        ) : selectedCategory !== "all" ? (
          <div className="flex items-center gap-2 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
            Category: {selectedCategory}
            <button
              onClick={() => handleCategoryChange("all")}
              className="ml-1 text-blue-600 hover:text-blue-800"
            >
              ✕
            </button>
          </div>
        ) : null}
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {loading ? (
          Array.from({ length: 8 }).map((_, index) => (
            <div
              key={index}
              className="bg-gray-200 animate-pulse rounded-lg h-64"
            ></div>
          ))
        ) : meds && meds.length > 0 ? (
          meds.map((med) => (
            <MedicationCard
              key={med.id}
              id={med.id}
              name={med.speciality}
              category={med.therapeutic_class}
              form={med.form}
              presentation={med.presentation}
              status={med.status}
              code={med.code}
            />
          ))
        ) : (
          <div className="col-span-full text-center py-12">
            <p className="text-slate-500">No medications found</p>
          </div>
        )}
      </div>

      {/* Pagination */}
      <div className="flex justify-center mt-12">
        <nav className="flex items-center gap-1">
          <Button 
            variant="outline" 
            size="icon" 
            disabled={currentPage === 1 || loading}
            onClick={() => handlePageChange(currentPage - 1)}
          >
            &lt;
          </Button>
          
          {generatePageNumbers().map((pageNum) => (
            <Button
              key={pageNum}
              variant="outline"
              size="icon"
              className={currentPage === pageNum ? "bg-teal-600 text-white" : ""}
              onClick={() => handlePageChange(pageNum)}
              disabled={loading}
            >
              {pageNum}
            </Button>
          ))}
          
          <Button 
            variant="outline" 
            size="icon"
            disabled={currentPage === totalPages || loading}
            onClick={() => handlePageChange(currentPage + 1)}
          >
            &gt;
          </Button>
        </nav>
      </div>

      {/* Page Info */}
      <div className="text-center mt-4 text-sm text-slate-600">
        {searchTerm ? (
          <>
            Page {currentPage} of {totalPages} • Showing {meds?.length || 0} of {filteredMeds?.length || 0} search results
          </>
        ) : (
          <>
            Page {currentPage} of {totalPages} • Showing {meds?.length || 0} medications
          </>
        )}
      </div>
    </div>
  );
}
