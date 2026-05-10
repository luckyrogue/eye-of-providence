import type { MemberStat, TeamMember } from "../../../entities/team";
import { CreateInviteBlock } from "../../../features/team-create-invite";
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
  const canInvite = role === "owner" || role === "admin";

  return (
    <>
      {canInvite && <CreateInviteBlock teamID={teamID} />}

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
