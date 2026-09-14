import type { ReactNode } from "react"
import Link from "next/link"

interface CategoryPillProps {
  label: string
  count: number
}

export function CategoryPill({label, count }: CategoryPillProps) {
  return (
    <Link href={`/categories/${label.toLowerCase().replace(/\s+/g, "-")}`}>
      <div className="flex items-center gap-2 px-4 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 rounded-full transition-colors">
        <div className="bg-teal-100 dark:bg-teal-900 p-1 rounded-full text-teal-600 dark:text-teal-400"></div>
        <span className="font-medium text-slate-900 dark:text-white">{label}</span>
        <span className="text-xs bg-slate-200 dark:bg-slate-700 px-2 py-0.5 rounded-full text-slate-600 dark:text-slate-300">
          {count}
        </span>
      </div>
    </Link>
  )
}
