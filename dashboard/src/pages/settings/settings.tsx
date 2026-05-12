import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, DangerZone } from "@eop/ui";
import { Globe, Languages, Shield, User } from "lucide-react";
import { useProfile } from "../../entities/user";
import { DeleteMyDataButton } from "../../features/delete-my-data";
import { LocaleSelect } from "../../features/locale-switch";
import { TimezoneSelect } from "../../features/timezone-switch";
import { APITokensWidget } from "../../widgets/api-tokens";
import { DevicesWidget } from "../../widgets/devices";
import { PushNotificationsWidget } from "../../widgets/push-notifications";
import { WebhooksWidget } from "../../widgets/webhooks";

export function Settings({ onWiped }: { onWiped: () => void }) {
  const { t } = useTranslation("common");
  const profile = useProfile();

  const privacyItems = [
    t("settings.privacy_files"),
    t("settings.privacy_keystrokes"),
    t("settings.privacy_private_windows"),
    t("settings.privacy_clipboard"),
  ];

  return (
    <div className="space-y-4">
      <div className="page-head">
        <div>
          <h1>{t("settings.title")}</h1>
          <div className="sub font-mono">{t("nav.settings")}</div>
        </div>
      </div>
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.profile_title")}</CardTitle>
          </div>
          <CardDescription>{t("settings.profile_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          {profile.data ? (
            <dl className="grid grid-cols-[8rem_1fr] gap-y-2 text-sm">
              <dt className="text-muted-foreground">User ID</dt>
              <dd className="font-mono text-xs break-all">{profile.data.user_id}</dd>
              {profile.data.email && (
                <>
                  <dt className="text-muted-foreground">Email</dt>
                  <dd>{profile.data.email}</dd>
                </>
              )}
              <dt className="text-muted-foreground">{t("settings.profile_provider")}</dt>
              <dd>{profile.data.provider ?? "—"}</dd>
              {profile.data.github_login && (
                <>
                  <dt className="text-muted-foreground">GitHub</dt>
                  <dd>@{profile.data.github_login}</dd>
                </>
              )}
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">
              {profile.isError ? t("settings.profile_load_failed") : t("settings.profile_loading")}
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Languages className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.language")}</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <LocaleSelect />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("settings.timezone")}</CardTitle>
          </div>
          <CardDescription>{t("settings.timezone_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          <TimezoneSelect />
        </CardContent>
      </Card>

      <PushNotificationsWidget />
      <DevicesWidget />
      <APITokensWidget />
      <WebhooksWidget />

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

      <DangerZone
        title={t("settings.danger_title")}
        description={t("settings.danger_lead")}
        action={<DeleteMyDataButton onWiped={onWiped} />}
      />
    </div>
  );
}
