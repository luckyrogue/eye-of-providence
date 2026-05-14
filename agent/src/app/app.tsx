import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { listen } from "@tauri-apps/api/event";
import { SimpleSelect, Tabs, TabsList, TabsTrigger } from "@eop/ui";
import { LOCALE_LABELS, SUPPORTED_LOCALES, type Locale } from "@eop/i18n";
import { preflightRun } from "../shared/api/tauri";
import { StatusPage } from "../pages/status";
import { OnboardingPage } from "../pages/onboarding";
import { SettingsPage } from "../pages/settings";
type Tab = "status" | "setup" | "settings";
const LOCALE_OPTIONS = SUPPORTED_LOCALES.map((lng) => ({
  value: lng,
  label: LOCALE_LABELS[lng],
}));
export function App() {
  const { t, i18n } = useTranslation("agent");
  const [tab, setTab] = useState<Tab>("status");
  const [hasIssues, setHasIssues] = useState(false);
  const lng = (i18n.resolvedLanguage?.split("-")[0] ?? i18n.language.split("-")[0]) as Locale;
  const localeValue = SUPPORTED_LOCALES.includes(lng) ? lng : "ru";
  useEffect(() => {
    preflightRun()
      .then((checks) => {
        const issues = checks.some((c) => c.status === "error" || c.status === "warn");
        setHasIssues(issues);
        if (issues) setTab("setup");
      })
      .catch(() => {});
  }, []);
  useEffect(() => {
    const unlisten = listen("auth-required", () => setTab("settings"));
    return () => {
      void unlisten.then((fn) => fn());
    };
  }, []);
  return (
    <main className="min-h-screen bg-background p-6">
      <div className="mx-auto max-w-2xl space-y-4">
        <header className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("app_title")}</h1>
            <p className="text-sm text-muted-foreground">{t("app_subtitle")}</p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <SimpleSelect
              value={localeValue}
              onValueChange={(v) => void i18n.changeLanguage(v)}
              options={LOCALE_OPTIONS}
              triggerClassName="w-[9.5rem]"
            />
            <Tabs value={tab} onValueChange={(v) => setTab(v as Tab)}>
              <TabsList>
                <TabsTrigger value="status">{t("tab_status")}</TabsTrigger>
                <TabsTrigger value="setup">
                  {t("tab_setup")}
                  {hasIssues ? t("tab_setup_marker") : ""}
                </TabsTrigger>
                <TabsTrigger value="settings">{t("tab_settings")}</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
        </header>

        {tab === "setup" ? (
          <OnboardingPage />
        ) : tab === "settings" ? (
          <SettingsPage />
        ) : (
          <StatusPage />
        )}
      </div>
    </main>
  );
}
