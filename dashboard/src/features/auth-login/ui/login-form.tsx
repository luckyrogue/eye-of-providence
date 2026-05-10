import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { login, type AuthResponse } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { loginSchema, type LoginValues } from "../../../shared/lib/schemas";

export function LoginForm({ onSuccess }: { onSuccess: (r: AuthResponse) => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const runToast = useMutationToast();
  const form = useForm<LoginValues>({
    defaultValues: { email: "", password: "" },
    resolver: zodResolver(loginSchema),
  });

  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: LoginValues) {
    const r = await runToast(login(values.email, values.password), {
      error: t("errors:auth_failed"),
    });
    if (r) onSuccess(r);
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
        <InputField
          control={form.control}
          name="password"
          label={t("auth:field_password")}
          type="password"
          autoComplete="current-password"
          translateError={tr}
        />
        <Button type="submit" disabled={form.formState.isSubmitting} className="w-full">
          {form.formState.isSubmitting ? "..." : t("auth:submit_login")}
        </Button>
      </form>
    </Form>
  );
}
