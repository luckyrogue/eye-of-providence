import { useEffect, useState } from "react";
import ReactDOM from "react-dom/client";
import { I18nextProvider, Trans, useTranslation } from "react-i18next";
import "@eop/ui/styles.css";
import { Button, Card, CardContent, CardHeader, CardTitle, SimpleSelect } from "@eop/ui";
import { LOCALE_LABELS, SUPPORTED_LOCALES, type Locale } from "@eop/i18n";
import i18n from "../shared/i18n";
import { fetchDevToken, setConfig, clearConfig } from "../shared/api/backend";

const DEFAULT_BACKEND_HOST = "eop.rysdavletov.org";

const LOCALE_OPTIONS = SUPPORTED_LOCALES.map((lng) => ({
  value: lng,
  label: LOCALE_LABELS[lng],
}));

function Popup() {
  const { t, i18n: i18nInstance } = useTranslation("popup");
  const [token, setToken] = useState<string | undefined>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const lng = (i18nInstance.resolvedLanguage?.split("-")[0] ??
    i18nInstance.language.split("-")[0]) as Locale;
  const localeValue = SUPPORTED_LOCALES.includes(lng) ? lng : "ru";

  useEffect(() => {
    chrome.storage.local.get(["eop_token"]).then((r) => {
      setToken(r.eop_token as string | undefined);
    });
  }, []);

  async function login() {
    setBusy(true);
    setError(null);
    try {
      const tok = await fetchDevToken();
      await setConfig(tok);
      setToken(tok);
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function logout() {
    await clearConfig();
    setToken(undefined);
  }

  async function flushNow() {
    setBusy(true);
    try {
      await chrome.runtime.sendMessage({ type: "flush-now" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="w-80 space-y-3 p-4">
      <div className="flex justify-end">
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
          {error && <div className="text-xs text-destructive">{error}</div>}
          {token ? (
            <>
              <div className="text-xs text-muted-foreground">
                {t("connected", { host: DEFAULT_BACKEND_HOST })}
              </div>
              <div className="flex gap-2">
                <Button size="sm" className="flex-1" onClick={flushNow} disabled={busy}>
                  {t("flush_now")}
                </Button>
                <Button size="sm" variant="outline" onClick={logout}>
                  {t("logout")}
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="text-xs text-muted-foreground">
                <Trans
                  ns="popup"
                  i18nKey="dev_login_lead"
                  values={{ host: DEFAULT_BACKEND_HOST }}
                  components={{ code: <code className="rounded bg-muted px-1" /> }}
                />
              </div>
              <Button size="sm" className="w-full" onClick={login} disabled={busy}>
                {busy ? t("get_token_busy") : t("get_token")}
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <I18nextProvider i18n={i18n}>
    <Popup />
  </I18nextProvider>,
);
