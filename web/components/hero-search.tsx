"use client"

import type React from "react"

import { useState } from "react"
import { Search, MapPin } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { useAuth } from "@/contexts/auth-context"

export function HeroSearch() {
  const [searchType, setSearchType] = useState<"medication" | "pharmacy">("medication")
  const [location, setLocation] = useState("")
  const [query, setQuery] = useState("")
  const { user, isAuthenticated } = useAuth()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    console.log(`Searching for ${searchType}: ${query} in ${location}`) 
  }

  return (
    <div className="space-y-4">
      {isAuthenticated && user && (
        <div className="mb-4">
          <p className="text-lg text-slate-600 dark:text-slate-300">
            Welcome back, <span className="font-semibold text-teal-600">{user.name.split(' ')[0]}</span>!
          </p>
        </div>
      )}
      
      <form onSubmit={handleSearch} className="flex flex-col gap-3 md:flex-row">
      <div className="relative flex-1">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="absolute left-0 top-0 h-full px-3 py-2 rounded-r-none border-r-0">
              {searchType === "medication" ? "Medication" : "Pharmacy"}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem onClick={() => setSearchType("medication")}>Medication</DropdownMenuItem>
            <DropdownMenuItem onClick={() => setSearchType("pharmacy")}>Pharmacy</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Input
          placeholder={searchType === "medication" ? "Search for medications..." : "Search for pharmacies..."}
          className="pl-32 h-12"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>
      <div className="relative md:w-1/3">
        <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
        <Input
          placeholder="Your location"
          className="pl-10 h-12"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
        />
      </div>
      <Button type="submit" className="h-12 bg-teal-600 hover:bg-teal-700">
        <Search className="mr-2 h-4 w-4" />Search
      </Button>
    </form>
    </div>
  )
}
