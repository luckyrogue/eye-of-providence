import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { pendingCount } from "../shared/api/tauri";
export function StatusPage() {
  const { t } = useTranslation("agent");
  const [pending, setPending] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  async function refresh() {
    try {
      setPending(await pendingCount());
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setPending(null);
    }
  }
  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 5000);
    return () => clearInterval(id);
  }, []);
  const description = error
    ? t("status_error")
    : pending === null
      ? t("status_loading")
      : pending === 0
        ? t("status_all_sent")
        : t("status_buffer", { count: pending });
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("status_title")}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {error && (
          <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">
            {error}
          </div>
        )}
        <p className="text-muted-foreground">{t("status_hint")}</p>
        <Button size="sm" variant="outline" onClick={refresh}>
          {t("status_refresh")}
        </Button>
      </CardContent>
    </Card>
  );
}
