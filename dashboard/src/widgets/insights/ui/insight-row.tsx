import { useTranslation } from "react-i18next";
import type { Insight } from "@/entities/user";
import { FALLBACK_INSIGHT_ICON, INSIGHT_ICON_MAP } from "../model/icon-map";

export function InsightRow({ insight }: { insight: Insight }) {
  const { t } = useTranslation("insights");
  const cfg = INSIGHT_ICON_MAP[insight.key] ?? FALLBACK_INSIGHT_ICON;
  const Icon = cfg.icon;

  // productive_day: backend шлёт dow (0..6), резолвим название дня тут.
  const vars: Record<string, string | number | boolean> = { ...(insight.vars ?? {}) };
  if (insight.key === "productive_day" && typeof vars.dow === "number") {
    vars.day = t(`dow_${vars.dow}`);
  }

  return (
    <li className="flex items-start gap-3 rounded-md border bg-muted/20 p-3 transition-colors hover:bg-muted/40">
      <Icon className={`h-4 w-4 mt-0.5 shrink-0 ${cfg.color}`} />
      <span className="text-sm">{t(insight.key, vars)}</span>
    </li>
  );
}
