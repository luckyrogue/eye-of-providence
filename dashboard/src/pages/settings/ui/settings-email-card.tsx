import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Mail } from "lucide-react";
import { ChangeEmailForm } from "../../../features/change-email";
export function SettingsEmailCard({
  hasPassword,
  currentEmail,
}: {
  hasPassword: boolean;
  currentEmail?: string;
}) {
  const { t } = useTranslation("common");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Mail className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.email_change_title")}</CardTitle>
        </div>
        <CardDescription>{t("settings.email_change_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        {hasPassword ? (
          <ChangeEmailForm currentEmail={currentEmail} />
        ) : (
          <p className="text-sm text-muted-foreground">{t("settings.no_password_set")}</p>
        )}
      </CardContent>
    </Card>
  );
}
