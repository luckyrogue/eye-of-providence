import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Tabs,
  TabsList,
  TabsTrigger,
} from "@eop/ui";
import { Sparkles } from "lucide-react";
import { useGenerateReport, type Report } from "../../../entities/report";
import { Markdown } from "../../../shared/lib/markdown";
import { formatDate } from "../../../shared/lib/tz";

// Если активный таб не выбран — показываем первый (свежий) отчёт из списка.
export function ReportsCard({ reports, tz }: { reports: Report[]; tz: string }) {
  const { t } = useTranslation("app");
  const genReport = useGenerateReport();
  const [activeReport, setActiveReport] = useState<Report | null>(null);
  const current = activeReport ?? reports[0] ?? null;

  return (
    <Card className="card-hover">
      <CardHeader className="flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="min-w-0 flex-1">
          <CardTitle className="flex items-center gap-2 font-display tracking-tight">
            <Sparkles className="h-4 w-4 shrink-0" style={{ color: "hsl(var(--accent))" }} />
            {t("dashboard.report_title")}
          </CardTitle>
          <CardDescription>{t("dashboard.report_lead")}</CardDescription>
        </div>
        <div className="flex gap-2 w-full sm:w-auto">
          <Button
            size="sm"
            onClick={() => genReport.mutate("weekly")}
            disabled={genReport.isPending}
            className="flex-1 sm:flex-initial"
          >
            {genReport.isPending ? "..." : t("dashboard.report_create_weekly")}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => genReport.mutate("monthly")}
            disabled={genReport.isPending}
            className="flex-1 sm:flex-initial"
          >
            {t("dashboard.report_create_monthly")}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {reports.length > 1 && (
          <Tabs
            value={current?.id ?? ""}
            onValueChange={(v) => {
              const r = reports.find((rr) => rr.id === v);
              if (r) setActiveReport(r);
            }}
          >
            <TabsList className="flex-wrap gap-2">
              {reports.map((r) => (
                <TabsTrigger key={r.id} value={r.id} className="text-xs">
                  {r.period}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        )}
        {current ? (
          <div className="rounded-lg border bg-card/50 p-5">
            <Markdown source={current.body_md} />
            <div className="mt-4 pt-4 border-t text-xs text-muted-foreground">
              <span className="font-mono">{current.model}</span> ·{" "}
              {formatDate(current.generated_at, tz)}
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{t("dashboard.report_empty")}</p>
        )}
      </CardContent>
    </Card>
  );
}
