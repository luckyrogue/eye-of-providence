import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Button } from "@eop/ui";
import { devLogin, fetchRecent, fetchSummary, ingestDemoEvent, type EventRow } from "./api";

export default function App() {
  const [userId, setUserId] = useState<string | null>(localStorage.getItem("eop_user_id"));
  const [events, setEvents] = useState<EventRow[]>([]);
  const [summary, setSummary] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function refresh() {
    setError(null);
    try {
      const [r, s] = await Promise.all([fetchRecent(20), fetchSummary(7)]);
      setEvents(r);
      setSummary(s);
    } catch (e) {
      setError(String(e));
    }
  }

  async function login() {
    setBusy(true);
    try {
      const uid = await devLogin();
      localStorage.setItem("eop_user_id", uid);
      setUserId(uid);
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function sendDemo() {
    setBusy(true);
    try {
      await ingestDemoEvent();
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    if (userId) refresh();
  }, [userId]);

  const totalMs = Object.values(summary).reduce((a, b) => a + b, 0);
  const aiMs = (summary["ai"] ?? 0);
  const aiRatio = totalMs ? Math.round((aiMs / totalMs) * 100) : 0;

  return (
    <main className="min-h-screen bg-background p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex items-end justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Eye of Providence</h1>
            <p className="text-muted-foreground">Phase 1 — vertical slice e2e</p>
          </div>
          <div className="text-sm text-muted-foreground">{userId ? `user: ${userId.slice(0, 8)}…` : "not logged in"}</div>
        </header>

        {error && (
          <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
        )}

        {!userId ? (
          <Card>
            <CardHeader>
              <CardTitle>Get started</CardTitle>
              <CardDescription>Phase 1 — dev login (без OAuth)</CardDescription>
            </CardHeader>
            <CardContent>
              <Button onClick={login} disabled={busy}>
                {busy ? "..." : "Get dev token"}
              </Button>
            </CardContent>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle>AI ratio (last 7 days)</CardTitle>
                <CardDescription>Доля времени с AI vs всё остальное</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="text-4xl font-bold">{aiRatio}%</div>
                <div className="text-sm text-muted-foreground">
                  AI: {(aiMs / 1000).toFixed(0)}s · total: {(totalMs / 1000).toFixed(0)}s
                </div>
                <div className="flex flex-wrap gap-2 pt-2">
                  {Object.entries(summary).map(([cat, ms]) => (
                    <span key={cat} className="rounded-full bg-secondary px-3 py-1 text-xs">
                      {cat}: {(ms as number / 1000).toFixed(0)}s
                    </span>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Recent events ({events.length})</CardTitle>
                <CardDescription>Live из in-memory store бэкенда</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2 mb-4">
                  <Button onClick={sendDemo} disabled={busy} size="sm">
                    Send demo events
                  </Button>
                  <Button onClick={refresh} disabled={busy} size="sm" variant="outline">
                    Refresh
                  </Button>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="text-left text-muted-foreground">
                      <tr>
                        <th className="py-2 pr-4">time</th>
                        <th className="py-2 pr-4">app</th>
                        <th className="py-2 pr-4">category</th>
                        <th className="py-2 pr-4">source</th>
                        <th className="py-2 pr-4">provider</th>
                        <th className="py-2 pr-4 text-right">duration</th>
                      </tr>
                    </thead>
                    <tbody>
                      {events.map((e, i) => (
                        <tr key={i} className="border-t border-border">
                          <td className="py-2 pr-4 font-mono text-xs">{new Date(e.ts).toLocaleTimeString()}</td>
                          <td className="py-2 pr-4">{e.app_bundle}</td>
                          <td className="py-2 pr-4">{e.category}</td>
                          <td className="py-2 pr-4">{e.source}</td>
                          <td className="py-2 pr-4">{e.ai_provider ?? "—"}</td>
                          <td className="py-2 pr-4 text-right">{(e.duration_ms / 1000).toFixed(1)}s</td>
                        </tr>
                      ))}
                      {events.length === 0 && (
                        <tr>
                          <td colSpan={6} className="py-6 text-center text-muted-foreground">
                            нет событий — нажми "Send demo events"
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </main>
  );
}
