import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useUnsubscribePush, type Subscription } from "@/entities/push";
import { useMutationToast } from "@/shared/hooks";

export function DisablePushButton({ sub }: { sub: Subscription }) {
  const { t } = useTranslation("pwa");
  const unsubscribe = useUnsubscribePush();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function disable() {
    const ok = await confirm({
      title: t("disable_confirm"),
      destructive: true,
    });
    if (!ok) return;
    // Unsubscribe браузер локально, чтобы permission entry не висел.
    try {
      const reg = await navigator.serviceWorker.ready;
      const browserSub = await reg.pushManager.getSubscription();
      if (browserSub && browserSub.endpoint === sub.endpoint) {
        await browserSub.unsubscribe();
      }
    } catch {
      // best-effort
    }
    await runToast(unsubscribe.mutateAsync(sub.endpoint), {});
  }

  return (
    <Button size="sm" variant="ghost" onClick={disable} disabled={unsubscribe.isPending}>
      <Trash2 className="h-3.5 w-3.5 mr-1" />
      {t("disable")}
    </Button>
  );
}
