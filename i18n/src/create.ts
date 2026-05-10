import i18next, { type i18n } from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import { LOCALE_STORAGE_KEY, SUPPORTED_LOCALES } from "./constants";
import type { CreateI18nOptions } from "./types";

/**
 * Отдельный экземпляр i18next (не singleton) + react-i18next + browser detector.
 * Вызывать один раз при старте приложения до первого render.
 */
export function createI18n(opts: CreateI18nOptions): i18n {
  const inst = i18next.createInstance();
  inst.use(LanguageDetector).use(initReactI18next);
  inst.init({
    resources: opts.resources,
    fallbackLng: opts.fallbackLng ?? "ru",
    supportedLngs: [...SUPPORTED_LOCALES],
    nonExplicitSupportedLngs: true,
    defaultNS: opts.defaultNS ?? opts.ns[0],
    ns: opts.ns,
    interpolation: { escapeValue: false },
    detection: {
      order: ["localStorage", "navigator", "htmlTag"],
      lookupLocalStorage: LOCALE_STORAGE_KEY,
      caches: ["localStorage"],
    },
  });
  return inst;
}
