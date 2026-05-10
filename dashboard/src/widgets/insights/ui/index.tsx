// Insights widget — narrative-карточки на dashboard.
//
// Backend (`GET /v1/me/insights`) возвращает [{key, vars}], frontend резолвит
// локализованную строку через `t(\`insights:${key}\`, vars)`.

import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Eyebrow,
  Skeleton,
} from "@eop/ui";
import { Sparkles } from "lucide-react";
import { useInsights } from "../../../entities/user";
import { InsightRow } from "./insight-row";

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
