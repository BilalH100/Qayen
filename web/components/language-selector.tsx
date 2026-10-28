"use client";

import { useAutoTranslation } from "@/contexts/auto-translation-context";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const languages = [
  { value: "en", label: "ENG" },
  { value: "fr", label: "FR" },
  { value: "ar", label: "AR" },
] as const;

function LanguageFlag({ language }: { language: "en" | "fr" | "ar" }) {
  const className = "h-4 w-6 shrink-0";
  if (language === "fr") {
    return (
      <svg viewBox="0 0 3 2" className={className} aria-hidden="true" focusable="false">
        <path fill="#002654" d="M0 0h1v2H0z" />
        <path fill="#fff" d="M1 0h1v2H1z" />
        <path fill="#ed2939" d="M2 0h1v2H2z" />
      </svg>
    );
  }
  if (language === "ar") {
    return (
      <svg viewBox="0 0 60 40" className={className} aria-hidden="true" focusable="false">
        <path fill="#c1272d" d="M0 0h60v40H0z" />
        <path d="M30 10 35.88 28.09 20.49 16.91 39.51 16.91 24.12 28.09Z" fill="none" stroke="#006233" strokeWidth="1.6" strokeLinejoin="round" />
      </svg>
    );
  }
  return (
    <svg viewBox="0 0 190 100" className={className} aria-hidden="true" focusable="false">
      <path fill="#b22234" d="M0 0h190v100H0z" />
      {Array.from({ length: 6 }, (_, stripe) => (
        <rect key={stripe} x="0" y={(stripe * 2 + 1) * 100 / 13} width="190" height={100 / 13} fill="#fff" />
      ))}
      <path fill="#3c3b6e" d="M0 0h76v53.846H0z" />
      {Array.from({ length: 9 }, (_, row) =>
        Array.from({ length: row % 2 === 0 ? 6 : 5 }, (_, column) => (
          <path key={`${row}-${column}`} fill="#fff" d="M0-3.1 .73-1 2.95-.96 1.18.38 1.82 2.51 0 1.24-1.82 2.51-1.18.38-2.95-.96-.73-1Z" transform={`translate(${(column + (row % 2 === 0 ? 0.5 : 1)) * 76 / 6} ${(row + 1) * 53.846 / 10})`} />
        )),
      )}
    </svg>
  );
}

export function LanguageSelector() {
  const { language, setLanguage, translationServiceUnavailable } = useAutoTranslation();

  const accessibleLabel = language === "fr" ? "Choisir la langue" : language === "ar" ? "اختر اللغة" : "Choose language";
  const serviceNote =
    language === "fr"
      ? "Certains contenus ne peuvent pas être traduits pour le moment. Réessayez plus tard."
      : language === "ar"
        ? "تعذرت ترجمة بعض المحتويات حاليًا. يرجى المحاولة لاحقًا."
        : "Some content cannot be translated right now. Please try again later.";

  return (
    <div data-translate-ignore translate="no" className="flex shrink-0 items-center">
      <label className="sr-only" htmlFor="qayen-language-selector">
        {accessibleLabel}
      </label>
      <Select
        dir="ltr"
        value={language}
        onValueChange={(value) => setLanguage(value as "en" | "fr" | "ar")}
      >
        <SelectTrigger
          id="qayen-language-selector"
          aria-label={accessibleLabel}
          title={translationServiceUnavailable ? serviceNote : accessibleLabel}
          className="h-9 w-[112px] border-border px-2 text-sm font-semibold"
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent data-translate-ignore translate="no" dir="ltr">
          {languages.map((item) => (
            <SelectItem key={item.value} value={item.value} textValue={item.label}>
              <span className="flex items-center gap-2">
                <LanguageFlag language={item.value} />
                <span>{item.label}</span>
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {translationServiceUnavailable && language !== "en" ? (
        <span className="sr-only" role="status">
          {serviceNote}
        </span>
      ) : null}
    </div>
  );
}
