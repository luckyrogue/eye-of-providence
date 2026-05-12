import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Fingerprint } from "lucide-react";
import { AddPasskeyButton, PasskeyList } from "../../../features/auth-passkey";

export function SettingsPasskeyCard() {
  const { t } = useTranslation("common");
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <Fingerprint className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.passkey_title")}</CardTitle>
          </div>
          <AddPasskeyButton />
        </div>
        <CardDescription>{t("settings.passkey_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <PasskeyList />
      </CardContent>
    </Card>
  );
}
