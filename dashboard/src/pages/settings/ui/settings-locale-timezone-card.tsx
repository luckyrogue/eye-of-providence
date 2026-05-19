import { useTranslation } from "react-i18next";
import { Card, CardContent, CardHeader, CardTitle } from "@eop/ui";
import { Globe } from "lucide-react";
import { LocaleSelect } from "./locale-select";
import { TimezoneSelect } from "./timezone-select";

export function SettingsLocaleTimezoneCard() {
  const { t } = useTranslation("common");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Globe className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.locale_timezone_title")}</CardTitle>
        </div>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-6 md:grid-cols-2 md:gap-3">
        <div className="space-y-2 min-w-0">
          <div className="text-sm font-medium">{t("settings.language")}</div>
          <LocaleSelect />
        </div>
        <div className="space-y-2 min-w-0">
          <div className="text-sm font-medium">{t("settings.timezone")}</div>
          <TimezoneSelect />
        </div>
      </CardContent>
    </Card>
  );
}
