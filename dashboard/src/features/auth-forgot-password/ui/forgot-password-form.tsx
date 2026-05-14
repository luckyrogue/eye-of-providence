import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { forgotPassword } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { forgotPasswordSchema, type ForgotPasswordValues } from "../../../shared/lib/schemas";
export function ForgotPasswordForm({ onSent }: { onSent: () => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const form = useForm<ForgotPasswordValues>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : msg);
  async function onSubmit(values: ForgotPasswordValues) {
    const ok = await runToast(forgotPassword(values.email), {
      error: t("errors:reset_email_failed"),
    });
    if (ok !== undefined) onSent();
  }
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <InputField
          control={form.control}
          name="email"
          label={t("auth:field_email")}
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          translateError={tr}
        />
        <Button type="submit" disabled={form.formState.isSubmitting} className="w-full">
          {form.formState.isSubmitting ? "..." : t("auth:forgot_submit")}
        </Button>
      </form>
    </Form>
  );
}
