import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { register as apiRegister, type AuthResponse } from "@/entities/user";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";
import { registerSchema, type RegisterValues } from "@/shared/lib/schemas";

export function RegisterForm({
  inviteCode,
  onSuccess,
}: {
  inviteCode?: string | null;
  onSuccess: (r: AuthResponse) => void;
}) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();

  const form = useForm<RegisterValues>({
    defaultValues: { email: "", password: "", displayName: "" },
    resolver: zodResolver(registerSchema),
  });

  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: RegisterValues) {
    const r = await runToast(
      apiRegister(values.email, values.password, values.displayName, inviteCode || undefined),
      { error: t("errors:register_failed") },
    );
    if (r) onSuccess(r);
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <InputField
          control={form.control}
          name="displayName"
          label={t("auth:field_name")}
          autoComplete="name"
          placeholder={t("auth:field_name_placeholder")}
          translateError={tr}
        />
        <InputField
          control={form.control}
          name="email"
          label={t("auth:field_email")}
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          translateError={tr}
        />
        <InputField
          control={form.control}
          name="password"
          label={t("auth:field_password_register")}
          type="password"
          autoComplete="new-password"
          translateError={tr}
        />
        <Button type="submit" disabled={form.formState.isSubmitting} className="w-full">
          {form.formState.isSubmitting ? "..." : t("auth:submit_register")}
        </Button>
      </form>
    </Form>
  );
}
