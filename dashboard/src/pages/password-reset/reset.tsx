import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { CheckCircle2, Lock } from "lucide-react";
import { resetPassword } from "../../entities/user";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";
import { resetPasswordSchema, type ResetPasswordValues } from "../../shared/lib/schemas";

export function ResetPasswordRoute() {
  const { t } = useTranslation(["auth", "errors"]);
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const token = params.get("token") ?? "";
  const [done, setDone] = useState(false);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordValues>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { password: "", confirmPassword: "" },
  });
  const tr = (msg?: string) => (msg ? t(msg as never) : undefined);

  async function onSubmit(values: ResetPasswordValues) {
    const ok = await runToast(resetPassword(token, values.password), {
      error: t("errors:reset_password_failed"),
    });
    if (ok !== undefined) {
      setDone(true);
      // Все старые сессии инвалидированы (token_version bumped) — отправляем
      // юзера логиниться заново.
      setTimeout(() => navigate("/login", { replace: true }), 1500);
    }
  }

  if (!token) {
    return (
      <div className="mx-auto max-w-md px-4 pt-12 sm:pt-16">
        <Card>
          <CardHeader>
            <CardTitle className="text-center">{t("auth:reset_invalid_link")}</CardTitle>
            <CardDescription className="text-center">
              {t("auth:reset_invalid_link_lead")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link to="/forgot-password" className="block text-center text-sm underline">
              {t("auth:reset_request_new")}
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="relative">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-md px-4 pt-12 sm:pt-16">
        <Card className="card-hover reveal">
          <CardHeader>
            <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              {done ? (
                <CheckCircle2 className="h-6 w-6 text-primary-foreground" />
              ) : (
                <Lock className="h-6 w-6 text-primary-foreground" />
              )}
            </div>
            <CardTitle className="text-center font-display tracking-tight">
              {done ? t("auth:reset_done_title") : t("auth:reset_title")}
            </CardTitle>
            <CardDescription className="text-center">
              {done ? t("auth:reset_done_lead") : t("auth:reset_lead")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {!done && (
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
                <Input
                  label={t("auth:reset_field_password")}
                  type="password"
                  autoComplete="new-password"
                  error={tr(errors.password?.message)}
                  {...register("password")}
                />
                <Input
                  label={t("auth:reset_field_password2")}
                  type="password"
                  autoComplete="new-password"
                  error={tr(errors.confirmPassword?.message)}
                  {...register("confirmPassword")}
                />
                <Button type="submit" disabled={isSubmitting} className="w-full">
                  {isSubmitting ? "..." : t("auth:reset_submit")}
                </Button>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
