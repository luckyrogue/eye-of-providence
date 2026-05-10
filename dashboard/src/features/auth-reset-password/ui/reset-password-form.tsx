import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { resetPassword } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { resetPasswordSchema, type ResetPasswordValues } from "../../../shared/lib/schemas";

export function ResetPasswordForm({ token, onDone }: { token: string; onDone: () => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const form = useForm<ResetPasswordValues>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { password: "", confirmPassword: "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: ResetPasswordValues) {
    const ok = await runToast(resetPassword(token, values.password), {
      error: t("errors:reset_password_failed"),
    });
    if (ok !== undefined) onDone();
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <InputField
          control={form.control}
          name="password"
          label={t("auth:reset_field_password")}
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
        <Button type="submit" disabled={form.formState.isSubmitting} className="w-full">
          {form.formState.isSubmitting ? "..." : t("auth:reset_submit")}
        </Button>
      </form>
    </Form>
  );
}
