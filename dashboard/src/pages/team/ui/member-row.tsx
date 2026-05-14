import { useTranslation } from "react-i18next";
import { Avatar, AvatarFallback, getInitials } from "@eop/ui";
import { Brain } from "lucide-react";
import type { MemberStat, TeamMember } from "../../../entities/team";
import { RemoveMemberButton } from "../../../features/team-remove-member";
import { MemberRoleSelect } from "../../../features/team-update-member-role";
import { isReadOnlyRole } from "../../../shared/ui/role-tooltip";
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
  const viewerCanMutate = !isReadOnlyRole(myRole);
  const canManage =
    viewerCanMutate && (myRole === "owner" || (myRole === "admin" && member.role !== "owner"));
  const canChangeRole = viewerCanMutate && myRole === "owner";
  return (
    <li className="flex items-center justify-between rounded-md border p-3 hover:bg-muted/30 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <Avatar>
          <AvatarFallback>{getInitials(member.display_name)}</AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <div className="text-sm font-medium truncate">{member.display_name}</div>
          <div className="text-xs text-muted-foreground truncate">{member.email}</div>
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        {stat && stat.total_ms > 0 && (
          <div className="text-right hidden sm:block">
            <div className="flex items-center gap-1 text-sm justify-end">
              <Brain className="h-3.5 w-3.5" style={{ color: "hsl(var(--accent))" }} />
              <span className="font-medium tabular-nums">{stat.ai_ratio}%</span>
            </div>
            <div className="text-[10px] text-muted-foreground tabular-nums font-mono">
              {t("team_detail.member_minutes_7d", { minutes: Math.round(stat.total_ms / 60000) })}
            </div>
          </div>
        )}
        {canChangeRole ? (
          <MemberRoleSelect teamID={teamID} userID={member.id} value={member.role} />
        ) : (
          <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
            {t(`team_detail.role.${member.role}` as const)}
          </span>
        )}
        {canManage && (
          <RemoveMemberButton
            teamID={teamID}
            userID={member.id}
            displayName={member.display_name}
          />
        )}
      </div>
    </li>
  );
}
