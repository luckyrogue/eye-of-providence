// Webhooks management — Settings page section.
//
// Симметрия с api-tokens widget: список + create-feature + delete-feature.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button, Card, CardContent, CardDescription, CardHeader, CardTitle,
} from "@eop/ui";
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
        <Button size="sm" onClick={() => setShowCreate(true)} className="w-full sm:w-auto shrink-0">
          <Plus className="h-3.5 w-3.5 mr-1" />
          {t("webhooks_create")}
        </Button>
      </CardHeader>
      <CardContent>
        {list.isPending ? (
          <p className="text-sm text-muted-foreground">…</p>
        ) : !list.data || list.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("webhooks_empty")}</p>
        ) : (
          <ul className="divide-y">
            {list.data.map((hook) => <WebhookRow key={hook.id} hook={hook} />)}
          </ul>
        )}
      </CardContent>

      <CreateWebhookDialog open={showCreate} onClose={() => setShowCreate(false)} />
    </Card>
  );
}
