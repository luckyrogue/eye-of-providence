import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { useChangeMyPassword } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { changePasswordSchema, type ChangePasswordValues } from "../../../shared/lib/schemas";

export function ChangePasswordForm() {
  const { t } = useTranslation(["common", "auth", "errors"]);
  const runToast = useMutationToast();
  const change = useChangeMyPassword();
  const form = useForm<ChangePasswordValues>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { currentPassword: "", newPassword: "", confirmPassword: "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: ChangePasswordValues) {
    const ok = await runToast(
      change.mutateAsync({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      }),
      {
        success: t("common:settings.password_change_success"),
        error: t("common:settings.password_change_failed"),
      },
    );
    if (ok !== undefined) form.reset();
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <InputField
            control={form.control}
            name="currentPassword"
            label={t("common:settings.current_password")}
            type="password"
            autoComplete="current-password"
            translateError={tr}
          />
          <div className="space-y-3">
            <InputField
              control={form.control}
              name="newPassword"
              label={t("common:settings.new_password")}
              type="password"
              autoComplete="new-password"
              translateError={tr}
            />
            <InputField
              control={form.control}
              name="confirmPassword"
              label={t("auth:reset_field_password2")}
              type="password"
              autoComplete="new-password"
              translateError={tr}
            />
          </div>
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
