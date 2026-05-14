import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Shield } from "lucide-react";
export function SettingsPrivacyCard() {
  const { t } = useTranslation("common");
  const privacyItems = [
    t("settings.privacy_files"),
    t("settings.privacy_keystrokes"),
    t("settings.privacy_private_windows"),
    t("settings.privacy_clipboard"),
  ];
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Shield className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.privacy_title")}</CardTitle>
        </div>
        <CardDescription>{t("settings.privacy_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
          {privacyItems.map((s) => (
            <li key={s}>{s}</li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
