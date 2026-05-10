import { useTranslation } from "react-i18next";
import type { Webhook } from "../../../entities/webhook";
import { DeleteWebhookButton } from "../../../features/webhook-delete";

export function WebhookRow({ hook }: { hook: Webhook }) {
  const { t } = useTranslation("developer");
  const last =
    hook.last_delivery_at && hook.last_status != null
      ? t("webhooks_last_delivery", {
          at: new Date(hook.last_delivery_at).toLocaleString(),
          status: hook.last_status === -1 ? t("webhooks_status_network") : hook.last_status,
        })
      : t("webhooks_no_deliveries");

  return (
    <li className="flex items-center justify-between py-3 gap-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs truncate">{hook.url}</span>
          {hook.format === "slack" && (
            <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-600 dark:text-purple-400">
              slack
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 mt-1 flex-wrap">
          {hook.events.map((e) => (
            <span key={e} className="text-[11px] uppercase tracking-wider px-1.5 py-0.5 rounded bg-muted">
              {e}
            </span>
          ))}
        </div>
        <div className="text-xs text-muted-foreground mt-1">{last}</div>
      </div>
      <DeleteWebhookButton webhook={hook} />
    </li>
  );
}
