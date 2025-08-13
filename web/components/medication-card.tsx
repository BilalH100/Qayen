import { Check, X, Pill } from "lucide-react";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";

interface MedicationCardProps {
  id: number;
  name: string;
  category?: string;
  form?: string;
  presentation?: string;
  status?: string;
  code?: string;
}

export function MedicationCard({
  id,
  name,
  category,
  form,
  presentation,
  status,
  code,
}: MedicationCardProps) {
  const slug = code || id.toString();

  const isAvailable =
    status?.toLowerCase() === "commercialisé" ||
    status?.toLowerCase() === "active";

  return (
    <Link href={`/medications/${slug}`}>
      <Card className="group overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
        <div className="relative h-48 w-full overflow-hidden bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-blue-900/20 dark:to-indigo-800/20 flex items-center justify-center transition-all duration-300 group-hover:from-blue-100 group-hover:to-indigo-200 dark:group-hover:from-blue-900/30 dark:group-hover:to-indigo-800/30">
          <div className="relative">
            <div className="absolute inset-0 bg-white/20 dark:bg-white/10 rounded-full blur-xl"></div>
            <Pill className="relative h-14 w-14 text-blue-600 dark:text-blue-400 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-12" />
          </div>
          {/* {category && (
            <Badge className="absolute top-3 right-3 bg-teal-600 text-white">
              {category}
            </Badge>
          )} */}
        </div>
        <CardContent className="p-4">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-white line-clamp-2">
            {name}
          </h3>
          {presentation && (
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
              {form && presentation
                ? `${form} - ${presentation}`
                : form || "Medication"}
            </p>
          )}
        </CardContent>
        <CardFooter className="p-4 pt-0">
          {isAvailable ? (
            <div className="flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400">
              <div className="bg-green-100 dark:bg-green-900/30 p-0.5 rounded-full">
                <Check className="h-3.5 w-3.5" />
              </div>
              <span>Available</span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400">
              <div className="bg-red-100 dark:bg-red-900/30 p-0.5 rounded-full">
                <X className="h-3.5 w-3.5" />
              </div>
              <span>Not Available</span>
            </div>
          )}
        </CardFooter>
      </Card>
    </Link>
  );
}
