import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";

type Status = "ok" | "warn" | "error";
type Check = {
  id: string;
  label: string;
  status: Status;
  message: string;
  action_url?: string;
  action_label?: string;
};

const STATUS_DOT: Record<Status, string> = {
  ok: "bg-green-500",
  warn: "bg-amber-500",
  error: "bg-red-500",
};

const STATUS_LABEL: Record<Status, string> = {
  ok: "OK",
  warn: "Warning",
  error: "Error",
};

export function Onboarding() {
  const [checks, setChecks] = useState<Check[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    setBusy(true);
    setError(null);
    try {
      const results = await invoke<Check[]>("preflight_run");
      setChecks(results);
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  const errors = checks.filter((c) => c.status === "error");
  const warns = checks.filter((c) => c.status === "warn");

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Setup checks</CardTitle>
          <CardDescription>
            {errors.length > 0
              ? `${errors.length} blocker(s), ${warns.length} warning(s) — исправь перед использованием`
              : warns.length > 0
              ? `Всё работает, но есть ${warns.length} предупреждений`
              : "Всё готово"}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">{error}</div>}

          <ul className="space-y-2">
            {checks.map((c) => (
              <li key={c.id} className="rounded-md border p-3">
                <div className="flex items-start gap-3">
                  <span className={`mt-1.5 h-2 w-2 rounded-full ${STATUS_DOT[c.status]}`} />
                  <div className="flex-1 space-y-1">
                    <div className="flex items-baseline justify-between gap-3">
                      <div className="font-medium text-sm">{c.label}</div>
                      <span className="text-xs text-muted-foreground uppercase tracking-wide">{STATUS_LABEL[c.status]}</span>
                    </div>
                    <p className="text-sm text-muted-foreground">{c.message}</p>
                    {c.action_url && (
                      <a
                        href={c.action_url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center text-xs text-primary hover:underline"
                      >
                        {c.action_label ?? "Open"} →
                      </a>
                    )}
                  </div>
                </div>
              </li>
            ))}
            {checks.length === 0 && !error && (
              <li className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
                {busy ? "Checking…" : "Нет результатов"}
              </li>
            )}
          </ul>

          <div className="flex gap-2">
            <Button size="sm" onClick={refresh} disabled={busy}>
              {busy ? "Checking…" : "Re-check"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
