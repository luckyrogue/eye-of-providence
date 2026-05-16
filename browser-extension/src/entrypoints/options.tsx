import { useEffect, useState } from "react";
import ReactDOM from "react-dom/client";
import { I18nextProvider, useTranslation } from "react-i18next";
import "@eop/ui/styles.css";
import { Button, Card, CardContent, CardHeader, CardTitle, SimpleSelect } from "@eop/ui";
import { LOCALE_LABELS, SUPPORTED_LOCALES, type Locale } from "@eop/i18n";
import i18n from "../shared/i18n";
import { backendDisplayHost, clearConfig } from "../shared/api/backend";
import { PairingWizard } from "../shared/ui/pairing-wizard";

const DEFAULT_BACKEND_HOST = backendDisplayHost();

const LOCALE_OPTIONS = SUPPORTED_LOCALES.map((lng) => ({
  value: lng,
  label: LOCALE_LABELS[lng],
}));

function Options() {
  const { t, i18n: i18nInstance } = useTranslation("popup");
  const [token, setToken] = useState<string | undefined>();

  const lng = (i18nInstance.resolvedLanguage?.split("-")[0] ??
    i18nInstance.language.split("-")[0]) as Locale;
  const localeValue = SUPPORTED_LOCALES.includes(lng) ? lng : "ru";

  useEffect(() => {
    void chrome.storage.local.get(["eop_token"]).then((r) => {
      setToken(r.eop_token as string | undefined);
    });
    const onChange = (changes: Record<string, chrome.storage.StorageChange>, area: string) => {
      if (area !== "local" || !("eop_token" in changes)) return;
      setToken(changes.eop_token.newValue as string | undefined);
    };
    chrome.storage.onChanged.addListener(onChange);
    return () => chrome.storage.onChanged.removeListener(onChange);
  }, []);

  async function logout() {
    await clearConfig();
    setToken(undefined);
  }

  return (
    <div className="mx-auto max-w-md p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t("title")}</h1>
        <SimpleSelect
          value={localeValue}
          onValueChange={(v) => void i18nInstance.changeLanguage(v)}
          options={LOCALE_OPTIONS}
          triggerClassName="w-[9.5rem]"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("title")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {token ? (
            <>
              <p className="text-sm">{t("connected", { host: DEFAULT_BACKEND_HOST })}</p>
              <Button size="sm" variant="outline" onClick={logout}>
                {t("logout")}
              </Button>
            </>
          ) : (
            <PairingWizard />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <I18nextProvider i18n={i18n}>
    <Options />
  </I18nextProvider>,
);
