import { useForm } from "react-hook-form";
import { Button, DangerZone, Input, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useUpdateTeam, useDeleteTeam, type Team } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

interface RenameForm {
  name: string;
}

export function SettingsTab({ team }: { team: Team }) {
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
      success: "Сохранено",
      error: "Не удалось переименовать",
    });
    if (ok !== null) reset({ name });
  }

  async function destroy() {
    const ok = await confirm({
      title: `Удалить «${team.name}»?`,
      description: "Уничтожит участников, проекты, инвайты и историю коммитов. Необратимо.",
      typeToConfirm: team.name,
      destructive: true,
      confirmText: "Удалить навсегда",
    });
    if (!ok) return;
    await runToast(del.mutateAsync(), {
      success: "Команда удалена",
      error: "Не удалось удалить",
    });
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit(onSave)} className="space-y-2">
        <Input
          label="Название"
          {...register("name", { required: true, maxLength: 100 })}
          disabled={update.isPending}
        />
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={update.isPending || !isDirty}>
            Сохранить
          </Button>
        </div>
      </form>

      <DangerZone
        title="Удалить команду"
        description="Уничтожит участников, проекты, инвайты и историю коммитов. Необратимо."
        action={
          <Button onClick={destroy} disabled={del.isPending} variant="destructive" size="sm">
            <Trash2 className="h-3.5 w-3.5 mr-1" /> Удалить
          </Button>
        }
      />
    </div>
  );
}
