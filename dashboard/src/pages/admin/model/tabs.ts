export const ADMIN_TABS = [
  { id: "overview", i18nKey: "admin.overview" },
  { id: "teams", i18nKey: "admin.teams" },
  { id: "users", i18nKey: "admin.users" },
  { id: "revenue", i18nKey: "admin.revenue" },
  { id: "sso", i18nKey: "admin.sso" },
  { id: "audit", i18nKey: "admin.audit" },
] as const;

export type AdminTabKey = (typeof ADMIN_TABS)[number]["id"];
