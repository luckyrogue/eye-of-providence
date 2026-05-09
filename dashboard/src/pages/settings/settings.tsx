import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card, CardContent, CardDescription, CardHeader, CardTitle,
  DangerZone, Select, useConfirm,
} from "@eop/ui";
import { Globe, Languages, Shield, Trash2, User } from "lucide-react";
import { useProfile, useDeleteMyData, updateLocale } from "../../entities/user";
import { getTz, setTz, UNIQUE_TIMEZONES } from "../../shared/lib/tz";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";
import { LOCALE_LABELS, SUPPORTED_LOCALES, LOCALE_STORAGE_KEY, type Locale } from "../../shared/i18n";

export function Settings({ onWiped }: { onWiped: () => void }) {
  const { t, i18n } = useTranslation("common");
  const profile = useProfile();
  const deleteData = useDeleteMyData();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const [tz, setLocalTz] = useState(getTz());
  const [locale, setLocaleState] = useState<Locale>(
    (i18n.resolvedLanguage as Locale) || "ru",
  );

  const privacyItems = [
    t("settings.privacy_files"),
    t("settings.privacy_keystrokes"),
    t("settings.privacy_private_windows"),
    t("settings.privacy_clipboard"),
  ];

  function changeLocale(next: Locale) {
    setLocaleState(next);
    localStorage.setItem(LOCALE_STORAGE_KEY, next);
    void i18n.changeLanguage(next);
    // Best-effort persist на backend, чтобы emails (invite/reset/digest)
    // приходили на нужном языке. Пропускаем ошибку — locale уже сохранён локально.
    void updateLocale(next).catch(() => undefined);
  }

  function changeTz(value: string) {
    setLocalTz(value);
    setTz(value);
  }

  async function destroy() {
    const ok = await confirm({
      title: t("settings.danger_confirm_title"),
      description: t("settings.danger_confirm_lead"),
      typeToConfirm: profile.data?.email ?? "delete",
      destructive: true,
      confirmText: t("settings.danger_confirm_btn"),
    });
    if (!ok) return;
    const r = await runToast(deleteData.mutateAsync(), {
      success: t("settings.delete_success"),
      error: t("settings.delete_failed"),
    });
    if (r !== null) onWiped();
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.profile_title")}</CardTitle>
          </div>
          <CardDescription>{t("settings.profile_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          {profile.data ? (
            <dl className="grid grid-cols-[8rem_1fr] gap-y-2 text-sm">
              <dt className="text-muted-foreground">User ID</dt>
              <dd className="font-mono text-xs break-all">{profile.data.user_id}</dd>
              {profile.data.email && (
                <>
                  <dt className="text-muted-foreground">Email</dt>
                  <dd>{profile.data.email}</dd>
                </>
              )}
              <dt className="text-muted-foreground">{t("settings.profile_provider")}</dt>
              <dd>{profile.data.provider ?? "—"}</dd>
              {profile.data.github_login && (
                <>
                  <dt className="text-muted-foreground">GitHub</dt>
                  <dd>@{profile.data.github_login}</dd>
                </>
              )}
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">
              {profile.isError ? t("settings.profile_load_failed") : t("settings.profile_loading")}
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Languages className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.language")}</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Select
            value={locale}
            onChange={(e) => changeLocale(e.target.value as Locale)}
            className="w-full max-w-sm px-3 py-2 text-sm"
          >
            {SUPPORTED_LOCALES.map((code) => (
              <option key={code} value={code}>
                {LOCALE_LABELS[code]}
              </option>
            ))}
          </Select>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.timezone")}</CardTitle>
          </div>
          <CardDescription>{t("settings.timezone_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          <Select value={tz} onChange={(e) => changeTz(e.target.value)} className="w-full max-w-sm px-3 py-2 text-sm">
            {!UNIQUE_TIMEZONES.find((t) => t.value === tz) && (
              <option value={tz}>{t("settings.timezone_current", { tz })}</option>
            )}
            {UNIQUE_TIMEZONES.map((tzo) => (
              <option key={tzo.value} value={tzo.value}>{tzo.label}</option>
            ))}
          </Select>
          <p className="mt-2 text-xs text-muted-foreground">
            {t("settings.timezone_now")} <span className="font-mono">{tz}</span>
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.privacy_title")}</CardTitle>
          </div>
          <CardDescription>{t("settings.privacy_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
            {privacyItems.map((s) => <li key={s}>{s}</li>)}
          </ul>
        </CardContent>
      </Card>

      <DangerZone
        title={t("settings.danger_title")}
        description={t("settings.danger_lead")}
        action={
          <Button variant="destructive" size="sm" onClick={destroy} disabled={deleteData.isPending}>
            <Trash2 className="h-3.5 w-3.5 mr-1" /> {t("actions.delete")}
          </Button>
        }
      />
    </div>
  );
}
