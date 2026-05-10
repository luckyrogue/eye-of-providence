import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { preflightRun, type PreflightCheck, type PreflightStatus } from "../shared/api/tauri";

const STATUS_DOT: Record<PreflightStatus, string> = {
  ok: "bg-success",
  warn: "bg-warning",
  error: "bg-destructive",
};

function badgeLabel(t: (key: string) => string, status: PreflightStatus): string {
  switch (status) {
    case "ok":
      return t("badge_ok");
    case "warn":
      return t("badge_warn");
    case "error":
      return t("badge_error");
  }
}

export function OnboardingPage() {
  const { t } = useTranslation("agent");
  const [checks, setChecks] = useState<PreflightCheck[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    setBusy(true);
    setError(null);
    try {
      setChecks(await preflightRun());
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    void refresh();
  }, []);

  const errors = checks.filter((c) => c.status === "error");
  const warns = checks.filter((c) => c.status === "warn");

  const desc =
    errors.length > 0
      ? t("setup_desc_blockers", { errors: errors.length, warnings: warns.length })
      : warns.length > 0
        ? t("setup_desc_warns", { count: warns.length })
        : t("setup_desc_ok");

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t("setup_title")}</CardTitle>
          <CardDescription>{desc}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {error && (
            <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">
              {error}
            </div>
          )}

          <ul className="space-y-2">
            {checks.map((c) => (
              <li key={c.id} className="rounded-md border p-3">
                <div className="flex items-start gap-3">
                  <span className={`mt-1.5 h-2 w-2 rounded-full ${STATUS_DOT[c.status]}`} />
                  <div className="flex-1 space-y-1">
                    <div className="flex items-baseline justify-between gap-3">
                      <div className="font-medium text-sm">{c.label}</div>
                      <span className="text-xs text-muted-foreground uppercase tracking-wide">
                        {badgeLabel(t, c.status)}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">{c.message}</p>
                    {c.action_url && (
                      <a
                        href={c.action_url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center text-xs text-primary hover:underline"
                      >
                        {c.action_label ?? t("setup_open")} →
                      </a>
                    )}
                  </div>
                </div>
              </li>
            ))}
            {checks.length === 0 && !error && (
              <li className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
                {busy ? t("setup_list_empty_busy") : t("setup_list_empty")}
              </li>
            )}
          </ul>

          <div className="flex gap-2">
            <Button size="sm" onClick={refresh} disabled={busy}>
              {busy ? t("setup_recheck_busy") : t("setup_recheck")}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
