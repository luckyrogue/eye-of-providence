import { createI18n } from "@eop/i18n";
import enPopup from "./locales/en/popup.json";
import ruPopup from "./locales/ru/popup.json";
import kkPopup from "./locales/kk/popup.json";
import esPopup from "./locales/es/popup.json";
const resources = {
  en: { popup: enPopup },
  ru: { popup: ruPopup },
  kk: { popup: kkPopup },
  es: { popup: esPopup },
};
const i18n = createI18n({
  resources,
  ns: ["popup"],
  defaultNS: "popup",
  fallbackLng: "ru",
});
export default i18n;
