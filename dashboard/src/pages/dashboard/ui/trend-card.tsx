import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Trend } from "../../../shared/ui/charts/trend";
import type { TrendPoint } from "../../../entities/event";
export function TrendCard({ points }: { points: TrendPoint[] }) {
  const { t } = useTranslation("app");
  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">{t("dashboard.trend_title")}</CardTitle>
        <CardDescription>{t("dashboard.trend_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Trend points={points} />
      </CardContent>
    </Card>
  );
}
