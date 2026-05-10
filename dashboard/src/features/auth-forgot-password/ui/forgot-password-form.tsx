import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Input } from "@eop/ui";
import { forgotPassword } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { forgotPasswordSchema, type ForgotPasswordValues } from "../../../shared/lib/schemas";

export function ForgotPasswordForm({ onSent }: { onSent: () => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordValues>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : undefined);

  async function onSubmit(values: ForgotPasswordValues) {
    const ok = await runToast(forgotPassword(values.email), {
      error: t("errors:reset_email_failed"),
    });
    // Backend всегда возвращает 200 (privacy by design), даже если email
    // несуществующий — показываем "проверьте почту" по дизайну.
    if (ok !== undefined) onSent();
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
      <Input
        label={t("auth:field_email")}
        type="email"
        autoComplete="email"
        placeholder="you@example.com"
        error={tr(errors.email?.message)}
        {...register("email")}
      />
      <Button type="submit" disabled={isSubmitting} className="w-full">
        {isSubmitting ? "..." : t("auth:forgot_submit")}
      </Button>
    </form>
  );
}
