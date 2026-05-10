import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@eop/ui";
import { Copy, Plus } from "lucide-react";
import { useCreateInvite } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// CreateInviteBlock — встраиваемый блок: «Сгенерировать invite-link»
// → показать read-only поле с URL и кнопкой Copy. Только для owner/admin
// (родительский гард решает).
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

  function copy() {
    if (!inviteUrl) return;
    navigator.clipboard.writeText(inviteUrl);
    toast.success(t("team_detail.invite_copied"));
  }

  return (
    <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Plus className="h-4 w-4" /> {t("team_detail.invite_section_title")}
      </div>
      {code ? (
        <div className="flex items-center gap-2">
          <input
            readOnly
            value={inviteUrl}
            className="flex-1 rounded-md border bg-background px-2 py-1.5 text-xs font-mono"
          />
          <Button size="sm" variant="outline" onClick={copy}>
            <Copy className="h-3.5 w-3.5 mr-1" /> {t("team_detail.invite_copy")}
          </Button>
        </div>
      ) : (
        <Button size="sm" onClick={generate} disabled={createInvite.isPending}>
          {t("team_detail.invite_create_link")}
        </Button>
      )}
      <p className="text-xs text-muted-foreground">{t("team_detail.invite_lead")}</p>
    </div>
  );
}
