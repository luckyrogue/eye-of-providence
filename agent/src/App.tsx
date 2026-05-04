import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Onboarding } from "./Onboarding";

type Tab = "status" | "setup";

export default function App() {
  const [tab, setTab] = useState<Tab>("status");
  const [pending, setPending] = useState<number | null>(null);
  const [hasIssues, setHasIssues] = useState(false);

  async function refreshStats() {
    try {
      const n = await invoke<number>("pending_count");
      setPending(n);
    } catch {
      setPending(null);
    }
  }

  useEffect(() => {
    refreshStats();
    const t = setInterval(refreshStats, 5000);
    return () => clearInterval(t);
  }, []);

  // Если в preflight есть errors — открыть Setup tab автоматически на первом запуске.
  useEffect(() => {
    invoke<{ status: string }[]>("preflight_run")
      .then((checks) => {
        const issues = checks.some((c) => c.status === "error" || c.status === "warn");
        setHasIssues(issues);
        if (issues) setTab("setup");
      })
      .catch(() => {});
  }, []);

  return (
    <main className="min-h-screen bg-background p-6">
      <div className="mx-auto max-w-2xl space-y-4">
        <header className="flex items-end justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Eye of Providence</h1>
            <p className="text-sm text-muted-foreground">desktop agent</p>
          </div>
          <nav className="flex gap-2 text-sm">
            <button
              onClick={() => setTab("status")}
              className={`rounded-md px-3 py-1 ${tab === "status" ? "bg-secondary" : "text-muted-foreground"}`}
            >
              Status
            </button>
            <button
              onClick={() => setTab("setup")}
              className={`rounded-md px-3 py-1 ${tab === "setup" ? "bg-secondary" : "text-muted-foreground"}`}
            >
              Setup{hasIssues ? " ●" : ""}
            </button>
          </nav>
        </header>

        {tab === "setup" ? (
          <Onboarding />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Tracking</CardTitle>
              <CardDescription>
                {pending === null
                  ? "loading…"
                  : pending === 0
                  ? "Все события отправлены"
                  : `В буфере: ${pending} событий ожидают отправки`}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
              <p className="text-muted-foreground">
                Агент работает в фоне и в системном tray. События пишутся локально и синхронизируются с backend каждые 15 секунд.
              </p>
              <Button size="sm" variant="outline" onClick={refreshStats}>
                Refresh
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
