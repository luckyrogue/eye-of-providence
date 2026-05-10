import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Button, Input } from "@eop/ui";
import { useUpdateTeam, type Team } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

interface RenameForm {
  name: string;
}

export function RenameTeamForm({ team }: { team: Team }) {
  const { t } = useTranslation(["app", "common"]);
  const update = useUpdateTeam(team.id);
  const runToast = useMutationToast();
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

  return (
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
  );
}
