import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Input } from "@eop/ui";
import { register as apiRegister, type AuthResponse } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { registerSchema, type RegisterValues } from "../../../shared/lib/schemas";

export function RegisterForm({
  inviteCode,
  onSuccess,
}: {
  inviteCode?: string | null;
  onSuccess: (r: AuthResponse) => void;
}) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({
    defaultValues: { email: "", password: "", displayName: "" },
    resolver: zodResolver(registerSchema),
  });

  const tr = (msg?: string) => (msg ? t(msg as never) : undefined);

  async function onSubmit(values: RegisterValues) {
    const r = await runToast(
      apiRegister(values.email, values.password, values.displayName, inviteCode || undefined),
      { error: t("errors:register_failed") },
    );
    if (r) onSuccess(r);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
      <Input
        label={t("auth:field_name")}
        placeholder={t("auth:field_name_placeholder")}
        autoComplete="name"
        error={tr(errors.displayName?.message)}
        {...register("displayName")}
      />
      <Input
        label={t("auth:field_email")}
        type="email"
        autoComplete="email"
        placeholder="you@example.com"
        error={tr(errors.email?.message)}
        {...register("email")}
      />
      <Input
        label={t("auth:field_password_register")}
        type="password"
        autoComplete="new-password"
        error={tr(errors.password?.message)}
        {...register("password")}
      />
      <Button type="submit" disabled={isSubmitting} className="w-full">
        {isSubmitting ? "..." : t("auth:submit_register")}
      </Button>
    </form>
  );
}
