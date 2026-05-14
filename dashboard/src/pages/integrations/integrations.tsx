import { useTranslation } from "react-i18next";
import { APITokensWidget } from "../../widgets/api-tokens";
import { DevicesWidget } from "../../widgets/devices";
import { PushNotificationsWidget } from "../../widgets/push-notifications";
import { WebhooksWidget } from "../../widgets/webhooks";
export function IntegrationsPage() {
  const { t } = useTranslation("common");
  return (
    <div className="space-y-4">
      <div className="page-head">
        <div>
          <h1>{t("integrations.title")}</h1>
          <div className="sub font-mono">{t("integrations.lead")}</div>
        </div>
      </div>

      <PushNotificationsWidget />
      <DevicesWidget />
      <APITokensWidget />
      <WebhooksWidget />
    </div>
  );
}
