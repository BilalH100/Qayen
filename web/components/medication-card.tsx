import { Check, X } from "lucide-react"
import { Card, CardContent, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"

interface MedicationCardProps {
  name: string
  image: string
  category: string
  price: number
  inStock: boolean
}

export function MedicationCard({ name, image, category, price, inStock }: MedicationCardProps) {
  return (
    <Link href={`/medications/${name.toLowerCase().replace(/\s+/g, "-")}`}>
      <Card className="overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
        <div className="relative h-48 w-full overflow-hidden bg-slate-50 dark:bg-slate-900 flex items-center justify-center p-4">
          <img
            src={image || "/placeholder.svg"}
            alt={name}
            className="max-h-full max-w-full object-contain transition-transform duration-300 hover:scale-110"
          />
          <Badge className="absolute top-3 right-3 bg-blue-600">{category}</Badge>
        </div>
        <CardContent className="p-4">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-white">{name}</h3>
          <div className="mt-2 text-2xl font-bold text-slate-900 dark:text-white">${price.toFixed(2)}</div>
        </CardContent>
        <CardFooter className="p-4 pt-0">
          {inStock ? (
            <div className="flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400">
              <div className="bg-green-100 dark:bg-green-900/30 p-0.5 rounded-full">
                <Check className="h-3.5 w-3.5" />
              </div>
              <span>In Stock</span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400">
              <div className="bg-red-100 dark:bg-red-900/30 p-0.5 rounded-full">
                <X className="h-3.5 w-3.5" />
              </div>
              <span>Out of Stock</span>
            </div>
          )}
        </CardFooter>
      </Card>
    </Link>
  )
}
