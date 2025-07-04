import { ArrowLeft, ShoppingCart, Heart, Share2, MapPin, Check, AlertCircle } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { MedicationCard } from "@/components/medication-card"

interface MedicationPageProps {
  params: {
    slug: string
  }
}

export default function MedicationPage({ params }: MedicationPageProps) {
  const medication = {
    name: params.slug
      .split("-")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" "),
    category: "Pain Relief",
    price: 9.99,
    description:
      "This medication is used to relieve pain from various conditions such as headache, dental pain, menstrual cramps, muscle aches, or arthritis. It may also be used to reduce fever.",
    dosage: "Take 1-2 tablets every 4-6 hours as needed. Do not exceed 6 tablets in 24 hours.",
    sideEffects:
      "Upset stomach, mild heartburn, nausea, vomiting, headache, diarrhea, constipation, dizziness, or drowsiness may occur.",
    warnings:
      "This medication may increase your risk of heart attack or stroke if you use it long term or have heart disease.",
    inStock: true,
    images: [
      "/placeholder.svg?height=400&width=400",
      "/placeholder.svg?height=400&width=400",
      "/placeholder.svg?height=400&width=400",
    ],
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
        {/* Product Images */}
        <div className="bg-slate-50 dark:bg-slate-900 rounded-xl p-8 flex items-center justify-center">
          <img
            src={medication.images[0] || "/placeholder.svg"}
            alt={medication.name}
            className="max-h-[400px] object-contain"
          />
        </div>

        {/* Product Info */}
        <div>
          <Badge className="mb-2 bg-teal-600">{medication.category}</Badge>
          <h1 className="text-3xl font-bold mb-2">{medication.name}</h1>
          <div className="text-2xl font-bold text-slate-900 dark:text-white mb-4">${medication.price.toFixed(2)}</div>

          <p className="text-slate-600 dark:text-slate-300 mb-6">{medication.description}</p>

          {medication.inStock ? (
            <div className="flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400 mb-6">
              <div className="bg-green-100 dark:bg-green-900/30 p-0.5 rounded-full">
                <Check className="h-3.5 w-3.5" />
              </div>
              <span>In Stock</span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400 mb-6">
              <div className="bg-red-100 dark:bg-red-900/30 p-0.5 rounded-full">
                <AlertCircle className="h-3.5 w-3.5" />
              </div>
              <span>Out of Stock</span>
            </div>
          )}

          <div className="flex gap-4 mb-8">
            <Button className="flex-1 bg-teal-600 hover:bg-teal-700">
              <ShoppingCart className="mr-2 h-4 w-4" /> Add to Cart
            </Button>
            <Button variant="outline" size="icon">
              <Heart className="h-4 w-4" />
            </Button>
            <Button variant="outline" size="icon">
              <Share2 className="h-4 w-4" />
            </Button>
          </div>

          <div className="border-t border-slate-200 dark:border-slate-700 pt-6">
            <h3 className="font-semibold mb-2">Find at nearby pharmacies</h3>
            <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300 mb-4">
              <MapPin className="h-4 w-4" />
              <span>New York, NY 10001</span>
              <Button variant="link" className="p-0 h-auto text-teal-600 dark:text-teal-400">
                Change
              </Button>
            </div>
            <div className="grid grid-cols-1 gap-4">
              <div className="flex justify-between items-center p-3 border rounded-lg">
                <div>
                  <div className="font-medium">MediCare Plus</div>
                  <div className="text-sm text-slate-500 dark:text-slate-400">0.8 miles away</div>
                </div>
                <div className="text-right">
                  <div className="font-bold">${medication.price.toFixed(2)}</div>
                  <div className="text-sm text-green-600 dark:text-green-400">In Stock</div>
                </div>
              </div>
              <div className="flex justify-between items-center p-3 border rounded-lg">
                <div>
                  <div className="font-medium">Wellness Pharmacy</div>
                  <div className="text-sm text-slate-500 dark:text-slate-400">1.2 miles away</div>
                </div>
                <div className="text-right">
                  <div className="font-bold">${(medication.price + 0.5).toFixed(2)}</div>
                  <div className="text-sm text-green-600 dark:text-green-400">In Stock</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Tabs defaultValue="details" className="mb-12">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="dosage">Dosage</TabsTrigger>
          <TabsTrigger value="side-effects">Side Effects</TabsTrigger>
          <TabsTrigger value="warnings">Warnings</TabsTrigger>
        </TabsList>
        <TabsContent value="details" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Product Details</h3>
          <p className="text-slate-600 dark:text-slate-300">{medication.description}</p>
          <ul className="list-disc list-inside mt-4 space-y-2 text-slate-600 dark:text-slate-300">
            <li>Active ingredient: Ibuprofen 200mg</li>
            <li>Non-steroidal anti-inflammatory drug (NSAID)</li>
            <li>Temporarily relieves minor aches and pains</li>
            <li>Reduces fever</li>
          </ul>
        </TabsContent>
        <TabsContent value="dosage" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Recommended Dosage</h3>
          <p className="text-slate-600 dark:text-slate-300">{medication.dosage}</p>
          <div className="mt-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-900 rounded-lg">
            <h4 className="font-semibold text-yellow-800 dark:text-yellow-400 mb-2">Important Note</h4>
            <p className="text-yellow-700 dark:text-yellow-300 text-sm">
              Always follow your doctor's instructions or the directions on the label. Do not take more than the
              recommended dose.
            </p>
          </div>
        </TabsContent>
        <TabsContent value="side-effects" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Possible Side Effects</h3>
          <p className="text-slate-600 dark:text-slate-300 mb-4">{medication.sideEffects}</p>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-4 border rounded-lg">
              <h4 className="font-semibold mb-2">Common Side Effects</h4>
              <ul className="list-disc list-inside space-y-1 text-slate-600 dark:text-slate-300 text-sm">
                <li>Upset stomach</li>
                <li>Mild heartburn</li>
                <li>Nausea</li>
                <li>Headache</li>
              </ul>
            </div>
            <div className="p-4 border rounded-lg">
              <h4 className="font-semibold mb-2">Serious Side Effects</h4>
              <ul className="list-disc list-inside space-y-1 text-slate-600 dark:text-slate-300 text-sm">
                <li>Allergic reactions</li>
                <li>Stomach bleeding</li>
                <li>Liver problems</li>
                <li>High blood pressure</li>
              </ul>
            </div>
          </div>
        </TabsContent>
        <TabsContent value="warnings" className="p-6 border rounded-b-lg">
          <h3 className="text-xl font-semibold mb-4">Warnings & Precautions</h3>
          <p className="text-slate-600 dark:text-slate-300 mb-4">{medication.warnings}</p>
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900 rounded-lg mb-4">
            <h4 className="font-semibold text-red-800 dark:text-red-400 mb-2">Do not use if:</h4>
            <ul className="list-disc list-inside space-y-1 text-red-700 dark:text-red-300 text-sm">
              <li>You are allergic to ibuprofen or any other NSAID</li>
              <li>You have had asthma, hives, or other allergic reactions after taking aspirin or other NSAIDs</li>
              <li>You are about to have heart surgery</li>
            </ul>
          </div>
          <p className="text-slate-600 dark:text-slate-300">
            Ask a doctor before use if you have stomach problems, heart disease, high blood pressure, or other medical
            conditions.
          </p>
        </TabsContent>
      </Tabs>

      {/* Related Products */}
      <section>
        <h2 className="text-2xl font-bold mb-6">Related Products</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <MedicationCard
            name="Aspirin"
            image="/placeholder.svg?height=200&width=200"
            category="Pain Relief"
            price={6.99}
            inStock={true}
          />
          <MedicationCard
            name="Acetaminophen"
            image="/placeholder.svg?height=200&width=200"
            category="Pain Relief"
            price={7.49}
            inStock={true}
          />
          <MedicationCard
            name="Naproxen"
            image="/placeholder.svg?height=200&width=200"
            category="Pain Relief"
            price={10.99}
            inStock={true}
          />
          <MedicationCard
            name="Diclofenac"
            image="/placeholder.svg?height=200&width=200"
            category="Pain Relief"
            price={12.99}
            inStock={false}
          />
        </div>
      </section>
    </div>
  )
}
