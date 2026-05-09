import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Button, DangerZone, Input, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useUpdateTeam, useDeleteTeam, type Team } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

interface RenameForm {
  name: string;
}

export function SettingsTab({ team }: { team: Team }) {
  const { t } = useTranslation(["app", "common"]);
  const update = useUpdateTeam(team.id);
  const del = useDeleteTeam(team.id);
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const { register, handleSubmit, formState: { isDirty }, reset } = useForm<RenameForm>({
    defaultValues: { name: team.name },
  });

  async function onSave(values: RenameForm) {
    const name = values.name.trim();
    if (!name || name === team.name) return;
    const ok = await runToast(update.mutateAsync(name), {
      success: t("app:team_detail.settings_save_success"),
      error: t("app:team_detail.settings_save_failed"),
    });
    if (ok !== null) reset({ name });
  }

  async function destroy() {
    const ok = await confirm({
      title: t("app:team_detail.settings_danger_confirm_title", { name: team.name }),
      description: t("app:team_detail.settings_danger_lead"),
      typeToConfirm: team.name,
      destructive: true,
      confirmText: t("app:team_detail.settings_danger_confirm"),
    });
    if (!ok) return;
    await runToast(del.mutateAsync(), {
      success: t("app:team_detail.settings_deleted"),
      error: t("app:team_detail.settings_delete_failed"),
    });
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit(onSave)} className="space-y-2">
        <Input
          label={t("app:team_detail.settings_label_name")}
          {...register("name", { required: true, maxLength: 100 })}
          disabled={update.isPending}
        />
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={update.isPending || !isDirty}>
            {t("common:actions.save")}
          </Button>
        </div>
      </form>

      <DangerZone
        title={t("app:team_detail.settings_danger_title")}
        description={t("app:team_detail.settings_danger_lead")}
        action={
          <Button onClick={destroy} disabled={del.isPending} variant="destructive" size="sm">
            <Trash2 className="h-3.5 w-3.5 mr-1" /> {t("common:actions.delete")}
          </Button>
        }
      />
    </div>
  );
}
