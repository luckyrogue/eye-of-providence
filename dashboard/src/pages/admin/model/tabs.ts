export const ADMIN_TABS = [
  { id: "overview", i18nKey: "admin.overview" },
  { id: "teams", i18nKey: "admin.teams" },
  { id: "users", i18nKey: "admin.users" },
] as const;

export type AdminTabKey = (typeof ADMIN_TABS)[number]["id"];
