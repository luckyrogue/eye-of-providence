// Webhooks management — Settings page section.
//
// Симметрия с api-tokens widget: список + create-feature + delete-feature.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Plus, Webhook as WebhookIcon } from "lucide-react";
import { useWebhooks } from "../../../entities/webhook";
import { CreateWebhookDialog } from "../../../features/webhook-create";
import { WebhookRow } from "./webhook-row";

export function WebhooksWidget() {
  const { t } = useTranslation("developer");
  const list = useWebhooks();
  const [showCreate, setShowCreate] = useState(false);

  return (
    <Card>
      <CardHeader className="flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <WebhookIcon className="h-4 w-4 text-muted-foreground shrink-0" />
            <CardTitle>{t("webhooks_title")}</CardTitle>
          </div>
          <CardDescription className="mt-1">{t("webhooks_lead")}</CardDescription>
        </div>
        <Button
          type="button"
          size="sm"
          variant="default"
          onClick={() => setShowCreate(true)}
          className="w-full shrink-0 justify-center gap-2 sm:h-10 sm:w-10 sm:min-w-10 sm:max-w-10 sm:gap-0 sm:px-0"
          title={t("webhooks_create")}
          aria-label={t("webhooks_create")}
        >
          <Plus className="h-4 w-4 shrink-0" aria-hidden />
          <span className="sm:hidden">{t("webhooks_create")}</span>
        </Button>
      </CardHeader>
      <CardContent>
        {list.isPending ? (
          <p className="text-sm text-muted-foreground">…</p>
        ) : !list.data || list.data.length === 0 ? null : (
          <ul className="divide-y">
            {list.data.map((hook) => (
              <WebhookRow key={hook.id} hook={hook} />
            ))}
          </ul>
        )}
      </CardContent>

      <CreateWebhookDialog open={showCreate} onClose={() => setShowCreate(false)} />
    </Card>
  );
}
