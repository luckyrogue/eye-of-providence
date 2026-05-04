import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { fetchProfile, deleteMyData, type Profile } from "./api";

export function Settings({ onWiped }: { onWiped: () => void }) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchProfile().then(setProfile).catch((e) => setError(String(e)));
  }, []);

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
          <CardTitle>Profile</CardTitle>
          <CardDescription>Что мы знаем о тебе</CardDescription>
        </CardHeader>
        <CardContent>
          {profile ? (
            <dl className="grid grid-cols-[8rem_1fr] gap-y-2 text-sm">
              <dt className="text-muted-foreground">user_id</dt>
              <dd className="font-mono text-xs">{profile.user_id}</dd>
              {profile.email && (
                <>
                  <dt className="text-muted-foreground">email</dt>
                  <dd>{profile.email}</dd>
                </>
              )}
              <dt className="text-muted-foreground">provider</dt>
              <dd>{profile.provider ?? "—"}</dd>
              {profile.github_login && (
                <>
                  <dt className="text-muted-foreground">github</dt>
                  <dd>@{profile.github_login}</dd>
                </>
              )}
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">{error ?? "loading…"}</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Privacy</CardTitle>
          <CardDescription>Что Eye of Providence НЕ собирает</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <ul className="list-disc pl-5 space-y-1 text-muted-foreground">
            <li>Содержимое файлов, промптов и ответов AI.</li>
            <li>Сами keystrokes — только counts.</li>
            <li>Заголовки приватных и incognito окон.</li>
            <li>Содержимое clipboard — только sha256 + размер.</li>
          </ul>
          <p className="text-xs text-muted-foreground pt-2">
            Подробности — в <code className="bg-secondary rounded px-1">docs/privacy.md</code> репозитория.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-destructive">Danger zone</CardTitle>
          <CardDescription>Удаление всех данных без возможности восстановления</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm">
            Удалит <strong>все</strong> события из ClickHouse, отчёты и user-row из Postgres.
            После этого нужно будет залогиниться заново.
          </p>
          {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">{error}</div>}
          {confirm ? (
            <div className="flex gap-2">
              <Button variant="destructive" onClick={doDelete} disabled={busy}>
                {busy ? "..." : "Yes, delete everything"}
              </Button>
              <Button variant="outline" onClick={() => setConfirm(false)} disabled={busy}>
                Cancel
              </Button>
            </div>
          ) : (
            <Button variant="destructive" onClick={() => setConfirm(true)}>
              Delete all my data
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
