import { useTranslation } from "react-i18next";
import { Eyebrow, StatTile } from "@eop/ui";
import { Activity, Brain, FileText } from "lucide-react";

export function StatsRow({
  aiRatio,
  totalMinutes,
  eventsCount,
  reportsCount,
}: {
  aiRatio: number;
  totalMinutes: number;
  eventsCount: number;
  reportsCount: number;
}) {
  const { t } = useTranslation("app");
  return (
    <div className="reveal">
      <div className="flex items-baseline justify-between mb-3">
        <Eyebrow>{t("dashboard.overview")}</Eyebrow>
        <span className="font-mono text-[11px] text-muted-foreground">
          {t("dashboard.last_7_days")}
        </span>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatTile
          label={t("dashboard.stat_ai_share")}
          value={`${aiRatio}%`}
          hint={t("dashboard.stat_ai_hint")}
          icon={<Brain className="h-4 w-4" style={{ color: "hsl(var(--accent))" }} />}
          accent="purple"
          className="reveal reveal-delay-1"
        />
        <StatTile
          label={t("dashboard.stat_active")}
          value={totalMinutes}
          unit={t("dashboard.stat_minutes")}
          hint={t("dashboard.stat_active_hint_events", { count: eventsCount })}
          icon={<Activity className="h-4 w-4 text-foreground/70" />}
          accent="blue"
          className="reveal reveal-delay-2"
        />
        <StatTile
          label={t("dashboard.stat_reports")}
          value={reportsCount}
          hint={t("dashboard.stat_reports_hint")}
          icon={<FileText className="h-4 w-4 text-warning" />}
          accent="amber"
          className="reveal reveal-delay-3"
        />
      </div>
    </div>
  );
}
