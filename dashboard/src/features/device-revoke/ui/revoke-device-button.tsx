import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useRevokeDevice, type Device } from "@/entities/device";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";

export function RevokeDeviceButton({ device }: { device: Device }) {
  const { t } = useTranslation("developer");
  const revoke = useRevokeDevice();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function doRevoke() {
    const ok = await confirm({
      title: t("devices_revoke_confirm"),
      description: t("devices_revoke_confirm_lead"),
      destructive: true,
      confirmText: t("devices_revoke_btn"),
    });
    if (!ok) return;
    await runToast(revoke.mutateAsync(device.id), {});
  }

  return (
    <Button variant="ghost" size="sm" onClick={doRevoke} disabled={revoke.isPending}>
      <Trash2 className="h-3.5 w-3.5 mr-1" />
      {t("devices_revoke")}
    </Button>
  );
}
