import { Star, MapPin, Clock } from "lucide-react"
import { Card, CardContent, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"

interface PharmacyCardProps {
  name: string
  image: string
  address: string
  rating: number
  openUntil: string
  distance: string
}

export function PharmacyCard({ name, image, address, rating, openUntil, distance }: PharmacyCardProps) {
  return (
    <Link href={`/pharmacies/${name.toLowerCase().replace(/\s+/g, "-")}`}>
      <Card className="overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
        <div className="relative h-48 w-full overflow-hidden">
          <img
            src={image || "/placeholder.svg"}
            alt={name}
            className="h-full w-full object-cover transition-transform duration-300 hover:scale-105"
          />
          <Badge className="absolute top-3 right-3 bg-teal-600">Open</Badge>
        </div>
        <CardContent className="p-4">
          <div className="flex justify-between items-start">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white">{name}</h3>
            <div className="flex items-center gap-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400 px-2 py-0.5 rounded-full text-sm">
              <Star className="h-3.5 w-3.5 fill-yellow-500 text-yellow-500" />
              <span>{rating}</span>
            </div>
          </div>
          <div className="flex items-center gap-1 text-slate-500 dark:text-slate-400 mt-2 text-sm">
            <MapPin className="h-4 w-4" />
            <span>{address}</span>
          </div>
        </CardContent>
        <CardFooter className="p-4 pt-0 flex justify-between text-sm text-slate-500 dark:text-slate-400">
          <div className="flex items-center gap-1">
            <Clock className="h-4 w-4" />
            <span>Open until {openUntil}</span>
          </div>
          <div>{distance}</div>
        </CardFooter>
      </Card>
    </Link>
  )
}
