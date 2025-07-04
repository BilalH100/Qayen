import { ArrowLeft, Star, Check, MapPin, Phone, Globe, Clock } from "lucide-react"
import Link from "next/link"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Avatar } from "@/components/ui/avatar"

const getServiceDescription = (service: string): string => {
  const descriptions: Record<string, string> = {
    "Prescription Filling": "Quick and accurate filling of all your prescription medications with personal consultation.",
    "Medication Counseling": "Expert advice on proper medication use, potential side effects, and drug interactions.",
    "Immunizations": "Convenient vaccinations for flu, pneumonia, shingles, and other preventable diseases.",
    "Health Screenings": "Regular health checks including blood pressure, cholesterol, and diabetes screenings.",
    "Compounding": "Custom preparation of medications tailored to your specific needs and preferences.",
    "Home Delivery": "Free medication delivery service to your doorstep for added convenience."
  }
  
  return descriptions[service] || "Professional pharmacy service to meet your healthcare needs."
}

const dummyReviews = [
  {
    name: "Sarah Johnson",
    avatar: "/placeholder-user.jpg",
    rating: 5,
    date: "May 15, 2025",
    comment: "Excellent service! The pharmacists are very knowledgeable and always take time to explain my medications. The home delivery option is so convenient."
  },
  {
    name: "Michael Chen",
    avatar: "/placeholder-user.jpg",
    rating: 4,
    date: "April 23, 2025",
    comment: "Very professional staff and quick service. I appreciate how they always check for potential drug interactions with my existing medications."
  },
  {
    name: "Emily Rodriguez",
    avatar: "/placeholder-user.jpg",
    rating: 5,
    date: "June 2, 2025",
    comment: "This pharmacy has been a lifesaver with their compounding services. They created a medication formulation that works perfectly for my child."
  }
]

interface PharmacyPageProps {
  params: {
    slug: string
  }
}

export default function PharmacyPage({ params }: PharmacyPageProps) {
  const pharmacy = {
    name: params.slug.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' '),
    address: "123 Health Street, New York, NY 10001",
    phone: "(123) 456-7890",
    website: "https://www.example.com",
    email: "info@examplepharmacy.com",
    location: { lat: 40.7128, lng: -74.006 },
    hours: {
      monday: "8:00 AM - 9:00 PM",
      tuesday: "8:00 AM - 9:00 PM",
      wednesday: "8:00 AM - 9:00 PM",
      thursday: "8:00 AM - 9:00 PM",
      friday: "8:00 AM - 9:00 PM",
      saturday: "9:00 AM - 7:00 PM",
      sunday: "10:00 AM - 6:00 PM"
    },
    services: [
      "Prescription Filling",
      "Medication Counseling",
      "Immunizations",
      "Health Screenings",
      "Compounding",
      "Home Delivery"
    ],
    rating: 4.8,
    reviews: 124,
    images: [
      "/placeholder.svg?height=600&width=1200",
      "/placeholder.svg?height=400&width=600",
      "/placeholder.svg?height=400&width=600"
    ]
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Link href="/pharmacies" className="flex items-center gap-2 text-teal-600 dark:text-teal-400 mb-6 hover:underline">
        <ArrowLeft className="h-4 w-4" />
        Back to Pharmacies
      </Link>
      
      <div className="relative h-64 md:h-80 rounded-xl overflow-hidden mb-8">
        <img 
          src={pharmacy.images[0] || "/placeholder.svg"} 
          alt={pharmacy.name} 
          className="w-full h-full object-cover"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/70 to-transparent flex flex-col justify-end p-6">
          <Badge className="mb-2 bg-teal-600 self-start">Open Now</Badge>
          <h1 className="text-3xl font-bold text-white mb-2">{pharmacy.name}</h1>
          <div className="flex items-center gap-2 text-white">
            <div className="flex items-center gap-1 bg-yellow-500/90 px-2 py-0.5 rounded-full text-sm">
              <Star className="h-3.5 w-3.5 fill-white text-white" />
              <span>{pharmacy.rating}</span>
            </div>
            <span className="text-white/80">({pharmacy.reviews} reviews)</span>
          </div>
        </div>
      </div>
      
      <div className="grid lg:grid-cols-3 gap-8 mb-12">
        <div className="lg:col-span-2">
          <Tabs defaultValue="about" className="mb-12">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="about">About</TabsTrigger>
              <TabsTrigger value="services">Services</TabsTrigger>
              <TabsTrigger value="reviews">Reviews</TabsTrigger>
            </TabsList>
            <TabsContent value="about" className="p-6 border rounded-b-lg">
              <div className="prose dark:prose-invert max-w-none">
                <h3 className="text-xl font-semibold mb-4">About {pharmacy.name}</h3>
                <p className="text-slate-600 dark:text-slate-300 mb-6">
                  {pharmacy.name} is a full-service pharmacy dedicated to providing personalized care and comprehensive pharmaceutical services to our community. Our team of experienced pharmacists and healthcare professionals are committed to helping you manage your health with expert advice and quality medications.
                </p>
                
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <h4 className="font-semibold mb-2">Our Mission</h4>
                    <p className="text-slate-600 dark:text-slate-300 text-sm">
                      To improve the health and wellness of our community by providing accessible, high-quality pharmaceutical care and personalized service.
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2">Our Values</h4>
                    <ul className="list-disc list-inside space-y-1 text-slate-600 dark:text-slate-300 text-sm">
                      <li>Patient-centered care</li>
                      <li>Professional excellence</li>
                      <li>Integrity and trust</li>
                      <li>Community engagement</li>
                    </ul>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-4 mt-6">
                  <img 
                    src={pharmacy.images[1] || "/placeholder.svg"} 
                    alt={`${pharmacy.name} interior`} 
                    className="rounded-lg"
                  />
                  <img 
                    src={pharmacy.images[2] || "/placeholder.svg"} 
                    alt={`${pharmacy.name} staff`} 
                    className="rounded-lg"
                  />
                </div>
              </div>
            </TabsContent>
            <TabsContent value="services" className="p-6 border rounded-b-lg">
              <h3 className="text-xl font-semibold mb-4">Our Services</h3>
              <div className="grid md:grid-cols-2 gap-4">
                {pharmacy.services.map((service, index) => (
                  <div key={index} className="flex items-start gap-3 p-4 border rounded-lg">
                    <div className="bg-teal-100 dark:bg-teal-900 p-2 rounded-full text-teal-600 dark:text-teal-400">
                      <Check className="h-5 w-5" />
                    </div>
                    <div>
                      <h4 className="font-medium">{service}</h4>
                      <p className="text-sm text-slate-500 dark:text-slate-400">
                        {getServiceDescription(service)}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </TabsContent>
            <TabsContent value="reviews" className="p-6 border rounded-b-lg">
              <h3 className="text-xl font-semibold mb-4">Customer Reviews</h3>
              <div className="flex items-center gap-2 mb-6">
                <div className="flex items-center gap-1 bg-yellow-500/90 px-2 py-1 rounded-full">
                  <Star className="h-4 w-4 fill-white text-white" />
                  <span className="text-white font-medium">{pharmacy.rating}</span>
                </div>
                <span className="text-slate-600 dark:text-slate-400">Based on {pharmacy.reviews} reviews</span>
              </div>
              
              <div className="space-y-6">
                {dummyReviews.map((review, index) => (
                  <Card key={index}>
                    <CardContent className="p-4">
                      <div className="flex items-start gap-4">
                        <Avatar>
                          <img src={review.avatar} alt={review.name} className="object-cover" />
                        </Avatar>
                        <div className="flex-1">
                          <div className="flex justify-between items-center mb-1">
                            <h4 className="font-medium">{review.name}</h4>
                            <span className="text-slate-500 text-sm">{review.date}</span>
                          </div>
                          <div className="flex items-center mb-2">
                            {[...Array(5)].map((_, i) => (
                              <Star key={i} className={`h-3.5 w-3.5 ${i < review.rating ? "fill-yellow-400 text-yellow-400" : "fill-gray-200 text-gray-200 dark:fill-gray-700 dark:text-gray-700"}`} />
                            ))}
                          </div>
                          <p className="text-slate-600 dark:text-slate-300 text-sm">{review.comment}</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
              
              <Button className="mt-6 w-full">Load More Reviews</Button>
            </TabsContent>
          </Tabs>
        </div>
        
        <div className="bg-white dark:bg-slate-900 border rounded-xl p-6 h-fit">
          <h3 className="text-lg font-semibold mb-4">Contact Information</h3>
          
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <MapPin className="h-5 w-5 text-teal-600 dark:text-teal-400 mt-0.5" />
              <div>
                <h4 className="font-medium text-sm mb-1">Address</h4>
                <p className="text-slate-600 dark:text-slate-300 text-sm">{pharmacy.address}</p>
                <Button variant="link" size="sm" className="p-0 h-auto text-teal-600 dark:text-teal-400">
                  View on map
                </Button>
              </div>
            </div>
            
            <div className="flex items-start gap-3">
              <Phone className="h-5 w-5 text-teal-600 dark:text-teal-400 mt-0.5" />
              <div>
                <h4 className="font-medium text-sm mb-1">Phone</h4>
                <p className="text-slate-600 dark:text-slate-300 text-sm">{pharmacy.phone}</p>
              </div>
            </div>
            
            <div className="flex items-start gap-3">
              <Globe className="h-5 w-5 text-teal-600 dark:text-teal-400 mt-0.5" />
              <div>
                <h4 className="font-medium text-sm mb-1">Website</h4>
                <a 
                  href={pharmacy.website} 
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-teal-600 dark:text-teal-400 text-sm hover:underline"
                >
                  {pharmacy.website.replace('https://', '')}
                </a>
              </div>
            </div>
            
            <div className="flex items-start gap-3">
              <Clock className="h-5 w-5 text-teal-600 dark:text-teal-400 mt-0.5" />
              <div>
                <h4 className="font-medium text-sm mb-1">Hours</h4>
                <ul className="space-y-1">
                  {Object.entries(pharmacy.hours).map(([day, hours]) => (
                    <li key={day} className="text-sm grid grid-cols-[100px_1fr]">
                      <span className="font-medium capitalize">{day}:</span>
                      <span className="text-slate-600 dark:text-slate-300">{hours}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
