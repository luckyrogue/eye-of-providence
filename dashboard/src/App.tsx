import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Button } from "@eop/ui";
import { Activity, Brain, Eye, FileText, LogOut, Settings as SettingsIcon, Sparkles, Users } from "lucide-react";
import {
  fetchRecent,
  fetchSummary,
  fetchLanguages,
  fetchHeatmap,
  fetchTrend,
  ingestDemoEvent,
  generateReport,
  listReports,
  type EventRow,
  type LangCell,
  type HeatmapCell,
  type TrendPoint,
  type Report,
} from "./api";
import { Markdown } from "./Markdown";
import { Heatmap } from "./Heatmap";
import { Languages } from "./Languages";
import { Settings } from "./Settings";
import { Trend } from "./Trend";
import { Auth } from "./Auth";
import { Teams } from "./Teams";
import { formatDate, formatTime, getTz } from "./tz";

type Tab = "dashboard" | "team" | "settings";

const CATEGORY_LABELS: Record<string, string> = {
  manual: "вручную",
  ai: "AI",
  refactor: "рефакторинг",
  idle: "простой",
  reading: "чтение",
  other: "прочее",
};

const SOURCE_LABELS: Record<string, string> = {
  os: "ОС",
  browser: "браузер",
  ide: "IDE",
  cli: "CLI",
};

export default function App() {
  const [userId, setUserId] = useState<string | null>(localStorage.getItem("eop_user_id"));
  const [events, setEvents] = useState<EventRow[]>([]);
  const [summary, setSummary] = useState<Record<string, number>>({});
  const [languages, setLanguages] = useState<LangCell[]>([]);
  const [heatmap, setHeatmap] = useState<HeatmapCell[]>([]);
  const [trend, setTrend] = useState<TrendPoint[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [active, setActive] = useState<Report | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>("dashboard");
  const [tz, setTz] = useState(getTz());

  function logout() {
    localStorage.removeItem("eop_token");
    localStorage.removeItem("eop_user_id");
    setUserId(null);
    setEvents([]);
    setSummary({});
    setLanguages([]);
    setHeatmap([]);
    setTrend([]);
    setReports([]);
    setActive(null);
    setTab("dashboard");
  }

  async function refresh() {
    setError(null);
    try {
      const [r, s, langs, hm, tr, rs] = await Promise.all([
        fetchRecent(20),
        fetchSummary(7),
        fetchLanguages(30),
        fetchHeatmap(30, tz),
        fetchTrend(30, tz),
        listReports(),
      ]);
      setEvents(r);
      setSummary(s);
      setLanguages(langs);
      setHeatmap(hm);
      setTrend(tr);
      setReports(rs);
      if (rs.length && !active) setActive(rs[0]);
    } catch (e) {
      setError(String(e));
    }
  }

  function onAuth(r: { user_id: string; display_name?: string }) {
    localStorage.setItem("eop_user_id", r.user_id);
    if (r.display_name) localStorage.setItem("eop_display_name", r.display_name);
    setUserId(r.user_id);
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
  }, [userId, tz]);

  const totalMs = Object.values(summary).reduce((a, b) => a + b, 0);
  const aiMs = summary["ai"] ?? 0;
  const aiRatio = totalMs ? Math.round((aiMs / totalMs) * 100) : 0;

  return (
    <main className="min-h-screen bg-gradient-to-br from-background via-background to-primary/5">
      <div className="border-b bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="mx-auto max-w-6xl px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              <Eye className="h-5 w-5 text-primary-foreground" />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight">Eye of Providence</h1>
              <p className="text-xs text-muted-foreground">Сколько ты пишешь сам, а сколько — с AI</p>
            </div>
          </div>
          {userId && (
            <nav className="flex items-center gap-1 text-sm">
              <button
                onClick={() => setTab("dashboard")}
                className={`flex items-center gap-1.5 rounded-md px-3 py-1.5 transition-colors ${tab === "dashboard" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-secondary"}`}
              >
                <Activity className="h-4 w-4" /> Дашборд
              </button>
              <button
                onClick={() => setTab("team")}
                className={`flex items-center gap-1.5 rounded-md px-3 py-1.5 transition-colors ${tab === "team" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-secondary"}`}
              >
                <Users className="h-4 w-4" /> Команда
              </button>
              <button
                onClick={() => setTab("settings")}
                className={`flex items-center gap-1.5 rounded-md px-3 py-1.5 transition-colors ${tab === "settings" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-secondary"}`}
              >
                <SettingsIcon className="h-4 w-4" /> Настройки
              </button>
              <span className="mx-2 h-5 w-px bg-border" />
              <button
                onClick={logout}
                className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted-foreground hover:bg-secondary"
              >
                <LogOut className="h-4 w-4" /> Выход
              </button>
            </nav>
          )}
        </div>
      </div>

      <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
        {error && (
          <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
        )}

        {!userId ? (
          <Auth onAuth={onAuth} />
        ) : tab === "settings" ? (
          <Settings onWiped={logout} onTzChange={setTz} />
        ) : tab === "team" ? (
          <Teams tz={tz} />
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <StatCard label="Доля AI" value={`${aiRatio}%`} hint="за 7 дней" icon={<Brain className="h-4 w-4 text-purple-500" />} accent="purple" />
              <StatCard label="Активное время" value={Math.round(totalMs / 60000)} unit="мин" hint={`${events.length} событий за период`} icon={<Activity className="h-4 w-4 text-blue-500" />} accent="blue" />
              <StatCard label="Отчёты" value={reports.length} hint="сгенерировано" icon={<FileText className="h-4 w-4 text-amber-500" />} accent="amber" />
            </div>

            <Card>
              <CardHeader>
                <CardTitle>Динамика за 30 дней</CardTitle>
                <CardDescription>Время вручную vs с AI по дням</CardDescription>
              </CardHeader>
              <CardContent>
                <Trend points={trend} />
              </CardContent>
            </Card>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle>Тепловая карта</CardTitle>
                  <CardDescription>За 30 дней, день недели × час · {tz}</CardDescription>
                </CardHeader>
                <CardContent>
                  <Heatmap cells={heatmap} />
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>По языкам</CardTitle>
                  <CardDescription>Доля AI и ручного кода по языкам</CardDescription>
                </CardHeader>
                <CardContent>
                  <Languages cells={languages} top={6} />
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader className="flex-row items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <Sparkles className="h-4 w-4 text-purple-500" />
                    AI-отчёт
                  </CardTitle>
                  <CardDescription>
                    Сгенерирован через Gemini (или mock-режим, если ключ не задан).
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" onClick={() => genReport("weekly")} disabled={busy !== null}>
                    {busy === "report" ? "..." : "Создать недельный"}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => genReport("monthly")} disabled={busy !== null}>
                    Месячный
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
                        className={`rounded-md px-3 py-1 text-xs transition-colors ${active?.id === r.id ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-secondary/80"}`}
                      >
                        {r.period}
                      </button>
                    ))}
                  </div>
                )}
                {active ? (
                  <div className="rounded-lg border bg-card/50 p-5">
                    <Markdown source={active.body_md} />
                    <div className="mt-4 pt-4 border-t text-xs text-muted-foreground">
                      <span className="font-mono">{active.model}</span> · {formatDate(active.generated_at, tz)}
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">Нажми «Создать недельный», чтобы получить первый отчёт.</p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Последние события</CardTitle>
                <CardDescription>Из event store. Время в часовом поясе из настроек.</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2 mb-4">
                  <Button onClick={sendDemo} disabled={busy !== null} size="sm">
                    Отправить демо-события
                  </Button>
                  <Button onClick={refresh} disabled={busy !== null} size="sm" variant="outline">
                    Обновить
                  </Button>
                </div>
                <div className="overflow-x-auto rounded-md border">
                  <table className="w-full text-sm">
                    <thead className="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
                      <tr>
                        <th className="py-2.5 px-3">Время</th>
                        <th className="py-2.5 px-3">Приложение</th>
                        <th className="py-2.5 px-3">Категория</th>
                        <th className="py-2.5 px-3">Источник</th>
                        <th className="py-2.5 px-3">Провайдер</th>
                        <th className="py-2.5 px-3 text-right">Длительность</th>
                      </tr>
                    </thead>
                    <tbody>
                      {events.map((e, i) => (
                        <tr key={i} className="border-t hover:bg-muted/30 transition-colors">
                          <td className="py-2 px-3 font-mono text-xs">{formatTime(e.ts, tz)}</td>
                          <td className="py-2 px-3">{e.app_bundle}</td>
                          <td className="py-2 px-3"><CategoryBadge cat={e.category} /></td>
                          <td className="py-2 px-3 text-muted-foreground">{SOURCE_LABELS[e.source] ?? e.source}</td>
                          <td className="py-2 px-3 text-muted-foreground">{e.ai_provider ?? "—"}</td>
                          <td className="py-2 px-3 text-right tabular-nums">{(e.duration_ms / 1000).toFixed(1)} с</td>
                        </tr>
                      ))}
                      {events.length === 0 && (
                        <tr>
                          <td colSpan={6} className="py-8 text-center text-muted-foreground">
                            Нет событий — нажми «Отправить демо-события»
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

function StatCard({ label, value, unit, hint, icon, accent }: {
  label: string; value: string | number; unit?: string; hint: string;
  icon: React.ReactNode; accent: "purple" | "blue" | "amber";
}) {
  const accents = {
    purple: "from-purple-500/20",
    blue: "from-blue-500/20",
    amber: "from-amber-500/20",
  };
  return (
    <Card className="overflow-hidden relative">
      <div className={`absolute right-0 top-0 h-24 w-24 rounded-bl-full bg-gradient-to-bl ${accents[accent]} to-transparent`} />
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xs uppercase tracking-wide text-muted-foreground">{label}</CardTitle>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-4xl font-bold tabular-nums">
          {value}{unit && <span className="text-base font-normal"> {unit}</span>}
        </div>
        <p className="text-xs text-muted-foreground mt-1">{hint}</p>
      </CardContent>
    </Card>
  );
}

const CATEGORY_COLORS: Record<string, string> = {
  manual: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30",
  ai: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/30",
  refactor: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30",
  idle: "bg-neutral-500/10 text-neutral-600 dark:text-neutral-400 border-neutral-500/30",
  other: "bg-neutral-500/10 text-neutral-600 dark:text-neutral-400 border-neutral-500/30",
};

function CategoryBadge({ cat }: { cat: string }) {
  return (
    <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium ${CATEGORY_COLORS[cat] ?? CATEGORY_COLORS.other}`}>
      {CATEGORY_LABELS[cat] ?? cat}
    </span>
  );
}
