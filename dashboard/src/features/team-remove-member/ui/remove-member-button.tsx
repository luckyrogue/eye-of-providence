import { useTranslation } from "react-i18next";
import { useConfirm } from "@eop/ui";
import { UserMinus } from "lucide-react";
import { useRemoveMember } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function RemoveMemberButton({
  teamID,
  userID,
  displayName,
  disabled,
}: {
  teamID: string;
  userID: string;
  displayName: string;
  disabled?: boolean;
}) {
  const { t } = useTranslation("app");
  const removeM = useRemoveMember(teamID);
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function remove() {
    const ok = await confirm({
      title: t("team_detail.member_remove_title", { name: displayName }),
      description: t("team_detail.member_remove_lead"),
      destructive: true,
      confirmText: t("team_detail.member_remove_confirm"),
    });
    if (!ok) return;
    await runToast(removeM.mutateAsync(userID), {
      success: t("team_detail.member_removed"),
      error: t("team_detail.member_remove_failed"),
    });
  }

  return (
    <button
      type="button"
      onClick={remove}
      disabled={disabled || removeM.isPending}
      title={t("team_detail.member_remove_btn_title")}
      aria-label={t("team_detail.member_remove_btn_title")}
      className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
    >
      <UserMinus className="h-3.5 w-3.5" />
    </button>
  );
}
