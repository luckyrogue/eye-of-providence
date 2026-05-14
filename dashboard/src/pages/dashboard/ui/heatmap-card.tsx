import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Heatmap } from "../../../shared/ui/charts/heatmap";
import type { HeatmapCell } from "../../../entities/event";
export function HeatmapCard({ cells, tz }: { cells: HeatmapCell[]; tz: string }) {
  const { t } = useTranslation("app");
  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">
          {t("dashboard.heatmap_title")}
        </CardTitle>
        <CardDescription>{t("dashboard.heatmap_lead", { tz })}</CardDescription>
      </CardHeader>
      <CardContent>
        <Heatmap cells={cells} />
      </CardContent>
    </Card>
  );
}
