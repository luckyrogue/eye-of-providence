// i18next setup для dashboard.
//
// Поддерживаемые локали: ru / en / kk / es.
// Detection: localStorage('eop_locale') > navigator.language > 'ru'.
//
// Для type-safety: типы ключей подсасываются из ru/common.json (как
// "канонический" namespace). Остальные локали должны иметь те же ключи.
//
// Lazy-load: пока тащим всё через import — namespaces маленькие. Если файлы
// раздуются (>50 KB), переключимся на i18next-http-backend.

import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";

import enCommon from "./locales/en/common.json";
import enErrors from "./locales/en/errors.json";
import enAuth from "./locales/en/auth.json";
import enLanding from "./locales/en/landing.json";
import enOnboarding from "./locales/en/onboarding.json";
import enApp from "./locales/en/app.json";
import enInsights from "./locales/en/insights.json";

import ruCommon from "./locales/ru/common.json";
import ruErrors from "./locales/ru/errors.json";
import ruAuth from "./locales/ru/auth.json";
import ruLanding from "./locales/ru/landing.json";
import ruOnboarding from "./locales/ru/onboarding.json";
import ruApp from "./locales/ru/app.json";
import ruInsights from "./locales/ru/insights.json";

import kkCommon from "./locales/kk/common.json";
import kkErrors from "./locales/kk/errors.json";
import kkAuth from "./locales/kk/auth.json";
import kkLanding from "./locales/kk/landing.json";
import kkOnboarding from "./locales/kk/onboarding.json";
import kkApp from "./locales/kk/app.json";
import kkInsights from "./locales/kk/insights.json";

import esCommon from "./locales/es/common.json";
import esErrors from "./locales/es/errors.json";
import esAuth from "./locales/es/auth.json";
import esLanding from "./locales/es/landing.json";
import esOnboarding from "./locales/es/onboarding.json";
import esApp from "./locales/es/app.json";
import esInsights from "./locales/es/insights.json";

export const SUPPORTED_LOCALES = ["ru", "en", "kk", "es"] as const;
export type Locale = (typeof SUPPORTED_LOCALES)[number];

export const LOCALE_LABELS: Record<Locale, string> = {
  ru: "Русский",
  en: "English",
  kk: "Қазақша",
  es: "Español",
};

export const LOCALE_STORAGE_KEY = "eop_locale";

const resources = {
  en: { common: enCommon, errors: enErrors, auth: enAuth, landing: enLanding, onboarding: enOnboarding, app: enApp, insights: enInsights },
  ru: { common: ruCommon, errors: ruErrors, auth: ruAuth, landing: ruLanding, onboarding: ruOnboarding, app: ruApp, insights: ruInsights },
  kk: { common: kkCommon, errors: kkErrors, auth: kkAuth, landing: kkLanding, onboarding: kkOnboarding, app: kkApp, insights: kkInsights },
  es: { common: esCommon, errors: esErrors, auth: esAuth, landing: esLanding, onboarding: esOnboarding, app: esApp, insights: esInsights },
};

// Init возвращает Promise, но мы не await'им — React-Suspense покажет fallback
// пока ресурсы не загружены. Учитывая, что resources уже импортированы статически,
// init фактически синхронный.
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: "ru",
    supportedLngs: SUPPORTED_LOCALES as unknown as string[],
    nonExplicitSupportedLngs: true, // 'ru-RU' → 'ru'
    defaultNS: "common",
    ns: ["common", "errors", "auth", "landing", "onboarding", "app", "insights"],
    interpolation: { escapeValue: false }, // React сам escape'ит
    detection: {
      order: ["localStorage", "navigator", "htmlTag"],
      lookupLocalStorage: LOCALE_STORAGE_KEY,
      caches: ["localStorage"],
    },
  });

export default i18n;
