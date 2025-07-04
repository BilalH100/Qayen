import Link from "next/link"
import Image from "next/image"
import { MapPin, ArrowRight, Pill, User } from "lucide-react"

import { Button } from "@/components/ui/button"
import { PharmacyCard } from "@/components/pharmacy-card"
import { MedicationCard } from "@/components/medication-card"
import { HeroSearch } from "@/components/hero-search"
import { CategoryPill } from "@/components/category-pill"
import logo from "../public/logo.png"

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col">
      <section className="relative bg-gradient-to-r from-teal-50 to-blue-50 dark:from-slate-900 dark:to-slate-800">
        <div className="container mx-auto px-4 py-20 md:py-32">
          <div className="grid gap-8 md:grid-cols-2 md:gap-12 items-center">
            <div className="space-y-6">
              <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white md:text-5xl lg:text-6xl">
                Find the right <span className="text-teal-600 dark:text-teal-400">medication</span> at the right{" "}
                <span className="text-teal-600 dark:text-teal-400">pharmacy</span>
              </h1>
              <p className="text-lg text-slate-600 dark:text-slate-300 md:text-xl">
                Access thousands of medications and find nearby pharmacies with real-time availability and pricing.
              </p>
              <HeroSearch />
            </div>
            <div className="hidden md:block relative">
              {/* <div className="absolute -top-6 -left-6 w-24 h-24 bg-teal-200 dark:bg-teal-900 rounded-full opacity-50"></div> */}
              {/* <div className="absolute -bottom-6 -right-6 w-32 h-32 bg-blue-200 dark:bg-blue-900 rounded-full opacity-50"></div> */}
              <Image
                src={logo}
                alt="Pharmacy illustration"
                className="relative z-10 rounded-2xl shadow-xl"
              />
            </div>
          </div>
        </div>
      </section>
      <section className="py-12 bg-white dark:bg-slate-950">
        <div className="container mx-auto px-4">
          <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-8">Browse by Category</h2>
          <div className="flex flex-wrap gap-3">
            <CategoryPill label="Pain Relief" count={124} />
            <CategoryPill label="Antibiotics" count={86} />
            <CategoryPill label="Vitamins" count={210} />
            <CategoryPill label="Diabetes" count={75} />
            <CategoryPill label="Heart Health" count={92} />
            <CategoryPill label="Allergies" count={118} />
            <CategoryPill label="Skin Care" count={156} />
          </div>
        </div>
      </section>
      <section className="py-16 bg-slate-50 dark:bg-slate-900">
        <div className="container mx-auto px-4">
          <div className="flex justify-between items-center mb-8">
            <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Featured Pharmacies</h2>
            <Link
              href="/pharmacies"
              className="text-teal-600 dark:text-teal-400 flex items-center gap-1 hover:underline"
            >
              View all <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
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
          </div>
        </div>
      </section>
      <section className="py-16 bg-white dark:bg-slate-950">
        <div className="container mx-auto px-4">
          <div className="flex justify-between items-center mb-8">
            <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Popular Medications</h2>
            <Link
              href="/medications"
              className="text-teal-600 dark:text-teal-400 flex items-center gap-1 hover:underline"
            >
              View all <ArrowRight className="h-4 w-4" />
            </Link>
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
          </div>
        </div>
      </section>

      <section className="py-16 bg-gradient-to-r from-teal-600 to-blue-600 dark:from-teal-900 dark:to-blue-900 text-white">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-12">Why Choose Our Platform</h2>
          <div className="grid gap-8 md:grid-cols-3">
            <div className="bg-white/10 backdrop-blur-sm p-6 rounded-xl">
              <div className="bg-teal-500 dark:bg-teal-700 w-12 h-12 rounded-full flex items-center justify-center mb-4">
                <MapPin className="h-6 w-6 text-white" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Locate Nearby Pharmacies</h3>
              <p className="text-white/80">
                Find pharmacies near you with real-time availability of medications you need.
              </p>
            </div>
            <div className="bg-white/10 backdrop-blur-sm p-6 rounded-xl">
              <div className="bg-teal-500 dark:bg-teal-700 w-12 h-12 rounded-full flex items-center justify-center mb-4">
                <Pill className="h-6 w-6 text-white" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Medication Information</h3>
              <p className="text-white/80">
                Access detailed information about medications, including usage, side effects, and alternatives.
              </p>
            </div>
            <div className="bg-white/10 backdrop-blur-sm p-6 rounded-xl">
              <div className="bg-teal-500 dark:bg-teal-700 w-12 h-12 rounded-full flex items-center justify-center mb-4">
                <User className="h-6 w-6 text-white" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Personalized Experience</h3>
              <p className="text-white/80">
                Get personalized recommendations based on your medication history and preferences.
              </p>
            </div>
          </div>
        </div>
      </section>
      <section className="py-16 bg-white dark:bg-slate-950">
        <div className="container mx-auto px-4 text-center">
          <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-4">Ready to find your medication?</h2>
          <p className="text-lg text-slate-600 dark:text-slate-300 mb-8 max-w-2xl mx-auto">
            Join thousands of users who find the right medications at the best prices every day.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" className="bg-teal-600 hover:bg-teal-700 text-white">
              Find Medications
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-teal-600 text-teal-600 hover:bg-teal-50 dark:border-teal-400 dark:text-teal-400 dark:hover:bg-teal-950"
            >
              Locate Pharmacies
            </Button>
          </div>
        </div>
      </section>
    </div>
  )
}
