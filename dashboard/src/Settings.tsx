import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Globe, Shield, Trash2, User } from "lucide-react";
import { fetchProfile, deleteMyData, type Profile } from "./api";
import { getTz, setTz, UNIQUE_TIMEZONES } from "./tz";

export function Settings({ onWiped, onTzChange }: { onWiped: () => void; onTzChange: (tz: string) => void }) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tz, setLocalTz] = useState(getTz());

  useEffect(() => {
    fetchProfile().then(setProfile).catch((e) => setError(String(e)));
  }, []);

  function changeTz(value: string) {
    setLocalTz(value);
    setTz(value);
    onTzChange(value);
  }

  async function doDelete() {
    setBusy(true);
    setError(null);
    try {
      await deleteMyData();
      onWiped();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
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
          {profile ? (
            <dl className="grid grid-cols-[8rem_1fr] gap-y-2 text-sm">
              <dt className="text-muted-foreground">User ID</dt>
              <dd className="font-mono text-xs break-all">{profile.user_id}</dd>
              {profile.email && (
                <>
                  <dt className="text-muted-foreground">Email</dt>
                  <dd>{profile.email}</dd>
                </>
              )}
              <dt className="text-muted-foreground">Провайдер</dt>
              <dd>{profile.provider ?? "—"}</dd>
              {profile.github_login && (
                <>
                  <dt className="text-muted-foreground">GitHub</dt>
                  <dd>@{profile.github_login}</dd>
                </>
              )}
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">{error ?? "загрузка…"}</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <CardTitle>Часовой пояс</CardTitle>
          </div>
          <CardDescription>Влияет на отображение времени в отчётах и таблице событий.</CardDescription>
        </CardHeader>
        <CardContent>
          <select
            value={tz}
            onChange={(e) => changeTz(e.target.value)}
            className="w-full max-w-sm rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          >
            {!UNIQUE_TIMEZONES.find((t) => t.value === tz) && (
              <option value={tz}>Текущий: {tz}</option>
            )}
            {UNIQUE_TIMEZONES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
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
        <CardContent className="space-y-2 text-sm">
          <ul className="list-disc pl-5 space-y-1 text-muted-foreground">
            <li>Содержимое файлов, промптов и ответов AI.</li>
            <li>Сами нажатия клавиш — только их количество.</li>
            <li>Заголовки приватных и incognito окон.</li>
            <li>Содержимое clipboard — только sha256 + размер.</li>
          </ul>
        </CardContent>
      </Card>

      <Card className="border-destructive/40">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Trash2 className="h-4 w-4 text-destructive" />
            <CardTitle className="text-destructive">Опасная зона</CardTitle>
          </div>
          <CardDescription>Удаление всех данных без возможности восстановления</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm">
            Удалит <strong>все</strong> события, отчёты и user-row. После этого нужно будет залогиниться заново.
          </p>
          {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">{error}</div>}
          {confirm ? (
            <div className="flex gap-2">
              <Button variant="destructive" onClick={doDelete} disabled={busy}>
                {busy ? "..." : "Да, удалить всё"}
              </Button>
              <Button variant="outline" onClick={() => setConfirm(false)} disabled={busy}>
                Отмена
              </Button>
            </div>
          ) : (
            <Button variant="destructive" onClick={() => setConfirm(true)}>
              Удалить все мои данные
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
