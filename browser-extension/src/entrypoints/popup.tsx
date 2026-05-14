import { useEffect, useState } from "react";
import ReactDOM from "react-dom/client";
import { I18nextProvider, useTranslation } from "react-i18next";
import "@eop/ui/styles.css";
import { Button, Card, CardContent, CardHeader, CardTitle, SimpleSelect } from "@eop/ui";
import { LOCALE_LABELS, SUPPORTED_LOCALES, type Locale } from "@eop/i18n";
import i18n from "../shared/i18n";
import { backendDisplayHost, clearConfig } from "../shared/api/backend";
import type {
  FlushNowMessage,
  GetStatusMessage,
  GetStatusResponse,
  SetPausedMessage,
} from "../shared/api/messages";
import { PairingWizard } from "../shared/ui/pairing-wizard";
const DEFAULT_BACKEND_HOST = backendDisplayHost();
const LOCALE_OPTIONS = SUPPORTED_LOCALES.map((lng) => ({
  value: lng,
  label: LOCALE_LABELS[lng],
}));
const DOWNTIME_WARNING_MS = 5 * 60 * 1000;
function Popup() {
  const { t, i18n: i18nInstance } = useTranslation("popup");
  const [token, setToken] = useState<string | undefined>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<GetStatusResponse | null>(null);
  const lng = (i18nInstance.resolvedLanguage?.split("-")[0] ??
    i18nInstance.language.split("-")[0]) as Locale;
  const localeValue = SUPPORTED_LOCALES.includes(lng) ? lng : "ru";
  const refreshStatus = () => {
    const msg: GetStatusMessage = { type: "get-status" };
    void chrome.runtime
      .sendMessage(msg)
      .then((r: GetStatusResponse | undefined) => {
        if (r) setStatus(r);
      })
      .catch(() => {});
  };
  useEffect(() => {
    void chrome.storage.local.get(["eop_token"]).then((r) => {
      setToken(r.eop_token as string | undefined);
    });
    refreshStatus();
    const id = window.setInterval(refreshStatus, 3000);
    const onChange = (changes: Record<string, chrome.storage.StorageChange>, area: string) => {
      if (area !== "local" || !("eop_token" in changes)) return;
      const next = changes.eop_token.newValue as string | undefined;
      setToken(next);
    };
    chrome.storage.onChanged.addListener(onChange);
    return () => {
      window.clearInterval(id);
      chrome.storage.onChanged.removeListener(onChange);
    };
  }, []);
  async function logout() {
    await clearConfig();
    setToken(undefined);
  }
  async function flushNow() {
    setBusy(true);
    setError(null);
    try {
      const msg: FlushNowMessage = { type: "flush-now" };
      await chrome.runtime.sendMessage(msg);
      refreshStatus();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }
  async function togglePause() {
    const next = !(status?.paused ?? false);
    const msg: SetPausedMessage = { type: "set-paused", paused: next };
    await chrome.runtime.sendMessage(msg);
    refreshStatus();
  }
  const pending = (status?.buffer ?? 0) + (status?.retry ?? 0);
  const downtime = status && status.lastSuccessTs > 0 ? Date.now() - status.lastSuccessTs : 0;
  const showDowntimeBanner =
    !!token && !status?.paused && downtime > DOWNTIME_WARNING_MS && pending > 0;
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
          {showDowntimeBanner && (
            <div className="rounded-md border border-destructive bg-destructive/10 px-2 py-1.5 text-[11px] text-destructive">
              {t("downtime_warning", {
                minutes: Math.round(downtime / 60000),
              })}
            </div>
          )}
          {token ? (
            <>
              <div className="text-xs text-muted-foreground">
                {t("connected", { host: DEFAULT_BACKEND_HOST })}
              </div>
              <div className="text-[11px] text-muted-foreground">
                {status?.paused
                  ? t("status_paused")
                  : pending === 0
                    ? t("status_idle")
                    : t("status_pending", { count: pending })}
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  className="flex-1"
                  onClick={flushNow}
                  disabled={busy || status?.paused}
                >
                  {t("flush_now")}
                </Button>
                <Button size="sm" variant="outline" onClick={togglePause}>
                  {status?.paused ? t("resume") : t("pause")}
                </Button>
                <Button size="sm" variant="ghost" onClick={logout}>
                  {t("logout")}
                </Button>
              </div>
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
    <Popup />
  </I18nextProvider>,
);
