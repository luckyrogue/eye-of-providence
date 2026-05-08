import { Avatar, Select, useConfirm } from "@eop/ui";
import { Brain, UserMinus } from "lucide-react";
import { useUpdateMemberRole, useRemoveMember, type MemberStat, type TeamMember } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/useMutationToast";
import { translateRole } from "../utils";

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
      success: "Роль обновлена",
      error: "Не удалось изменить роль",
    });
  }

  async function remove() {
    const ok = await confirm({
      title: `Удалить ${member.display_name}?`,
      description: "Участник потеряет доступ к этой команде. Историю событий это не затронет.",
      destructive: true,
      confirmText: "Удалить",
    });
    if (!ok) return;
    await runToast(removeM.mutateAsync(member.id), {
      success: "Участник удалён",
      error: "Не удалось удалить",
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
              {Math.round(stat.total_ms / 60000)} мин · 7д
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
            <option value="owner">владелец</option>
            <option value="admin">админ</option>
            <option value="member">участник</option>
          </Select>
        ) : (
          <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
            {translateRole(member.role)}
          </span>
        )}
        {canManage && (
          <button
            type="button"
            onClick={remove}
            disabled={busy}
            title="Удалить из команды"
            aria-label="Удалить из команды"
            className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
          >
            <UserMinus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    </li>
  );
}
