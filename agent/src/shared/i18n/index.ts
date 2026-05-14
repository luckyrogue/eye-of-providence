import { createI18n } from "@eop/i18n";
import enAgent from "./locales/en/agent.json";
import ruAgent from "./locales/ru/agent.json";
import kkAgent from "./locales/kk/agent.json";
import esAgent from "./locales/es/agent.json";
const resources = {
  en: { agent: enAgent },
  ru: { agent: ruAgent },
  kk: { agent: kkAgent },
  es: { agent: esAgent },
};
const i18n = createI18n({
  resources,
  ns: ["agent"],
  defaultNS: "agent",
  fallbackLng: "ru",
});
export default i18n;
