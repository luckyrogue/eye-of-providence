import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@eop/ui";
import { Copy, Plus } from "lucide-react";
import { useCreateInvite, type MemberStat, type TeamMember } from "../api/teams";
import { useMutationToast } from "../hooks/useMutationToast";
import { MemberRow } from "./MemberRow";

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
  const createInvite = useCreateInvite(teamID);
  const runToast = useMutationToast();
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function makeInvite() {
    const r = await runToast(createInvite.mutateAsync(), { error: "Не удалось создать invite" });
    if (r) setInviteCode(r.code);
  }

  function copyInvite() {
    if (!inviteUrl) return;
    navigator.clipboard.writeText(inviteUrl);
    toast.success("Скопировано");
  }

  return (
    <>
      {(role === "owner" || role === "admin") && (
        <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Plus className="h-4 w-4" /> Пригласить участника
          </div>
          {inviteCode ? (
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={inviteUrl}
                className="flex-1 rounded-md border bg-background px-2 py-1.5 text-xs font-mono"
              />
              <Button size="sm" variant="outline" onClick={copyInvite}>
                <Copy className="h-3.5 w-3.5 mr-1" /> Скопировать
              </Button>
            </div>
          ) : (
            <Button size="sm" onClick={makeInvite} disabled={createInvite.isPending}>
              Создать ссылку
            </Button>
          )}
          <p className="text-xs text-muted-foreground">
            Ссылка работает 7 дней, до 10 регистраций. Отправь её тому, кого хочешь добавить.
          </p>
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
