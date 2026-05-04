import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Button } from "@eop/ui";
import {
  devLogin,
  fetchRecent,
  fetchSummary,
  ingestDemoEvent,
  generateReport,
  listReports,
  type EventRow,
  type Report,
} from "./api";
import { Markdown } from "./Markdown";

export default function App() {
  const [userId, setUserId] = useState<string | null>(localStorage.getItem("eop_user_id"));
  const [events, setEvents] = useState<EventRow[]>([]);
  const [summary, setSummary] = useState<Record<string, number>>({});
  const [reports, setReports] = useState<Report[]>([]);
  const [active, setActive] = useState<Report | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  async function refresh() {
    setError(null);
    try {
      const [r, s, rs] = await Promise.all([fetchRecent(20), fetchSummary(7), listReports()]);
      setEvents(r);
      setSummary(s);
      setReports(rs);
      if (rs.length && !active) setActive(rs[0]);
    } catch (e) {
      setError(String(e));
    }
  }

  async function login() {
    setBusy("login");
    try {
      const uid = await devLogin();
      localStorage.setItem("eop_user_id", uid);
      setUserId(uid);
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(null);
    }
  }

  async function sendDemo() {
    setBusy("demo");
    try {
      await ingestDemoEvent();
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(null);
    }
  }

  async function genReport(period: "weekly" | "monthly") {
    setBusy("report");
    setError(null);
    try {
      const r = await generateReport(period);
      setActive(r);
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(null);
    }
  }

  useEffect(() => {
    if (userId) refresh();
  }, [userId]);

  const totalMs = Object.values(summary).reduce((a, b) => a + b, 0);
  const aiMs = summary["ai"] ?? 0;
  const aiRatio = totalMs ? Math.round((aiMs / totalMs) * 100) : 0;

  return (
    <main className="min-h-screen bg-background p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex items-end justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Eye of Providence</h1>
            <p className="text-muted-foreground">Phase 5 — Gemini reports + dashboard polish</p>
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
              <Button onClick={login} disabled={busy !== null}>
                {busy === "login" ? "..." : "Get dev token"}
              </Button>
            </CardContent>
          </Card>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm font-medium text-muted-foreground">AI ratio</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-4xl font-bold">{aiRatio}%</div>
                  <p className="text-xs text-muted-foreground mt-1">last 7 days</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm font-medium text-muted-foreground">Active time</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-4xl font-bold">{Math.round(totalMs / 60000)}<span className="text-base font-normal"> min</span></div>
                  <p className="text-xs text-muted-foreground mt-1">events: {events.length}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm font-medium text-muted-foreground">Reports</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-4xl font-bold">{reports.length}</div>
                  <p className="text-xs text-muted-foreground mt-1">generated</p>
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader className="flex-row items-center justify-between">
                <div>
                  <CardTitle>AI report</CardTitle>
                  <CardDescription>
                    Сгенерировано через Gemini (или mock, если нет API key).
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" onClick={() => genReport("weekly")} disabled={busy !== null}>
                    {busy === "report" ? "..." : "Generate weekly"}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => genReport("monthly")} disabled={busy !== null}>
                    Monthly
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {reports.length > 1 && (
                  <div className="flex gap-2 flex-wrap">
                    {reports.map((r) => (
                      <button
                        key={r.id}
                        onClick={() => setActive(r)}
                        className={`rounded-md px-3 py-1 text-xs ${
                          active?.id === r.id ? "bg-primary text-primary-foreground" : "bg-secondary"
                        }`}
                      >
                        {r.period}
                      </button>
                    ))}
                  </div>
                )}
                {active ? (
                  <div className="rounded-md border bg-card p-4">
                    <Markdown source={active.body_md} />
                    <div className="mt-4 text-xs text-muted-foreground">
                      {active.model} · {new Date(active.generated_at).toLocaleString()}
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">Нажми "Generate weekly", чтобы создать первый отчёт.</p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Recent events</CardTitle>
                <CardDescription>Live из in-memory store</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2 mb-4">
                  <Button onClick={sendDemo} disabled={busy !== null} size="sm">
                    Send demo events
                  </Button>
                  <Button onClick={refresh} disabled={busy !== null} size="sm" variant="outline">
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
