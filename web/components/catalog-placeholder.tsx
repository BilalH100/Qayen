import Link from "next/link";
import { Info, Mail, Package, Stethoscope, Store } from "lucide-react";
import { Button } from "@/components/ui/button";

const sectionIcons = {
  store: Store,
  package: Package,
  medical: Stethoscope,
  info: Info,
  mail: Mail,
};

interface CatalogPlaceholderProps {
  title: string;
  description: string;
  icon: keyof typeof sectionIcons;
  message?: string;
}

export function CatalogPlaceholder({ title, description, icon, message = "We're preparing this section. Check back soon to discover the full catalog." }: CatalogPlaceholderProps) {
  const Icon = sectionIcons[icon];

  return (
    <section className="container mx-auto max-w-5xl px-4 py-12 md:py-16">
      <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white md:text-4xl">
        {title}
      </h1>
      <p className="mt-3 text-lg text-slate-600 dark:text-slate-300">{description}</p>

      <div className="mt-10 flex flex-col items-center rounded-2xl border border-dashed border-teal-200 bg-teal-50/50 px-6 py-16 text-center dark:border-teal-800 dark:bg-teal-950/20">
        <div className="mb-6 rounded-full bg-teal-100 p-5 text-teal-600 dark:bg-teal-900/50 dark:text-teal-400">
          <Icon className="h-10 w-10" aria-hidden="true" />
        </div>
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">Coming soon</h2>
        <p className="mt-3 max-w-md text-slate-600 dark:text-slate-300">
          {message}
        </p>
        <Button asChild variant="outline" className="mt-8">
          <Link href="/">Go Home</Link>
        </Button>
      </div>
    </section>
  );
}
