import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { useUpdateTeam, type Team } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

interface RenameForm {
  name: string;
}

export function RenameTeamForm({ team }: { team: Team }) {
  const { t } = useTranslation(["app", "common"]);
  const update = useUpdateTeam(team.id);
  const runToast = useMutationToast();
  const form = useForm<RenameForm>({
    defaultValues: { name: team.name },
  });

  async function onSave(values: RenameForm) {
    const name = values.name.trim();
    if (!name || name === team.name) return;
    const ok = await runToast(update.mutateAsync(name), {
      success: t("app:team_detail.settings_save_success"),
      error: t("app:team_detail.settings_save_failed"),
    });
    if (ok !== null) form.reset({ name });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSave)} className="space-y-2">
        <InputField
          control={form.control}
          name="name"
          label={t("app:team_detail.settings_label_name")}
          disabled={update.isPending}
          hideMessage
        />
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={update.isPending || !form.formState.isDirty}>
            {t("common:actions.save")}
          </Button>
        </div>
      </form>
    </Form>
  );
}
