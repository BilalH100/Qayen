import { Search, Filter } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { MedicationCard } from "@/components/medication-card"
import { CategoryPill } from "@/components/category-pill"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

export default function MedicationsPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Medications</h1>

      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
          <Input placeholder="Search medications..." className="pl-10" />
        </div>
        <div className="flex gap-2">
          <Select defaultValue="all">
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Category" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Categories</SelectItem>
              <SelectItem value="pain-relief">Pain Relief</SelectItem>
              <SelectItem value="antibiotics">Antibiotics</SelectItem>
              <SelectItem value="vitamins">Vitamins</SelectItem>
              <SelectItem value="diabetes">Diabetes</SelectItem>
            </SelectContent>
          </Select>
          <Select defaultValue="relevance">
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="relevance">Relevance</SelectItem>
              <SelectItem value="price-low">Price: Low to High</SelectItem>
              <SelectItem value="price-high">Price: High to Low</SelectItem>
              <SelectItem value="name-asc">Name: A to Z</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon">
            <Filter className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <div className="flex flex-wrap gap-3 mb-8">
        <CategoryPill  label="All" count={1024} />
        <CategoryPill  label="Pain Relief" count={124} />
        <CategoryPill  label="Antibiotics" count={86} />
        <CategoryPill  label="Vitamins" count={210} />
        <CategoryPill  label="Diabetes" count={75} />
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <MedicationCard
          name="Ibuprofen"
          image="/placeholder.svg?height=200&width=200"
          category="Pain Relief"
          price={9.99}
          inStock={true}
        />
        <MedicationCard
          name="Amoxicillin"
          image="/placeholder.svg?height=200&width=200"
          category="Antibiotics"
          price={14.5}
          inStock={true}
        />
        <MedicationCard
          name="Vitamin D3"
          image="/placeholder.svg?height=200&width=200"
          category="Vitamins"
          price={12.99}
          inStock={true}
        />
        <MedicationCard
          name="Metformin"
          image="/placeholder.svg?height=200&width=200"
          category="Diabetes"
          price={8.75}
          inStock={false}
        />
        <MedicationCard
          name="Aspirin"
          image="/placeholder.svg?height=200&width=200"
          category="Pain Relief"
          price={6.99}
          inStock={true}
        />
        <MedicationCard
          name="Lisinopril"
          image="/placeholder.svg?height=200&width=200"
          category="Heart Health"
          price={11.25}
          inStock={true}
        />
        <MedicationCard
          name="Atorvastatin"
          image="/placeholder.svg?height=200&width=200"
          category="Heart Health"
          price={15.99}
          inStock={true}
        />
        <MedicationCard
          name="Cetirizine"
          image="/placeholder.svg?height=200&width=200"
          category="Allergies"
          price={8.49}
          inStock={true}
        />
      </div>

      {/* Pagination */}
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
