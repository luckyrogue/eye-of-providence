import { useCallback, useEffect, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import {
  accountInfo,
  getAutostart,
  isPaused,
  logout,
  openLogsFolder,
  setAutostart,
  setBackendUrl,
  setPaused,
  type AccountInfo,
} from "../shared/api/tauri";
import { PairingWizard } from "../shared/ui/pairing-wizard";

export function SettingsPage() {
  const { t } = useTranslation("agent");
  const [account, setAccount] = useState<AccountInfo | null>(null);
  const [backend, setBackend] = useState("");
  const [backendSaved, setBackendSaved] = useState(false);
  const [paused, setPausedState] = useState(false);
  const [autostart, setAutostartState] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const info = await accountInfo();
      setAccount(info);
      setBackend(info.backend_url ?? "");
    } catch (e) {
      console.warn("[eop] account_info failed", e);
    }
    try {
      setPausedState(await isPaused());
    } catch (e) {
      console.warn("[eop] is_paused failed", e);
    }
    try {
      setAutostartState(await getAutostart());
    } catch (e) {
      console.warn("[eop] get_autostart failed", e);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const onSaveBackend = useCallback(async () => {
    try {
      await setBackendUrl(backend);
      setBackendSaved(true);
      window.setTimeout(() => setBackendSaved(false), 1500);
      await refresh();
    } catch (e) {
      console.warn("[eop] set_backend_url failed", e);
    }
  }, [backend, refresh]);

  const onLogout = useCallback(async () => {
    try {
      await logout();
      await refresh();
    } catch (e) {
      console.warn("[eop] logout failed", e);
    }
  }, [refresh]);

  const onTogglePause = useCallback(async () => {
    try {
      await setPaused(!paused);
      setPausedState(!paused);
    } catch (e) {
      console.warn("[eop] set_paused failed", e);
    }
  }, [paused]);

  const onToggleAutostart = useCallback(async () => {
    try {
      await setAutostart(!autostart);
      setAutostartState(!autostart);
    } catch (e) {
      console.warn("[eop] set_autostart failed", e);
    }
  }, [autostart]);

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t("settings_account_title")}</CardTitle>
          <CardDescription>
            {account?.paired && account.user_id ? (
              <Trans
                ns="agent"
                i18nKey="settings_paired_as"
                values={{ user_id: account.user_id }}
                components={{ code: <code className="rounded bg-muted px-1 font-mono" /> }}
              />
            ) : (
              t("settings_not_paired")
            )}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {account?.paired ? (
            <Button size="sm" variant="outline" onClick={() => void onLogout()}>
              {t("settings_logout")}
            </Button>
          ) : (
            <PairingWizard
              backend={account?.backend_url ?? backend}
              onClaimed={() => void refresh()}
            />
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("settings_backend_title")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <label className="text-xs text-muted-foreground" htmlFor="backend-url">
            {t("settings_backend_label")}
          </label>
          <div className="flex gap-2">
            <Input
              id="backend-url"
              value={backend}
              onChange={(e) => setBackend(e.target.value)}
              placeholder="https://api.eop.example"
              className="flex-1"
            />
            <Button size="sm" onClick={() => void onSaveBackend()}>
              {backendSaved ? t("settings_backend_saved") : t("settings_backend_save")}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("settings_tracking_title")}</CardTitle>
          <CardDescription>
            {paused ? t("settings_tracking_paused") : t("settings_tracking_active")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            size="sm"
            variant={paused ? "default" : "outline"}
            onClick={() => void onTogglePause()}
          >
            {paused ? t("settings_tracking_resume") : t("settings_tracking_pause")}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("settings_autostart_title")}</CardTitle>
          <CardDescription>
            {autostart ? t("settings_autostart_on") : t("settings_autostart_off")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            size="sm"
            variant={autostart ? "outline" : "default"}
            onClick={() => void onToggleAutostart()}
          >
            {autostart ? t("settings_autostart_disable") : t("settings_autostart_enable")}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("settings_logs_title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Button size="sm" variant="outline" onClick={() => void openLogsFolder()}>
            {t("settings_logs_open")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
