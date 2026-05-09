import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card, CardContent, CardDescription, CardHeader, CardTitle,
  DangerZone, Select, useConfirm,
} from "@eop/ui";
import { Globe, Languages, Shield, Trash2, User } from "lucide-react";
import { useProfile, useDeleteMyData } from "../../entities/user";
import { getTz, setTz, UNIQUE_TIMEZONES } from "../../shared/lib/tz";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";
import { LOCALE_LABELS, SUPPORTED_LOCALES, LOCALE_STORAGE_KEY, type Locale } from "../../shared/i18n";

const PRIVACY_NOT_COLLECTED = [
  "Содержимое файлов, промптов и ответов AI.",
  "Сами нажатия клавиш — только их количество.",
  "Заголовки приватных и incognito окон.",
  "Содержимое clipboard — только sha256 + размер.",
];

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

  function changeLocale(next: Locale) {
    setLocaleState(next);
    localStorage.setItem(LOCALE_STORAGE_KEY, next);
    void i18n.changeLanguage(next);
  }

  function changeTz(value: string) {
    setLocalTz(value);
    setTz(value);
  }

  async function destroy() {
    const ok = await confirm({
      title: "Удалить все данные?",
      description:
        "Стерт user-row, все события и отчёты. После этого нужно будет залогиниться заново.",
      typeToConfirm: profile.data?.email ?? "delete",
      destructive: true,
      confirmText: "Удалить навсегда",
    });
    if (!ok) return;
    const r = await runToast(deleteData.mutateAsync(), {
      success: "Данные удалены",
      error: "Не удалось удалить",
    });
    if (r !== null) onWiped();
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            <CardTitle>Профиль</CardTitle>
          </div>
          <CardDescription>Что мы знаем о тебе</CardDescription>
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
              <dt className="text-muted-foreground">Провайдер</dt>
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
              {profile.isError ? "не удалось загрузить" : "загрузка…"}
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
          <CardDescription>Влияет на отображение времени в отчётах и таблице событий.</CardDescription>
        </CardHeader>
        <CardContent>
          <Select value={tz} onChange={(e) => changeTz(e.target.value)} className="w-full max-w-sm px-3 py-2 text-sm">
            {!UNIQUE_TIMEZONES.find((t) => t.value === tz) && <option value={tz}>Текущий: {tz}</option>}
            {UNIQUE_TIMEZONES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </Select>
          <p className="mt-2 text-xs text-muted-foreground">
            Сейчас: <span className="font-mono">{tz}</span>
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-4 w-4 text-muted-foreground" />
            <CardTitle>Приватность</CardTitle>
          </div>
          <CardDescription>Что Eye of Providence НЕ собирает</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
            {PRIVACY_NOT_COLLECTED.map((s) => <li key={s}>{s}</li>)}
          </ul>
        </CardContent>
      </Card>

      <DangerZone
        title="Удалить все мои данные"
        description="Стерт user-row, все события и отчёты. Необратимо. Нужна будет повторная регистрация."
        action={
          <Button variant="destructive" size="sm" onClick={destroy} disabled={deleteData.isPending}>
            <Trash2 className="h-3.5 w-3.5 mr-1" /> Удалить
          </Button>
        }
      />
    </div>
  );
}
