"use client";

import { useAutoTranslation } from "@/contexts/auto-translation-context";

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
    <div data-translate-ignore className="flex shrink-0 items-center">
      <label className="sr-only" htmlFor="qayen-language-selector">
        {accessibleLabel}
      </label>
      <select
        id="qayen-language-selector"
        aria-label={accessibleLabel}
        title={translationServiceUnavailable ? serviceNote : accessibleLabel}
        value={language}
        onChange={(event) => setLanguage(event.target.value as "en" | "fr" | "ar")}
        className="h-9 w-[64px] rounded-md border border-border bg-background px-2 text-sm font-semibold text-foreground outline-none focus:ring-2 focus:ring-ring"
      >
        <option value="en">ENG</option>
        <option value="fr">FR</option>
        <option value="ar">AR</option>
      </select>
      {translationServiceUnavailable && language !== "en" ? (
        <span className="sr-only" role="status">
          {serviceNote}
        </span>
      ) : null}
    </div>
  );
}
