import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Input } from "@eop/ui";
import { login, type AuthResponse } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { loginSchema, type LoginValues } from "../../../shared/lib/schemas";

export function LoginForm({ onSuccess }: { onSuccess: (r: AuthResponse) => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({
    defaultValues: { email: "", password: "" },
    resolver: zodResolver(loginSchema),
  });

  // tr — переводит i18n-key из zod errors. Schemas хранят key'и (не строки),
  // чтобы один schema работал во всех 4 локалях.
  const tr = (msg?: string) => (msg ? t(msg as never) : undefined);

  async function onSubmit(values: LoginValues) {
    const r = await runToast(login(values.email, values.password), {
      error: t("errors:auth_failed"),
    });
    if (r) onSuccess(r);
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
      <Input
        label={t("auth:field_password")}
        type="password"
        autoComplete="current-password"
        error={tr(errors.password?.message)}
        {...register("password")}
      />
      <Button type="submit" disabled={isSubmitting} className="w-full">
        {isSubmitting ? "..." : t("auth:submit_login")}
      </Button>
    </form>
  );
}
