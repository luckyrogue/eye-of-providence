import { useTranslation } from "react-i18next";
import type { MemberStat, TeamMember } from "@/entities/team";
import { CreateInviteBlock } from "@/features/team-create-invite";
import { ObserverHint } from "./role-hint";
import { isReadOnlyRole } from "../lib/roles";
import { MemberRow } from "./member-row";

export function MembersTab({
  teamID,
  role,
  members,
  stats,
}: {
  teamID: string;
  role: string;
  members: TeamMember[];
  stats: MemberStat[];
}) {
  const { t } = useTranslation("app");
  // Observer is strictly read-only; admin/owner mutate.
  const canInvite = (role === "owner" || role === "admin") && !isReadOnlyRole(role);
  const readOnly = isReadOnlyRole(role);

  return (
    <>
      {canInvite && <CreateInviteBlock teamID={teamID} />}
      {readOnly && (
        <div className="flex justify-end">
          <ObserverHint
            label={t("team.role_badge.observer")}
            hint={t("team.role_badge.observer_tooltip")}
          />
        </div>
      )}

      <ul className="space-y-2">
        {members.map((m) => (
          <MemberRow
            key={m.id}
            member={m}
            stat={stats.find((s) => s.id === m.id)}
            myRole={role}
            teamID={teamID}
          />
        ))}
      </ul>
    </>
  );
}
