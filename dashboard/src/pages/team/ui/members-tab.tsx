import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@eop/ui";
import { Copy, Plus } from "lucide-react";
import { useCreateInvite, type MemberStat, type TeamMember } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
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
  const createInvite = useCreateInvite(teamID);
  const runToast = useMutationToast();
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function makeInvite() {
    const r = await runToast(createInvite.mutateAsync(), { error: t("team_detail.invite_create_failed") });
    if (r) setInviteCode(r.code);
  }

  function copyInvite() {
    if (!inviteUrl) return;
    navigator.clipboard.writeText(inviteUrl);
    toast.success(t("team_detail.invite_copied"));
  }

  return (
    <>
      {(role === "owner" || role === "admin") && (
        <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Plus className="h-4 w-4" /> {t("team_detail.invite_section_title")}
          </div>
          {inviteCode ? (
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={inviteUrl}
                className="flex-1 rounded-md border bg-background px-2 py-1.5 text-xs font-mono"
              />
              <Button size="sm" variant="outline" onClick={copyInvite}>
                <Copy className="h-3.5 w-3.5 mr-1" /> {t("team_detail.invite_copy")}
              </Button>
            </div>
          ) : (
            <Button size="sm" onClick={makeInvite} disabled={createInvite.isPending}>
              {t("team_detail.invite_create_link")}
            </Button>
          )}
          <p className="text-xs text-muted-foreground">{t("team_detail.invite_lead")}</p>
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
