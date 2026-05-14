export const SUPPORTED_LOCALES = ["ru", "en", "kk", "es"] as const;
export type Locale = (typeof SUPPORTED_LOCALES)[number];
export const LOCALE_LABELS: Record<Locale, string> = {
  ru: "Русский",
  en: "English",
  kk: "Қазақша",
  es: "Español",
};
export const LOCALE_STORAGE_KEY = "eop_locale";
