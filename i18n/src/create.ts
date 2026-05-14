import i18next, { type i18n } from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import { LOCALE_STORAGE_KEY, SUPPORTED_LOCALES } from "./constants";
import type { CreateI18nOptions } from "./types";
export function createI18n(opts: CreateI18nOptions): i18n {
  const inst = i18next.createInstance();
  inst.use(LanguageDetector).use(initReactI18next);
  void inst.init({
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
  if (typeof document !== "undefined") {
    const apply = (lng: string) => {
      const base = lng.split("-")[0];
      document.documentElement.lang = base;
    };
    apply(inst.language || (opts.fallbackLng ?? "ru"));
    inst.on("languageChanged", apply);
  }
  return inst;
}
