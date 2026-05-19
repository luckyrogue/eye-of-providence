import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { useChangeMyName } from "@/entities/user";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";
import { changeNameSchema, type ChangeNameValues } from "@/shared/lib/schemas";

export function ChangeNameForm({
  displayName,
  lastName,
}: {
  displayName?: string;
  lastName?: string;
}) {
  const { t } = useTranslation(["common", "auth", "errors"]);
  const runToast = useMutationToast();
  const change = useChangeMyName();
  const form = useForm<ChangeNameValues>({
    resolver: zodResolver(changeNameSchema),
    defaultValues: { displayName: displayName ?? "", lastName: lastName ?? "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: ChangeNameValues) {
    const trimmed = values.lastName.trim();
    await runToast(
      change.mutateAsync({
        displayName: values.displayName,
        lastName: trimmed === "" ? null : trimmed,
      }),
      {
        success: t("common:settings.name_change_success"),
        error: t("common:settings.name_change_failed"),
      },
    );
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <InputField
            control={form.control}
            name="displayName"
            label={t("common:settings.first_name")}
            autoComplete="given-name"
            translateError={tr}
          />
          <InputField
            control={form.control}
            name="lastName"
            label={t("common:settings.last_name")}
            autoComplete="family-name"
            translateError={tr}
          />
        </div>
        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={form.formState.isSubmitting || change.isPending}
            size="sm"
          >
            {form.formState.isSubmitting || change.isPending ? "..." : t("common:actions.save")}
          </Button>
        </div>
      </form>
    </Form>
  );
}
