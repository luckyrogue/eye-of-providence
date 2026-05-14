import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { useDeleteMyData, useProfile } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
export function DeleteMyDataButton({ onWiped }: { onWiped: () => void }) {
  const { t } = useTranslation("common");
  const profile = useProfile();
  const deleteData = useDeleteMyData();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  async function destroy() {
    const ok = await confirm({
      title: t("settings.danger_confirm_title"),
      description: t("settings.danger_confirm_lead"),
      typeToConfirm: profile.data?.email ?? "delete",
      destructive: true,
      confirmText: t("settings.danger_confirm_btn"),
    });
    if (!ok) return;
    const r = await runToast(deleteData.mutateAsync(), {
      success: t("settings.delete_success"),
      error: t("settings.delete_failed"),
    });
    if (r !== null) onWiped();
  }
  return (
    <Button variant="destructive" size="sm" onClick={destroy} disabled={deleteData.isPending}>
      {t("actions.delete")}
    </Button>
  );
}
