import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { Button, Form, InputField } from "@eop/ui";
import { useChangeMyEmail } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { changeEmailSchema, type ChangeEmailValues } from "../../../shared/lib/schemas";

export function ChangeEmailForm({ currentEmail }: { currentEmail?: string }) {
  const { t } = useTranslation(["common", "auth", "errors"]);
  const runToast = useMutationToast();
  const change = useChangeMyEmail();
  const form = useForm<ChangeEmailValues>({
    resolver: zodResolver(changeEmailSchema),
    defaultValues: { currentPassword: "", email: currentEmail ?? "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : msg);

  async function onSubmit(values: ChangeEmailValues) {
    const ok = await runToast(
      change.mutateAsync({ currentPassword: values.currentPassword, email: values.email }),
      {
        success: t("common:settings.email_change_success"),
        error: t("common:settings.email_change_failed"),
      },
    );
    if (ok !== undefined) form.reset({ currentPassword: "", email: values.email });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <InputField
            control={form.control}
            name="email"
            label={t("auth:field_email")}
            type="email"
            autoComplete="email"
            translateError={tr}
          />
          <InputField
            control={form.control}
            name="currentPassword"
            label={t("common:settings.current_password")}
            type="password"
            autoComplete="current-password"
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
