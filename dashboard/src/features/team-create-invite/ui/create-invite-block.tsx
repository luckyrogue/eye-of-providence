import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "@eop/ui";
import { Button, SecretField } from "@eop/ui";
import { Plus } from "lucide-react";
import { useCreateInvite } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// Видимость по роли решает родитель: owner/admin-гард тут не дублируется.
export function CreateInviteBlock({ teamID }: { teamID: string }) {
  const { t } = useTranslation("app");
  const createInvite = useCreateInvite(teamID);
  const runToast = useMutationToast();
  const [code, setCode] = useState<string | null>(null);
  const inviteUrl = code ? `${window.location.origin}/?invite=${code}` : "";

  async function generate() {
    const r = await runToast(createInvite.mutateAsync(), {
      error: t("team_detail.invite_create_failed"),
    });
    if (r) setCode(r.code);
  }

  return (
    <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Plus className="h-4 w-4" /> {t("team_detail.invite_section_title")}
      </div>
      {code ? (
        <SecretField
          value={inviteUrl}
          copyLabel={t("team_detail.invite_copy")}
          copiedLabel={t("team_detail.invite_copied")}
          onCopy={() => toast.success(t("team_detail.invite_copied"))}
        />
      ) : (
        <Button size="sm" onClick={generate} disabled={createInvite.isPending}>
          {t("team_detail.invite_create_link")}
        </Button>
      )}
      <p className="text-xs text-muted-foreground">{t("team_detail.invite_lead")}</p>
    </div>
  );
}
