import { Check, X } from "lucide-react";
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
      <Card className="overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
        <div className="relative h-48 w-full overflow-hidden bg-slate-50 dark:bg-slate-900 flex items-center justify-center p-4">
          <div className="text-center">
            <div className="text-6xl text-slate-300 dark:text-slate-600 mb-2">
              💊
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400"></p>
          </div>
          {category && (
            <Badge className="absolute top-3 right-3 bg-teal-600 text-white">
              {category}
            </Badge>
          )}
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
