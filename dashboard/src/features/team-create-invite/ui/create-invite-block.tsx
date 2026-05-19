import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, SecretField, toast } from "@eop/ui";
import { Plus } from "lucide-react";
import { useCreateInvite } from "@/entities/team";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";

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
    <div className="rounded-lg border bg-muted/30 p-3">
      <div className="flex flex-row items-start justify-between gap-3">
        <div className="min-w-0 flex-1 space-y-1">
          <div className="text-sm font-medium">{t("team_detail.invite_section_title")}</div>
          <p className="text-xs text-muted-foreground">{t("team_detail.invite_lead")}</p>
        </div>
        {!code && (
          <Button
            type="button"
            size="icon"
            variant="default"
            onClick={() => void generate()}
            disabled={createInvite.isPending}
            className="shrink-0"
            aria-label={t("team_detail.invite_create_link")}
            title={t("team_detail.invite_create_link")}
          >
            <Plus className="h-4 w-4" aria-hidden />
          </Button>
        )}
      </div>
      {code ? (
        <div className="mt-3">
          <SecretField
            value={inviteUrl}
            copyLabel={t("team_detail.invite_copy")}
            copiedLabel={t("team_detail.invite_copied")}
            onCopy={() => toast.success(t("team_detail.invite_copied"))}
          />
        </div>
      ) : null}
    </div>
  );
}
