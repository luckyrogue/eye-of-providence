import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useRevokeToken, type APIToken } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
export function RevokeTokenButton({ token }: { token: APIToken }) {
  const { t } = useTranslation("developer");
  const revoke = useRevokeToken();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  async function doRevoke() {
    const ok = await confirm({
      title: t("tokens_revoke_confirm"),
      description: t("tokens_revoke_confirm_lead"),
      destructive: true,
      confirmText: t("tokens_revoke_btn"),
    });
    if (!ok) return;
    await runToast(revoke.mutateAsync(token.id), {});
  }
  return (
    <Button variant="ghost" size="sm" onClick={doRevoke} disabled={revoke.isPending}>
      <Trash2 className="h-3.5 w-3.5 mr-1" />
      {t("tokens_revoke")}
    </Button>
  );
}
