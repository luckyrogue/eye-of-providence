import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { User } from "lucide-react";
import { ChangeNameForm } from "./change-name-form";

export function SettingsNameCard({
  displayName,
  lastName,
}: {
  displayName: string;
  lastName: string;
}) {
  const { t } = useTranslation("common");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <User className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.name_change_title")}</CardTitle>
        </div>
        <CardDescription>{t("settings.name_change_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <ChangeNameForm displayName={displayName} lastName={lastName} />
      </CardContent>
    </Card>
  );
}
