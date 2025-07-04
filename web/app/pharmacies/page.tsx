import { Search, Filter, MapPin } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { PharmacyCard } from "@/components/pharmacy-card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function PharmaciesPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Pharmacies</h1>
      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input placeholder="Search pharmacies..." className="pl-10" />
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
        <TabsList>
          <TabsTrigger value="grid">Grid View</TabsTrigger>
          <TabsTrigger value="map">Map View</TabsTrigger>
        </TabsList>
        <TabsContent value="grid">
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <PharmacyCard
              name="MediCare Plus"
              image="/placeholder.svg?height=200&width=300"
              address="123 Health Street, New York"
              rating={4.8}
              openUntil="9:00 PM"
              distance="0.8 miles"
            />
            <PharmacyCard
              name="Wellness Pharmacy"
              image="/placeholder.svg?height=200&width=300"
              address="456 Care Avenue, New York"
              rating={4.6}
              openUntil="10:00 PM"
              distance="1.2 miles"
            />
            <PharmacyCard
              name="City Health Dispensary"
              image="/placeholder.svg?height=200&width=300"
              address="789 Medical Blvd, New York"
              rating={4.9}
              openUntil="8:00 PM"
              distance="0.5 miles"
            />
            <PharmacyCard
              name="Central Pharmacy"
              image="/placeholder.svg?height=200&width=300"
              address="101 Main Street, New York"
              rating={4.5}
              openUntil="9:30 PM"
              distance="1.5 miles"
            />
            <PharmacyCard
              name="QuickMeds"
              image="/placeholder.svg?height=200&width=300"
              address="202 Fast Lane, New York"
              rating={4.3}
              openUntil="11:00 PM"
              distance="2.1 miles"
            />
            <PharmacyCard
              name="Family Care Pharmacy"
              image="/placeholder.svg?height=200&width=300"
              address="303 Community Road, New York"
              rating={4.7}
              openUntil="8:30 PM"
              distance="1.8 miles"
            />
          </div>
        </TabsContent>
        <TabsContent value="map">
          <div className="bg-slate-100 dark:bg-slate-800 rounded-lg h-[600px] flex items-center justify-center">
            <p className="text-slate-500 dark:text-slate-400">Map view would be displayed here</p>
          </div>
        </TabsContent>
      </Tabs>
      <div className="flex justify-center mt-12">
        <nav className="flex items-center gap-1">
          <Button variant="outline" size="icon" disabled>
            &lt;
          </Button>
          <Button variant="outline" size="icon" className="bg-teal-600 text-white">
            1
          </Button>
          <Button variant="outline" size="icon">
            2
          </Button>
          <Button variant="outline" size="icon">
            3
          </Button>
          <Button variant="outline" size="icon">
            &gt;
          </Button>
        </nav>
      </div>
    </div>
  )
}
