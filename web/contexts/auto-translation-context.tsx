"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  LANGUAGE_STORAGE_KEY,
  startAutomaticTranslation,
  type AppLanguage,
} from "@/lib/auto-translate-dom";

type AutoTranslationContextValue = {
  language: AppLanguage;
  setLanguage: (language: AppLanguage) => void;
  translationServiceUnavailable: boolean;
};

const AutoTranslationContext = createContext<AutoTranslationContextValue | null>(null);
const supportedLanguages: AppLanguage[] = ["en", "fr", "ar"];

export function AutoTranslationProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<AppLanguage>("en");
  const [initialized, setInitialized] = useState(false);
  const [translationServiceUnavailable, setTranslationServiceUnavailable] = useState(false);

  useEffect(() => {
    try {
      const savedLanguage = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
      if (supportedLanguages.includes(savedLanguage as AppLanguage)) {
        setLanguageState(savedLanguage as AppLanguage);
      }
    } catch {
      // Default to English if browser storage is unavailable.
    }
    setInitialized(true);
  }, []);

  const setLanguage = useCallback((nextLanguage: AppLanguage) => {
    setLanguageState(nextLanguage);
    try {
      window.localStorage.setItem(LANGUAGE_STORAGE_KEY, nextLanguage);
    } catch {
      // The selection remains active for the current session if storage is unavailable.
    }
  }, []);

  useEffect(() => {
    if (!initialized) return;

    setTranslationServiceUnavailable(false);
    const stopTranslation = startAutomaticTranslation(language, setTranslationServiceUnavailable);
    return stopTranslation;
  }, [initialized, language]);

  const value = useMemo(
    () => ({ language, setLanguage, translationServiceUnavailable }),
    [language, setLanguage, translationServiceUnavailable],
  );

  return <AutoTranslationContext.Provider value={value}>{children}</AutoTranslationContext.Provider>;
}

export function useAutoTranslation(): AutoTranslationContextValue {
  const context = useContext(AutoTranslationContext);
  if (!context) {
    throw new Error("useAutoTranslation must be used inside AutoTranslationProvider.");
  }
  return context;
}
