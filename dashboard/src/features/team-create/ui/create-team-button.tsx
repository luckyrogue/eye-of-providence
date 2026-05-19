import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, cn, PromptDialog } from "@eop/ui";
import { Plus } from "lucide-react";
import { useCreateTeam } from "@/entities/team";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";

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
        type="button"
        size="sm"
        onClick={() => setOpen(true)}
        disabled={create.isPending || disabled}
        className={cn(
          "justify-center gap-2 sm:h-10 sm:w-10 sm:min-w-10 sm:max-w-10 sm:shrink-0 sm:gap-0 sm:px-0",
          className,
        )}
        title={blockedReason ?? t("app:teams.new_team")}
        aria-label={t("app:teams.new_team")}
      >
        <Plus className="hidden h-4 w-4 shrink-0 sm:block" aria-hidden />
        <span className="sm:hidden">{t("common:actions.add")}</span>
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
