import {
  createI18n,
  LOCALE_LABELS,
  LOCALE_STORAGE_KEY,
  SUPPORTED_LOCALES,
  type Locale,
} from "@eop/i18n";
import enCommon from "./locales/en/common.json";
import enErrors from "./locales/en/errors.json";
import enAuth from "./locales/en/auth.json";
import enLanding from "./locales/en/landing.json";
import enOnboarding from "./locales/en/onboarding.json";
import enApp from "./locales/en/app.json";
import enInsights from "./locales/en/insights.json";
import enDeveloper from "./locales/en/developer.json";
import enChangelog from "./locales/en/changelog.json";
import enPwa from "./locales/en/pwa.json";
import ruCommon from "./locales/ru/common.json";
import ruErrors from "./locales/ru/errors.json";
import ruAuth from "./locales/ru/auth.json";
import ruLanding from "./locales/ru/landing.json";
import ruOnboarding from "./locales/ru/onboarding.json";
import ruApp from "./locales/ru/app.json";
import ruInsights from "./locales/ru/insights.json";
import ruDeveloper from "./locales/ru/developer.json";
import ruChangelog from "./locales/ru/changelog.json";
import ruPwa from "./locales/ru/pwa.json";
import kkCommon from "./locales/kk/common.json";
import kkErrors from "./locales/kk/errors.json";
import kkAuth from "./locales/kk/auth.json";
import kkLanding from "./locales/kk/landing.json";
import kkOnboarding from "./locales/kk/onboarding.json";
import kkApp from "./locales/kk/app.json";
import kkInsights from "./locales/kk/insights.json";
import kkDeveloper from "./locales/kk/developer.json";
import kkChangelog from "./locales/kk/changelog.json";
import kkPwa from "./locales/kk/pwa.json";
import esCommon from "./locales/es/common.json";
import esErrors from "./locales/es/errors.json";
import esAuth from "./locales/es/auth.json";
import esLanding from "./locales/es/landing.json";
import esOnboarding from "./locales/es/onboarding.json";
import esApp from "./locales/es/app.json";
import esInsights from "./locales/es/insights.json";
import esDeveloper from "./locales/es/developer.json";
import esChangelog from "./locales/es/changelog.json";
import esPwa from "./locales/es/pwa.json";
export { SUPPORTED_LOCALES, LOCALE_LABELS, LOCALE_STORAGE_KEY, type Locale };
const resources = {
  en: {
    common: enCommon,
    errors: enErrors,
    auth: enAuth,
    landing: enLanding,
    onboarding: enOnboarding,
    app: enApp,
    insights: enInsights,
    developer: enDeveloper,
    changelog: enChangelog,
    pwa: enPwa,
  },
  ru: {
    common: ruCommon,
    errors: ruErrors,
    auth: ruAuth,
    landing: ruLanding,
    onboarding: ruOnboarding,
    app: ruApp,
    insights: ruInsights,
    developer: ruDeveloper,
    changelog: ruChangelog,
    pwa: ruPwa,
  },
  kk: {
    common: kkCommon,
    errors: kkErrors,
    auth: kkAuth,
    landing: kkLanding,
    onboarding: kkOnboarding,
    app: kkApp,
    insights: kkInsights,
    developer: kkDeveloper,
    changelog: kkChangelog,
    pwa: kkPwa,
  },
  es: {
    common: esCommon,
    errors: esErrors,
    auth: esAuth,
    landing: esLanding,
    onboarding: esOnboarding,
    app: esApp,
    insights: esInsights,
    developer: esDeveloper,
    changelog: esChangelog,
    pwa: esPwa,
  },
};
const i18n = createI18n({
  resources,
  fallbackLng: "ru",
  defaultNS: "common",
  ns: [
    "common",
    "errors",
    "auth",
    "landing",
    "onboarding",
    "app",
    "insights",
    "developer",
    "changelog",
    "pwa",
  ],
});
export default i18n;
