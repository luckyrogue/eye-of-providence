import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { pendingCount } from "../shared/api/tauri";

export function StatusPage() {
  const { t } = useTranslation("agent");
  const [pending, setPending] = useState<number | null>(null);

  async function refresh() {
    try {
      setPending(await pendingCount());
    } catch {
      setPending(null);
    }
  }

  useEffect(() => {
    refresh();
    const id = setInterval(refresh, 5000);
    return () => clearInterval(id);
  }, []);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("status_title")}</CardTitle>
        <CardDescription>
          {pending === null
            ? t("status_loading")
            : pending === 0
              ? t("status_all_sent")
              : t("status_buffer", { count: pending })}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        <p className="text-muted-foreground">{t("status_hint")}</p>
        <Button size="sm" variant="outline" onClick={refresh}>
          {t("status_refresh")}
        </Button>
      </CardContent>
    </Card>
  );
}
