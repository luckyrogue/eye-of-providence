import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useDeleteWebhook, type Webhook } from "../../../entities/webhook";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function DeleteWebhookButton({ webhook }: { webhook: Webhook }) {
  const { t } = useTranslation("developer");
  const del = useDeleteWebhook();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function doDelete() {
    const ok = await confirm({
      title: t("webhooks_delete_confirm"),
      description: t("webhooks_delete_confirm_lead"),
      destructive: true,
    });
    if (!ok) return;
    await runToast(del.mutateAsync(webhook.id), {});
  }

  return (
    <Button variant="ghost" size="sm" onClick={doDelete} disabled={del.isPending}>
      <Trash2 className="h-3.5 w-3.5 mr-1" />
      {t("tokens_revoke")}
    </Button>
  );
}
