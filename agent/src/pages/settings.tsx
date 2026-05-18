import { useCallback, useEffect, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { getCurrentWindow } from "@tauri-apps/api/window";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import {
  accountInfo,
  logout,
  openPermissionsSettings,
  permissionsStatus,
  type AccountInfo,
  type PermissionsStatus,
} from "../shared/api/tauri";
import { PairingWizard } from "../shared/ui/pairing-wizard";

export function SettingsPage() {
  const { t } = useTranslation("agent");
  const [account, setAccount] = useState<AccountInfo | null>(null);
  const [perms, setPerms] = useState<PermissionsStatus | null>(null);

  const refresh = useCallback(async () => {
    try {
      setAccount(await accountInfo());
    } catch (e) {
      console.warn("[eop] account_info failed", e);
    }
    try {
      setPerms(await permissionsStatus());
    } catch (e) {
      console.warn("[eop] permissions_status failed", e);
    }
  }, []);

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 3000);
    return () => clearInterval(id);
  }, [refresh]);

  useEffect(() => {
    const win = getCurrentWindow();
    const unlisten = win.onFocusChanged(({ payload: focused }) => {
      if (focused) void refresh();
    });
    return () => {
      void unlisten.then((fn) => fn());
    };
  }, [refresh]);

  const accessibility = perms?.accessibility ?? null;

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
            <Button size="sm" variant="outline" onClick={() => void logout().then(refresh)}>
              {t("settings_logout")}
            </Button>
          ) : (
            <PairingWizard backend={account?.backend_url ?? ""} onClaimed={() => void refresh()} />
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("settings_permissions_title")}</CardTitle>
          <CardDescription>{t("settings_permissions_hint")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between rounded-md border p-3 text-sm">
            <span>{t("settings_permissions_accessibility")}</span>
            <span className={accessibility ? "text-success" : "text-muted-foreground"}>
              {accessibility === null
                ? t("conn_loading")
                : accessibility
                  ? t("settings_permissions_granted")
                  : t("settings_permissions_missing")}
            </span>
          </div>
          <Button size="sm" onClick={() => void openPermissionsSettings()}>
            {t("settings_permissions_open")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
