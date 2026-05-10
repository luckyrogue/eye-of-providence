import { useTranslation } from "react-i18next";
import { useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useAdminDeleteTeam } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { IconButton } from "../../../shared/ui/icon-button";

export function DeleteTeamButton({ teamID, name }: { teamID: string; name: string }) {
  const { t } = useTranslation(["app", "common"]);
  const deleteTeam = useAdminDeleteTeam();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function destroy() {
    const ok = await confirm({
      title: t("admin.team_delete_confirm_title", { name, defaultValue: `Delete "${name}"?` }),
      description: t("admin.team_delete_confirm_lead", {
        defaultValue: "Will erase all members, projects, invites, commit history. Irreversible.",
      }),
      typeToConfirm: name,
      destructive: true,
      confirmText: t("team_detail.settings_danger_confirm"),
    });
    if (!ok) return;
    await runToast(deleteTeam.mutateAsync(teamID), {
      success: t("admin.team_deleted", { defaultValue: "Team deleted" }),
      error: t("admin.team_delete_failed", { defaultValue: "Failed to delete" }),
    });
  }

  return (
    <IconButton
      danger
      title={t("admin.team_delete_btn", { defaultValue: "Delete company" })}
      onClick={destroy}
      disabled={deleteTeam.isPending}
    >
      <Trash2 className="h-3.5 w-3.5" />
    </IconButton>
  );
}
