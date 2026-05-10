import { useTranslation } from "react-i18next";
import { IconButton, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useAdminDeleteUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function DeleteUserButton({ userID, email }: { userID: string; email: string }) {
  const { t } = useTranslation("app");
  const del = useAdminDeleteUser();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function destroy() {
    const ok = await confirm({
      title: t("admin.user_delete_confirm_title", { email, defaultValue: `Delete ${email}?` }),
      description: t("admin.user_delete_confirm_lead", {
        defaultValue: "All events and reports will be erased. Irreversible.",
      }),
      destructive: true,
      confirmText: t("admin.users_delete"),
    });
    if (!ok) return;
    await runToast(del.mutateAsync(userID), {
      success: t("admin.users_deleted"),
      error: t("admin.users_delete_failed"),
    });
  }

  return (
    <IconButton
      danger
      title={t("admin.user_delete_btn", { defaultValue: "Delete user" })}
      onClick={destroy}
      disabled={del.isPending}
    >
      <Trash2 className="h-3.5 w-3.5" />
    </IconButton>
  );
}
