import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { KeyRound } from "lucide-react";
import { ChangePasswordForm } from "./change-password-form";

export function SettingsPasswordCard({ hasPassword }: { hasPassword: boolean }) {
  const { t } = useTranslation("common");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <KeyRound className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.password_change_title")}</CardTitle>
        </div>
        <CardDescription>{t("settings.password_change_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        {hasPassword ? (
          <ChangePasswordForm />
        ) : (
          <p className="text-sm text-muted-foreground">{t("settings.no_password_set")}</p>
        )}
      </CardContent>
    </Card>
  );
}
