import type { InitOptions } from "i18next";
export type CreateI18nOptions = {
  resources: NonNullable<InitOptions["resources"]>;
  ns: string[];
  defaultNS?: string;
  fallbackLng?: string;
};
