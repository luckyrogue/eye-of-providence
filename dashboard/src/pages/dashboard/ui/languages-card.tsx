import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Languages } from "../../../shared/ui/charts/languages";
import type { LangCell } from "../../../entities/event";
export function LanguagesCard({ cells }: { cells: LangCell[] }) {
  const { t } = useTranslation("app");
  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">
          {t("dashboard.languages_title")}
        </CardTitle>
        <CardDescription>{t("dashboard.languages_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Languages cells={cells} top={6} />
      </CardContent>
    </Card>
  );
}
