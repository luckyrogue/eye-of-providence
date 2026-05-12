import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, PromptDialog } from "@eop/ui";
import { Plus } from "lucide-react";
import { useCreateTeam } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// Side-effects после создания (выбор активной команды, switchTeam, …)
// решает родитель через `onCreated(teamID)` — фича сама не знает, что
// делать с id.
export function CreateTeamButton({
  disabled,
  blockedReason,
  onCreated,
  className,
}: {
  disabled?: boolean;
  blockedReason?: string;
  onCreated?: (teamID: string) => void;
  className?: string;
}) {
  const { t } = useTranslation(["app", "errors", "common"]);
  const create = useCreateTeam();
  const runToast = useMutationToast();
  const [open, setOpen] = useState(false);

  async function submit(name: string) {
    try {
      const r = await create.mutateAsync(name);
      await runToast(Promise.resolve(r), { success: t("app:teams.created_toast") });
      onCreated?.(r.id);
      setOpen(false);
    } catch (e) {
      const code = (e as { code?: string }).code;
      const errorMsg = code
        ? t(`errors:${code}`, { defaultValue: t("errors:generic") })
        : t("errors:generic");
      await runToast(Promise.reject(new Error(errorMsg)), {
        error: t("errors:team_create_failed"),
      });
    }
  }

  return (
    <>
      <Button
        size="sm"
        onClick={() => setOpen(true)}
        disabled={create.isPending || disabled}
        className={className}
        title={blockedReason}
      >
        <Plus className="h-3.5 w-3.5 mr-1" /> {t("app:teams.new_team")}
      </Button>

      <PromptDialog
        open={open}
        title={t("app:teams.create_dialog_title")}
        description={t("app:teams.create_dialog_lead")}
        label={t("app:teams.create_dialog_label")}
        placeholder={t("app:teams.create_dialog_placeholder")}
        confirmText={t("common:actions.create")}
        busy={create.isPending}
        onClose={() => setOpen(false)}
        onConfirm={submit}
      />
    </>
  );
}
