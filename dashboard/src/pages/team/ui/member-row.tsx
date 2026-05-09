import { useTranslation } from "react-i18next";
import { Avatar, Select, useConfirm } from "@eop/ui";
import { Brain, UserMinus } from "lucide-react";
import { useUpdateMemberRole, useRemoveMember, type MemberStat, type TeamMember } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function MemberRow({
  member,
  stat,
  myRole,
  teamID,
}: {
  member: TeamMember;
  stat?: MemberStat;
  myRole: string;
  teamID: string;
}) {
  const { t } = useTranslation("app");
  const updateRole = useUpdateMemberRole(teamID);
  const removeM = useRemoveMember(teamID);
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const canManage = myRole === "owner" || (myRole === "admin" && member.role !== "owner");
  const canChangeRole = myRole === "owner";
  const busy = updateRole.isPending || removeM.isPending;

  async function changeRole(role: string) {
    if (role === member.role) return;
    await runToast(updateRole.mutateAsync({ userID: member.id, role }), {
      success: t("team_detail.member_role_updated"),
      error: t("team_detail.member_role_update_failed"),
    });
  }

  async function remove() {
    const ok = await confirm({
      title: t("team_detail.member_remove_title", { name: member.display_name }),
      description: t("team_detail.member_remove_lead"),
      destructive: true,
      confirmText: t("team_detail.member_remove_confirm"),
    });
    if (!ok) return;
    await runToast(removeM.mutateAsync(member.id), {
      success: t("team_detail.member_removed"),
      error: t("team_detail.member_remove_failed"),
    });
  }

  return (
    <li className="flex items-center justify-between rounded-md border p-3 hover:bg-muted/30 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <Avatar name={member.display_name} />
        <div className="min-w-0">
          <div className="text-sm font-medium truncate">{member.display_name}</div>
          <div className="text-xs text-muted-foreground truncate">{member.email}</div>
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        {stat && stat.total_ms > 0 && (
          <div className="text-right hidden sm:block">
            <div className="flex items-center gap-1 text-sm justify-end">
              <Brain className="h-3.5 w-3.5 text-purple-500" />
              <span className="font-medium tabular-nums">{stat.ai_ratio}%</span>
            </div>
            <div className="text-[10px] text-muted-foreground tabular-nums font-mono">
              {t("team_detail.member_minutes_7d", { minutes: Math.round(stat.total_ms / 60000) })}
            </div>
          </div>
        )}
        {canChangeRole ? (
          <Select
            mono
            value={member.role}
            disabled={busy}
            onChange={(e) => changeRole(e.target.value)}
            className="px-2 py-1 text-xs"
          >
            <option value="owner">{t("team_detail.role.owner")}</option>
            <option value="admin">{t("team_detail.role.admin")}</option>
            <option value="member">{t("team_detail.role.member")}</option>
          </Select>
        ) : (
          <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
            {t(`team_detail.role.${member.role}` as const, { defaultValue: member.role })}
          </span>
        )}
        {canManage && (
          <button
            type="button"
            onClick={remove}
            disabled={busy}
            title={t("team_detail.member_remove_btn_title")}
            aria-label={t("team_detail.member_remove_btn_title")}
            className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
          >
            <UserMinus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    </li>
  );
}
