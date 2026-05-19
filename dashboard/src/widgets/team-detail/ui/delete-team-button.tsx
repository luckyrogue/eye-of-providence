import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useDeleteTeam, type Team } from "@/entities/team";
import { useMutationToast } from "@/shared/hooks";

export function DeleteTeamButton({ team }: { team: Team }) {
  const { t } = useTranslation(["app", "common"]);
  const del = useDeleteTeam(team.id);
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function destroy() {
    const ok = await confirm({
      title: t("app:team_detail.settings_danger_confirm_title", { name: team.name }),
      description: t("app:team_detail.settings_danger_lead"),
      typeToConfirm: team.name,
      destructive: true,
      confirmText: t("app:team_detail.settings_danger_confirm"),
    });
    if (!ok) return;
    await runToast(del.mutateAsync(), {
      success: t("app:team_detail.settings_deleted"),
      error: t("app:team_detail.settings_delete_failed"),
    });
  }

  return (
    <Button onClick={destroy} disabled={del.isPending} variant="destructive" size="sm">
      <Trash2 className="h-3.5 w-3.5 mr-1" /> {t("common:actions.delete")}
    </Button>
  );
}
