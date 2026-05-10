// Insights widget — narrative-карточки на dashboard.
//
// Backend (`GET /v1/me/insights`) возвращает [{key, vars}], frontend резолвит
// локализованную строку через `t(\`insights:${key}\`, vars)`. День недели для
// productive_day mapпится через `dow_0..dow_6` keys в той же ns.
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Eyebrow, Skeleton } from "@eop/ui";
import { Sparkles, TrendingDown, TrendingUp, Minus, Zap, Languages, Calendar, Clock } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useInsights, type Insight } from "../../entities/user";

const ICON_MAP: Record<string, { icon: LucideIcon; color: string }> = {
  ai_trend_up: { icon: TrendingUp, color: "text-purple-500" },
  ai_trend_down: { icon: TrendingDown, color: "text-blue-500" },
  ai_trend_flat: { icon: Minus, color: "text-muted-foreground" },
  ai_trend_started: { icon: Sparkles, color: "text-purple-500" },
  ai_ratio: { icon: Zap, color: "text-amber-500" },
  top_lang: { icon: Languages, color: "text-emerald-500" },
  productive_day: { icon: Calendar, color: "text-blue-500" },
  total_activity: { icon: Clock, color: "text-muted-foreground" },
};

export function InsightsWidget() {
  const { t } = useTranslation("insights");
  const { data: insights, isPending } = useInsights();

  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <div className="flex items-baseline justify-between">
          <Eyebrow>{t("title")}</Eyebrow>
        </div>
        <CardTitle className="font-display tracking-tight mt-2 flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-purple-500" />
          {t("lead")}
        </CardTitle>
        <CardDescription className="sr-only">{t("lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        {isPending ? (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : !insights || insights.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("empty")}</p>
        ) : (
          <ul className="space-y-2">
            {insights.map((ins, i) => (
              <InsightRow key={`${ins.key}-${i}`} insight={ins} />
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function InsightRow({ insight }: { insight: Insight }) {
  const { t } = useTranslation("insights");
  const cfg = ICON_MAP[insight.key];
  const Icon = cfg?.icon ?? Sparkles;
  const color = cfg?.color ?? "text-muted-foreground";

  // productive_day: backend шлёт dow (0..6), резолвим название дня тут.
  const vars = { ...(insight.vars ?? {}) };
  if (insight.key === "productive_day" && typeof vars.dow === "number") {
    vars.day = t(`dow_${vars.dow}`);
  }

  return (
    <li className="flex items-start gap-3 rounded-md border bg-muted/20 p-3 transition-colors hover:bg-muted/40">
      <Icon className={`h-4 w-4 mt-0.5 shrink-0 ${color}`} />
      <span className="text-sm">{t(insight.key, vars)}</span>
    </li>
  );
}
